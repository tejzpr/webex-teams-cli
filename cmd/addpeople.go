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

	log "github.com/sirupsen/logrus"

	"github.com/gammazero/workerpool"
	webexteams "github.com/jbogarin/go-cisco-webex-teams/sdk"
	"github.com/urfave/cli/v2"
)

// AddPeopleApplication struct
type AddPeopleApplication struct {
	*Application
	PeopleCSVPath string
	Access        string
}

// AddPeopleCMD function
func (app *Application) AddPeopleCMD() *cli.Command {
	return &cli.Command{
		Name:        "addmembers",
		Aliases:     []string{"am"},
		Description: "Add members to a room(s)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "memberscsv",
				Aliases:  []string{"mcsv"},
				Value:    "",
				Usage:    "Path to CSV with list of email addresses formatted as : email, moderator",
				Required: true,
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
				Usage:    "Members will be added to rooms for which you have specified permissions of either  'a' (include all),'o' (owner), 'm' (moderator) or 'om' (owner and moderator). Default is owner and moderator.",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "roomsidscsv",
				Aliases:  []string{"rcsv"},
				Value:    "",
				Usage:    "Path to a CSV containing a list of RoomID's to which members will be added to.",
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
					fmt.Print("Continue to add members to all rooms that you have moderator access to? (y/n) : ")
					text, _ := reader.ReadString('\n')
					if !strings.HasSuffix("y", strings.TrimSpace(strings.ToLower(text))) {
						return nil
					}
				}
			}

			csvPath := c.String("memberscsv")
			if csvPath != "" {
				roomUtilsApp := &AddPeopleApplication{Application: app, PeopleCSVPath: csvPath, Access: access}
				err := roomUtilsApp.AddPeopleToRoom(roomIDs)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func (app *AddPeopleApplication) checkAccess(me *webexteams.Person, room *webexteams.Room, membership webexteams.Membership) bool {
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

// AddPeopleToRoom function
func (app *AddPeopleApplication) AddPeopleToRoom(roomIDs []string) error {

	errChan := make(chan error, 10)
	go func() {
		for err := range errChan {
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	var rwg sync.WaitGroup
	rwp := workerpool.New(2)
	for _, roomID := range roomIDs {
		rwg.Add(1)
		func(roomID string) {
			rwp.Submit(func() {
				defer rwg.Done()
				if roomID == "" {
					membershipQueryParams := &webexteams.ListMembershipsQueryParams{}
					memberships, _, err := app.Client.Memberships.ListMemberships(membershipQueryParams)
					if err != nil {
						errChan <- err
						return
					}
					if len(memberships.Items) > 0 {
						var wg sync.WaitGroup
						for _, membership := range memberships.Items {
							wg.Add(1)
							go func(membership webexteams.Membership) {
								defer wg.Done()
								room, _, err := app.Client.Rooms.GetRoom(membership.RoomID)
								if err != nil {
									errChan <- err
									return
								}

								if room.Title != "" && app.checkAccess(app.Me, room, membership) {
									err := app.processAddPeople(room)
									if err == nil {
										log.Println("Added members to: ", room.Title)
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
					room, _, err := app.Client.Rooms.GetRoom(roomID)
					if err != nil {
						errChan <- err
						return
					}
					membershipQueryParams := &webexteams.ListMembershipsQueryParams{
						RoomID:      room.ID,
						PersonEmail: app.Email,
					}

					memberships, _, err := app.Client.Memberships.ListMemberships(membershipQueryParams)
					if err != nil {
						errChan <- err
						return
					}

					if len(memberships.Items) > 0 {
						if room.Title != "" && app.checkAccess(app.Me, room, memberships.Items[0]) {
							err := app.processAddPeople(room)
							if err == nil {
								log.Println("Added members to: ", room.Title)
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

func (app *AddPeopleApplication) processAddPeople(room *webexteams.Room) error {

	if room.RoomType == "direct" {
		return errors.New("Cannot add new members to a 1:1 room")
	}

	membershipQueryParams := &webexteams.ListMembershipsQueryParams{
		PersonEmail: app.Email,
		RoomID:      room.ID,
	}

	memberships, _, err := app.Client.Memberships.ListMemberships(membershipQueryParams)
	if err != nil {
		return err
	}

	// Has membership
	if len(memberships.Items) > 0 {
		membershipID := memberships.Items[0].ID
		_ = membershipID
		// fmt.Println("IS Moderator:", memberships.Items[0].IsModerator)
		importPeopleCSVPath := app.PeopleCSVPath
		if !strings.HasSuffix(importPeopleCSVPath, ".csv") {
			return errors.New("Only CSV files are supported")
		}
		if strings.HasPrefix(importPeopleCSVPath, "~") {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			tmpFilename := importPeopleCSVPath[len("~"):]
			importPeopleCSVPath = path.Join(home, tmpFilename)
		}

		absFilePath, err := filepath.Abs(importPeopleCSVPath)
		if err != nil {
			return err
		}

		csvfile, err := os.Open(absFilePath)
		if err != nil {
			return err
		}

		defer csvfile.Close()
		c := ParseUsersCSV(csvfile)
		var wg sync.WaitGroup
		wp := workerpool.New(10)
		for v := range c {
			if v.Err == nil {
				wg.Add(1)
				func(room *webexteams.Room, v UserCSVReturn) {
					wp.Submit(func() {
						defer wg.Done()
						app.createMember(room, v.Value.Email, v.Value.IsModerator)
					})
				}(room, v)
			} else {
				return v.Err
			}
		}

		wg.Wait()

	} else {
		return errors.New("You are not a member of this room")
	}
	return nil
}
