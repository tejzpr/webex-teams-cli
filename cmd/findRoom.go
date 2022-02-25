package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	webexteams "github.com/jbogarin/go-cisco-webex-teams/sdk"
	"github.com/urfave/cli/v2"
)

func (app *Application) getRooms(max int, roomType string) ([]webexteams.Room, error) {
	roomsQueryParams := &webexteams.ListRoomsQueryParams{
		Max:      max,
		TeamID:   "",
		RoomType: roomType,
		Paginate: false,
		SortBy:   "lastactivity",
	}

	rooms, _, err := app.Client.Rooms.ListRooms(roomsQueryParams)
	if err != nil {
		return make([]webexteams.Room, 0), err
	}
	return rooms.Items, nil
}

// FindRoomCMD function
func (app *Application) FindRoomCMD() *cli.Command {
	return &cli.Command{
		Name:    "findroom",
		Aliases: []string{"fr"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "title",
				Aliases:  []string{"t"},
				Value:    "",
				Usage:    "Webex room title of the room",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "roomType",
				Aliases:  []string{"rt"},
				Value:    "",
				Usage:    "Filter rooms by room type - group / direct. Defaults to all.",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			rooms, err := app.getRooms(1000, c.String("roomType"))
			if err != nil {
				return err
			}
			for _, room := range rooms {
				if strings.ToLower(room.Title) == strings.ToLower(c.String("title")) {
					m, err := json.Marshal(room)
					if err != nil {
						return err
					}
					fmt.Printf("%s", string(m))
				}
			}
			return nil
		},
	}
}
