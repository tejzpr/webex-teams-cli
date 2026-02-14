package cmd

import (
	"fmt"
	"testing"

	"github.com/tejzpr/webex-go-sdk/v2/memberships"
	"github.com/tejzpr/webex-go-sdk/v2/people"
	"github.com/tejzpr/webex-go-sdk/v2/rooms"
)

// Helper to create test data
func createTestPerson(id string) *people.Person {
	return &people.Person{
		ID: id,
	}
}

func createTestRoom(id, creatorID string) *rooms.Room {
	return &rooms.Room{
		ID:        id,
		CreatorID: creatorID,
		Type:      "group",
	}
}

func createTestMembership(roomID, personEmail string, isModerator bool) memberships.Membership {
	return memberships.Membership{
		RoomID:      roomID,
		PersonEmail: personEmail,
		IsModerator: isModerator,
	}
}

// --- AddPeopleApplication checkAccess tests ---

func TestAddPeopleCheckAccess(t *testing.T) {
	me := createTestPerson("user123")
	room := createTestRoom("room123", "user123")                          // I am the creator
	membership := createTestMembership("room123", "me@example.com", true) // I am moderator

	tests := []struct {
		name   string
		access string
		want   bool
	}{
		{"access all", "a", true},
		{"access owner - I am creator", "o", true},
		{"access moderator - I am moderator", "m", true},
		{"access owner+moderator - both true", "om", true},
		{"access owner - I am not creator", "o", false},
		{"access moderator - I am not moderator", "m", false},
		{"access owner+moderator - neither", "om", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &AddPeopleApplication{
				Application: &Application{},
				Access:      tt.access,
			}

			// Modify test data for negative cases
			testMe := me
			testRoom := room
			testMembership := membership

			if tt.name == "access owner - I am not creator" {
				testRoom = createTestRoom("room123", "other123")
			}
			if tt.name == "access moderator - I am not moderator" {
				testMembership = createTestMembership("room123", "me@example.com", false)
			}
			if tt.name == "access owner+moderator - neither" {
				testRoom = createTestRoom("room123", "other123")
				testMembership = createTestMembership("room123", "me@example.com", false)
			}

			got := app.checkAccess(testMe, testRoom, testMembership)
			if got != tt.want {
				t.Errorf("checkAccess() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- BroadcastToRoomsApplication checkAccess tests ---

func TestBroadcastToRoomsCheckAccess(t *testing.T) {
	me := createTestPerson("user123")
	room := createTestRoom("room123", "user123")                          // I am the creator
	membership := createTestMembership("room123", "me@example.com", true) // I am moderator

	tests := []struct {
		name   string
		access string
		want   bool
	}{
		{"access all", "a", true},
		{"access owner - I am creator", "o", true},
		{"access moderator - I am moderator", "m", true},
		{"access owner+moderator - both true", "om", true},
		{"access owner - I am not creator", "o", false},
		{"access moderator - I am not moderator", "m", false},
		{"access owner+moderator - neither", "om", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &BroadcastToRoomsApplication{
				Application: &Application{},
				Access:      tt.access,
			}

			// Modify test data for negative cases
			testMe := me
			testRoom := room
			testMembership := membership

			if tt.name == "access owner - I am not creator" {
				testRoom = createTestRoom("room123", "other123")
			}
			if tt.name == "access moderator - I am not moderator" {
				testMembership = createTestMembership("room123", "me@example.com", false)
			}
			if tt.name == "access owner+moderator - neither" {
				testRoom = createTestRoom("room123", "other123")
				testMembership = createTestMembership("room123", "me@example.com", false)
			}

			got := app.checkAccess(testMe, testRoom, testMembership)
			if got != tt.want {
				t.Errorf("checkAccess() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- RemovePeopleApplication checkAccess tests ---

func TestRemovePeopleCheckAccess(t *testing.T) {
	me := createTestPerson("user123")
	room := createTestRoom("room123", "user123")                          // I am the creator
	membership := createTestMembership("room123", "me@example.com", true) // I am moderator

	tests := []struct {
		name   string
		access string
		want   bool
	}{
		{"access all", "a", true},
		{"access owner - I am creator", "o", true},
		{"access moderator - I am moderator", "m", true},
		{"access owner+moderator - both true", "om", true},
		{"access owner - I am not creator", "o", false},
		{"access moderator - I am not moderator", "m", false},
		{"access owner+moderator - neither", "om", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &RemovePeopleApplication{
				Application: &Application{},
				Access:      tt.access,
			}

			// Modify test data for negative cases
			testMe := me
			testRoom := room
			testMembership := membership

			if tt.name == "access owner - I am not creator" {
				testRoom = createTestRoom("room123", "other123")
			}
			if tt.name == "access moderator - I am not moderator" {
				testMembership = createTestMembership("room123", "me@example.com", false)
			}
			if tt.name == "access owner+moderator - neither" {
				testRoom = createTestRoom("room123", "other123")
				testMembership = createTestMembership("room123", "me@example.com", false)
			}

			got := app.checkAccess(testMe, testRoom, testMembership)
			if got != tt.want {
				t.Errorf("checkAccess() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- Edge cases ---

func TestCheckAccessInvalidAccessValue(t *testing.T) {
	me := createTestPerson("user123")
	room := createTestRoom("room123", "user123")
	membership := createTestMembership("room123", "me@example.com", true)

	// Test with invalid access values - should default to false
	apps := []interface {
		checkAccess(*people.Person, *rooms.Room, memberships.Membership) bool
	}{
		&AddPeopleApplication{Application: &Application{}, Access: "invalid"},
		&BroadcastToRoomsApplication{Application: &Application{}, Access: "invalid"},
		&RemovePeopleApplication{Application: &Application{}, Access: "invalid"},
	}

	for i, app := range apps {
		t.Run(fmt.Sprintf("invalid access %d", i), func(t *testing.T) {
			got := app.checkAccess(me, room, membership)
			if got != false {
				t.Errorf("checkAccess() with invalid access should return false, got %v", got)
			}
		})
	}
}

func TestCheckAccessEmptyAccessValue(t *testing.T) {
	me := createTestPerson("user123")
	room := createTestRoom("room123", "user123")
	membership := createTestMembership("room123", "me@example.com", true)

	// Test with empty access values - should default to false
	apps := []interface {
		checkAccess(*people.Person, *rooms.Room, memberships.Membership) bool
	}{
		&AddPeopleApplication{Application: &Application{}, Access: ""},
		&BroadcastToRoomsApplication{Application: &Application{}, Access: ""},
		&RemovePeopleApplication{Application: &Application{}, Access: ""},
	}

	for i, app := range apps {
		t.Run(fmt.Sprintf("empty access %d", i), func(t *testing.T) {
			got := app.checkAccess(me, room, membership)
			if got != false {
				t.Errorf("checkAccess() with empty access should return false, got %v", got)
			}
		})
	}
}
