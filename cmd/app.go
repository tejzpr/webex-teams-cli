package cmd

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	webex "github.com/tejzpr/webex-go-sdk/v2"
	"github.com/tejzpr/webex-go-sdk/v2/contents"
	"github.com/tejzpr/webex-go-sdk/v2/memberships"
	"github.com/tejzpr/webex-go-sdk/v2/messages"
	"github.com/tejzpr/webex-go-sdk/v2/people"
	"github.com/tejzpr/webex-go-sdk/v2/rooms"
)

// Application struct
type Application struct {
	AccessToken    string
	Email          string
	DownloadsDir   string
	Me             *people.Person
	Client         *webex.WebexClient
	ContentsClient *contents.Client
}

type email string

func (app *Application) createMember(room *rooms.Room, email email, isModerator bool) error {
	membership := &memberships.Membership{
		RoomID:      room.ID,
		PersonEmail: string(email),
		IsModerator: isModerator,
	}

	_, err := app.Client.Memberships().Create(membership)
	if err != nil {
		return fmt.Errorf("error adding %s: %w", email, err)
	}

	log.Infof("Added %s to %s", email, room.Title)
	return nil
}

func (app *Application) removeMember(room *rooms.Room, email email) error {
	opts := &memberships.ListOptions{
		PersonEmail: string(email),
		Max:         1,
		RoomID:      room.ID,
	}

	page, err := app.Client.Memberships().List(opts)
	if err != nil {
		return err
	}
	if len(page.Items) > 0 {
		err := app.Client.Memberships().Delete(page.Items[0].ID)
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
func (app *Application) SendMessage2Room(params *SendMessageParams) (*messages.Message, error) {

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

	// Build the message
	msg := &messages.Message{}

	if params.Text != "" {
		msg.Markdown = params.Text
	}

	if parsedRoomID != "" {
		room, err := app.Client.Rooms().Get(parsedRoomID)
		if err != nil {
			return nil, err
		}
		msg.RoomID = room.ID
	} else if params.PersonID != "" {
		msg.ToPersonID = params.PersonID
	} else if params.PersonEmail != "" {
		msg.ToPersonEmail = params.PersonEmail
	}

	// Handle file attachment
	if params.Filename != "" {
		fileUpload, err := app.resolveFile(params)
		if err != nil {
			return nil, err
		}
		return app.Client.Messages().CreateWithAttachment(msg, fileUpload)
	}

	return app.Client.Messages().Create(msg)
}

// resolveFile resolves a filename (local path or remote URL) into a FileUpload
func (app *Application) resolveFile(params *SendMessageParams) (*messages.FileUpload, error) {
	if app.isValidUrl(params.Filename) {
		return app.resolveRemoteFile(params)
	}
	return app.resolveLocalFile(params)
}

func (app *Application) resolveRemoteFile(params *SendMessageParams) (*messages.FileUpload, error) {
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
	defer httpResponse.Body.Close()

	fileBytes, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}

	fileName := fmt.Sprintf("%s%s", fileID, fileExt)
	return &messages.FileUpload{
		FileName:  fileName,
		FileBytes: fileBytes,
	}, nil
}

func (app *Application) resolveLocalFile(params *SendMessageParams) (*messages.FileUpload, error) {
	filename := params.Filename

	if strings.HasPrefix(filename, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		tmpFilename := filename[len("~"):]
		filename = path.Join(home, tmpFilename)
	}

	absFilePath, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(absFilePath); os.IsNotExist(err) {
		return nil, err
	}

	fileBytes, err := os.ReadFile(absFilePath)
	if err != nil {
		return nil, err
	}

	baseName := path.Base(absFilePath)
	return &messages.FileUpload{
		FileName:  baseName,
		FileBytes: fileBytes,
	}, nil
}

// GetMessagesForRoomFromRoomID method
func (app *Application) GetMessagesForRoomFromRoomID(roomID string, max int) ([]messages.Message, error) {
	empty := make([]messages.Message, 0)
	const noMessage = "No messages found"
	if max == 0 {
		max = 10
	}

	opts := &messages.ListOptions{
		RoomID: roomID,
		Max:    max,
	}

	page, err := app.Client.Messages().List(opts)
	if err != nil {
		return empty, err
	}

	if len(page.Items) > 0 {
		return page.Items, nil
	}
	return empty, errors.New(noMessage)
}

// GetMessagesForRoom method
func (app *Application) GetMessagesForRoom(room *rooms.Room, max int) ([]messages.Message, error) {
	roomID := room.ID
	empty := make([]messages.Message, 0)
	const noMessage = "No messages found"
	if max == 0 {
		max = 10
	}

	opts := &messages.ListOptions{
		RoomID: roomID,
		Max:    max,
	}

	if room.Type == "group" {
		opts.MentionedPeople = "me"
	}

	page, err := app.Client.Messages().List(opts)
	if err != nil {
		return empty, err
	}

	if len(page.Items) > 0 {
		return page.Items, nil
	}
	return empty, errors.New(noMessage)
}

// GetEmail returns the authenticated user's email
func (app *Application) GetEmail() string {
	return app.Email
}

// GetDownloadsDir returns the configured downloads directory
func (app *Application) GetDownloadsDir() string {
	return app.DownloadsDir
}

// GetClient returns the Webex SDK client
func (app *Application) GetClient() *webex.WebexClient {
	return app.Client
}

// GetContentsClient returns the contents API client
func (app *Application) GetContentsClient() *contents.Client {
	return app.ContentsClient
}

// GetRooms retrieves rooms sorted by last activity
func (app *Application) GetRooms(max int, roomType string) ([]rooms.Room, error) {
	opts := &rooms.ListOptions{
		Max:    max,
		Type:   roomType,
		SortBy: "lastactivity",
	}

	page, err := app.Client.Rooms().List(opts)
	if err != nil {
		return make([]rooms.Room, 0), err
	}
	return page.Items, nil
}
