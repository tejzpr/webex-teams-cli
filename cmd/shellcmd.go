package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/google/uuid"
	webexteams "github.com/jbogarin/go-cisco-webex-teams/sdk"
	cmap "github.com/orcaman/concurrent-map"
	log "github.com/sirupsen/logrus"

	"github.com/urfave/cli/v2"
)

// ShellCMDApplication struct
type ShellCMDApplication struct {
	*Application
	AuthorizedUsers            []webexteams.Person
	AuthorizedUserEmails       map[string]string
	AuthorizedUserEmailsString string
	Room                       *webexteams.Room
	LogFile                    *os.File
	LogChannel                 chan logmessage
	LogWG                      sync.WaitGroup
	ErrorChannel               chan error
	ErrorsWG                   sync.WaitGroup
	ResponseChannel            chan rmessage
	ResponseWG                 sync.WaitGroup
	ProcessingChannel          chan command
	ProcessingWG               sync.WaitGroup
	ProcessingCMDWorkerPool    *workerpool.WorkerPool
	ShellToUse                 string
	ProcessTimeout             time.Duration
	SupportedCommands          map[string]bool
	ExecutedCommands           cmap.ConcurrentMap
	TerminateCommandIDChannel  chan string
}

type logmessage struct {
	MSG   string
	EMAIL string
}

type rmessage struct {
	MSG  string
	FILE *os.File
}

type command struct {
	CMD              string
	TRIGGEREDBYEMAIL string
	ID               string
}

// ShellCMD function
func (app *Application) ShellCMD() *cli.Command {
	return &cli.Command{
		Name:        "shell",
		Aliases:     []string{"sh"},
		Description: "Start a shell listener.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "personEmails",
				Aliases:  []string{"pes"},
				Value:    "",
				Usage:    "Emails of people who can issue command requests",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "roomID",
				Aliases:  []string{"rid"},
				Value:    "",
				Usage:    "Webex room ID from which the user can issue commands",
				Required: true,
				EnvVars:  []string{"WEBEX_ROOM_ID"},
			},
			&cli.StringFlag{
				Name:     "logfile",
				Aliases:  []string{"log"},
				Value:    "shell.log",
				Usage:    "Path of logfile",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "shelltouse",
				Aliases:  []string{"sh"},
				Value:    "bash",
				Usage:    "Shell to use for executing commands",
				Required: false,
			},
			&cli.DurationFlag{
				Name:     "timeout",
				Aliases:  []string{"tm"},
				Value:    time.Second * 0,
				Usage:    "Kill a process after a period of time in seconds (pass 0s to disable timeout)",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {

			shellApp := &ShellCMDApplication{Application: app}

			if app.Me.PersonType != "bot" {
				return errors.New("shell cmd can be started only if the webex access token is for a bot")
			}

			log.Infof("Shell Bot is : %s", app.Me.DisplayName)

			tmpUserEmails := c.String("personEmails")

			if tmpUserEmails == "" {
				return errors.New("person emails is a required parameter, commands can be executed only by this user")
			}

			userEmails := strings.Split(tmpUserEmails, ",")
			shellApp.AuthorizedUserEmails = make(map[string]string, len(userEmails))
			tmpEmailsList := make([]string, 0)
			for _, userEmail := range userEmails {
				trmEmail := strings.TrimSpace(userEmail)
				// Get Authorized User's details
				queryParams := &webexteams.ListPeopleQueryParams{
					Email: trmEmail,
					Max:   1,
				}
				tmpEmailsList = append(tmpEmailsList, trmEmail)
				people, _, err := app.Client.People.ListPeople(queryParams)

				if err != nil {
					log.Fatal(err)
				} else if len(people.Items) > 0 {
					shellApp.AuthorizedUsers = append(shellApp.AuthorizedUsers, people.Items[0])
					shellApp.AuthorizedUserEmails[people.Items[0].Emails[0]] = people.Items[0].Emails[0]
				} else {
					return fmt.Errorf("No person found with email : %s", trmEmail)
				}
			}

			shellApp.AuthorizedUserEmailsString = strings.Join(tmpEmailsList, ", ")

			log.Infof("Authorized Users are: %v", shellApp.AuthorizedUserEmailsString)

			var parsedRoomID string
			var err error
			roomID := c.String("roomID")
			if roomID != "" {
				parsedRoomID, err = app.parseRoomID(roomID)
				if err != nil {
					return err
				}
			} else {
				return errors.New("roomID is required")
			}

			if parsedRoomID != "" {
				room, _, err := app.Client.Rooms.GetRoom(parsedRoomID)
				if err != nil {
					return err
				}
				shellApp.Room = room
				log.Infof("Accepting messages from room : %s", shellApp.Room.Title)
			} else {
				return errors.New("roomID is not valid")
			}

			logFile := c.String("logfile")

			if strings.HasPrefix(logFile, "~") {
				home, err := os.UserHomeDir()
				if err != nil {
					return err
				}
				tmpFilename := logFile[len("~"):]
				logFile = path.Join(home, tmpFilename)
			}

			shellApp.LogFile, err = os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			log.Infof("Writing access logs to: %s", logFile)

			shellApp.ExecutedCommands = cmap.New()
			shellApp.ProcessTimeout = c.Duration("timeout")
			shellApp.ShellToUse = c.String("shelltouse")
			shellApp.LogChannel = make(chan logmessage, 100)
			shellApp.ResponseChannel = make(chan rmessage, 10)
			shellApp.ProcessingChannel = make(chan command, 10)
			shellApp.TerminateCommandIDChannel = make(chan string, 10)
			shellApp.ErrorChannel = make(chan error, 1)
			shellApp.ProcessingCMDWorkerPool = workerpool.New(100)
			shellApp.process()
			return nil
		},
	}
}

