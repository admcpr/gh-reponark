package ui

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

func TestHelp(t *testing.T) {
	up := key.NewBinding(key.WithKeys("k"), key.WithHelp("k", "up"))
	down := key.NewBinding(key.WithKeys("j"), key.WithHelp("j", "down"))

	short := Help{Short: []key.Binding{up}}
	assert.Equal(t, []key.Binding{up}, short.ShortHelp())
	assert.Equal(t, [][]key.Binding{{up}}, short.FullHelp(), "the short line stands in for full help")

	full := Help{Short: []key.Binding{up}, Full: [][]key.Binding{{up, down}}}
	assert.Equal(t, [][]key.Binding{{up, down}}, full.FullHelp())
}

func TestCombine(t *testing.T) {
	up := key.NewBinding(key.WithKeys("up", "k"))
	down := key.NewBinding(key.WithKeys("down", "j"))

	both := Combine("j/k", "repo", down, up)

	assert.Equal(t, []string{"down", "j", "up", "k"}, both.Keys())
	assert.Equal(t, "j/k", both.Help().Key)
	assert.Equal(t, "repo", both.Help().Desc)
}

func TestWithHelp(t *testing.T) {
	original := key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "switch view"))

	renamed := WithHelp(original, "matrix")

	assert.Equal(t, "matrix", renamed.Help().Desc)
	assert.Equal(t, "switch view", original.Help().Desc, "the original is unchanged")
}

func TestHint(t *testing.T) {
	hint := Hint("type", "to search")

	assert.True(t, hint.Enabled(), "hints show in the footer")
	assert.False(t, key.Matches(tea.KeyPressMsg{Code: 't', Text: "t"}, hint), "but match no key")
}

func TestHelp_String(t *testing.T) {
	h := Help{Short: []key.Binding{
		key.NewBinding(key.WithKeys("j"), key.WithHelp("j/k", "repo")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	}}
	assert.Equal(t, "j/k repo  esc back", h.String())
}
