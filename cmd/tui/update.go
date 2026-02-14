package tui

import (
	"bytes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/blacktop/go-termimg"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// makeChatEntry builds a chatEntry from message fields
func (m *Model) makeChatEntry(personEmail, text, parentID, messageID string, files []string, created interface{}) chatEntry {
	entry := chatEntry{
		senderEmail: personEmail,
		text:        text,
		isSelf:      personEmail == m.app.GetEmail(),
		hasFiles:    len(files) > 0,
		fileURLs:    files,
		parentID:    parentID,
		messageID:   messageID,
	}
	if len(files) > 0 {
		entry.fileNames = fileNamesFromURLs(files)
	}
	return entry
}

// getReplyParentID returns the parentID to use when sending, based on replyToMsg
func (m *Model) getReplyParentID() string {
	if m.replyToMsg == nil {
		return ""
	}
	// If replying to a thread reply, use its parentID (the thread root)
	if m.replyToMsg.parentID != "" {
		return m.replyToMsg.parentID
	}
	// Otherwise reply to this message as thread root
	return m.replyToMsg.messageID
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		chatW := m.width - sidebarWidth - 1
		if chatW < 10 {
			chatW = 10
		}
		vpHeight := m.mainHeight() - 4
		if vpHeight < 1 {
			vpHeight = 1
		}
		m.viewport.Width = chatW
		m.viewport.Height = vpHeight
		m.textInput.Width = chatW - 6
		m.sidebarInput.Width = sidebarWidth - 4
		if m.currentRoomID != "" {
			m.viewport.SetContent(m.renderChatContent())
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case msgRoomsLoaded:
		m.allRooms = msg.rooms
		m.filterRooms()

	case msgRoomSelected:
		m.currentRoomID = msg.room.ID
		m.currentRoomTitle = msg.room.Title
		m.chatEntries = nil
		m.seenIDs = make(map[string]bool)
		m.chatCursor = 0
		m.replyToMsg = nil
		m.listening = false
		m.focus = focusInput
		m.sidebarInput.Blur()
		m.textInput.Focus()
		cmds = append(cmds, m.loadHistoryCmd)
		cmds = append(cmds, m.startListenerCmd())

	case msgHistoryLoaded:
		for i := len(msg.msgs) - 1; i >= 0; i-- {
			item := msg.msgs[i]
			if m.seenIDs[item.ID] {
				continue
			}
			m.seenIDs[item.ID] = true
			entry := m.makeChatEntry(item.PersonEmail, item.Text, item.ParentID, item.ID, item.Files, item.Created)
			entry.created = item.Created
			m.chatEntries = append(m.chatEntries, entry)
			// Queue async filename resolution for messages with files
			if len(item.Files) > 0 {
				cmds = append(cmds, m.resolveFileNamesCmd(item.ID, item.Files))
			}
		}
		m.chatCursor = len(m.chatEntries) - 1
		if m.chatCursor < 0 {
			m.chatCursor = 0
		}
		m.viewport.SetContent(m.renderChatContent())
		m.viewport.GotoBottom()
		cmds = append(cmds, m.waitForMessageCmd())

	case msgNewMessage:
		if msg.msg != nil && !m.seenIDs[msg.msg.ID] {
			m.seenIDs[msg.msg.ID] = true
			entry := m.makeChatEntry(msg.msg.PersonEmail, msg.msg.Text, msg.msg.ParentID, msg.msg.ID, msg.msg.Files, msg.msg.Created)
			entry.created = msg.msg.Created
			m.chatEntries = append(m.chatEntries, entry)
			m.viewport.SetContent(m.renderChatContent())
			m.viewport.GotoBottom()
			// Resolve filenames for new messages with attachments
			if len(msg.msg.Files) > 0 {
				cmds = append(cmds, m.resolveFileNamesCmd(msg.msg.ID, msg.msg.Files))
			}
		}
		cmds = append(cmds, m.waitForMessageCmd())

	case msgSendResult:
		if msg.err != nil {
			m.errMsg = msg.err.Error()
		} else {
			m.errMsg = ""
			m.replyToMsg = nil
			if msg.msg != nil && !m.seenIDs[msg.msg.ID] {
				m.seenIDs[msg.msg.ID] = true
				entry := m.makeChatEntry(msg.msg.PersonEmail, msg.msg.Text, msg.msg.ParentID, msg.msg.ID, msg.msg.Files, msg.msg.Created)
				entry.isSelf = true
				entry.created = msg.msg.Created
				m.chatEntries = append(m.chatEntries, entry)
				m.viewport.SetContent(m.renderChatContent())
				m.viewport.GotoBottom()
			}
		}

	case msgFileNamesResolved:
		// Update the chatEntry with resolved real filenames
		for i := range m.chatEntries {
			if m.chatEntries[i].messageID == msg.messageID {
				m.chatEntries[i].fileNames = msg.names
				break
			}
		}
		m.viewport.SetContent(m.renderChatContent())

	case msgFileReady:
		if msg.err != nil {
			m.errMsg = "Fetch failed: " + msg.err.Error()
		} else if isImageContentType(msg.contentType) {
			// Image — display inline with termimg, suspend alt screen
			m.pendingFileData = msg.data
			m.pendingFileName = msg.fileName
			cmds = append(cmds, m.viewImageCmd(msg.data, msg.fileName))
		} else {
			// Non-image file — open directory picker to choose save location
			m.pendingFileData = msg.data
			m.pendingFileName = msg.fileName
			home, err := os.UserHomeDir()
			if err != nil {
				home = "."
			}
			entries, err := loadFilePickerDir(home)
			if err != nil {
				m.errMsg = "Cannot open save picker: " + err.Error()
			} else {
				m.prevFocus = m.focus
				m.focus = focusFilePicker
				m.filePickerFn = filePickerModeSave
				m.filePickerDir = home
				m.filePickerEntries = entries
				m.filePickerCursor = 0
				m.textInput.Blur()
			}
		}

	case msgImageRendered:
		// Show image as a dismissable overlay
		m.showingImage = true
		m.imageText = msg.text
		m.imageFileName = msg.fileName
		m.imageData = m.pendingFileData
		m.pendingFileData = nil
		m.pendingFileName = ""
		m.prevFocus = m.focus
		m.focus = focusImageViewer
		m.textInput.Blur()

	case msgImageViewed:
		// Returned from image viewer, clear state
		m.showingImage = false
		m.imageText = ""
		m.imageData = nil
		m.imageFileName = ""

	case msgListening:
		m.listening = true

	case msgFileDownloaded:
		if msg.err != nil {
			m.errMsg = "Download failed: " + msg.err.Error()
		} else {
			m.errMsg = ""
			m.chatEntries = append(m.chatEntries, chatEntry{
				senderEmail: "system",
				text:        "Downloaded: " + msg.path,
				isSelf:      false,
			})
			m.viewport.SetContent(m.renderChatContent())
			m.viewport.GotoBottom()
		}

	case msgErr:
		m.errMsg = msg.err.Error()
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys
	if key.Matches(msg, m.keys.Quit) {
		return m, tea.Quit
	}

	if key.Matches(msg, m.keys.Help) {
		m.showHelp = !m.showHelp
		return m, nil
	}

	// Image viewer intercepts all keys when active
	if m.focus == focusImageViewer {
		return m.handleImageViewerKey(msg)
	}

	// File picker intercepts all keys when active
	if m.focus == focusFilePicker {
		return m.handleFilePickerKey(msg)
	}

	// "/" focuses sidebar search from anywhere (except file picker and search itself)
	if key.Matches(msg, m.keys.Search) && m.focus != focusSidebarSearch && m.focus != focusInput {
		m.focus = focusSidebarSearch
		m.textInput.Blur()
		m.sidebarInput.Focus()
		return m, nil
	}

	// ctrl+f opens file picker from input or chat
	if key.Matches(msg, m.keys.AttachFile) && (m.focus == focusInput || m.focus == focusChat) && m.currentRoomID != "" {
		return m.openFilePicker()
	}

	// ctrl+r sets reply target from input or chat
	if key.Matches(msg, m.keys.Reply) && (m.focus == focusInput || m.focus == focusChat) {
		if len(m.chatEntries) > 0 && m.chatCursor >= 0 && m.chatCursor < len(m.chatEntries) {
			entry := m.chatEntries[m.chatCursor]
			m.replyToMsg = &entry
			m.focus = focusInput
			m.textInput.Focus()
		}
		return m, nil
	}

	// ctrl+d fetches attachment from chat-focused message
	if key.Matches(msg, m.keys.Download) && (m.focus == focusChat || m.focus == focusInput) {
		if len(m.chatEntries) > 0 && m.chatCursor >= 0 && m.chatCursor < len(m.chatEntries) {
			entry := m.chatEntries[m.chatCursor]
			if entry.hasFiles && len(entry.fileURLs) > 0 {
				fileURL := entry.fileURLs[0]
				return m, m.fetchFileCmd(fileURL)
			}
		}
		return m, nil
	}

	// Escape cancels reply mode
	if key.Matches(msg, m.keys.Escape) {
		if m.replyToMsg != nil {
			m.replyToMsg = nil
			return m, nil
		}
	}

	// Tab cycles focus: sidebar → input → chat → sidebar
	if key.Matches(msg, m.keys.Tab) {
		m.sidebarInput.Blur()
		m.textInput.Blur()
		switch m.focus {
		case focusSidebar, focusSidebarSearch:
			m.focus = focusInput
			m.textInput.Focus()
		case focusInput:
			m.focus = focusChat
		case focusChat:
			m.focus = focusSidebar
		}
		return m, nil
	}

	// Dispatch to focused pane
	switch m.focus {
	case focusSidebar:
		return m.handleSidebarKey(msg)
	case focusSidebarSearch:
		return m.handleSidebarSearchKey(msg)
	case focusInput:
		return m.handleInputKey(msg)
	case focusChat:
		return m.handleChatKey(msg)
	}

	return m, nil
}

func (m Model) handleSidebarKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.String() == "a":
		m.roomTypeFilter = "all"
		m.filterRooms()
		return m, nil
	case msg.String() == "g":
		m.roomTypeFilter = "group"
		m.filterRooms()
		return m, nil
	case msg.String() == "d":
		m.roomTypeFilter = "direct"
		m.filterRooms()
		return m, nil

	case key.Matches(msg, m.keys.ScrollUp):
		if m.roomCursor > 0 {
			m.roomCursor--
		}
		return m, nil

	case key.Matches(msg, m.keys.ScrollDn):
		if m.roomCursor < len(m.filteredRooms)-1 {
			m.roomCursor++
		}
		return m, nil

	case key.Matches(msg, m.keys.Send):
		if len(m.filteredRooms) > 0 && m.roomCursor < len(m.filteredRooms) {
			selected := m.filteredRooms[m.roomCursor]
			return m, func() tea.Msg {
				return msgRoomSelected{room: selected}
			}
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleSidebarSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Escape):
		m.focus = focusSidebar
		m.sidebarInput.Blur()
		return m, nil

	case key.Matches(msg, m.keys.Send):
		m.focus = focusSidebar
		m.sidebarInput.Blur()
		return m, nil

	default:
		var cmd tea.Cmd
		m.sidebarInput, cmd = m.sidebarInput.Update(msg)
		m.filterRooms()
		return m, cmd
	}
}

