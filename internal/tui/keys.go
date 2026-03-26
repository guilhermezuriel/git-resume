package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap holds the global key bindings for the TUI.
type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Back   key.Binding
	Quit   key.Binding
	PageUp key.Binding
	PageDn key.Binding
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "u"),
			key.WithHelp("pgup/u", "page up"),
		),
		PageDn: key.NewBinding(
			key.WithKeys("pgdown", "d"),
			key.WithHelp("pgdn/d", "page down"),
		),
	}
}

// Keys is the package-level key map used by all screens.
var Keys = DefaultKeyMap()
