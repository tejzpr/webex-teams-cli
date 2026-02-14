package cmd

import (
	"io"
	"log"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v2"

	"github.com/tejzpr/webex-teams-cli/cmd/tui"
)

// ChatCMD returns the top-level "chat" command that launches the Bubbletea TUI
func (app *Application) ChatCMD() *cli.Command {
	return &cli.Command{
		Name:    "chat",
		Aliases: []string{"c"},
		Usage:   "Launch interactive Webex chat TUI",
		Action: func(c *cli.Context) error {
			// Redirect Go's standard log output away from stderr so SDK
			// log.Printf calls don't corrupt the Bubbletea TUI display.
			logFile, err := os.OpenFile(filepath.Join(app.DownloadsDir, "webex-tui.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				// If we can't open a log file, just discard
				log.SetOutput(io.Discard)
			} else {
				log.SetOutput(logFile)
				defer logFile.Close()
			}

			model := tui.NewModel(app)
			p := tea.NewProgram(model, tea.WithAltScreen())
			_, err = p.Run()
			return err
		},
	}
}
