package accounts

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestMonogram(t *testing.T) {
	assert.Equal(t, "O ", ansi.Strip(monogram("octocat")))
	assert.Equal(t, "Ü ", ansi.Strip(monogram("über")))
	assert.Equal(t, "? ", ansi.Strip(monogram("")))
	assert.Equal(t, monogram("acme"), monogram("acme"), "the colour is stable")
}
