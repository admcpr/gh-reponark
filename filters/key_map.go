package filters

import "charm.land/bubbles/v2/key"

// filterKeyMap holds the bindings for the filter screen. The search input is
// wired to the same bindings, so the help footer always matches behaviour.
type filterKeyMap struct {
	Select         key.Binding
	Complete       key.Binding
	NextSuggestion key.Binding
	PrevSuggestion key.Binding
	Back           key.Binding
}

func newFilterKeyMap() filterKeyMap {
	return filterKeyMap{
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "add filter"),
		),
		Complete: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "complete"),
		),
		NextSuggestion: key.NewBinding(
			key.WithKeys("down", "ctrl+n"),
			key.WithHelp("↓", "next"),
		),
		PrevSuggestion: key.NewBinding(
			key.WithKeys("up", "ctrl+p"),
			key.WithHelp("↑", "prev"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "ctrl+enter"),
			key.WithHelp("esc", "back"),
		),
	}
}

func (k filterKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Select, k.Complete, k.NextSuggestion, k.PrevSuggestion, k.Back}
}

func (k filterKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

// editorKeyMap holds the bindings shared by the filter editors. Each editor
// handles, and advertises, the subset that applies to it.
type editorKeyMap struct {
	Apply     key.Binding
	Back      key.Binding
	NextField key.Binding
	Toggle    key.Binding
	Yes       key.Binding
	No        key.Binding
}

func newEditorKeyMap() editorKeyMap {
	return editorKeyMap{
		Apply: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "apply"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		NextField: key.NewBinding(
			key.WithKeys("tab", "shift+tab"),
			key.WithHelp("tab", "next field"),
		),
		Toggle: key.NewBinding(
			key.WithKeys("left", "right"),
			key.WithHelp("←/→", "toggle"),
		),
		Yes: key.NewBinding(
			key.WithKeys("y", "Y"),
			key.WithHelp("y", "yes"),
		),
		No: key.NewBinding(
			key.WithKeys("n", "N"),
			key.WithHelp("n", "no"),
		),
	}
}