func (app *ShellCMDApplication) sendToLog(msg string, email string) string {
	app.LogWG.Add(1)
	app.LogChannel <- logmessage{msg, email}
	return msg
}

func (app *ShellCMDApplication) sendError(err error) error {
	app.ErrorsWG.Add(1)
	app.ErrorChannel <- err
	return err
}

func (app *ShellCMDApplication) sendShellResponseMessage(msg string, fileobj *os.File) string {
	app.ResponseWG.Add(1)
	app.ResponseChannel <- rmessage{msg, fileobj}
	return msg
}

func (app *ShellCMDApplication) sendForProcessing(cmd command) {
	app.ProcessingWG.Add(1)
	app.ProcessingChannel <- cmd
}

type executedCommand struct {
	cmd    *exec.Cmd
	cmdObj command
}

func (app *ShellCMDApplication) stdHandler(cmd command, stdout string, stderr string) {
	app.sendToLog(fmt.Sprintf("Response sent for command (ID: %s): `%s`", cmd.ID, cmd.CMD), cmd.TRIGGEREDBYEMAIL)
	app.sendShellResponseMessage(fmt.Sprintf("Response for command (ID: %s): `%s`", cmd.ID, cmd.CMD), nil)

	if stdout != "" {
		app.sendShellResponseMessage(fmt.Sprintf(`
	STDOUT 
	%v
	`, stdout), nil)
	}

	if stderr != "" {
		app.sendShellResponseMessage(fmt.Sprintf(`
	STDERR 
	%v
	`, stderr), nil)
	}
}

func (app *ShellCMDApplication) runCMD(cmd command) {
	defer app.ProcessingWG.Done()
	app.shellout(cmd)
}

