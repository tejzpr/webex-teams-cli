package tui

import (
	"os"
	"path/filepath"
	"testing"

	webex "github.com/tejzpr/webex-go-sdk/v2"
	"github.com/tejzpr/webex-go-sdk/v2/contents"
	"github.com/tejzpr/webex-go-sdk/v2/messages"
	"github.com/tejzpr/webex-go-sdk/v2/rooms"
)

// --- mock app for testing ---

type mockApp struct {
	email string
}

func (m *mockApp) GetRooms(max int, roomType string) ([]rooms.Room, error) {
	return nil, nil
}
func (m *mockApp) GetMessagesForRoomFromRoomID(roomID string, max int) ([]messages.Message, error) {
	return nil, nil
}
func (m *mockApp) GetEmail() string              { return m.email }
func (m *mockApp) GetDownloadsDir() string       { return "/tmp" }
func (m *mockApp) GetClient() *webex.WebexClient { return nil }
func (m *mockApp) GetContentsClient() *contents.Client {
	return nil
}

// --- parseContentDisposition tests ---

func TestParseContentDisposition(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{"empty header", "", ""},
		{"simple filename", `attachment; filename="report.pdf"`, "report.pdf"},
		{"filename without quotes", `attachment; filename=report.pdf`, "report.pdf"},
		{"filename with single quotes", `attachment; filename='report.pdf'`, "report.pdf"},
		{"multiple parts", `inline; filename="image.png"; size=12345`, "image.png"},
		{"no filename", `inline; size=12345`, ""},
		{"filename with spaces", `attachment; filename="my file.txt"`, "my file.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseContentDisposition(tt.header)
			if got != tt.want {
				t.Errorf("parseContentDisposition(%q) = %q, want %q", tt.header, got, tt.want)
			}
		})
	}
}

// --- loadFilePickerDir tests ---

func TestLoadFilePickerDir(t *testing.T) {
	entries, err := loadFilePickerDir(t.TempDir())
	if err != nil {
		t.Fatalf("loadFilePickerDir on temp dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries in empty temp dir, got %d", len(entries))
	}

	_, err = loadFilePickerDir("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("expected error for nonexistent directory, got nil")
	}
}

func TestLoadFilePickerDirFiltersHidden(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".hidden"), []byte("hidden"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "visible.txt"), []byte("visible"), 0644); err != nil {
		t.Fatal(err)
	}

	entries, err := loadFilePickerDir(dir)
	if err != nil {
		t.Fatalf("loadFilePickerDir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 visible entry, got %d", len(entries))
	}
	if entries[0].Name() != "visible.txt" {
		t.Errorf("expected visible.txt, got %s", entries[0].Name())
	}
}

// --- saveFileCmd tests ---

func TestSaveFileCmd(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "saved.txt")
	data := []byte("test content")

	cmd := saveFileCmd(data, savePath)
	msg := cmd()

	result, ok := msg.(msgFileDownloaded)
	if !ok {
		t.Fatalf("expected msgFileDownloaded, got %T", msg)
	}
	if result.err != nil {
		t.Fatalf("unexpected error: %v", result.err)
	}
	if result.path != savePath {
		t.Errorf("expected path %q, got %q", savePath, result.path)
	}

	// Verify file was written
	got, err := os.ReadFile(savePath)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}
	if string(got) != "test content" {
		t.Errorf("expected 'test content', got %q", string(got))
	}
}

func TestSaveFileCmdError(t *testing.T) {
	cmd := saveFileCmd([]byte("data"), "/nonexistent/dir/file.txt")
	msg := cmd()

	result, ok := msg.(msgFileDownloaded)
	if !ok {
		t.Fatalf("expected msgFileDownloaded, got %T", msg)
	}
	if result.err == nil {
		t.Error("expected error for invalid path, got nil")
	}
}

// --- makeChatEntry tests ---

func TestMakeChatEntry(t *testing.T) {
	m := &Model{
		app: &mockApp{email: "me@test.com"},
	}

	entry := m.makeChatEntry("me@test.com", "hello", "parent1", "msg1", []string{"url1"}, nil)

	if !entry.isSelf {
		t.Error("expected isSelf=true for matching email")
	}
	if entry.text != "hello" {
		t.Errorf("expected text 'hello', got %q", entry.text)
	}
	if entry.parentID != "parent1" {
		t.Errorf("expected parentID 'parent1', got %q", entry.parentID)
	}
	if entry.messageID != "msg1" {
		t.Errorf("expected messageID 'msg1', got %q", entry.messageID)
	}
	if !entry.hasFiles {
		t.Error("expected hasFiles=true")
	}
	if len(entry.fileNames) != 1 || entry.fileNames[0] != "File attachment 1" {
		t.Errorf("unexpected fileNames: %v", entry.fileNames)
	}

	entry2 := m.makeChatEntry("other@test.com", "world", "", "msg2", nil, nil)
	if entry2.isSelf {
		t.Error("expected isSelf=false for different email")
	}
	if entry2.hasFiles {
		t.Error("expected hasFiles=false for no files")
	}
}

