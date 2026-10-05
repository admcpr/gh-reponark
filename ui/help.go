package ui

import (
	"strings"

	"charm.land/bubbles/v2/key"
)

// Help is what a screen tells the footer about its keys: a short line that
// always shows, and columns of every binding for the full view behind "?".
// It satisfies help.KeyMap.
type Help struct {
	Short []key.Binding
	Full  [][]key.Binding
}

func (h Help) ShortHelp() []key.Binding { return h.Short }

// FullHelp falls back to the short line for screens with nothing more to say.
func (h Help) FullHelp() [][]key.Binding {
	if len(h.Full) == 0 {
		return [][]key.Binding{h.Short}
	}
	return h.Full
}

// String is the short line as plain text, e.g. "j/k repo  esc back".
func (h Help) String() string {
	entries := make([]string, len(h.Short))
	for i, b := range h.Short {
		entries[i] = b.Help().Key + " " + b.Help().Desc
	}
	return strings.Join(entries, "  ")
}

// HelpProvider screens describe their keys for the footer.
type HelpProvider interface {
	Help() Help
}

// Typing screens report when a text field has focus, so keys that would
// otherwise be shortcuts, such as "?", are typed instead.
type Typing interface {
	Typing() bool
}

// Combine merges bindings that do the same thing in different directions into
// one footer entry, e.g. "j/k repo" for down and up.
func Combine(keys, desc string, bindings ...key.Binding) key.Binding {
	var all []string
	for _, b := range bindings {
		all = append(all, b.Keys()...)
	}
	return key.NewBinding(key.WithKeys(all...), key.WithHelp(keys, desc))
}

// WithHelp returns a copy of binding with a different description.
func WithHelp(binding key.Binding, desc string) key.Binding {
	binding.SetHelp(binding.Help().Key, desc)
	return binding
}

// Hint is a footer entry for something that is not a key press, such as
// "type to search". It matches no key.
func Hint(label, desc string) key.Binding {
	return key.NewBinding(key.WithKeys("\x00hint"), key.WithHelp(label, desc))
}
