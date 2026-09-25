package repo

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

// plain renders a view to a string with all ANSI styling removed so tests can
// assert on the visible text.
func plain(v tea.View) string {
	return ansi.Strip(fmt.Sprint(v.Content))
}

func newTestRepoConfig() RepoConfig {
	return NewRepoConfig(Repository{
		Name:           "test-repo",
		Description:    "Widgets for everyone",
		Url:            "https://github.com/owner/test-repo",
		IsArchived:     true,
		IsPrivate:      true,
		StargazerCount: 42,
		CreatedAt:      time.Date(2024, 3, 4, 0, 0, 0, 0, time.UTC),
	})
}

func newSelectedModel() Model {
	m := NewModel(80, 24)
	m.SelectRepo(newTestRepoConfig())
	return m
}

func TestNewModel(t *testing.T) {
	m := NewModel(80, 24)

	assert.Equal(t, 80, m.width)
	assert.Equal(t, 24, m.height)
	assert.Equal(t, 0, m.ActiveTab())
	assert.Equal(t, 0, m.ActiveProperty())
	assert.NotNil(t, m.repository.Properties)
}

func TestModel_SetDimensions(t *testing.T) {
	m := NewModel(80, 24)

	m.SetDimensions(120, 40)

	assert.Equal(t, 120, m.width)
	assert.Equal(t, 40, m.height)
	assert.Equal(t, 120, m.help.Width())
}

func TestModel_Init(t *testing.T) {
	m := NewModel(80, 24)
	assert.Nil(t, m.Init())
}

func TestModel_SelectRepo_KeepsTabAndProperty(t *testing.T) {
	m := newSelectedModel()
	m.SelectTab(2)
	m.SelectProperty(3)

	m.SelectRepo(NewRepoConfig(Repository{Name: "other"}))

	assert.Equal(t, "other", m.repository.Name)
	assert.Equal(t, 2, m.ActiveTab(), "the same group stays open")
	assert.Equal(t, 3, m.ActiveProperty(), "the same property stays focused, for comparing repos")
}

func TestModel_SelectTab(t *testing.T) {
	m := newSelectedModel()
	m.SelectProperty(2)

	m.SelectTab(3)

	assert.Equal(t, 3, m.ActiveTab())
	assert.Equal(t, Groups()[3], m.ActiveGroup())
	assert.Equal(t, 0, m.ActiveProperty(), "a new group starts at its first property")
}

func TestModel_SelectTab_Wraps(t *testing.T) {
	m := newSelectedModel()
	last := len(Groups()) - 1

	m.SelectTab(-1)
	assert.Equal(t, last, m.ActiveTab())

	m.SelectTab(last + 1)
	assert.Equal(t, 0, m.ActiveTab())
}

func TestModel_SelectProperty_Clamps(t *testing.T) {
	m := newSelectedModel()
	last := len(m.ActiveGroup().Properties) - 1

	m.SelectProperty(-3)
	assert.Equal(t, 0, m.ActiveProperty())

	m.SelectProperty(last + 5)
	assert.Equal(t, last, m.ActiveProperty())
}

func TestModel_SelectProperty_ScrollsGrid(t *testing.T) {
	m := newSelectedModel()
	m.SetDimensions(80, inspectorChrome+3) // three rows of properties
	last := len(m.ActiveGroup().Properties) - 1

	m.SelectProperty(last)

	assert.Equal(t, last-2, m.propertyOffset)
	content := plain(m.View())
	assert.Contains(t, content, m.ActiveGroup().Properties[last].Name)
	assert.NotContains(t, content, "Database ID", "properties scrolled off the top are hidden")
}

func TestModel_Update_Navigation(t *testing.T) {
	tab := tea.KeyPressMsg{Code: tea.KeyTab}
	shiftTab := tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	last := len(Groups()) - 1

	tests := []struct {
		name         string
		startTab     int
		startProp    int
		key          tea.KeyPressMsg
		wantTab      int
		wantProperty int
	}{
		{name: "tab moves forward", startTab: 0, key: tab, wantTab: 1},
		{name: "tab wraps to first", startTab: last, key: tab, wantTab: 0},
		{name: "shift+tab moves backward", startTab: 2, key: shiftTab, wantTab: 1},
		{name: "shift+tab wraps to last", startTab: 0, key: shiftTab, wantTab: last},
		{name: "tab resets the property", startTab: 0, startProp: 2, key: tab, wantTab: 1},
		{name: "other keys leave the property alone", startTab: 1, startProp: 2, key: tea.KeyPressMsg{Code: 'j', Text: "j"}, wantTab: 1, wantProperty: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newSelectedModel()
			m.SelectTab(tt.startTab)
			m.SelectProperty(tt.startProp)

			updated, cmd := m.Update(tt.key)
			got := updated.(Model)

			assert.Nil(t, cmd)
			assert.Equal(t, tt.wantTab, got.ActiveTab())
			assert.Equal(t, tt.wantProperty, got.ActiveProperty())
		})
	}
}

func TestModel_Update_IgnoresOtherKeys(t *testing.T) {
	m := newSelectedModel()

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})

	assert.Nil(t, cmd)
	assert.Equal(t, 0, updated.(Model).ActiveTab())
}

func TestModel_View(t *testing.T) {
	m := newSelectedModel()
	m.SelectProperty(2)
	m.SetFocused(true)

	content := plain(m.View())

	assert.Contains(t, content, "test-repo  private archived", "name and badges")
	assert.Contains(t, content, "Widgets for everyone", "repository description")
	assert.Contains(t, content, "Overview", "tab bar")
	assert.Contains(t, content, "▸ Name ", "focused property is marked")
	assert.Contains(t, content, "The name of the repository.", "focused property is described")
}

func TestModel_View_FocusChangesHighlight(t *testing.T) {
	m := newSelectedModel()
	unfocused := m.View()

	m.SetFocused(true)
	assert.True(t, m.Focused())
	focused := m.View()

	assert.Contains(t, plain(unfocused), "▸ Id", "the focused property is marked even without focus")
	assert.NotEqual(t, fmt.Sprint(unfocused.Content), fmt.Sprint(focused.Content), "focus changes the highlight")
	assert.Equal(t, plain(unfocused), plain(focused), "but not the text")
}

func TestModel_PageSize(t *testing.T) {
	m := NewModel(80, inspectorChrome+5)
	assert.Equal(t, 5, m.PageSize())
}

func TestModel_View_NoDescription(t *testing.T) {
	m := NewModel(80, 24)
	m.SelectRepo(NewRepoConfig(Repository{Name: "bare"}))

	assert.Contains(t, plain(m.View()), "No description")
}

func TestModel_View_FitsDimensions(t *testing.T) {
	m := newSelectedModel()
	m.SetDimensions(40, 15)

	lines := strings.Split(plain(m.View()), "\n")

	assert.Len(t, lines, 15)
	for _, line := range lines {
		assert.Equal(t, 40, lipgloss.Width(line), "%q", line)
	}
}

func TestModel_HelpView(t *testing.T) {
	m := NewModel(80, 24)

	content := plain(m.HelpView())

	assert.Contains(t, content, "next tab")
	assert.Contains(t, content, "prev tab")
	assert.NotContains(t, content, "quit", "only keys the pane actually handles are advertised")
}

func TestModel_Keys(t *testing.T) {
	m := NewModel(80, 24)

	keys := m.Keys()

	assert.Equal(t, []string{"tab"}, keys.NextTab.Keys())
	assert.Equal(t, []string{"shift+tab"}, keys.PrevTab.Keys())
}
