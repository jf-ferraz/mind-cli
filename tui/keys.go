package tui

import "github.com/charmbracelet/bubbles/key"

// GlobalKeyMap defines keys that work in every tab.
type GlobalKeyMap struct {
	Quit      key.Binding
	ForceQuit key.Binding
	Help      key.Binding
	Refresh   key.Binding
	Tab1      key.Binding
	Tab2      key.Binding
	Tab3      key.Binding
	Tab4      key.Binding
	Tab5      key.Binding
	NextTab   key.Binding
	PrevTab   key.Binding
}

var globalKeys = GlobalKeyMap{
	Quit:      key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	ForceQuit: key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "force quit")),
	Help:      key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle help")),
	Refresh:   key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh data")),
	Tab1:      key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "Status tab")),
	Tab2:      key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "Docs tab")),
	Tab3:      key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "Iterations tab")),
	Tab4:      key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "Checks tab")),
	Tab5:      key.NewBinding(key.WithKeys("5"), key.WithHelp("5", "Quality tab")),
	NextTab:   key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next tab")),
	PrevTab:   key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev tab")),
}

// NavigationKeyMap defines up/down navigation shared by multiple tabs.
type NavigationKeyMap struct {
	Up   key.Binding
	Down key.Binding
}

var navKeys = NavigationKeyMap{
	Up:   key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down: key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
}
