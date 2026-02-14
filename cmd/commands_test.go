package cmd

import (
	"testing"

	"github.com/urfave/cli/v2"
)

// Helper function to find a flag by name
func getFlagByName(flags []cli.Flag, name string) cli.Flag {
	for _, flag := range flags {
		if flag.Names()[0] == name {
			return flag
		}
	}
	return nil
}

// --- Test RoomCMD structure ---

func TestRoomCMDStructure(t *testing.T) {
	app := &Application{}
	cmd := app.RoomCMD()

	if cmd.Name != "room" {
		t.Errorf("Expected command name 'room', got %q", cmd.Name)
	}

	expectedAliases := []string{"r"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	} else {
		for i, alias := range expectedAliases {
			if cmd.Aliases[i] != alias {
				t.Errorf("Expected alias %q, got %q", alias, cmd.Aliases[i])
			}
		}
	}

	// Check flags
	expectedFlags := []string{"roomID", "toPersonID", "toPersonEmail"}
	if len(cmd.Flags) != len(expectedFlags) {
		t.Errorf("Expected %d flags, got %d", len(expectedFlags), len(cmd.Flags))
	}

	for _, flagName := range expectedFlags {
		flag := getFlagByName(cmd.Flags, flagName)
		if flag == nil {
			t.Errorf("Expected flag %q not found", flagName)
		}
	}

	// Check subcommands
	expectedSubcommands := []string{"message", "addmembers", "exportmembers", "removemembers", "broadcast"}
	if len(cmd.Subcommands) != len(expectedSubcommands) {
		t.Errorf("Expected %d subcommands, got %d", len(expectedSubcommands), len(cmd.Subcommands))
	}

	for i, expected := range expectedSubcommands {
		if cmd.Subcommands[i].Name != expected {
			t.Errorf("Expected subcommand %q, got %q", expected, cmd.Subcommands[i].Name)
		}
	}
}

// --- Test ChatCMD structure ---

func TestChatCMDStructure(t *testing.T) {
	app := &Application{}
	cmd := app.ChatCMD()

	if cmd.Name != "chat" {
		t.Errorf("Expected command name 'chat', got %q", cmd.Name)
	}

	expectedAliases := []string{"c"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	} else {
		for i, alias := range expectedAliases {
			if cmd.Aliases[i] != alias {
				t.Errorf("Expected alias %q, got %q", alias, cmd.Aliases[i])
			}
		}
	}

	if cmd.Usage != "Launch interactive Webex chat TUI" {
		t.Errorf("Expected usage 'Launch interactive Webex chat TUI', got %q", cmd.Usage)
	}
}

// --- Test WebexUtils structure ---

func TestWebexUtilsStructure(t *testing.T) {
	app := &Application{}
	cmd := app.WebexUtils()

	if cmd.Name != "utils" {
		t.Errorf("Expected command name 'utils', got %q", cmd.Name)
	}

	expectedAliases := []string{"u"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	} else {
		for i, alias := range expectedAliases {
			if cmd.Aliases[i] != alias {
				t.Errorf("Expected alias %q, got %q", alias, cmd.Aliases[i])
			}
		}
	}

	// Check subcommands
	expectedSubcommands := []string{"findroom", "listrooms"}
	if len(cmd.Subcommands) != len(expectedSubcommands) {
		t.Errorf("Expected %d subcommands, got %d", len(expectedSubcommands), len(cmd.Subcommands))
	}

	for i, expected := range expectedSubcommands {
		if cmd.Subcommands[i].Name != expected {
			t.Errorf("Expected subcommand %q, got %q", expected, cmd.Subcommands[i].Name)
		}
	}
}

// --- Test SendMessageCMD structure ---

