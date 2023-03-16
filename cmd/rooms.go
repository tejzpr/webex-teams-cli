package cmd

import (
	"github.com/urfave/cli/v2"
)

// RoomCMD function
func (app *Application) RoomCMD() *cli.Command {
	return &cli.Command{
		Name:    "room",
		Aliases: []string{"room"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "roomID",
				Aliases:  []string{"rid"},
				Value:    "",
				Usage:    "Webex room ID to send the message to",
				Required: false,
				EnvVars:  []string{"WEBEX_ROOM_ID"},
			},
			&cli.StringFlag{
				Name:     "toPersonID",
				Aliases:  []string{"pid"},
				Value:    "",
				Usage:    "Webex person ID to send the message to",
				Required: false,
				EnvVars:  []string{"WEBEX_PERSON_ID"},
			},
			&cli.StringFlag{
				Name:     "toPersonEmail",
				Aliases:  []string{"pe"},
				Value:    "",
				Usage:    "Webex person Email to send the message to",
				Required: false,
				EnvVars:  []string{"WEBEX_PERSON_EMAIL"},
			},
			&cli.StringFlag{
				Name:     "interactive",
				Aliases:  []string{"i"},
				Value:    "N",
				Usage:    "Y/N to start an interactive session",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "console",
				Aliases:  []string{"c"},
				Value:    "N",
				Usage:    "Y/N to start an console based session",
				Required: false,
			},
		},
		Subcommands: []*cli.Command{
			app.SendMessageToRoomCMD(),
			app.AddPeopleCMD(),
			app.ExportPeopleCMD(),
			app.RemovePeopleCMD(),
			app.BroadcastToRoomsCMD(),
		},
		Action: func(c *cli.Context) error {
			return nil
		},
	}
}
