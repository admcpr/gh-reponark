package shared

import (
	"testing"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewColors(t *testing.T) {
	light := NewColors(false)
	dark := NewColors(true)

	assert.Equal(t, "PencilDark", light.name)
	assert.Equal(t, lipgloss.Color("#212121"), light.Background)
	assert.Equal(t, lipgloss.Color("#f1f1f1"), light.Foreground)

	assert.Equal(t, "3024 Night", dark.name)
	assert.Equal(t, lipgloss.Color("#090300"), dark.Background)
	assert.Equal(t, lipgloss.Color("#a5a2a2"), dark.Foreground)

	// The base palette is shared between both modes.
	assert.Equal(t, light.Blue, dark.Blue)
	assert.Equal(t, light.Cyan, dark.Cyan)
	assert.Equal(t, light.Red, dark.Red)
}

type styleTestKeyMap struct{}

func (k styleTestKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit"))}
}

func (k styleTestKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

func TestNewHelpModel(t *testing.T) {
	tests := []struct {
		name      string
		width     int
		wantWidth int
	}{
		{name: "positive width is applied", width: 80, wantWidth: 80},
		{name: "zero width leaves default", width: 0, wantWidth: 0},
		{name: "negative width leaves default", width: -5, wantWidth: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewHelpModel(tt.width)
			assert.Equal(t, tt.wantWidth, m.Width())
			assert.Contains(t, m.View(styleTestKeyMap{}), "quit")
		})
	}
}

func TestBuildDefaultDelegate(t *testing.T) {
	d := BuildDefaultDelegate()

	assert.Equal(t, AppColors.Cyan, d.Styles.SelectedTitle.GetForeground())
	assert.Equal(t, AppColors.Cyan, d.Styles.SelectedTitle.GetBorderLeftForeground())
	assert.Equal(t, d.Styles.SelectedTitle.GetForeground(), d.Styles.SelectedDesc.GetForeground())
}