func (m Model) handleInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Escape) {
		if m.replyToMsg != nil {
			m.replyToMsg = nil
			return m, nil
		}
	}

	if key.Matches(msg, m.keys.Send) {
		text := strings.TrimSpace(m.textInput.Value())
		if text == "" {
			return m, nil
		}
		m.textInput.Reset()
		m.errMsg = ""

		parentID := m.getReplyParentID()

		// Check for sendfile command
		if strings.HasPrefix(text, "sendfile ") {
			filePath := strings.TrimSpace(text[len("sendfile "):])
			return m, m.sendFileCmd(filePath, parentID)
		}

		return m, m.sendMessageCmd(text, parentID)
	}

	if key.Matches(msg, m.keys.ClearChat) {
		m.chatEntries = nil
		m.chatCursor = 0
		m.replyToMsg = nil
		m.viewport.SetContent(m.renderChatContent())
		return m, nil
	}

	// Pass key to textinput
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m Model) handleChatKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.ScrollUp):
		if m.chatCursor > 0 {
			m.chatCursor--
		}
		m.viewport.SetContent(m.renderChatContent())
		m.viewport.LineUp(1)
		return m, nil
	case key.Matches(msg, m.keys.ScrollDn):
		if m.chatCursor < len(m.chatEntries)-1 {
			m.chatCursor++
		}
		m.viewport.SetContent(m.renderChatContent())
		m.viewport.LineDown(1)
		return m, nil
	case key.Matches(msg, m.keys.PageUp):
		m.viewport.HalfViewUp()
		return m, nil
	case key.Matches(msg, m.keys.PageDown):
		m.viewport.HalfViewDown()
		return m, nil
	case key.Matches(msg, m.keys.Send):
		// Enter on a message in chat mode → set as reply target and switch to input
		if len(m.chatEntries) > 0 && m.chatCursor >= 0 && m.chatCursor < len(m.chatEntries) {
			entry := m.chatEntries[m.chatCursor]
			m.replyToMsg = &entry
			m.focus = focusInput
			m.textInput.Focus()
		}
		return m, nil
	case key.Matches(msg, m.keys.ClearChat):
		m.chatEntries = nil
		m.chatCursor = 0
		m.replyToMsg = nil
		m.viewport.SetContent(m.renderChatContent())
		return m, nil
	}
	return m, nil
}

