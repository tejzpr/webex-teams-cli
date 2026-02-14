package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gammazero/workerpool"
	"github.com/tejzpr/webex-go-sdk/v2/memberships"
	"github.com/tejzpr/webex-go-sdk/v2/people"
	"github.com/tejzpr/webex-go-sdk/v2/rooms"
	"github.com/urfave/cli/v2"
)

// BroadcastToRoomsApplication struct
type BroadcastToRoomsApplication struct {
	*Application
	PeopleCSVPath string
	Access        string
	UserEmail     string
	BroadcastText string
	BroadcastFile string
}

// BroadcastToRoomsCMD function
func (app *Application) BroadcastToRoomsCMD() *cli.Command {
	return &cli.Command{
		Name:        "broadcast",
		Aliases:     []string{"bc"},
		Description: "Add members to a room(s)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "text",
				Aliases:  []string{"t"},
				Value:    "",
				Usage:    "Text to broadcast supports markdown formatting.",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "file",
				Aliases:  []string{"f"},
				Value:    "",
				Usage:    "Local file path or Remote URI to be broadcasted.",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "confirm",
				Aliases:  []string{"c"},
				Value:    "",
				Usage:    "Continue without confirmation? Allowed values are 'y' or 'n' ",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "access",
				Aliases:  []string{"a"},
				Value:    "om",
				Usage:    "Message will be broadcasted to rooms for which you have specified permissions of either 'a' (send to all), 'o' (owner), 'm' (moderator) or 'om' (owner and moderator). Default is owner and moderator.",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "roomsidscsv",
				Aliases:  []string{"rcsv"},
				Value:    "",
				Usage:    "Path to a CSV containing a list of RoomID's to which message will be broadcasted to.",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			roomID := c.String("roomID")
			var roomIDs []string

			roomsCSV := c.String("roomsidscsv")
			if roomsCSV != "" {
				roomsCSVPath := roomsCSV
				if !strings.HasSuffix(roomsCSVPath, ".csv") {
					return errors.New("Only CSV files are supported")
				}
				if strings.HasPrefix(roomsCSVPath, "~") {
					home, err := os.UserHomeDir()
					if err != nil {
						return err
					}
					tmpFilename := roomsCSVPath[len("~"):]
					roomsCSVPath = path.Join(home, tmpFilename)
				}

				absFilePath, err := filepath.Abs(roomsCSVPath)
				if err != nil {
					return err
				}

				csvfile, err := os.Open(absFilePath)
				if err != nil {
					return err
				}

				defer csvfile.Close()
				c := ParseRoomIDsCSV(csvfile)
				for v := range c {
					parsedRoomID, err := app.parseRoomID(v.Value.RoomID)
					if err != nil {
						return err
					}
					roomIDs = append(roomIDs, parsedRoomID)
				}
			} else if roomID != "" {
				roomIDs = append(roomIDs, roomID)
			}

			access := c.String("access")
			if access != "a" && access != "o" && access != "m" && access != "om" {
				return errors.New("Allowed valued for access flag are a, o, m and om")
			}

			confirm := c.String("confirm")
			if len(roomIDs) <= 0 {
				if confirm != "" {
					if !strings.HasSuffix("y", strings.TrimSpace(strings.ToLower(confirm))) {
						return nil
					}
				} else {
					reader := bufio.NewReader(os.Stdin)
					fmt.Print("Continue to broadcast to all rooms that you have moderator access to? (y/n) : ")
					text, _ := reader.ReadString('\n')
					if !strings.HasSuffix("y", strings.TrimSpace(strings.ToLower(text))) {
						return nil
					}
				}
			}

			broadcasttext := c.String("text")
			broadcastfile := c.String("file")
			if broadcasttext != "" || broadcastfile != "" {
				roomUtilsApp := &BroadcastToRoomsApplication{Application: app, BroadcastFile: broadcastfile, BroadcastText: broadcasttext, Access: access}
				err := roomUtilsApp.BroadcastToRoom(roomIDs)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func (app *BroadcastToRoomsApplication) checkAccess(me *people.Person, room *rooms.Room, membership memberships.Membership) bool {
	if app.Access == "a" {
		return true
	} else if app.Access == "o" {
		if room.CreatorID == me.ID {
			return true
		}
	} else if app.Access == "m" {
		if membership.IsModerator == true {
			return true
		}
	} else if app.Access == "om" {
		if room.CreatorID == me.ID && membership.IsModerator == true {
			return true
		}
	}
	return false
}

// BroadcastToRoom function
func (app *BroadcastToRoomsApplication) BroadcastToRoom(roomIDs []string) error {
	me, err := app.Client.People().GetMe()
	if err != nil {
		log.Fatal(err)
	}
	if len(me.Emails) > 0 {
		app.UserEmail = me.Emails[0]
	} else {
		return errors.New("Invalid user email")
	}

	errChan := make(chan error, 10)
	go func() {
		for err := range errChan {
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	var rwg sync.WaitGroup
	rwp := workerpool.New(10)
	for _, roomID := range roomIDs {
		rwg.Add(1)
		func(roomID string) {
			rwp.Submit(func() {
				defer rwg.Done()
				if roomID == "" {
					membershipQueryParams := &memberships.ListOptions{}
					mbrPage, err := app.Client.Memberships().List(membershipQueryParams)
					if err != nil {
						errChan <- err
						return
					}
					if len(mbrPage.Items) > 0 {
						var wg sync.WaitGroup
						for _, membership := range mbrPage.Items {
							wg.Add(1)
							go func(membership memberships.Membership) {
								defer wg.Done()
								room, err := app.Client.Rooms().Get(membership.RoomID)
								if err != nil {
									errChan <- err
									return
								}

								if room.Title != "" && app.checkAccess(me, room, membership) {
									err := app.sendBroadCastToRoom(room)
									if err == nil {
										log.Println("To Room: ", room.Title)
									} else {
										errChan <- err
										return
									}
								}
								errChan <- nil
							}(membership)
						}
						wg.Wait()
					}
				} else {
					room, err := app.Client.Rooms().Get(roomID)
					if err != nil {
						errChan <- err
						return
					}
					membershipQueryParams := &memberships.ListOptions{
						RoomID:      room.ID,
						PersonEmail: app.UserEmail,
					}

					mbrPage, err := app.Client.Memberships().List(membershipQueryParams)
					if err != nil {
						errChan <- err
						return
					}

					if len(mbrPage.Items) > 0 {
						if room.Title != "" && app.checkAccess(me, room, mbrPage.Items[0]) {
							err := app.sendBroadCastToRoom(room)
							if err == nil {
								log.Println("Message sent to: ", room.Title)
							} else {
								errChan <- err
								return
							}
						}
					}
				}
				errChan <- nil
			})
		}(roomID)
	}
	rwg.Wait()
	return nil
}

func (app *BroadcastToRoomsApplication) sendBroadCastToRoom(room *rooms.Room) error {

	membershipQueryParams := &memberships.ListOptions{
		PersonEmail: app.UserEmail,
		RoomID:      room.ID,
	}

	mbrPage, err := app.Client.Memberships().List(membershipQueryParams)
	if err != nil {
		return err
	}

	// Has membership
	if len(mbrPage.Items) > 0 {
		membershipID := mbrPage.Items[0].ID
		_ = membershipID

		params := &SendMessageParams{
			RoomID:                   room.ID,
			Text:                     app.BroadcastText,
			Filename:                 app.BroadcastFile,
			RemoteFileRequestTimeout: time.Duration(10),
		}

		sentMessage, err := app.SendMessage2Room(params)
		if err != nil {
			return err
		} else {
			log.Infof("Sent message: %s", sentMessage.ID)
		}

	} else {
		return errors.New("You are not a member of this room")
	}
	return nil
}
