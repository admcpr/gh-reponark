// Package uitest holds helpers shared by the screen packages' tests.
package uitest

import (
	"fmt"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// Plain renders a view to a string with all ANSI styling removed so tests can
// assert on the visible text.
func Plain(v tea.View) string {
	return ansi.Strip(fmt.Sprint(v.Content))
}

// LoadMsg runs cmd and returns the message its loading command produced,
// skipping the spinner tick that Init batches alongside it.
func LoadMsg(cmd tea.Cmd) tea.Msg {
	switch msg := cmd().(type) {
	case tea.BatchMsg:
		for _, c := range msg {
			if m := LoadMsg(c); m != nil {
				return m
			}
		}
		return nil
	case spinner.TickMsg:
		return nil
	default:
		return msg
	}
}

// StripAll removes ANSI styling from every line.
func StripAll(lines []string) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = ansi.Strip(l)
	}
	return out
}