// openFilePicker opens the file picker overlay
func (m Model) openFilePicker() (tea.Model, tea.Cmd) {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	entries, err := loadFilePickerDir(home)
	if err != nil {
		m.errMsg = "Cannot open file picker: " + err.Error()
		return m, nil
	}
	m.prevFocus = m.focus
	m.focus = focusFilePicker
	m.filePickerDir = home
	m.filePickerEntries = entries
	m.filePickerCursor = 0
	m.textInput.Blur()
	return m, nil
}

func (m Model) handleFilePickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Escape):
		// Close file picker, return to previous focus
		m.focus = m.prevFocus
		if m.focus == focusInput {
			m.textInput.Focus()
		}
		m.filePickerEntries = nil
		return m, nil

	case key.Matches(msg, m.keys.ScrollUp):
		if m.filePickerCursor > 0 {
			m.filePickerCursor--
		}
		return m, nil

	case key.Matches(msg, m.keys.ScrollDn):
		if m.filePickerCursor < len(m.filePickerEntries)-1 {
			m.filePickerCursor++
		}
		return m, nil

	case msg.String() == "backspace":
		// Go up one directory
		parent := filepath.Dir(m.filePickerDir)
		if parent != m.filePickerDir {
			entries, err := loadFilePickerDir(parent)
			if err != nil {
				m.errMsg = err.Error()
				return m, nil
			}
			m.filePickerDir = parent
			m.filePickerEntries = entries
			m.filePickerCursor = 0
		}
		return m, nil

	case key.Matches(msg, m.keys.Send):
		if len(m.filePickerEntries) == 0 {
			return m, nil
		}
		selected := m.filePickerEntries[m.filePickerCursor]
		fullPath := filepath.Join(m.filePickerDir, selected.Name())

		if selected.IsDir() {
			// Enter directory
			entries, err := loadFilePickerDir(fullPath)
			if err != nil {
				m.errMsg = err.Error()
				return m, nil
			}
			m.filePickerDir = fullPath
			m.filePickerEntries = entries
			m.filePickerCursor = 0
			return m, nil
		}

		if m.filePickerFn == filePickerModeSave {
			// In save mode, only directories can be selected — save file here
			// But user selected a file, ignore (they need to select a dir or press save in a dir)
			return m, nil
		}

		// File selected — send it
		parentID := m.getReplyParentID()
		m.focus = m.prevFocus
		if m.focus == focusInput {
			m.textInput.Focus()
		}
		m.filePickerEntries = nil
		m.filePickerFn = filePickerModeSend
		return m, m.sendFileCmd(fullPath, parentID)

	case msg.String() == "s" && m.filePickerFn == filePickerModeSave:
		// "s" in save mode = save file to current directory
		if m.pendingFileData == nil {
			return m, nil
		}
		savePath := path.Join(m.filePickerDir, m.pendingFileName)
		m.focus = m.prevFocus
		if m.focus == focusInput {
			m.textInput.Focus()
		}
		m.filePickerEntries = nil
		data := m.pendingFileData
		m.pendingFileData = nil
		m.pendingFileName = ""
		m.filePickerFn = filePickerModeSend
		return m, saveFileCmd(data, savePath)
	}

	return m, nil
}

