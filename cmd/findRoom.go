package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
)

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
			rooms, err := app.GetRooms(1000, c.String("roomType"))
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
