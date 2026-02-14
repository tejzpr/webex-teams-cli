package tui

import (
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Send       key.Binding
	Quit       key.Binding
	Search     key.Binding
	Help       key.Binding
	ClearChat  key.Binding
	ScrollUp   key.Binding
	ScrollDn   key.Binding
	PageUp     key.Binding
	PageDown   key.Binding
	Tab        key.Binding
	Escape     key.Binding
	Reply      key.Binding
	AttachFile key.Binding
	Download   key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Send: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select/send"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search rooms"),
		),
		Help: key.NewBinding(
			key.WithKeys("ctrl+h"),
			key.WithHelp("ctrl+h", "help"),
		),
		ClearChat: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "clear chat"),
		),
		ScrollUp: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "up"),
		),
		ScrollDn: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("pgdn", "page down"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch pane"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Reply: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "reply"),
		),
		AttachFile: key.NewBinding(
			key.WithKeys("ctrl+f"),
			key.WithHelp("ctrl+f", "attach file"),
		),
		Download: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "download"),
		),
	}
}

// ShortHelp returns key bindings for the mini help bar
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Tab, k.Search, k.Reply, k.AttachFile, k.Download, k.Help, k.Quit}
}

// FullHelp returns key bindings for the full help view
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Send, k.Search, k.Tab, k.Reply},
		{k.AttachFile, k.Download, k.ClearChat},
		{k.Help, k.Quit, k.Escape},
		{k.ScrollUp, k.ScrollDn, k.PageUp, k.PageDown},
	}
}
