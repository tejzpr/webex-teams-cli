package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/tejzpr/webex-go-sdk/v2/memberships"
	"github.com/urfave/cli/v2"
)

// MessageRelayServerApplication struct
type MessageRelayServerApplication struct {
	*Application
	MessagerelayKey string
}

// MessageRelayServer function
func (app *Application) MessageRelayServer() *cli.Command {
	return &cli.Command{
		Name:        "messagerelayserver",
		Aliases:     []string{"mrserver"},
		Description: "Start a server which can relay incoming messages to a webex room.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "port",
				Aliases:  []string{"p"},
				Value:    "8000",
				Usage:    "Port on which to run the server. Default is 8000",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "messagerelaykey",
				Aliases:  []string{"mrkey"},
				Value:    "",
				Usage:    "A key of length greater than 256, that would be used to establish authenticity of calls",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {

			port := c.Int64("port")

			messagerelaykey := c.String("messagerelaykey")
			if len(messagerelaykey) < 256 {
				return errors.New("The messagerelaykey should be longer than 255 characters")
			}

			relayApp := &MessageRelayServerApplication{Application: app, MessagerelayKey: messagerelaykey}
			r := chi.NewRouter()
			r.Use(middleware.RequestID)
			r.Use(middleware.Logger)
			r.Use(middleware.Recoverer)
			r.Get("/", relayApp.index)
			r.Post("/{webexroom}", relayApp.sendMessagePOST)

			log.Println(fmt.Sprintf("Started server on :%d", port))
			http.ListenAndServe(fmt.Sprintf(":%d", port), r)
			return nil
		},
	}
}

func (app *MessageRelayServerApplication) index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("Hi, I can add you to webex rooms maintained by %s", app.Me.DisplayName)))
}

func (app *MessageRelayServerApplication) authCheck(r *http.Request) error {
	messageKey := r.Header.Get("X-Message-Key")
	if messageKey == "" {
		return fmt.Errorf("Not Authorized")
	} else if messageKey != app.MessagerelayKey {
		return fmt.Errorf("Not Authorized")
	}
	return nil
}

func (app *MessageRelayServerApplication) sendMessagePOST(w http.ResponseWriter, r *http.Request) {

	err := app.authCheck(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}

	tmproom := chi.URLParam(r, "webexroom")
	webexroom, err := app.parseRoomID(tmproom)
	if err != nil {
		log.Debugf("The room %s does not valid", webexroom)
		log.Debug(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Invalid Room")))
		return
	}

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Invalid Message Body")))
		return
	}

	message := string(body)

	if message == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Empty Message")))
		return
	}

	room, err := app.Client.Rooms().Get(webexroom)
	if err != nil {
		log.Debugf("The room %s does not exists", webexroom)
		log.Debug(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Invalid Room")))
		return
	}

	membershipQueryParams := &memberships.ListOptions{
		RoomID:      room.ID,
		PersonEmail: app.Email,
	}

	mbrPage, err := app.Client.Memberships().List(membershipQueryParams)
	if err != nil {
		log.Debugf("Error getting membership for user for room %s", webexroom)
		log.Debug(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Server Error (memberships)")))
		return
	}

	if len(mbrPage.Items) <= 0 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("I am unable to send a message to the room, because the configured User / Bot is not a member of the room")))
		return
	}

	messageParams := &SendMessageParams{RoomID: room.ID, Text: message}

	sentMsg, err := app.SendMessage2Room(messageParams)
	if err != nil {
		log.Debugf("Error sending message to room %s", webexroom)
		log.Debug(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("I am unable to send a message to the room.")))
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(fmt.Sprintf("Sent message %s to room", sentMsg.ID)))
	return
}