func TestSendMessageCMDStructure(t *testing.T) {
	app := &Application{}
	cmd := app.SendMessageToRoomCMD()

	if cmd.Name != "message" {
		t.Errorf("Expected command name 'message', got %q", cmd.Name)
	}

	expectedAliases := []string{"msg"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	} else {
		for i, alias := range expectedAliases {
			if cmd.Aliases[i] != alias {
				t.Errorf("Expected alias %q, got %q", alias, cmd.Aliases[i])
			}
		}
	}

	// Check specific flags
	expectedFlags := []string{"text", "file", "remoteFileRequestTimeout"}
	for _, flagName := range expectedFlags {
		flag := getFlagByName(cmd.Flags, flagName)
		if flag == nil {
			t.Errorf("Expected flag %q not found", flagName)
		}
	}

	// Check flag defaults
	textFlag := getFlagByName(cmd.Flags, "text")
	if textFlag == nil || textFlag.(*cli.StringFlag).Value != "" {
		t.Errorf("Expected text flag default to be empty string")
	}

	fileFlag := getFlagByName(cmd.Flags, "file")
	if fileFlag == nil || fileFlag.(*cli.StringFlag).Value != "" {
		t.Errorf("Expected file flag default to be empty string")
	}

	timeoutFlag := getFlagByName(cmd.Flags, "remoteFileRequestTimeout")
	if timeoutFlag == nil || timeoutFlag.(*cli.StringFlag).Value != "" {
		t.Errorf("Expected remoteFileRequestTimeout flag default to be empty string")
	}
}

// --- Test AddPeopleCMD structure ---

func TestAddPeopleCMDStructure(t *testing.T) {
	app := &Application{}
	cmd := app.AddPeopleCMD()

	if cmd.Name != "addmembers" {
		t.Errorf("Expected command name 'addmembers', got %q", cmd.Name)
	}

	expectedAliases := []string{"am"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	}

	// Check required flags
	requiredFlags := []string{"memberscsv"}
	for _, flagName := range requiredFlags {
		flag := getFlagByName(cmd.Flags, flagName)
		if flag == nil {
			t.Errorf("Expected required flag %q not found", flagName)
		}
		if !flag.(*cli.StringFlag).Required {
			t.Errorf("Expected flag %q to be required", flagName)
		}
	}

	// Check optional flags
	optionalFlags := []string{"confirm", "access", "roomsidscsv"}
	for _, flagName := range optionalFlags {
		flag := getFlagByName(cmd.Flags, flagName)
		if flag == nil {
			t.Errorf("Expected optional flag %q not found", flagName)
		}
	}

	// Check default values
	accessFlag := getFlagByName(cmd.Flags, "access")
	if accessFlag == nil || accessFlag.(*cli.StringFlag).Value != "om" {
		t.Errorf("Expected access flag default to be 'om'")
	}

	confirmFlag := getFlagByName(cmd.Flags, "confirm")
	if confirmFlag == nil || confirmFlag.(*cli.StringFlag).Value != "n" {
		t.Errorf("Expected confirm flag default to be 'n'")
	}
}

// --- Test RemovePeopleCMD structure ---

func TestRemovePeopleCMDStructure(t *testing.T) {
	app := &Application{}
	cmd := app.RemovePeopleCMD()

	if cmd.Name != "removemembers" {
		t.Errorf("Expected command name 'removemembers', got %q", cmd.Name)
	}

	expectedAliases := []string{"rm"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	}

	// Check required flags
	requiredFlags := []string{"memberscsv"}
	for _, flagName := range requiredFlags {
		flag := getFlagByName(cmd.Flags, flagName)
		if flag == nil {
			t.Errorf("Expected required flag %q not found", flagName)
		}
		if !flag.(*cli.StringFlag).Required {
			t.Errorf("Expected flag %q to be required", flagName)
		}
	}

	// Check default values
	accessFlag := getFlagByName(cmd.Flags, "access")
	if accessFlag == nil || accessFlag.(*cli.StringFlag).Value != "om" {
		t.Errorf("Expected access flag default to be 'om'")
	}
}

// --- Test BroadcastCMD structure ---

func TestBroadcastCMDStructure(t *testing.T) {
	app := &Application{}
	cmd := app.BroadcastToRoomsCMD()

	if cmd.Name != "broadcast" {
		t.Errorf("Expected command name 'broadcast', got %q", cmd.Name)
	}

	expectedAliases := []string{"bc"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	}

	// Check required flags
	requiredFlags := []string{"text"}
	for _, flagName := range requiredFlags {
		flag := getFlagByName(cmd.Flags, flagName)
		if flag == nil {
			t.Errorf("Expected required flag %q not found", flagName)
		}
		if !flag.(*cli.StringFlag).Required {
			t.Errorf("Expected flag %q to be required", flagName)
		}
	}

	// Check optional flags
	optionalFlags := []string{"file", "confirm", "access", "roomsidscsv"}
	for _, flagName := range optionalFlags {
		flag := getFlagByName(cmd.Flags, flagName)
		if flag == nil {
			t.Errorf("Expected optional flag %q not found", flagName)
		}
	}

	// Check default values
	accessFlag := getFlagByName(cmd.Flags, "access")
	if accessFlag == nil || accessFlag.(*cli.StringFlag).Value != "om" {
		t.Errorf("Expected access flag default to be 'om'")
	}
}

