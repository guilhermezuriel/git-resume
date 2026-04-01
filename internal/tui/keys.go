package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap holds the global key bindings for the TUI.
type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Enter    key.Binding
	Back     key.Binding
	Quit     key.Binding
	Tab      key.Binding
	PageUp   key.Binding
	PageDn   key.Binding
	Generate key.Binding
	Delete   key.Binding
	Refresh  key.Binding
	Search   key.Binding
	Copy     key.Binding
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
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left panel"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right panel"),
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
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch panel"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "u"),
			key.WithHelp("pgup/u", "page up"),
		),
		PageDn: key.NewBinding(
			key.WithKeys("pgdown", "d"),
			key.WithHelp("pgdn/d", "page down"),
		),
		Generate: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new summary"),
		),
		Delete: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "delete"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Copy: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "copy"),
		),
	}
}

// Keys is the package-level key map used by all screens.
var Keys = DefaultKeyMap()
