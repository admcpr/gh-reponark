package shared

import (
	"testing"

	"charm.land/bubbles/v2/key"
	"github.com/stretchr/testify/assert"
)

func TestKeyBindings(t *testing.T) {
	back := key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back"))
	quit := key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit"))
	bindings := KeyBindings{back, quit}

	short := bindings.ShortHelp()
	assert.Len(t, short, 2)
	assert.Equal(t, back.Keys(), short[0].Keys())
	assert.Equal(t, quit.Keys(), short[1].Keys())

	full := bindings.FullHelp()
	assert.Len(t, full, 1)
	assert.Equal(t, short, full[0])
}

func TestKeyBindings_RenderInHelp(t *testing.T) {
	bindings := KeyBindings{key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back"))}

	assert.Contains(t, NewHelpModel(80).View(bindings), "back")
}

func TestKeyBindings_Empty(t *testing.T) {
	assert.Empty(t, KeyBindings{}.ShortHelp())
	assert.Len(t, KeyBindings{}.FullHelp(), 1)
}