// --- getReplyParentID tests ---

func TestGetReplyParentID(t *testing.T) {
	m := &Model{}

	if id := m.getReplyParentID(); id != "" {
		t.Errorf("expected empty, got %q", id)
	}

	m.replyToMsg = &chatEntry{messageID: "root1", parentID: ""}
	if id := m.getReplyParentID(); id != "root1" {
		t.Errorf("expected 'root1', got %q", id)
	}

	m.replyToMsg = &chatEntry{messageID: "reply1", parentID: "root1"}
	if id := m.getReplyParentID(); id != "root1" {
		t.Errorf("expected 'root1' (thread root), got %q", id)
	}
}

// --- filterRooms tests ---

func TestFilterRooms(t *testing.T) {
	m := &Model{
		allRooms: []rooms.Room{
			{ID: "1", Title: "Engineering Team", Type: "group"},
			{ID: "2", Title: "John Doe", Type: "direct"},
			{ID: "3", Title: "Design Team", Type: "group"},
			{ID: "4", Title: "Jane Smith", Type: "direct"},
		},
		roomTypeFilter: "all",
	}

	m.filterRooms()
	if len(m.filteredRooms) != 4 {
		t.Errorf("all filter: expected 4 rooms, got %d", len(m.filteredRooms))
	}

	m.roomTypeFilter = "group"
	m.filterRooms()
	if len(m.filteredRooms) != 2 {
		t.Errorf("group filter: expected 2 rooms, got %d", len(m.filteredRooms))
	}
	for _, r := range m.filteredRooms {
		if r.Type != "group" {
			t.Errorf("group filter: got room type %q", r.Type)
		}
	}

	m.roomTypeFilter = "direct"
	m.filterRooms()
	if len(m.filteredRooms) != 2 {
		t.Errorf("direct filter: expected 2 rooms, got %d", len(m.filteredRooms))
	}

	m.roomTypeFilter = "all"
	m.sidebarInput.SetValue("eng")
	m.filterRooms()
	if len(m.filteredRooms) != 1 || m.filteredRooms[0].Title != "Engineering Team" {
		t.Errorf("search filter: expected 1 room 'Engineering Team', got %v", m.filteredRooms)
	}
}

// --- mainHeight tests ---

func TestMainHeight(t *testing.T) {
	m := Model{height: 40, showHelp: false}
	h := m.mainHeight()
	if h != 38 {
		t.Errorf("expected mainHeight=38, got %d", h)
	}

	m.showHelp = true
	h = m.mainHeight()
	if h != 35 {
		t.Errorf("expected mainHeight=35 with help, got %d", h)
	}

	m.height = 2
	m.showHelp = false
	h = m.mainHeight()
	if h != 4 {
		t.Errorf("expected minimum mainHeight=4, got %d", h)
	}
}

// --- filePickerMode tests ---

func TestFilePickerModeConstants(t *testing.T) {
	if filePickerModeSend != 0 {
		t.Errorf("expected filePickerModeSend=0, got %d", filePickerModeSend)
	}
	if filePickerModeSave != 1 {
		t.Errorf("expected filePickerModeSave=1, got %d", filePickerModeSave)
	}
}

// --- focus constants tests ---

func TestFocusConstants(t *testing.T) {
	if focusSidebar != 0 {
		t.Error("focusSidebar should be 0")
	}
	if focusImageViewer != 5 {
		t.Errorf("focusImageViewer should be 5, got %d", focusImageViewer)
	}
}

// --- NewModel tests ---

func TestNewModel(t *testing.T) {
	app := &mockApp{email: "test@test.com"}
	m := NewModel(app)

	if m.focus != focusSidebar {
		t.Errorf("expected initial focus=focusSidebar, got %d", m.focus)
	}
	if m.roomTypeFilter != "all" {
		t.Errorf("expected initial roomTypeFilter='all', got %q", m.roomTypeFilter)
	}
	if m.seenIDs == nil {
		t.Error("expected seenIDs to be initialized")
	}
	if m.incomingCh == nil {
		t.Error("expected incomingCh to be initialized")
	}
}

// --- message type tests ---

func TestMsgFileReady(t *testing.T) {
	msg := msgFileReady{
		data:        []byte("imagedata"),
		contentType: "image/png",
		fileName:    "photo.png",
	}
	if msg.contentType != "image/png" {
		t.Errorf("expected contentType 'image/png', got %q", msg.contentType)
	}
	if msg.fileName != "photo.png" {
		t.Errorf("expected fileName 'photo.png', got %q", msg.fileName)
	}
	if !isImageContentType(msg.contentType) {
		t.Error("expected image content type to be detected")
	}
}

func TestMsgImageRendered(t *testing.T) {
	msg := msgImageRendered{
		text:     "rendered image text",
		fileName: "test.png",
	}
	if msg.text != "rendered image text" {
		t.Errorf("unexpected text: %q", msg.text)
	}
	if msg.fileName != "test.png" {
		t.Errorf("unexpected fileName: %q", msg.fileName)
	}
}
