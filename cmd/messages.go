package cmd

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

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

			// One-shot send
			sentMessage, err := app.SendMessage2Room(params)
			if err != nil {
				log.Error(err.Error())
				return nil
			}
			log.Infof("Sent message: %s", sentMessage.ID)
			return nil
		},
	}
}
