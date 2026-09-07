package shared

import (
	"errors"
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestNewErrorModel(t *testing.T) {
	err := errors.New("boom")
	m := NewErrorModel(err, 80, 24)

	assert.Equal(t, err, m.Err())
	assert.Equal(t, 80, m.width)
	assert.Equal(t, 24, m.height)
	assert.Equal(t, 80, m.help.Width())
}

func TestErrorModel_SetDimensions(t *testing.T) {
	m := NewErrorModel(errors.New("boom"), 80, 24)

	m.SetDimensions(120, 40)

	assert.Equal(t, 120, m.width)
	assert.Equal(t, 40, m.height)
	assert.Equal(t, 120, m.help.Width())
}

func TestErrorModel_Init(t *testing.T) {
	m := NewErrorModel(errors.New("boom"), 80, 24)
	assert.Nil(t, m.Init())
}

func TestErrorModel_Update_GoesBack(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyPressMsg
	}{
		{name: "esc", key: tea.KeyPressMsg{Code: tea.KeyEscape}},
		{name: "enter", key: tea.KeyPressMsg{Code: tea.KeyEnter}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewErrorModel(errors.New("boom"), 80, 24)

			updated, cmd := m.Update(tt.key)

			assert.Same(t, m, updated)
			if assert.NotNil(t, cmd) {
				prev, ok := cmd().(PreviousMsg)
				assert.True(t, ok, "expected a PreviousMsg")
				assert.Nil(t, prev.Message)
			}
		})
	}
}

func TestErrorModel_Update_IgnoresOtherMessages(t *testing.T) {
	m := NewErrorModel(errors.New("boom"), 80, 24)

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	assert.Nil(t, cmd)

	_, cmd = m.Update(tea.WindowSizeMsg{Width: 1, Height: 1})
	assert.Nil(t, cmd)
}

func TestErrorModel_View(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "with error", err: errors.New("fetching user octocat: boom"), want: "fetching user octocat: boom"},
		{name: "nil error", err: nil, want: "Unknown error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewErrorModel(tt.err, 80, 24)

			content := ansi.Strip(fmt.Sprint(m.View().Content))

			assert.Contains(t, content, "Something went wrong")
			assert.Contains(t, content, tt.want)
		})
	}
}

func TestErrorModel_HeaderAndHelpViews(t *testing.T) {
	m := NewErrorModel(errors.New("boom"), 80, 24)

	assert.Contains(t, ansi.Strip(fmt.Sprint(m.HeaderView().Content)), "Error")
	assert.Contains(t, ansi.Strip(fmt.Sprint(m.HelpView().Content)), "back")
}

func TestErrorKeyMap(t *testing.T) {
	keymap := errorKeyMap{}

	short := keymap.ShortHelp()
	assert.Len(t, short, 1)

	full := keymap.FullHelp()
	assert.Len(t, full, 1)
	assert.Equal(t, short, full[0])
}
