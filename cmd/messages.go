package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gabriel-vasile/mimetype"
	webexteams "github.com/jbogarin/go-cisco-webex-teams/sdk"
	"github.com/jroimartin/gocui"
	cmap "github.com/orcaman/concurrent-map"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// MessagesApplication struct
type MessagesApplication struct {
	*Application
	IsHelpShown  bool
	IsRoomsShown bool
}

// SendMessageToRoomCMD function
func (app *Application) SendMessageToRoomCMD() *cli.Command {
	return &cli.Command{
		Name:    "message",
		Aliases: []string{"msg"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "text",
				Aliases:  []string{"t"},
				Value:    "",
				Usage:    "Text to be sent to the room",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "file",
				Aliases:  []string{"f"},
				Value:    "",
				Usage:    "Local file path or Remote URI to be sent to the room",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "remoteFileRequestTimeout",
				Aliases:  []string{"rfrt"},
				Value:    "",
				Usage:    "Remote file get request timeout in seconds",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			interactive := c.Bool("interactive")
			console := c.Bool("console")
			roomID := c.String("roomID")
			toPersonID := c.String("toPersonID")
			toPersonEmail := c.String("toPersonEmail")
			fileName := c.String("file")
			txt := c.String("text")
			remoteFileRequestTimeout := c.Int64("remoteFileRequestTimeout")
			if remoteFileRequestTimeout == 0 {
				remoteFileRequestTimeout = 10
			}
			params := &SendMessageParams{
				RoomID:                   roomID,
				PersonID:                 toPersonID,
				PersonEmail:              toPersonEmail,
				Text:                     txt,
				Filename:                 fileName,
				RemoteFileRequestTimeout: time.Duration(remoteFileRequestTimeout),
			}
			var err error
			sendMsgApp := &MessagesApplication{Application: app}
			if interactive {
				err = sendMsgApp.sendMessage2RoomInteractive(params)
				if err != nil {
					log.Error(err.Error())
				}
			} else if console {
				err = sendMsgApp.sendMessage2RoomConsole(params)
				if err != nil {
					log.Error(err.Error())
				}
			} else {
				sentMessage, err := sendMsgApp.SendMessage2Room(params)
				if err != nil {
					log.Error(err.Error())
				} else {
					log.Infof("Sent message: %s", sentMessage.ID)
				}
			}
			if err != nil {
				log.Error(err.Error())
			}
			return nil
		},
	}
}

func (app *MessagesApplication) getRooms(max int) ([]webexteams.Room, error) {
	roomsQueryParams := &webexteams.ListRoomsQueryParams{
		Max:      max,
		TeamID:   "",
		Paginate: false,
		SortBy:   "lastactivity",
	}

	rooms, _, err := app.Client.Rooms.ListRooms(roomsQueryParams)
	if err != nil {
		return make([]webexteams.Room, 0), err
	}
	return rooms.Items, nil
}

var (
	viewArr = []string{"chat", "input"}
	active  = 0
)

func (app *MessagesApplication) setCurrentViewOnTop(g *gocui.Gui, name string) (*gocui.View, error) {
	if _, err := g.SetCurrentView(name); err != nil {
		return nil, err
	}
	return g.SetViewOnTop(name)
}

func (app *MessagesApplication) uiQuit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func (app *MessagesApplication) uiNextView(g *gocui.Gui, v *gocui.View) error {
	nextIndex := (active + 1) % len(viewArr)
	name := viewArr[nextIndex]

	if _, err := app.setCurrentViewOnTop(g, name); err != nil {
		return err
	}

	if nextIndex == 1 {
		g.Cursor = true
		g.Mouse = false
	} else {
		g.Cursor = false
		g.Mouse = true
	}

	active = nextIndex
	return nil
}

func (app *MessagesApplication) uiScrollView(v *gocui.View, dy int) error {
	if v != nil {
		v.Autoscroll = false
		ox, oy := v.Origin()
		if err := v.SetOrigin(ox, oy+dy); err != nil {
			return err
		}
	}
	return nil
}

