package repo

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"github.com/stretchr/testify/assert"
)

func TestNewRepoKeyMap(t *testing.T) {
	k := NewRepoKeyMap()

	assert.Equal(t, []string{"tab"}, k.NextTab.Keys())
	assert.Equal(t, "next tab", k.NextTab.Help().Desc)
	assert.Equal(t, []string{"shift+tab"}, k.PrevTab.Keys())
	assert.Equal(t, "prev tab", k.PrevTab.Help().Desc)
}

func TestKeyMap_MatchesKeys(t *testing.T) {
	k := NewRepoKeyMap()

	assert.True(t, key.Matches(tea.KeyPressMsg{Code: tea.KeyTab}, k.NextTab))
	assert.True(t, key.Matches(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}, k.PrevTab))
	assert.False(t, key.Matches(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}, k.NextTab))
	assert.False(t, key.Matches(tea.KeyPressMsg{Code: tea.KeyLeft}, k.NextTab, k.PrevTab))
}

func TestKeyMap_ShortHelp(t *testing.T) {
	k := NewRepoKeyMap()

	short := k.ShortHelp()

	assert.Len(t, short, 2)
	assert.Equal(t, k.NextTab.Keys(), short[0].Keys())
	assert.Equal(t, k.PrevTab.Keys(), short[1].Keys())
}

func TestKeyMap_FullHelp(t *testing.T) {
	k := NewRepoKeyMap()

	full := k.FullHelp()

	assert.Len(t, full, 1)
	assert.Equal(t, k.ShortHelp(), full[0])
}
