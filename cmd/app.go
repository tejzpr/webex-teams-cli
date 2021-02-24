package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	webexteams "github.com/jbogarin/go-cisco-webex-teams/sdk"
)

// Application struct
type Application struct {
	AccessToken  string
	Email        string
	DownloadsDir string
	Me           *webexteams.Person
	Client       *webexteams.Client
}

type email string

func (app *Application) createMember(room *webexteams.Room, email email, isModerator bool) error {
	membershipRequest := &webexteams.MembershipCreateRequest{
		RoomID:      room.ID,
		PersonEmail: string(email),
		IsModerator: isModerator,
	}

	_, _, err := app.Client.Memberships.CreateMembership(membershipRequest)
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) removeMember(room *webexteams.Room, email email) error {

	membershipQueryParams := &webexteams.ListMembershipsQueryParams{
		PersonEmail: string(email),
		Max:         1,
		RoomID:      room.ID,
	}

	memberships, _, err := app.Client.Memberships.ListMemberships(membershipQueryParams)
	if err != nil {
		return err
	}
	if len(memberships.Items) > 0 {
		membershipID := memberships.Items[0].ID
		_, err := app.Client.Memberships.DeleteMembership(membershipID)
		if err != nil {
			return err
		}
	}
	return nil
}

// SendMessageParams struct
type SendMessageParams struct {
	RoomID                   string
	PersonID                 string
	PersonEmail              string
	Text                     string
	Filename                 string
	RemoteFileRequestTimeout time.Duration
}

// SendMessage2Room send a message
func (app *Application) SendMessage2Room(params *SendMessageParams) (*webexteams.Message, error) {

	var parsedRoomID string
	var err error

	if params.RoomID == "" && params.PersonID == "" && params.PersonEmail == "" {
		return nil, errors.New("roomID or PersonID or PersonEmail is required")
	}

	if params.RoomID != "" {
		parsedRoomID, err = app.parseRoomID(params.RoomID)
		if err != nil {
			return nil, err
		}
	}

	var fileToSend webexteams.File

	if params.Filename != "" {
		if app.isValidUrl(params.Filename) {

			var netTransport = &http.Transport{
				Dial: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).Dial,
				TLSHandshakeTimeout: 5 * time.Second,
			}

			var netClient = &http.Client{
				Timeout:   time.Second * params.RemoteFileRequestTimeout,
				Transport: netTransport,
			}

			req, err := http.NewRequest("GET", params.Filename, nil)
			if err != nil {
				return nil, err
			}
			baseFileName := path.Base(req.URL.Path)
			fileExt := path.Ext(req.URL.Path)
			fileID := baseFileName[0 : len(baseFileName)-len(fileExt)]

			if fileID == "" {
				fileID = app.getMD5Hash(params.Filename)
			}
			httpResponse, err := netClient.Do(req)
			if err != nil {
				return nil, err
			}

			resp, err := ioutil.ReadAll(httpResponse.Body)
			if err != nil {
				return nil, err
			}
			_ = httpResponse.Body.Close()

			mime := mimetype.Detect(resp)

			reader := bytes.NewReader(resp)

			extn := mime.Extension()
			if fileExt != "" && strings.HasPrefix(mime.String(), "text/plain") {
				extn = fileExt
			}

			fileToSend = webexteams.File{
				Name:        fmt.Sprintf("%s%s", fileID, extn),
				Reader:      reader,
				ContentType: mime.String(),
			}
		} else {
			fileID, fileExt := app.getFilenameWithoutExtension(params.Filename)

			if strings.HasPrefix(params.Filename, "~") {
				home, err := os.UserHomeDir()
				if err != nil {
					return nil, err
				}
				tmpFilename := params.Filename[len("~"):]
				params.Filename = path.Join(home, tmpFilename)
			}

			absFilePath, err := filepath.Abs(params.Filename)
			if err != nil {
				return nil, err
			}

			if fileID == "" {
				fileID = app.getMD5Hash(absFilePath)
			}

			if _, err := os.Stat(absFilePath); os.IsNotExist(err) {
				return nil, err
			}

			resp, err := os.Open(absFilePath)
			if err != nil {
				return nil, err
			}
			defer resp.Close()

			mime, err := mimetype.DetectReader(resp)
			if err != nil {
				return nil, err
			}

			extn := mime.Extension()
			if fileExt != "" && strings.HasPrefix(mime.String(), "text/plain") {
				extn = fileExt
			}
			// Go to beginning of file because mimetype checker would have
			resp.Seek(0, io.SeekStart)
			fileToSend = webexteams.File{
				Name:        fmt.Sprintf("%s%s", fileID, extn),
				Reader:      resp,
				ContentType: mime.String(),
			}
		}
	}

	message := &webexteams.MessageCreateRequest{}

	if params.Text != "" {
		message.Markdown = params.Text
	}

	if fileToSend.Name != "" {
		message.Files = []webexteams.File{fileToSend}
	}

	if parsedRoomID != "" {
		room, _, err := app.Client.Rooms.GetRoom(parsedRoomID)
		if err != nil {
			return nil, err
		}
		message.RoomID = room.ID
	} else if params.PersonID != "" {
		message.ToPersonID = params.PersonID
	} else if params.PersonEmail != "" {
		message.ToPersonEmail = params.PersonEmail
	}

	newTextMessage, _, err := app.Client.Messages.CreateMessage(message)
	if err != nil {
		return nil, err
	}

	return newTextMessage, nil
}

// GetMessagesForRoomFromRoomID method
func (app *Application) GetMessagesForRoomFromRoomID(roomID string, max int) ([]webexteams.Message, error) {
	empty := make([]webexteams.Message, 0)
	const noMessage = "No messages found"
	if max == 0 {
		max = 10
	}

	messageQueryParams := &webexteams.ListMessagesQueryParams{
		RoomID: roomID,
		Max:    max,
	}

	response, _, err := app.Client.Messages.ListMessages(messageQueryParams)
	if err != nil {
		return empty, err
	}

	if len(response.Items) > 0 {
		return response.Items, nil
	}
	return empty, errors.New(noMessage)
}

// GetMessagesForRoom method
func (app *Application) GetMessagesForRoom(room *webexteams.Room, max int) ([]webexteams.Message, error) {
	roomID := room.ID
	empty := make([]webexteams.Message, 0)
	const noMessage = "No messages found"
	if max == 0 {
		max = 10
	}

	messageQueryParams := &webexteams.ListMessagesQueryParams{
		RoomID: roomID,
		Max:    max,
	}

	if room.RoomType == "group" {
		messageQueryParams.MentionedPeople = "me"
	}

	response, _, err := app.Client.Messages.ListMessages(messageQueryParams)
	if err != nil {
		return empty, err
	}

	if len(response.Items) > 0 {
		return response.Items, nil
	}
	return empty, errors.New(noMessage)
}