func (app *MessagesApplication) removeRooms(g *gocui.Gui) error {
	viewArr = []string{"chat", "input"}
	g.DeleteKeybindings("roominput")
	g.DeleteKeybindings("rooms")
	_ = g.DeleteView("roominput")
	return g.DeleteView("rooms")
}

func (app *MessagesApplication) showRoomsAndGetUserInput(g *gocui.Gui, selectedSpace chan webexteams.Room) error {
	viewArr = []string{"rooms", "roominput"}
	var rooms []webexteams.Room
	var err error

	var loadedSpaces []webexteams.Room
	maxX, maxY := g.Size()
	vRoom, err := g.SetView("rooms", 0, 0, maxX-2, maxY-6)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		if _, err = app.setCurrentViewOnTop(g, "rooms"); err != nil {
			return err
		}
		vRoom.Title = "Spaces"
		vRoom.Wrap = false
		vRoom.Autoscroll = false
		fmt.Fprintln(vRoom, "Loading spaces..")
		loadedSpacesChan := make(chan []webexteams.Room)
		go func(g *gocui.Gui) {
			rooms, err = app.getRooms(50)
			if err != nil {
				fmt.Fprintln(vRoom, err.Error())
			}
			loadedSpacesChan <- rooms
			vRoom.Clear()
			g.Update(func(g *gocui.Gui) error {
				for idx, room := range rooms {
					vRoom, err := g.View("rooms")
					if err != nil {
						// fmt.Fprintln(vRoom, err.Error())
						vRoom.Clear()
						vRoom.Write([]byte(fmt.Sprintf("\n\033[31;1m<-\033[0m (%s): \n %s \n", app.Email, err.Error())))
						return err
					}
					fmt.Fprintln(vRoom, fmt.Sprintf("(%d) %s [%s]", idx+1, room.Title, room.RoomType))
				}
				return nil
			})
		}(g)
		loadedSpaces = <-loadedSpacesChan
	}
	vRoomInput, err := g.SetView("roominput", 0, maxY-5, maxX-2, maxY-2)
	vRoomInput.Title = "Input the Number for the Space you would like to join and press Enter"
	vRoomInput.Wrap = false
	vRoomInput.Editable = true
	vRoomInput.Autoscroll = false

	if _, err = app.setCurrentViewOnTop(g, "roominput"); err != nil {
		return err
	}

	if err := g.SetKeybinding("roominput", gocui.KeyEnter, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		inputView, _ := g.View("roominput")
		txt := inputView.Buffer()
		txt = strings.TrimSuffix(txt, "\n")
		txt = strings.TrimSuffix(txt, "\r")
		index, err := strconv.Atoi(strings.TrimSpace(txt))
		if err != nil {
			inputView.Clear()
			inputView.SetCursor(0, 0)
			return nil
		}
		inputView.Clear()
		inputView.SetCursor(0, 0)
		go func(index int) {
			if index >= 0 && index <= len(loadedSpaces) {
				selectedSpace <- loadedSpaces[index-1]
			}
		}(index)
		return nil
	}); err != nil {
		return err
	}

	if err := g.SetKeybinding("rooms", gocui.KeyArrowUp, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			app.uiScrollView(v, -1)
			return nil
		}); err != nil {
		return err
	}
	if err := g.SetKeybinding("rooms", gocui.KeyArrowDown, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			app.uiScrollView(v, 1)
			return nil
		}); err != nil {
		return err
	}
	return nil
}

func (app *MessagesApplication) removeHelp(g *gocui.Gui, v *gocui.View) error {
	return g.DeleteView("help")
}

func (app *MessagesApplication) showHelp(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := g.Size()
	vHelp, err := g.SetView("help", maxX/2, 1, maxX-4, maxY-8)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		g.SetViewOnTop("help")
		vHelp.Title = "Help"
		vHelp.Wrap = true
		fmt.Fprintln(vHelp, `1. Crtl + R to switch rooms`)
		fmt.Fprintln(vHelp, `2. TAB to switch panes`)
		fmt.Fprintln(vHelp, `3. type in "sendfile <filename> or remote URI" to a send file`)
		fmt.Fprintln(vHelp, `4. type in "getfile <file id> to download a file`)
	}
	return nil
}

