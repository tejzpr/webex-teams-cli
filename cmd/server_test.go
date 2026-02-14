package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tejzpr/webex-go-sdk/v2/people"
	"github.com/urfave/cli/v2"
)

// --- AddUserToRoomServer tests ---

func TestAddUserServerIndex(t *testing.T) {
	app := &AddUserToRoomServerApplication{
		Application: &Application{
			Me: &people.Person{DisplayName: "Test User"},
		},
	}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	app.index(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	expectedBody := "Hi, I can add you to webex rooms maintained by Test User"
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, w.Body.String())
	}
}

// --- MessageRelayServer tests ---

func TestMessageRelayServerIndex(t *testing.T) {
	app := &MessageRelayServerApplication{
		Application: &Application{
			Me: &people.Person{DisplayName: "Test User"},
		},
	}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	app.index(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	expectedBody := "Hi, I can add you to webex rooms maintained by Test User"
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, w.Body.String())
	}
}

func TestMessageRelayAuthCheck(t *testing.T) {
	app := &MessageRelayServerApplication{
		MessagerelayKey: "test-key-12345",
	}

	tests := []struct {
		name          string
		headerValue   string
		expectedError bool
	}{
		{
			name:          "missing header",
			headerValue:   "",
			expectedError: true,
		},
		{
			name:          "wrong key",
			headerValue:   "wrong-key",
			expectedError: true,
		},
		{
			name:          "correct key",
			headerValue:   "test-key-12345",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/room123", nil)
			if tt.headerValue != "" {
				req.Header.Set("X-Message-Key", tt.headerValue)
			}

			err := app.authCheck(req)
			if (err != nil) != tt.expectedError {
				t.Errorf("authCheck() error = %v, expectedError %v", err, tt.expectedError)
			}

			if tt.expectedError && err != nil {
				expectedMsg := "Not Authorized"
				if err.Error() != expectedMsg {
					t.Errorf("Expected error %q, got %q", expectedMsg, err.Error())
				}
			}
		})
	}
}

// --- Test server creation (without actually starting servers) ---

// Helper function to find a flag by name (same as in commands_test.go)
func getFlagByNameServer(flags []cli.Flag, name string) cli.Flag {
	for _, flag := range flags {
		if flag.Names()[0] == name {
			return flag
		}
	}
	return nil
}

func TestAddUserToRoomServerCommand(t *testing.T) {
	app := &Application{}

	cmd := app.AddUserToRoomServer()

	// Test command structure
	if cmd.Name != "adduserserver" {
		t.Errorf("Expected command name 'adduserserver', got %q", cmd.Name)
	}

	if len(cmd.Aliases) != 1 || cmd.Aliases[0] != "auserver" {
		t.Errorf("Expected alias 'auserver', got %v", cmd.Aliases)
	}

	// Check required flags
	portFlag := getFlagByNameServer(cmd.Flags, "port")
	if portFlag == nil {
		t.Error("Expected 'port' flag")
	} else if portFlag.(*cli.StringFlag).Value != "8000" {
		t.Errorf("Expected default port 8000, got %s", portFlag.(*cli.StringFlag).Value)
	}

	emailDomainFlag := getFlagByNameServer(cmd.Flags, "emaildomain")
	if emailDomainFlag == nil {
		t.Error("Expected 'emaildomain' flag")
	} else if !emailDomainFlag.(*cli.StringFlag).Required {
		t.Error("Expected emaildomain flag to be required")
	}
}

func TestMessageRelayServerCommand(t *testing.T) {
	app := &Application{}

	cmd := app.MessageRelayServer()

	// Test command structure
	if cmd.Name != "messagerelayserver" {
		t.Errorf("Expected command name 'messagerelayserver', got %q", cmd.Name)
	}

	if len(cmd.Aliases) != 1 || cmd.Aliases[0] != "mrserver" {
		t.Errorf("Expected alias 'mrserver', got %v", cmd.Aliases)
	}

	// Check required flags
	portFlag := getFlagByNameServer(cmd.Flags, "port")
	if portFlag == nil {
		t.Error("Expected 'port' flag")
	} else if portFlag.(*cli.StringFlag).Value != "8000" {
		t.Errorf("Expected default port 8000, got %s", portFlag.(*cli.StringFlag).Value)
	}

	keyFlag := getFlagByNameServer(cmd.Flags, "messagerelaykey")
	if keyFlag == nil {
		t.Error("Expected 'messagerelaykey' flag")
	} else if !keyFlag.(*cli.StringFlag).Required {
		t.Error("Expected messagerelaykey flag to be required")
	}
}

// --- Test sendMessagePOST auth rejection ---

func TestMessageRelaySendMessageUnauthorized(t *testing.T) {
	app := &MessageRelayServerApplication{
		Application: &Application{
			Me: &people.Person{DisplayName: "Test User"},
		},
		MessagerelayKey: "test-key",
	}

	// Missing auth key → 401
	req := httptest.NewRequest("POST", "/room123", strings.NewReader("Hello"))
	w := httptest.NewRecorder()

	app.sendMessagePOST(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
	if !strings.Contains(w.Body.String(), "Not Authorized") {
		t.Errorf("Expected body to contain 'Not Authorized', got %q", w.Body.String())
	}
}

func TestMessageRelaySendMessageEmptyBody(t *testing.T) {
	app := &MessageRelayServerApplication{
		Application: &Application{
			Me: &people.Person{DisplayName: "Test User"},
		},
		MessagerelayKey: "test-key",
	}

	// Authenticated but empty body → 400
	req := httptest.NewRequest("POST", "/room123", strings.NewReader(""))
	req.Header.Set("X-Message-Key", "test-key")
	w := httptest.NewRecorder()

	app.sendMessagePOST(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "Empty Message") {
		t.Errorf("Expected body to contain 'Empty Message', got %q", w.Body.String())
	}
}
