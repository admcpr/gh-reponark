package repos

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOrgKeyMap(t *testing.T) {
	keymap := newOrgKeyMap()

	assert.Equal(t, []string{"up", "k"}, keymap.Up.Keys())
	assert.Equal(t, []string{"down", "j"}, keymap.Down.Keys())
	assert.Equal(t, []string{"f", "F"}, keymap.Filters.Keys())
	assert.Equal(t, []string{"esc"}, keymap.Back.Keys())
	assert.Equal(t, []string{"v"}, keymap.ToggleView.Keys())
	assert.Equal(t, []string{"enter"}, keymap.Inspect.Keys())
	assert.Equal(t, []string{"left", "h"}, keymap.Left.Keys())
	assert.Equal(t, []string{"right", "l"}, keymap.Right.Keys())
}