func (app *MessagesApplication) sendMessage2RoomInteractive(params *SendMessageParams) error {
	if app.Me.PersonType == "bot" {
		return errors.New("shell cmd can be started only if the webex access token is for a user (not a bot)")
	}
	const noMessage = "No messages found"
	var messages []webexteams.Message
	var err error
	var roomID = params.RoomID
	var roomTitle = ""
	var titleTemplate = "Webex Teams Chat [%s] (Ctrl + H for Help)"
	proceesedMessages := cmap.New()
	proceesedFiles := cmap.New()
	chatviewChannel := make(chan string, 1)
	updateBlockerChan := make(chan bool, 1)

	if roomID == "" && (params.PersonID != "" || params.PersonEmail != "") {
		directMessageRequest := webexteams.DirectMessagesQueryParams{
			PersonID:    params.PersonID,
			PersonEmail: params.PersonEmail,
		}
		response, _, err := app.Client.Messages.GetDirectMessages(&directMessageRequest)
		if err != nil {
			return err
		}
		if len(response.Items) > 0 {
			messages = response.Items
			roomID = messages[0].RoomID
		} else {
			return errors.New(noMessage)
		}
	} else if roomID != "" {
		messages, err = app.GetMessagesForRoomFromRoomID(roomID, 1)
		if err != nil {
			return err
		}
	} else {

	}

	roomDetails, _, err := app.Client.Rooms.GetRoom(roomID)
	if err != nil {
		return err
	}
	roomTitle = roomDetails.Title

	uiInputToChat := func(g *gocui.Gui, v *gocui.View) error {
		inputView, _ := g.View("input")
		chatView, _ := g.View("chat")
		chatView.Autoscroll = true
		text := inputView.Buffer()
		newParams := params
		sendFileIdentifier := "sendfile"
		getFileIdentifier := "getfile"
		if strings.HasPrefix(text, getFileIdentifier) {
			fileHash := strings.TrimSpace(text[len(getFileIdentifier):])
			if contentURL, ok := proceesedFiles.Get(fileHash); ok {
				contentURLStr := contentURL.(string)
				contentID := strings.TrimSpace(contentURLStr[len("https://api.ciscospark.com/v1/contents/"):])
				chatviewChannel <- fmt.Sprintf("\n\033[31;1m--\033[0m (%s): \n %s %s \n", app.Email, "Downloading Attachment", fileHash)
				go func(contentID string) {
					fileResp, err := app.Client.Contents.GetContent(contentID)
					if err != nil {
						chatviewChannel <- fmt.Sprintf("\n\033[31;1m--\033[0m (%s): \n %s %s \n", app.Email, "Error while downloading Attachment: ", err.Error())
					} else {
						resp := fileResp.Body()
						mime := mimetype.Detect(resp)
						extn := mime.Extension()
						downloadsFile := path.Join(app.DownloadsDir, fmt.Sprintf("%s%s", fileHash, extn))
						err := ioutil.WriteFile(downloadsFile, resp, 0644)
						if err != nil {
							chatviewChannel <- fmt.Sprintf("\n\033[31;1m--\033[0m (%s): \n %s %s \n", app.Email, "Error while downloading Attachment: ", err.Error())
						} else {
							chatviewChannel <- fmt.Sprintf("\n\033[31;1m--\033[0m (%s): \n %s %s \n", app.Email, "Downloaded Attachment", fileHash)
						}
					}
				}(contentID)
			} else {
				chatviewChannel <- fmt.Sprintf("\n\033[31;1m--\033[0m (%s): \n %s \n", app.Email, "Invalid Attachment ID")
			}
			inputView.Clear()
			inputView.SetCursor(0, 0)
		} else {
			if strings.HasPrefix(text, sendFileIdentifier) {
				newParams.Filename = strings.TrimSpace(text[len(sendFileIdentifier):])
			} else {
				newParams.Filename = ""
				newParams.Text = text
			}
			var err error

			updateBlockerChan <- true
			msg, err := app.SendMessage2Room(newParams)
			if err != nil {
				chatviewChannel <- fmt.Sprintf("\n\033[31;1m--\033[0m (%s): \n %s \n", app.Email, err.Error())
			} else {
				proceesedMessages.Set(msg.ID, true)
				chatviewChannel <- fmt.Sprintf("\n\033[32;1m<-\033[0m (%s): \n %s \n", app.Email, text)
			}

			inputView.Clear()
			inputView.SetCursor(0, 0)
			updateBlockerChan <- false
		}
		return nil
	}

	// START
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return err
	}
	defer g.Close()

	updaterFunc := func(g *gocui.Gui) error {
		if err != nil {
			return err
		}

		now := time.Now()
		for range time.Tick(2 * time.Second) {
			blockUpdate := false

			select {
			case status := <-updateBlockerChan:
				blockUpdate = status
			default:
				blockUpdate = false
			}

			if roomID != "" && !blockUpdate {
				// Get messages after you start interacting
				messages, err = app.GetMessagesForRoomFromRoomID(roomID, 10)
				if err != nil {
					return err
				}
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
						duplicateCheck := false
						if tmp, ok := proceesedMessages.Get(item.ID); ok {
							duplicateCheck = tmp.(bool)
						}
						if !duplicateCheck {
							proceesedMessages.Set(item.ID, true)
							if len(item.Files) > 0 {
								g.Update(func(g *gocui.Gui) error {
									chatView, err := g.View("chat")
									for _, rfile := range item.Files {
										fileHash := app.getAdlerHash(rfile)
										proceesedFiles.Set(fileHash, rfile)

										if item.ParentID == "" {
											chatviewChannel <- fmt.Sprintf("\n\033[34;1m->\033[0m (%s): \n Attachment (Use command 'getfile %s' to download attachment) \n %s \n", item.PersonEmail, fileHash, item.Text)
										} else {
											chatviewChannel <- fmt.Sprintf("\n\033[34;1m->->\033[0m (%s): \n Attachment (Use command 'getfile %s' to download attachment) \n %s \n", item.PersonEmail, fileHash, item.Text)
										}
									}

									chatView.Autoscroll = true
									if err != nil {
										return err
									}
									return nil
								})
							} else {
								g.Update(func(g *gocui.Gui) error {
									chatView, err := g.View("chat")
									if item.ParentID == "" {
										_, err = chatView.Write([]byte(fmt.Sprintf("\n\033[34;1m->\033[0m (%s): \n %s \n", item.PersonEmail, item.Text)))
									} else {
										_, err = chatView.Write([]byte(fmt.Sprintf("\n\033[34;1m->->\033[0m (%s): \n %s \n", item.PersonEmail, item.Text)))
									}
									chatView.Autoscroll = true
									if err != nil {
										return err
									}
									return nil
								})
							}
						}
					}
				}
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	listRoomsAndAskUserToSelectARoom := func(g *gocui.Gui) error {
		if !app.IsRoomsShown {
			app.IsRoomsShown = true

			go func() {
				selectedSpace := make(chan webexteams.Room)
				err := app.showRoomsAndGetUserInput(g, selectedSpace)
				if err != nil {
					close(selectedSpace)
					chatviewChannel <- fmt.Sprintf("\n\033[31;1m--\033[0m (%s): \n %s \n", app.Email, err.Error())
					return
				}
				newRoom := <-selectedSpace

				g.Update(func(g *gocui.Gui) error {
					app.setCurrentViewOnTop(g, "chat")
					app.setCurrentViewOnTop(g, "input")
					app.IsRoomsShown = false
					app.removeRooms(g)
					if newRoom.ID != "" {
						chatView, _ := g.View("chat")
						roomTitle = newRoom.Title
						chatView.Title = fmt.Sprintf(titleTemplate, roomTitle)
						chatView.Clear()
						inputView, _ := g.View("input")
						inputView.Clear()
						roomID = newRoom.ID
						params.RoomID = roomID
					}
					return nil
				})

			}()
		} else {
			app.IsRoomsShown = false
			app.removeRooms(g)
		}
		return nil
	}

	uiLayout := func(g *gocui.Gui) error {
		maxX, maxY := g.Size()

		if v, err := g.SetView("chat", 0, 0, maxX-2, maxY-6); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = fmt.Sprintf(titleTemplate, roomTitle)
			v.Wrap = true
			v.Autoscroll = true

			go func(g *gocui.Gui) {
				chatView, _ := g.View("chat")
				for chatItem := range chatviewChannel {
					g.Update(func(g *gocui.Gui) error {
						// updateBlockerChan <- true
						_, err := chatView.Write([]byte(chatItem))
						if err != nil {
							chatView.Write([]byte(err.Error()))
						}
						// updateBlockerChan <- false
						return nil
					})
				}
			}(g)

		}
		if v, err := g.SetView("input", 0, maxY-5, maxX-2, maxY-2); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.Title = "Input (Enter to send)"
			v.Editable = true
			v.Wrap = false
			if _, err = app.setCurrentViewOnTop(g, "input"); err != nil {
				return err
			}
		}
		go updaterFunc(g)

		return nil
	}

	g.SetManagerFunc(uiLayout)

	g.Highlight = true
	g.Cursor = true
	g.SelFgColor = gocui.ColorGreen

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, app.uiQuit); err != nil {
		return err
	}

	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, app.uiNextView); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlH, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if !app.IsHelpShown {
			app.IsHelpShown = true
			app.showHelp(g, v)
		} else {
			app.IsHelpShown = false
			app.removeHelp(g, v)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlR, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return listRoomsAndAskUserToSelectARoom(g)
	}); err != nil {
		return err
	}

	if err := g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, uiInputToChat); err != nil {
		return err
	}

	if err := g.SetKeybinding("chat", gocui.KeyArrowUp, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			app.uiScrollView(v, -1)
			return nil
		}); err != nil {
		return err
	}
	if err := g.SetKeybinding("chat", gocui.KeyArrowDown, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			app.uiScrollView(v, 1)
			return nil
		}); err != nil {
		return err
	}

	if roomID == "" {
		g.Update(func(g *gocui.Gui) error {
			return listRoomsAndAskUserToSelectARoom(g)
		})
	}

	errorChan := make(chan error)
	go func() {
		if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
			errorChan <- err
		} else if err == gocui.ErrQuit {
			errorChan <- nil
		}
	}()

	return <-errorChan
}

