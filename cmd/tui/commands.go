package tui

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tejzpr/webex-go-sdk/v2/messages"
)

// loadRoomsCmd fetches the room list asynchronously
func (m *Model) loadRoomsCmd() tea.Msg {
	roomList, err := m.app.GetRooms(200, "")
	if err != nil {
		return msgErr{err: err}
	}
	return msgRoomsLoaded{rooms: roomList}
}

// loadHistoryCmd fetches message history for the current room
func (m *Model) loadHistoryCmd() tea.Msg {
	if m.currentRoomID == "" {
		return nil
	}
	msgs, err := m.app.GetMessagesForRoomFromRoomID(m.currentRoomID, 50)
	if err != nil {
		return msgErr{err: err}
	}
	return msgHistoryLoaded{msgs: msgs}
}

// sendMessageCmd sends a text message to the current room, optionally as a thread reply
func (m *Model) sendMessageCmd(text string, parentID string) tea.Cmd {
	return func() tea.Msg {
		if m.currentRoomID == "" {
			return msgErr{err: fmt.Errorf("no room selected")}
		}
		msg := &messages.Message{
			RoomID:   m.currentRoomID,
			Markdown: text,
		}
		if parentID != "" {
			msg.ParentID = parentID
		}
		result, err := m.app.GetClient().Messages().Create(msg)
		return msgSendResult{msg: result, err: err}
	}
}

// sendFileCmd sends a file to the current room, optionally as a thread reply
func (m *Model) sendFileCmd(filePath string, parentID string) tea.Cmd {
	return func() tea.Msg {
		if m.currentRoomID == "" {
			return msgErr{err: fmt.Errorf("no room selected")}
		}
		fileBytes, err := os.ReadFile(filePath)
		if err != nil {
			return msgSendResult{err: err}
		}
		baseName := path.Base(filePath)
		msg := &messages.Message{
			RoomID: m.currentRoomID,
		}
		if parentID != "" {
			msg.ParentID = parentID
		}
		upload := &messages.FileUpload{
			FileName:  baseName,
			FileBytes: fileBytes,
		}
		result, err := m.app.GetClient().Messages().CreateWithAttachment(msg, upload)
		return msgSendResult{msg: result, err: err}
	}
}

// startListenerCmd starts the WebSocket message listener
func (m *Model) startListenerCmd() tea.Cmd {
	return func() tea.Msg {
		err := m.app.GetClient().Messages().Listen(func(msg *messages.Message) {
			if msg.RoomID == m.currentRoomID {
				m.incomingCh <- msg
			}
		})
		if err != nil {
			return msgErr{err: err}
		}
		return msgListening{}
	}
}

// waitForMessageCmd waits for the next incoming message from the channel
func (m *Model) waitForMessageCmd() tea.Cmd {
	return func() tea.Msg {
		msg := <-m.incomingCh
		return msgNewMessage{msg: msg}
	}
}

// fetchFileCmd downloads a file attachment and returns it with content type info
func (m *Model) fetchFileCmd(contentURL string) tea.Cmd {
	return func() tea.Msg {
		fileInfo, err := m.app.GetContentsClient().DownloadFromURL(contentURL)
		if err != nil {
			return msgFileReady{err: err}
		}
		fileName := parseContentDisposition(fileInfo.ContentDisposition)
		if fileName == "" {
			fileName = "attachment"
		}
		return msgFileReady{
			data:        fileInfo.Data,
			contentType: fileInfo.ContentType,
			fileName:    fileName,
		}
	}
}

// saveFileCmd saves file data to a specific path on disk
func saveFileCmd(data []byte, savePath string) tea.Cmd {
	return func() tea.Msg {
		err := os.WriteFile(savePath, data, 0644)
		if err != nil {
			return msgFileDownloaded{err: err}
		}
		return msgFileDownloaded{path: savePath}
	}
}

// isImageContentType checks if a content type is an image
func isImageContentType(ct string) bool {
	return strings.HasPrefix(ct, "image/")
}

// isVideoContentType checks if a content type is a video
func isVideoContentType(ct string) bool {
	return strings.HasPrefix(ct, "video/")
}

// sendParams holds parameters for sending a message
type sendParams struct {
	roomID string
	text   string
}

// fileNamesFromURLs returns placeholder labels for Webex content URLs.
// Webex file URLs look like /v1/contents/<base64id> — the path segment
// is not a human-readable filename. We show "File attachment" as a
// placeholder until the real name is resolved asynchronously.
func fileNamesFromURLs(urls []string) []string {
	names := make([]string, 0, len(urls))
	for i := range urls {
		names = append(names, fmt.Sprintf("File attachment %d", i+1))
	}
	return names
}

// resolveFileNamesCmd fetches real filenames for a message's attachments
// by downloading metadata from the Webex Contents API.
func (m *Model) resolveFileNamesCmd(messageID string, fileURLs []string) tea.Cmd {
	return func() tea.Msg {
		names := make([]string, len(fileURLs))
		for i, fileURL := range fileURLs {
			fileInfo, err := m.app.GetContentsClient().DownloadFromURL(fileURL)
			if err != nil {
				names[i] = fmt.Sprintf("File attachment %d", i+1)
				continue
			}
			name := parseContentDisposition(fileInfo.ContentDisposition)
			if name == "" {
				names[i] = fmt.Sprintf("File attachment %d", i+1)
			} else {
				names[i] = name
			}
		}
		return msgFileNamesResolved{messageID: messageID, names: names}
	}
}

// parseContentDisposition extracts filename from Content-Disposition header
func parseContentDisposition(header string) string {
	if header == "" {
		return ""
	}
	// Look for filename="..." or filename=...
	for _, part := range strings.Split(header, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "filename=") {
			name := strings.TrimPrefix(part, "filename=")
			name = strings.Trim(name, "\"' ")
			return name
		}
	}
	return ""
}

// loadFilePickerDir reads directory entries for the file picker
func loadFilePickerDir(dir string) ([]os.DirEntry, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(absDir)
	if err != nil {
		return nil, err
	}
	// Filter out hidden files
	var visible []os.DirEntry
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), ".") {
			visible = append(visible, e)
		}
	}
	return visible, nil
}
