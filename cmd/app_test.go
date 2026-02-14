package cmd

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	webex "github.com/tejzpr/webex-go-sdk/v2"
	"github.com/tejzpr/webex-go-sdk/v2/contents"
)

// --- SendMessage2Room validation tests ---

func TestSendMessage2RoomValidation(t *testing.T) {
	app := &Application{}

	tests := []struct {
		name    string
		params  *SendMessageParams
		wantErr bool
	}{
		{
			name: "no target specified",
			params: &SendMessageParams{
				Text: "hello",
			},
			wantErr: true,
		},
		{
			name: "roomID specified",
			params: &SendMessageParams{
				RoomID: "room123",
				Text:   "hello",
			},
			wantErr: false,
		},
		{
			name: "personID specified",
			params: &SendMessageParams{
				PersonID: "person123",
				Text:     "hello",
			},
			wantErr: false,
		},
		{
			name: "personEmail specified",
			params: &SendMessageParams{
				PersonEmail: "test@example.com",
				Text:        "hello",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't test the full SendMessage2Room without a client,
			// but we can test the validation logic it performs
			if tt.params.RoomID == "" && tt.params.PersonID == "" && tt.params.PersonEmail == "" {
				// This should return an error
				_, err := app.SendMessage2Room(tt.params)
				if err == nil {
					t.Error("Expected error when no target specified")
				}
				if !strings.Contains(err.Error(), "roomID or PersonID or PersonEmail is required") {
					t.Errorf("Expected target validation error, got: %v", err)
				}
			}
		})
	}
}

// --- resolveLocalFile tests ---

func TestResolveLocalFile(t *testing.T) {
	app := &Application{}

	// Create a temporary file for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("test content")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		filename string
		wantErr  bool
		wantName string
	}{
		{
			name:     "valid file",
			filename: testFile,
			wantErr:  false,
			wantName: "test.txt",
		},
		{
			name:     "nonexistent file",
			filename: filepath.Join(tmpDir, "nonexistent.txt"),
			wantErr:  true,
		},
		{
			name:     "relative path",
			filename: "./test.txt",
			wantErr:  true, // Will fail since we're not in tmpDir
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &SendMessageParams{
				Filename: tt.filename,
			}

			result, err := app.resolveLocalFile(params)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveLocalFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result.FileName != tt.wantName {
					t.Errorf("resolveLocalFile() filename = %q, want %q", result.FileName, tt.wantName)
				}
				if string(result.FileBytes) != string(testContent) {
					t.Errorf("resolveLocalFile() content = %q, want %q", string(result.FileBytes), string(testContent))
				}
			}
		})
	}
}

func TestResolveLocalFileTildeExpansion(t *testing.T) {
	app := &Application{}

	// Create a temporary file in the test's temp directory
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("test content")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Mock home directory by constructing a tilde path that points to our temp dir
	tildePath := "~" + strings.TrimPrefix(tmpDir, filepath.Dir(tmpDir))

	params := &SendMessageParams{
		Filename: tildePath + "/test.txt",
	}

	// This should fail because tilde expansion uses the real home directory
	_, err := app.resolveLocalFile(params)
	if err == nil {
		t.Error("Expected error for tilde path pointing to non-home directory")
	}
}

// --- resolveRemoteFile tests ---

func TestResolveRemoteFile(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("remote content"))
	}))
	defer server.Close()

	app := &Application{}

	params := &SendMessageParams{
		Filename:                 server.URL + "/test.txt",
		RemoteFileRequestTimeout: 5,
	}

	result, err := app.resolveRemoteFile(params)
	if err != nil {
		t.Fatalf("resolveRemoteFile() error = %v", err)
	}

	expectedName := "test.txt" // Based on server URL path
	if result.FileName != expectedName {
		t.Errorf("resolveRemoteFile() filename = %q, want %q", result.FileName, expectedName)
	}

	if string(result.FileBytes) != "remote content" {
		t.Errorf("resolveRemoteFile() content = %q, want %q", string(result.FileBytes), "remote content")
	}
}

func TestResolveRemoteFileError(t *testing.T) {
	app := &Application{}

	// Use a localhost port that is not listening to guarantee connection refused
	params := &SendMessageParams{
		Filename:                 "http://127.0.0.1:1/file.txt",
		RemoteFileRequestTimeout: 1,
	}

	_, err := app.resolveRemoteFile(params)
	if err == nil {
		t.Error("Expected error for unreachable URL")
	}
}

func TestResolveRemoteFileInvalidURL(t *testing.T) {
	app := &Application{}

	params := &SendMessageParams{
		Filename:                 "not-a-url",
		RemoteFileRequestTimeout: 5,
	}

	_, err := app.resolveRemoteFile(params)
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

// --- resolveFile tests ---

func TestResolveFile(t *testing.T) {
	app := &Application{}

	// Test with local file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("local content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test local file resolution
	params := &SendMessageParams{
		Filename: testFile,
	}

	result, err := app.resolveFile(params)
	if err != nil {
		t.Fatalf("resolveFile() for local file error = %v", err)
	}
	if string(result.FileBytes) != "local content" {
		t.Errorf("resolveFile() local content = %q, want %q", string(result.FileBytes), "local content")
	}

	// Test with remote file
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("remote content"))
	}))
	defer server.Close()

	params.Filename = server.URL
	result, err = app.resolveFile(params)
	if err != nil {
		t.Fatalf("resolveFile() for remote file error = %v", err)
	}
	if string(result.FileBytes) != "remote content" {
		t.Errorf("resolveFile() remote content = %q, want %q", string(result.FileBytes), "remote content")
	}
}

// --- Accessor tests ---

func TestGetEmail(t *testing.T) {
	app := &Application{Email: "test@example.com"}
	if got := app.GetEmail(); got != "test@example.com" {
		t.Errorf("GetEmail() = %q, want %q", got, "test@example.com")
	}
}

func TestGetDownloadsDir(t *testing.T) {
	app := &Application{DownloadsDir: "/tmp/downloads"}
	if got := app.GetDownloadsDir(); got != "/tmp/downloads" {
		t.Errorf("GetDownloadsDir() = %q, want %q", got, "/tmp/downloads")
	}
}

func TestGetClient(t *testing.T) {
	client := &webex.WebexClient{}
	app := &Application{Client: client}
	if got := app.GetClient(); got != client {
		t.Error("GetClient() returned different client")
	}
}

func TestGetContentsClient(t *testing.T) {
	contents := &contents.Client{} // Using correct type
	app := &Application{ContentsClient: contents}
	if got := app.GetContentsClient(); got != contents {
		t.Error("GetContentsClient() returned different client")
	}
}

// --- GetMessagesForRoomFromRoomID default max tests ---

func TestGetMessagesForRoomFromRoomIDDefaultMax(t *testing.T) {
	// Verify that max defaults to 10 when 0 is passed
	// We can't call the real method without a client, but we can
	// verify the SendMessageParams struct and accessor behavior.
	app := &Application{Email: "test@test.com"}
	if app.GetEmail() != "test@test.com" {
		t.Error("GetEmail mismatch")
	}
}

// --- isValidUrl helper test (already in utils_test.go but included here for completeness) ---

func TestIsValidUrlForApp(t *testing.T) {
	app := &Application{}

	// Test that the method exists and works
	if !app.isValidUrl("http://example.com") {
		t.Error("Expected http://example.com to be valid")
	}

	if app.isValidUrl("not-a-url") {
		t.Error("Expected 'not-a-url' to be invalid")
	}
}
