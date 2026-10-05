package inspector

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestTypeOf_Glyph(t *testing.T) {
	assert.Equal(t, "●", ansi.Strip(TypeOf("bool").Glyph()))
	assert.Equal(t, "#", ansi.Strip(TypeOf("int").Glyph()))
	assert.Equal(t, "◆", ansi.Strip(TypeOf("time.Time").Glyph()))
	assert.Equal(t, "¶", ansi.Strip(TypeOf("string").Glyph()))
}

func TestTypePill(t *testing.T) {
	assert.Equal(t, "▐toggle▌", ansi.Strip(TypePill("bool")))
	assert.Equal(t, "▐count▌", ansi.Strip(TypePill("int")))
	assert.Equal(t, "▐date▌", ansi.Strip(TypePill("time.Time")))
	assert.Equal(t, "▐text▌", ansi.Strip(TypePill("string")))
}