// --- Test ExportPeopleCMD structure ---

func TestExportPeopleCMDStructure(t *testing.T) {
	app := &Application{}
	cmd := app.ExportPeopleCMD()

	if cmd.Name != "exportmembers" {
		t.Errorf("Expected command name 'exportmembers', got %q", cmd.Name)
	}

	expectedAliases := []string{"em"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	}

	// Check required flags
	requiredFlags := []string{"memberscsv"}
	for _, flagName := range requiredFlags {
		flag := getFlagByName(cmd.Flags, flagName)
		if flag == nil {
			t.Errorf("Expected required flag %q not found", flagName)
		}
		if !flag.(*cli.StringFlag).Required {
			t.Errorf("Expected flag %q to be required", flagName)
		}
	}
}

// --- Test FindRoomCMD structure ---

func TestFindRoomCMDStructure(t *testing.T) {
	app := &Application{}
	cmd := app.FindRoomCMD()

	if cmd.Name != "findroom" {
		t.Errorf("Expected command name 'findroom', got %q", cmd.Name)
	}

	expectedAliases := []string{"fr"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	}

	// Check required flags
	requiredFlags := []string{"title"}
	for _, flagName := range requiredFlags {
		flag := getFlagByName(cmd.Flags, flagName)
		if flag == nil {
			t.Errorf("Expected required flag %q not found", flagName)
		}
		if !flag.(*cli.StringFlag).Required {
			t.Errorf("Expected flag %q to be required", flagName)
		}
	}

	// Check optional flags
	optionalFlags := []string{"roomType"}
	for _, flagName := range optionalFlags {
		flag := getFlagByName(cmd.Flags, flagName)
		if flag == nil {
			t.Errorf("Expected optional flag %q not found", flagName)
		}
	}
}

// --- Test ListRoomsCMD structure ---

func TestListRoomsCMDStructure(t *testing.T) {
	app := &Application{}
	cmd := app.ListRoomsCMD()

	if cmd.Name != "listrooms" {
		t.Errorf("Expected command name 'listrooms', got %q", cmd.Name)
	}

	expectedAliases := []string{"lr"}
	if len(cmd.Aliases) != len(expectedAliases) {
		t.Errorf("Expected %d aliases, got %d", len(expectedAliases), len(cmd.Aliases))
	}

	// Check optional flags
	optionalFlags := []string{"roomType"}
	for _, flagName := range optionalFlags {
		flag := getFlagByName(cmd.Flags, flagName)
		if flag == nil {
			t.Errorf("Expected optional flag %q not found", flagName)
		}
	}

	// Check default values
	roomTypeFlag := getFlagByName(cmd.Flags, "roomType")
	if roomTypeFlag == nil || roomTypeFlag.(*cli.StringFlag).Value != "" {
		t.Errorf("Expected roomType flag default to be empty string")
	}
}

// --- Test command consistency ---

func TestCommandConsistency(t *testing.T) {
	app := &Application{}

	// Test that all commands have proper structure
	commands := []cli.Command{
		*app.RoomCMD(),
		*app.ChatCMD(),
		*app.WebexUtils(),
		*app.SendMessageToRoomCMD(),
		*app.AddPeopleCMD(),
		*app.RemovePeopleCMD(),
		*app.BroadcastToRoomsCMD(),
		*app.ExportPeopleCMD(),
		*app.FindRoomCMD(),
		*app.ListRoomsCMD(),
	}

	for i, cmd := range commands {
		if cmd.Name == "" {
			t.Errorf("Command %d has empty name", i)
		}
		if len(cmd.Aliases) == 0 && cmd.Name != "utils" { // utils might not need aliases
			t.Errorf("Command %q has no aliases", cmd.Name)
		}
		if cmd.Action == nil {
			t.Errorf("Command %q has no action", cmd.Name)
		}
	}
}