// handleImageViewerKey handles keys when the image viewer overlay is active
func (m Model) handleImageViewerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Escape):
		// Close image viewer
		m.showingImage = false
		m.imageText = ""
		m.imageData = nil
		m.imageFileName = ""
		m.focus = m.prevFocus
		if m.focus == focusInput {
			m.textInput.Focus()
		}
		return m, nil

	case msg.String() == "s":
		// Save image to local filesystem — open directory picker
		if m.imageData == nil {
			return m, nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		entries, err := loadFilePickerDir(home)
		if err != nil {
			m.errMsg = "Cannot open save picker: " + err.Error()
			return m, nil
		}
		// Transfer image data to pending file state for the save picker
		m.pendingFileData = m.imageData
		m.pendingFileName = m.imageFileName
		// Close image viewer overlay
		m.showingImage = false
		m.imageText = ""
		m.imageData = nil
		m.imageFileName = ""
		// Open file picker in save mode
		m.focus = focusFilePicker
		m.filePickerFn = filePickerModeSave
		m.filePickerDir = home
		m.filePickerEntries = entries
		m.filePickerCursor = 0
		m.textInput.Blur()
		return m, nil
	}

	return m, nil
}

// viewImageCmd renders an image as text (halfblocks/sixel/kitty) and displays it in the chat viewport
func (m *Model) viewImageCmd(data []byte, fileName string) tea.Cmd {
	return func() tea.Msg {
		// Decode image from bytes
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return msgErr{err: err}
		}

		// Render using go-termimg (auto-detects best protocol, falls back to halfblocks)
		widget := termimg.NewImageWidgetFromImage(img)
		widget.SetSize(60, 20)
		rendered, err := widget.Render()
		if err != nil {
			return msgErr{err: err}
		}

		return msgImageRendered{text: rendered, fileName: fileName}
	}
}
