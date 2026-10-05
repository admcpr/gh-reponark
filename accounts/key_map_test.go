package accounts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserKeyMap(t *testing.T) {
	keymap := newUserKeyMap()

	assert.Equal(t, []string{"up", "k"}, keymap.Up.Keys())
	assert.Equal(t, []string{"down", "j"}, keymap.Down.Keys())
	assert.Equal(t, []string{"left", "h", "pgup", "b", "u"}, keymap.PageUp.Keys())
	assert.Equal(t, []string{"right", "l", "pgdown", "f", "d"}, keymap.PageDown.Keys())
	assert.Equal(t, []string{"home", "g"}, keymap.Top.Keys())
	assert.Equal(t, []string{"end", "G"}, keymap.Bottom.Keys())
	assert.Equal(t, []string{"enter"}, keymap.Select.Keys())
	assert.Equal(t, []string{"esc"}, keymap.Back.Keys())
}
