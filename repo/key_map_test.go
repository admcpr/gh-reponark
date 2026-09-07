package repo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRepoKeyMap(t *testing.T) {
	k := NewRepoKeyMap()

	tests := []struct {
		name     string
		keys     []string
		wantKeys []string
		help     string
	}{
		{name: "Up", keys: k.Up.Keys(), wantKeys: []string{"up", "k"}, help: k.Up.Help().Desc},
		{name: "Down", keys: k.Down.Keys(), wantKeys: []string{"down", "j"}, help: k.Down.Help().Desc},
		{name: "Left", keys: k.Left.Keys(), wantKeys: []string{"left", "h"}, help: k.Left.Help().Desc},
		{name: "Right", keys: k.Right.Keys(), wantKeys: []string{"right", "l"}, help: k.Right.Help().Desc},
		{name: "Filter", keys: k.Filter.Keys(), wantKeys: []string{"/"}, help: k.Filter.Help().Desc},
		{name: "Esc", keys: k.Esc.Keys(), wantKeys: []string{"esc"}, help: k.Esc.Help().Desc},
		{name: "Quit", keys: k.Quit.Keys(), wantKeys: []string{"q"}, help: k.Quit.Help().Desc},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantKeys, tt.keys)
			assert.NotEmpty(t, tt.help)
		})
	}
}

func TestKeyMap_ShortHelp(t *testing.T) {
	k := NewRepoKeyMap()

	short := k.ShortHelp()

	assert.Len(t, short, 4)
	assert.Equal(t, k.Left.Keys(), short[0].Keys())
	assert.Equal(t, k.Right.Keys(), short[1].Keys())
	assert.Equal(t, k.Esc.Keys(), short[2].Keys())
	assert.Equal(t, k.Quit.Keys(), short[3].Keys())
}

func TestKeyMap_FullHelp(t *testing.T) {
	k := NewRepoKeyMap()

	full := k.FullHelp()

	assert.Len(t, full, 1)
	assert.Equal(t, k.ShortHelp(), full[0])
}
