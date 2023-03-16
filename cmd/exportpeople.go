package cmd

import (
	"encoding/csv"
	"os"

	webexteams "github.com/jbogarin/go-cisco-webex-teams/sdk"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// ExportPeopleCMD function
func (app *Application) ExportPeopleCMD() *cli.Command {
	return &cli.Command{
		Name:        "exportmembers",
		Aliases:     []string{"em"},
		Description: "Export members of a room to a CSV file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "memberscsv",
				Aliases:  []string{"csv"},
				Value:    "",
				Usage:    "Path to CSV to export to",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			roomID := c.String("roomID")

			parsedRoomID, err := app.parseRoomID(roomID)
			if err != nil {
				return err
			}
			csvPath := c.String("memberscsv")
			if csvPath != "" {
				roomUtilsApp := &ExportPeopleApplication{Application: app, MemberCSVPath: csvPath}
				err := roomUtilsApp.Export(parsedRoomID)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
}

// ExportPeopleApplication struct
type ExportPeopleApplication struct {
	*Application
	MemberCSVPath string
}

// Export function
func (app *ExportPeopleApplication) Export(roomID string) error {

	room, _, err := app.Client.Rooms.GetRoom(roomID)
	if err != nil {
		return err
	}

	log.Info(room.Title)
	membershipQueryParams := &webexteams.ListMembershipsQueryParams{
		RoomID: room.ID,
	}

	memberships, _, err := app.Client.Memberships.ListMemberships(membershipQueryParams)
	if err != nil {
		return err
	}
	log.Info(room.ID)
	if len(memberships.Items) > 0 {
		csvFile, err := os.Create(app.MemberCSVPath)
		if err != nil {
			log.Fatal(err)
		}
		defer csvFile.Close()

		csvWriter := csv.NewWriter(csvFile)
		defer csvWriter.Flush()
		csvWriter.Write([]string{"email", "moderator"})
		// Save membership to CSV file
		for _, membership := range memberships.Items {
			moderator := "false"
			if membership.IsModerator {
				moderator = "true"
			}
			emailData := []string{membership.PersonEmail, moderator}
			csvWriter.Write(emailData)
		}
	}

	return nil
}
