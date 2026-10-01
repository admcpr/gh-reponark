package filters

import "charm.land/bubbles/v2/key"

// filterKeyMap holds the bindings for the filter screen. Update matches
// against these and the help footer renders them.
type filterKeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Top       key.Binding
	Bottom    key.Binding
	Search    key.Binding
	Edit      key.Binding
	EndSearch key.Binding
	Toggle    key.Binding
	Clear     key.Binding
	ClearAll  key.Binding
	Done      key.Binding
	Cancel    key.Binding
	Back      key.Binding
}

func newFilterKeyMap() filterKeyMap {
	return filterKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Top: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "first"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "last"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Edit: key.NewBinding(
			key.WithKeys("enter", "l", "right"),
			key.WithHelp("l/enter", "edit"),
		),
		// Only keys that cannot be typed end a search, so every letter,
		// j included, goes into the query.
		EndSearch: key.NewBinding(
			key.WithKeys("enter", "down"),
			key.WithHelp("↓/enter", "done"),
		),
		Toggle: key.NewBinding(
			key.WithKeys("space"),
			key.WithHelp("space", "any/yes/no"),
		),
		Clear: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "clear"),
		),
		ClearAll: key.NewBinding(
			key.WithKeys("X"),
			key.WithHelp("X", "clear all"),
		),
		Done: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "done"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "done"),
		),
	}
}