func (app *ShellCMDApplication) process() error {
	proceesedMessages := cmap.New()

	defer func() {
		if err := app.LogFile.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		for logMsg := range app.LogChannel {
			timestamp := time.Now().UTC()
			msg := fmt.Sprintf("[%s] %s on room %s : %s \n", timestamp, logMsg.EMAIL, app.Room.Title, logMsg.MSG)
			if _, err := app.LogFile.Write([]byte(msg)); err != nil {
				log.Fatal(err)
			}
			app.LogWG.Done()
		}
	}()

	go func() {
		for respMsg := range app.ResponseChannel {
			func(respMsg rmessage) {
				tmpFilename := ""
				if respMsg.FILE != nil {
					tmpFilename = respMsg.FILE.Name()
				}

				if tmpFilename != "" {
					defer os.Remove(tmpFilename)
				}

				sendparams := &SendMessageParams{
					RoomID:   app.Room.ID,
					Text:     respMsg.MSG,
					Filename: tmpFilename,
				}
				_, err := app.SendMessage2Room(sendparams)

				if err != nil {
					app.sendError(err)
				}
				app.ResponseWG.Done()
			}(respMsg)
		}
	}()

	go func() {
		for processingCMD := range app.ProcessingChannel {
			func(processingCMD command) {
				app.ProcessingCMDWorkerPool.Submit(func() {
					app.runCMD(processingCMD)
				})
			}(processingCMD)
		}
	}()

	app.startUserInitiatedKillRoutine()

	go func() {
		now := time.Now()
		for range time.Tick(2 * time.Second) {
			messages, err := app.GetMessagesForRoom(app.Room, 10)
			if err != nil {
				// app.sendError(err)
			} else {
				var currentMessageIDX = -1
				for i := len(messages) - 1; i >= 0; i-- {
					if messages[i].Created.After(now) {
						currentMessageIDX = i
						break
					}
				}

				if currentMessageIDX >= 0 {
					recentMessages := messages[:currentMessageIDX]
					if currentMessageIDX == 0 {
						recentMessages = []webexteams.Message{messages[0]}
					}

					for i := len(recentMessages) - 1; i >= 0; i-- {
						item := recentMessages[i]
						if app.AuthorizedUserEmails[item.PersonEmail] == item.PersonEmail {
							duplicateCheck := false
							if tmp, ok := proceesedMessages.Get(item.ID); ok {
								duplicateCheck = tmp.(bool)
							}

							if !duplicateCheck {
								proceesedMessages.Set(item.ID, true)

								if app.Room.RoomType == "group" && strings.HasPrefix(item.Text, app.Me.NickName) {
									item.Text = strings.TrimSpace(item.Text[len(app.Me.NickName):])
								}

								if strings.ToLower(item.Text) == "help" {
									app.sendShellResponseMessage(fmt.Sprintf("Hi I am currently accepting commands from:  **%s** ", app.AuthorizedUserEmailsString), nil)
									app.sendShellResponseMessage("**Supported commands** ", nil)
									app.sendShellResponseMessage(`
* *cmd* : To issue a command type in 'cmd' followed by the command you would like to run. Usage: **cmd [allowed system command]**
* *kill* : Kills a command using its ID. Usage: **kill [command ID]**
* *list active* : Lists all active commands. Usage: **list active**
* *list completed* : Lists all completed commands. Usage: **list completed**
									`, nil)

								} else if strings.HasPrefix(strings.ToLower(item.Text), "cmd") {
									cmdTxt := strings.TrimSpace(item.Text[len("cmd"):])
									cmd := command{}
									cmd.CMD = cmdTxt
									cmdUUID := uuid.New()
									cmd.ID = cmdUUID.String()
									cmd.TRIGGEREDBYEMAIL = item.PersonEmail
									app.sendShellResponseMessage(app.sendToLog(fmt.Sprintf("Recieved command (ID: %s): `%s`", cmd.ID, cmd.CMD), item.PersonEmail), nil)
									app.sendForProcessing(cmd)
								} else if strings.HasPrefix(strings.ToLower(item.Text), "kill") {
									pid := strings.TrimSpace(item.Text[len("kill"):])
									_, err := uuid.Parse(pid)
									if err != nil {
										app.sendShellResponseMessage(app.sendToLog(fmt.Sprintf("Invalid process ID: `%s`", pid), item.PersonEmail), nil)
									} else {
										app.sendShellResponseMessage(app.sendToLog(fmt.Sprintf("Recieved kill for command ID: `%s`", pid), item.PersonEmail), nil)
										app.TerminateCommandIDChannel <- pid
									}
								} else if strings.HasPrefix(strings.ToLower(item.Text), "list") {
									subCommand := strings.TrimSpace(item.Text[len("list"):])
									switch strings.ToLower(subCommand) {
									case "active":
										var activeCommands = make([]string, 0)
										for item := range app.ExecutedCommands.Iter() {
											id := item.Key
											exeCmd := item.Val.(executedCommand)
											if exeCmd.cmd.ProcessState == nil {
												activeCommands = append(activeCommands, fmt.Sprintf("%s (%s) > cmd[%s]", id, exeCmd.cmdObj.TRIGGEREDBYEMAIL, exeCmd.cmdObj.CMD))
											}
										}
										if len(activeCommands) > 0 {
											app.sendShellResponseMessage("Active Commands : ", nil)
											app.sendShellResponseMessage(fmt.Sprintf("\n``` \n %s\n``` ", strings.Join(activeCommands, "\n")), nil)
										} else {
											app.sendShellResponseMessage("No active commands found", nil)
										}
									case "completed":
										var completedCommands = make([]string, 0)
										for item := range app.ExecutedCommands.Iter() {
											id := item.Key
											exeCmd := item.Val.(executedCommand)
											if exeCmd.cmd.ProcessState != nil {
												completedCommands = append(completedCommands, fmt.Sprintf("%s (%s) Status: %d > cmd[%s]", id, exeCmd.cmdObj.TRIGGEREDBYEMAIL, exeCmd.cmd.ProcessState.ExitCode(), exeCmd.cmdObj.CMD))
											}
										}
										if len(completedCommands) > 0 {
											app.sendShellResponseMessage("Completed Commands : ", nil)
											app.sendShellResponseMessage(fmt.Sprintf("\n``` \n %s\n``` ", strings.Join(completedCommands, "\n")), nil)
										} else {
											app.sendShellResponseMessage("No completed / killed commands found", nil)
										}
									default:
										app.sendShellResponseMessage("Command not found", nil)
									}
								} else {
									app.sendShellResponseMessage("Command not found", nil)
								}
							}
						}
					}
				}
			}
		}
	}()

	for errorMsg := range app.ErrorChannel {
		log.Error(errorMsg)
		app.ErrorsWG.Done()
	}

	app.ProcessingWG.Wait()
	app.ResponseWG.Wait()
	app.LogWG.Wait()

	return nil
}
