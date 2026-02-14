package tui

import (
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tejzpr/webex-go-sdk/v2/contents"
	"github.com/tejzpr/webex-go-sdk/v2/messages"
	"github.com/tejzpr/webex-go-sdk/v2/rooms"

	webex "github.com/tejzpr/webex-go-sdk/v2"
)

// AppProvider is the interface the TUI needs from the application layer.
// This avoids an import cycle between cmd and cmd/tui.
type AppProvider interface {
	GetRooms(max int, roomType string) ([]rooms.Room, error)
	GetMessagesForRoomFromRoomID(roomID string, max int) ([]messages.Message, error)
	GetEmail() string
	GetDownloadsDir() string
	GetClient() *webex.WebexClient
	GetContentsClient() *contents.Client
}

// focus tracks which pane has keyboard focus
type focus int

const (
	focusSidebar focus = iota
	focusSidebarSearch
	focusInput
	focusChat
	focusFilePicker
	focusImageViewer
)

const sidebarWidth = 28

// filePickerMode controls what the file picker does on selection
type filePickerMode int

const (
	filePickerModeSend filePickerMode = iota // select a file to send as attachment
	filePickerModeSave                       // select a directory to save a downloaded file
)

// Model is the top-level Bubbletea model for the interactive TUI
type Model struct {
	// Application reference (provides SDK client, helpers)
	app AppProvider

	// UI components
	viewport     viewport.Model
	textInput    textinput.Model
	sidebarInput textinput.Model
	spinner      spinner.Model
	keys         keyMap

	// Room state
	allRooms         []rooms.Room
	filteredRooms    []rooms.Room
	roomCursor       int
	roomTypeFilter   string // "all", "group", "direct"
	currentRoomID    string
	currentRoomTitle string

	// Chat state
	chatEntries []chatEntry
	seenIDs     map[string]bool
	chatCursor  int
	replyToMsg  *chatEntry

	// File picker state
	filePickerDir     string
	filePickerEntries []os.DirEntry
	filePickerCursor  int
	filePickerFn      filePickerMode
	prevFocus         focus

	// Pending file for save/view
	pendingFileData []byte
	pendingFileName string

	// Image viewer overlay
	showingImage  bool
	imageText     string
	imageData     []byte
	imageFileName string

	// UI state
	focus     focus
	showHelp  bool
	listening bool
	width     int
	height    int
	errMsg    string

	// Channels
	incomingCh chan *messages.Message
}

// NewModel creates a fully initialised Model ready for tea.NewProgram
func NewModel(app AppProvider) Model {
	ti := textinput.New()
	ti.Placeholder = "Type a message…"
	ti.CharLimit = 4096
	ti.Width = 60

	si := textinput.New()
	si.Placeholder = "Search rooms…"
	si.CharLimit = 128
	si.Width = sidebarWidth - 4

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = spinnerStyle

	vp := viewport.New(60, 20)
	vp.SetContent("")

	return Model{
		app:            app,
		textInput:      ti,
		sidebarInput:   si,
		spinner:        sp,
		viewport:       vp,
		keys:           defaultKeyMap(),
		seenIDs:        make(map[string]bool),
		roomTypeFilter: "all",
		focus:          focusSidebar,
		incomingCh:     make(chan *messages.Message, 64),
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.spinner.Tick,
		m.loadRoomsCmd,
	)
}

// filterRooms applies the current search text and type filter to allRooms
func (m *Model) filterRooms() {
	query := strings.ToLower(m.sidebarInput.Value())
	var result []rooms.Room
	for _, r := range m.allRooms {
		// Type filter
		if m.roomTypeFilter != "all" && r.Type != m.roomTypeFilter {
			continue
		}
		// Search filter
		if query != "" && !strings.Contains(strings.ToLower(r.Title), query) {
			continue
		}
		result = append(result, r)
	}
	m.filteredRooms = result
	if m.roomCursor >= len(m.filteredRooms) {
		m.roomCursor = 0
	}
}
