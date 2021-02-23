package main

/*
@author: Tejus Pratap
*/

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"webex-teams-cli/cmd"

	webexteams "github.com/jbogarin/go-cisco-webex-teams/sdk"
	"github.com/urfave/cli/v2"

	log "github.com/sirupsen/logrus"
)

func main() {

	appWebex := &cmd.Application{}
	// Print the current version
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("%s\n", c.App.Version)
	}
	// Initialize a new CLI application
	app := &cli.App{
		Name: "Webex Instateams CLI",
		Usage: `
		Description:
		A CLI tool to send messages to / interact with Cisco Webex Teams.

		How to use:

		Set Env variable WEBEX_ACCESS_TOKEN, which can get retrieved from https://developer.webex.com/docs/api/getting-started

		export WEBEX_ACCESS_TOKEN="<access_token>"

		Set Env variable WEBEX_ROOM_ID, is the Space ID that you can get by visiting https://teams.webex.com/ and clicking on a room.

		export WEBEX_ROOM_ID="<roomid>"

		Then you can send a message to the room by running the command

		webex-teams-cli room msg -t "message text" -f <file>
		`,
		Version: "v0.1b",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "accessToken",
				Aliases:  []string{"a"},
				Value:    "",
				Usage:    "The access token to used for interaction with Cisco Webex Teams",
				Required: true,
				EnvVars:  []string{"WEBEX_ACCESS_TOKEN"},
			},
			&cli.StringFlag{
				Name:     "downloadsDir",
				Aliases:  []string{"dd"},
				Value:    "./downloads",
				Usage:    "Directory to store any downloads to",
				Required: false,
			},
		},
		Commands: []*cli.Command{
			appWebex.RoomCMD(),
			appWebex.AddUserToRoomServer(),
			appWebex.MessageRelayServer(),
			appWebex.ShellCMD(),
		},
		Before: func(c *cli.Context) error {
			accessToken := c.String("accessToken")
			Client := webexteams.NewClient()
			Client.SetAuthToken(accessToken)
			appWebex.AccessToken = accessToken
			appWebex.Client = Client
			me, _, err := Client.People.GetMe()
			if err != nil {
				return err
			}
			appWebex.Me = me

			if len(appWebex.Me.Emails) > 0 {
				appWebex.Email = appWebex.Me.Emails[0]
			} else {
				return errors.New("Could not resolve user's email")
			}

			downloadsDir := c.String("downloadsDir")
			if strings.HasPrefix(downloadsDir, "~") {
				home, err := os.UserHomeDir()
				if err != nil {
					return err
				}
				tmpFilename := downloadsDir[len("~"):]
				downloadsDir = path.Join(home, tmpFilename)
			}

			absFilePath, err := filepath.Abs(downloadsDir)
			if err != nil {
				return err
			}
			if _, err := os.Stat(absFilePath); os.IsNotExist(err) {
				os.Mkdir(absFilePath, 0766)
			}

			appWebex.DownloadsDir = downloadsDir

			return nil
		},
		Action: func(c *cli.Context) error {
			return nil
		},
	}

	// Run the CLI application
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err.Error())
	}
}
