package cmd

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	webexteams "github.com/jbogarin/go-cisco-webex-teams/sdk"
	"github.com/urfave/cli/v2"
)

// AddUserToRoomServerApplication struct
type AddUserToRoomServerApplication struct {
	*Application
	EmailDomain string
}

// AddUserToRoomServer function
func (app *Application) AddUserToRoomServer() *cli.Command {
	return &cli.Command{
		Name:        "adduserserver",
		Aliases:     []string{"auserver"},
		Description: "Start a server which can add users to a room to which the default user has access to.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "port",
				Aliases:  []string{"p"},
				Value:    "8000",
				Usage:    "Port on which to run the server. Default is 8000",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "emaildomain",
				Aliases:  []string{"ed"},
				Value:    "email.com",
				Usage:    "Domain for user email address. Defaults to email.com",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {

			port := c.Int64("port")
			emaildomain := c.String("emaildomain")

			onboardApp := &AddUserToRoomServerApplication{Application: app, EmailDomain: emaildomain}
			r := chi.NewRouter()
			r.Use(middleware.RequestID)
			r.Use(middleware.Logger)
			r.Use(middleware.Recoverer)
			r.Get("/", onboardApp.index)
			r.Get("/{webexroom}", onboardApp.addUser)

			log.Println(fmt.Sprintf("Started server on :%d", port))
			http.ListenAndServe(fmt.Sprintf(":%d", port), r)
			return nil
		},
	}
}

func (app *AddUserToRoomServerApplication) index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("Hi, I can add you to webex rooms maintained by %s", app.Me.DisplayName)))
}

func (app *AddUserToRoomServerApplication) addUser(w http.ResponseWriter, r *http.Request) {
	tmproom := chi.URLParam(r, "webexroom")
	webexroom, err := app.parseRoomID(tmproom)
	if err != nil {
		log.Debugf("The room %s does not valid", webexroom)
		log.Debugf(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Invalid Room")))
		return
	}

	authSSOUser := r.Header.Get("auth_user")
	if authSSOUser == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Invalid User")))
		return
	}
	emailAddress := fmt.Sprintf("%s@%s", authSSOUser, app.EmailDomain)
	room, _, err := app.Client.Rooms.GetRoom(webexroom)
	if err != nil {
		log.Debugf("The room %s does not exists", webexroom)
		log.Debugf(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Invalid Room")))
		return
	}

	membershipQueryParams := &webexteams.ListMembershipsQueryParams{
		RoomID:      room.ID,
		PersonEmail: app.Email,
	}

	memberships, _, err := app.Client.Memberships.ListMemberships(membershipQueryParams)
	if err != nil {
		log.Debugf("Error getting membership for user for room %s", webexroom)
		log.Debugf(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Server Error (memberships)")))
		return
	}

	if len(memberships.Items) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("I am unable to add you to the room, because the configured User / Bot is not a member of the room")))
		return
	}

	err = app.createMember(room, email(emailAddress), false)
	if err != nil {
		log.Debugf("Error adding user to room %s", webexroom)
		log.Debugf(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Unable to add user to room")))
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(fmt.Sprintf("Added user to room")))
	return
}
