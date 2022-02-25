package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v2"
)

// ListRoomsCMD function
func (app *Application) ListRoomsCMD() *cli.Command {
	return &cli.Command{
		Name:    "listrooms",
		Aliases: []string{"lr"},
		Flags: []cli.Flag{
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
			m, err := json.Marshal(rooms)
			if err != nil {
				return err
			}
			fmt.Printf("%s", string(m))
			return nil
		},
	}
}
