package org

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOrgKeyMap(t *testing.T) {
	k := NewOrgKeyMap()

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
		{name: "Enter", keys: k.Enter.Keys(), wantKeys: []string{"enter"}, help: k.Enter.Help().Desc},
		{name: "Esc", keys: k.Esc.Keys(), wantKeys: []string{"esc"}, help: k.Esc.Help().Desc},
		{name: "Help", keys: k.Help.Keys(), wantKeys: []string{"?"}, help: k.Help.Help().Desc},
		{name: "Quit", keys: k.Quit.Keys(), wantKeys: []string{"q"}, help: k.Quit.Help().Desc},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantKeys, tt.keys)
			assert.NotEmpty(t, tt.help)
		})
	}
}

func TestOrgKeyMap_ShortHelp(t *testing.T) {
	k := NewOrgKeyMap()

	short := k.ShortHelp()

	assert.Len(t, short, 5)
	assert.Equal(t, k.Up.Keys(), short[0].Keys())
	assert.Equal(t, k.Down.Keys(), short[1].Keys())
	assert.Equal(t, k.Enter.Keys(), short[2].Keys())
	assert.Equal(t, k.Esc.Keys(), short[3].Keys())
	assert.Equal(t, k.Quit.Keys(), short[4].Keys())
}

func TestOrgKeyMap_FullHelp(t *testing.T) {
	k := NewOrgKeyMap()

	full := k.FullHelp()

	assert.Len(t, full, 2)
	assert.Len(t, full[0], 4)
	assert.Len(t, full[1], 4)
	assert.Equal(t, k.Up.Keys(), full[0][0].Keys())
	assert.Equal(t, k.Right.Keys(), full[0][3].Keys())
	assert.Equal(t, k.Enter.Keys(), full[1][0].Keys())
	assert.Equal(t, k.Quit.Keys(), full[1][3].Keys())
}
