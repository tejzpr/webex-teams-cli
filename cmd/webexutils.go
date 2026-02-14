package cmd

import (
	"github.com/urfave/cli/v2"
)

// WebexUtils function
func (app *Application) WebexUtils() *cli.Command {
	return &cli.Command{
		Name:    "utils",
		Usage:   "Utils for Webex",
		Aliases: []string{"u"},
		Flags:   []cli.Flag{},
		Subcommands: []*cli.Command{
			app.FindRoomCMD(),
			app.ListRoomsCMD(),
		},
		Action: func(c *cli.Context) error {
			return nil
		},
	}
}
