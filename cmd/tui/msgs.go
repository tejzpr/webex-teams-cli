package tui

import (
	"time"

	"github.com/tejzpr/webex-go-sdk/v2/messages"
	"github.com/tejzpr/webex-go-sdk/v2/rooms"
)

// --- Bubbletea messages (events) ---

// msgRoomsLoaded is sent when the room list has been fetched
type msgRoomsLoaded struct {
	rooms []rooms.Room
}

// msgRoomSelected is sent when the user picks a room from the list
type msgRoomSelected struct {
	room rooms.Room
}

// msgHistoryLoaded is sent when initial message history arrives
type msgHistoryLoaded struct {
	msgs []messages.Message
}

// msgNewMessage is sent when a real-time message arrives via WebSocket
type msgNewMessage struct {
	msg *messages.Message
}

// msgSendResult is sent after attempting to send a message
type msgSendResult struct {
	msg *messages.Message
	err error
}

// msgErr is a generic error event
type msgErr struct {
	err error
}

// msgListening is sent when the WebSocket listener is connected
type msgListening struct{}

// msgFileDownloaded is sent when a file download completes (save to disk)
type msgFileDownloaded struct {
	path string
	err  error
}

// msgFileReady is sent when a file attachment has been fetched and is ready for viewing/saving
type msgFileReady struct {
	data        []byte
	contentType string
	fileName    string
	err         error
}

// msgImageViewed is sent after the user dismisses the inline image viewer
type msgImageViewed struct{}

// msgImageRendered is sent when an image has been rendered to text for display in the viewport
type msgImageRendered struct {
	text     string
	fileName string
}

// msgFileNamesResolved is sent when real filenames are fetched for a message's attachments
type msgFileNamesResolved struct {
	messageID string
	names     []string
}

// chatEntry is a single rendered line in the chat viewport
type chatEntry struct {
	senderEmail string
	text        string
	isSelf      bool
	hasFiles    bool
	fileURLs    []string
	fileNames   []string
	parentID    string
	messageID   string
	created     *time.Time
}