func (app *MessagesApplication) sendMessage2RoomConsole(params *SendMessageParams) error {
	if app.Me.PersonType == "bot" {
		return errors.New("shell cmd can be started only if the webex access token is for a user (not a bot)")
	}
	const noMessage = "No messages found"
	var messages []webexteams.Message
	var err error
	var roomID = params.RoomID
	proceesedMessages := cmap.New()
	updateBlockerChan := make(chan bool, 1)
	proceesedFiles := cmap.New()

	if roomID == "" {
		directMessageRequest := webexteams.DirectMessagesQueryParams{
			PersonID:    params.PersonID,
			PersonEmail: params.PersonEmail,
		}
		response, _, err := app.Client.Messages.GetDirectMessages(&directMessageRequest)
		if err != nil {
			return err
		}
		if len(response.Items) > 0 {
			messages = response.Items
			roomID = messages[0].RoomID
		} else {
			return errors.New(noMessage)
		}
	} else {
		messages, err = app.GetMessagesForRoomFromRoomID(roomID, 1)
		if err != nil {
			return err
		}
	}

	end := make(chan error)

	go func() {
		reader := bufio.NewReader(os.Stdin)
		writer := bufio.NewWriter(os.Stdout)

		mainPrompt := func() {
			fmt.Fprint(writer, fmt.Sprintf("<- (%s): ", app.Email))
			writer.Flush()
		}
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			end <- nil
		}()

		go func() {
			now := time.Now()
			for range time.Tick(2 * time.Second) {
				blockUpdate := false

				select {
				case status := <-updateBlockerChan:
					blockUpdate = status
				default:
					blockUpdate = false
				}

				if !blockUpdate {
					// Get messages after you start interacting
					messages, err = app.GetMessagesForRoomFromRoomID(roomID, 10)
					if err != nil {
						end <- err
					}
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

							duplicateCheck := false
							if tmp, ok := proceesedMessages.Get(item.ID); ok {
								duplicateCheck = tmp.(bool)
							}

							if !duplicateCheck {
								proceesedMessages.Set(item.ID, true)
								if len(item.Files) > 0 {
									for _, rfile := range item.Files {
										fileHash := app.getAdlerHash(rfile)
										proceesedFiles.Set(fileHash, rfile)
										fmt.Fprint(writer, "\n")
										fmt.Fprint(writer, fmt.Sprintf("-> (%s): Attachment (Use command 'getfile %s' to download attachment) \n %s ", item.PersonEmail, fileHash, item.Text))
										fmt.Fprint(writer, "\n")
										writer.Flush()
										mainPrompt()
									}
								} else {
									fmt.Fprint(writer, "\n")
									fmt.Fprint(writer, fmt.Sprintf("-> (%s): %s ", item.PersonEmail, item.Text))
									fmt.Fprint(writer, "\n")
									writer.Flush()
									mainPrompt()
								}
							}
						}
					}
					if err != nil {
						end <- err
					}
				}
			}
		}()

		for {
			mainPrompt()
			text, _ := reader.ReadString('\n')

			if runtime.GOOS == "windows" {
				text = strings.Replace(text, "\r\n", "", -1)
			} else {
				text = strings.Replace(text, "\n", "", -1)
			}
			newParams := params
			sendFileIdentifier := "sendfile"

			getFileIdentifier := "getfile"
			if strings.HasPrefix(text, getFileIdentifier) {
				fileHash := strings.TrimSpace(text[len(getFileIdentifier):])
				if contentURL, ok := proceesedFiles.Get(fileHash); ok {
					contentURLStr := contentURL.(string)
					contentID := strings.TrimSpace(contentURLStr[len("https://api.ciscospark.com/v1/contents/"):])
					fmt.Fprint(writer, fmt.Sprintf("-- (%s): %s %s \n", app.Email, "Downloading Attachment: ", fileHash))
					go func(contentID string) {
						fileResp, err := app.Client.Contents.GetContent(contentID)
						if err != nil {
							fmt.Fprint(writer, fmt.Sprintf("-- (%s): Error while downloading Attachment: %s \n", app.Email, err.Error()))
						} else {
							resp := fileResp.Body()
							mime := mimetype.Detect(resp)
							extn := mime.Extension()
							downloadsFile := path.Join(app.DownloadsDir, fmt.Sprintf("%s%s", fileHash, extn))
							err := ioutil.WriteFile(downloadsFile, resp, 0644)
							if err != nil {
								fmt.Fprint(writer, fmt.Sprintf("-- (%s): Error while downloading Attachment: %s \n", app.Email, err.Error()))
							} else {
								fmt.Fprint(writer, fmt.Sprintf("-- (%s): %s %s \n", app.Email, "Downloaded Attachment", fileHash))
							}
						}
					}(contentID)
				} else {
					fmt.Fprint(writer, fmt.Sprintf("-- (%s): %s \n", app.Email, "Invalid Attachment ID"))
				}
			} else {
				if strings.HasPrefix(text, sendFileIdentifier) {
					newParams.Filename = strings.TrimSpace(text[len(sendFileIdentifier):])
				} else {
					newParams.Filename = ""
					newParams.Text = text
				}
				var err error
				updateBlockerChan <- true
				msg, err := app.SendMessage2Room(newParams)
				if err != nil {
					fmt.Fprint(writer, fmt.Sprintf("-- (%s): %s \n", app.Email, err.Error()))
				} else {
					proceesedMessages.Set(msg.ID, true)
				}
				updateBlockerChan <- false

			}
		}
	}()

	endErr := <-end
	return endErr
}
