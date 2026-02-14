package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	primaryColor   = lipgloss.AdaptiveColor{Light: "#5B5FC7", Dark: "#7B7FE0"}
	secondaryColor = lipgloss.AdaptiveColor{Light: "#00A3BF", Dark: "#00C4DB"}
	accentColor    = lipgloss.AdaptiveColor{Light: "#FF6B35", Dark: "#FF8C5A"}
	successColor   = lipgloss.AdaptiveColor{Light: "#2E7D32", Dark: "#66BB6A"}
	errorColor     = lipgloss.AdaptiveColor{Light: "#C62828", Dark: "#EF5350"}
	mutedColor     = lipgloss.AdaptiveColor{Light: "#9E9E9E", Dark: "#757575"}
	bgColor        = lipgloss.AdaptiveColor{Light: "#FAFAFA", Dark: "#1E1E2E"}
	selfBubbleBg   = lipgloss.AdaptiveColor{Light: "#E3F2FD", Dark: "#1A3A5C"}
	otherBubbleBg  = lipgloss.AdaptiveColor{Light: "#F5F5F5", Dark: "#2D2D3F"}

	// Header
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(primaryColor).
			Padding(0, 1)

	// Status indicators
	connectedStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	disconnectedStyle = lipgloss.NewStyle().
				Foreground(errorColor).
				Bold(true)

	// Chat bubbles
	selfBubbleStyle = lipgloss.NewStyle().
			Background(selfBubbleBg).
			Padding(0, 1).
			MarginLeft(4).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor)

	otherBubbleStyle = lipgloss.NewStyle().
				Background(otherBubbleBg).
				Padding(0, 1).
				MarginRight(4).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(secondaryColor)

	// Sender name
	selfNameStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	otherNameStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	// Timestamp
	timestampStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	// File attachment badge
	fileBadgeStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	// Input area
	inputBorderStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(0, 1)

	// Help bar
	helpBarStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Background(bgColor).
			Padding(0, 1)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Spinner
	spinnerStyle = lipgloss.NewStyle().
			Foreground(primaryColor)

	// Error
	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// Room list item
	roomItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#333333", Dark: "#CCCCCC"})

	roomItemSelectedStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	// Sidebar
	sidebarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderRight(true).
			BorderForeground(mutedColor).
			Padding(0, 1)

	sidebarSearchStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(secondaryColor).
				Padding(0, 0).
				MarginBottom(1)

	sidebarFilterStyle = lipgloss.NewStyle().
				Foreground(mutedColor)

	sidebarFilterActiveStyle = lipgloss.NewStyle().
					Foreground(primaryColor).
					Bold(true).
					Underline(true)

	sidebarItemActiveStyle = lipgloss.NewStyle().
				Foreground(successColor).
				Bold(true)

	// Chat header (room title in chat pane)
	chatHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(secondaryColor).
			Padding(0, 1)

	// Thread line — left border for threaded replies
	threadLineStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderLeft(true).
			BorderForeground(primaryColor).
			PaddingLeft(1).
			MarginLeft(2)

	// Reply bar — "Replying to…" indicator above input
	replyBarStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Italic(true).
			Padding(0, 1)

	// File picker overlay
	filePickerStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	filePickerSelectedStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	filePickerDirStyle = lipgloss.NewStyle().
				Foreground(secondaryColor).
				Bold(true)

	// Chat cursor highlight
	chatCursorStyle = lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "#E0E0E0", Dark: "#3A3A4F"})

	// Overlay
	overlayStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)
)
