package inspector

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"gh-reponark/repo"
	"gh-reponark/ui"
	"gh-reponark/ui/uitest"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRepoConfig() repo.RepoConfig {
	return repo.NewRepoConfig(repo.Repository{
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

// sizeForRows sets the pane to width and the shortest height that gives it
// exactly rows property rows, so tests can size the grid without knowing
// how much header the pane keeps at that height.
func sizeForRows(t *testing.T, m *Model, width, rows int) {
	t.Helper()
	for height := 1; height <= 200; height++ {
		m.SetDimensions(width, height)
		if m.PageSize() == rows {
			return
		}
	}
	t.Fatalf("no height gives %d rows at width %d", rows, width)
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
}

func TestModel_Init(t *testing.T) {
	m := NewModel(80, 24)
	assert.Nil(t, m.Init())
}

func TestModel_SelectRepo_KeepsTabAndProperty(t *testing.T) {
	m := newSelectedModel()
	m.SelectTab(2)
	m.SelectProperty(3)

	m.SelectRepo(repo.NewRepoConfig(repo.Repository{Name: "other"}))

	assert.Equal(t, "other", m.repository.Name)
	assert.Equal(t, 2, m.ActiveTab(), "the same group stays open")
	assert.Equal(t, 3, m.ActiveProperty(), "the same property stays focused, for comparing repos")
}

func TestModel_SelectTab(t *testing.T) {
	m := newSelectedModel()
	m.SelectProperty(2)

	m.SelectTab(3)

	assert.Equal(t, 3, m.ActiveTab())
	assert.Equal(t, repo.Groups()[3], m.ActiveGroup())
	assert.Equal(t, 0, m.ActiveProperty(), "a new group starts at its first property")
}

func TestModel_SelectTab_Wraps(t *testing.T) {
	m := newSelectedModel()
	last := len(repo.Groups()) - 1

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
	sizeForRows(t, &m, 80, 3)
	last := len(m.ActiveGroup().Properties) - 1

	m.SelectProperty(last)

	assert.Equal(t, last-2, m.property.Offset)
	content := uitest.Plain(m.View())
	assert.Contains(t, content, m.ActiveGroup().Properties[last].Name)
	assert.NotContains(t, content, "Database ID", "properties scrolled off the top are hidden")
}

func TestModel_Update_Navigation(t *testing.T) {
	tab := tea.KeyPressMsg{Code: tea.KeyTab}
	shiftTab := tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	last := len(repo.Groups()) - 1

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
	m.SetDimensions(80, 36)
	m.SelectProperty(2)
	m.SetFocused(true)

	content := uitest.Plain(m.View())

	assert.Contains(t, content, "test-repo  ▐private▌ ▐archived▌", "name and badge pills")
	assert.Contains(t, content, "Widgets for everyone", "repository description")
	assert.Contains(t, content, "42 ★ · 0 forks", "key facts")
	assert.Contains(t, content, "▐Archived▌", "posture chips")
	assert.Contains(t, content, "★ STARS", "stat tiles")
	assert.Contains(t, content, "▐Overview▌", "segmented group control")
	assert.Contains(t, content, "▸ Name ", "focused property is marked")
	assert.Contains(t, content, "Name ▐text▌", "the footer card names the property and its type")
	assert.Contains(t, content, "3/16", "and its place in the group")
	assert.Contains(t, content, "The name of the repository.", "focused property is described")
	assert.NotContains(t, content, "HEALTH", "no health score")
}

func TestModel_View_FocusChangesHighlight(t *testing.T) {
	m := newSelectedModel()
	unfocused := m.View()

	m.SetFocused(true)
	assert.True(t, m.Focused())
	focused := m.View()

	assert.Contains(t, uitest.Plain(unfocused), "▸ Id", "the focused property is marked even without focus")
	assert.NotEqual(t, fmt.Sprint(unfocused.Content), fmt.Sprint(focused.Content), "focus changes the highlight")
	assert.Equal(t, uitest.Plain(unfocused), uitest.Plain(focused), "but not the text")
}

func TestModel_View_LongDescription(t *testing.T) {
	m := NewModel(49, 36)
	m.SelectRepo(repo.NewRepoConfig(repo.Repository{Name: "long", Description: "Command line for the Cobalt platform that is rather long so it wraps"}))

	line := strings.Split(uitest.Plain(m.View()), "\n")[1]
	assert.Equal(t, "Command line for the Cobalt platform that is rat…", line, "one line, cut with an ellipsis and no padding before it")

	m.SetDimensions(57, 44)
	lines := strings.Split(uitest.Plain(m.View()), "\n")
	assert.Equal(t, "Command line for the Cobalt platform that is rather long ", lines[1], "two lines at forty rows or more")
	assert.Equal(t, "so it wraps                                              ", lines[2])
}

func TestModel_View_NoDescription(t *testing.T) {
	m := NewModel(80, 36)
	m.SelectRepo(repo.NewRepoConfig(repo.Repository{Name: "bare"}))

	assert.Contains(t, uitest.Plain(m.View()), "No description")
}

func TestModel_View_FitsDimensions(t *testing.T) {
	m := newSelectedModel()
	m.SetDimensions(40, 15)

	lines := strings.Split(uitest.Plain(m.View()), "\n")

	assert.Len(t, lines, 15)
	for _, line := range lines {
		assert.Equal(t, 40, lipgloss.Width(line), "%q", line)
	}
}

func TestModel_View_TinyPanes(t *testing.T) {
	m := newSelectedModel()
	for _, size := range [][2]int{{1, 1}, {8, 3}, {20, 5}, {11, 14}} {
		m.SetDimensions(size[0], size[1])
		lines := strings.Split(uitest.Plain(m.View()), "\n")
		assert.Len(t, lines, size[1], "%v", size)
		for _, line := range lines {
			assert.Equal(t, size[0], lipgloss.Width(line), "%v: %q", size, line)
		}
	}
}

func TestModel_View_NeverPanics(t *testing.T) {
	for _, size := range [][2]int{{1, 1}, {60, 1}, {1, 40}, {40, 3}, {0, 0}, {50, 24}, {64, 40}} {
		for _, focused := range []bool{false, true} {
			m := newSelectedModel()
			m.SetOrgRepos([]repo.RepoConfig{newTestRepoConfig()})
			m.SetFocused(focused)
			m.SetDimensions(size[0], size[1])
			for tab := range repo.Groups() {
				m.SelectTab(tab)
				m.SelectProperty(1)
				require.NotPanics(t, func() { m.View() }, "%v focused=%v tab=%d", size, focused, tab)
				lines := strings.Split(uitest.Plain(m.View()), "\n")
				assert.Len(t, lines, ui.Max(1, size[1]), "%v", size)
				for _, line := range lines {
					assert.Equal(t, ui.Max(1, size[0]), lipgloss.Width(line), "%v: %q", size, line)
				}
			}
		}
	}
}

func TestModel_View_PillsAreShedNotCut(t *testing.T) {
	m := NewModel(30, 36)
	r := repo.Repository{Name: "legacy-billing", IsPrivate: true, IsArchived: true, IsFork: true, IsTemplate: true}
	r.PrimaryLanguage.Name = "PHP"
	m.SelectRepo(repo.NewRepoConfig(r))

	title := strings.Split(uitest.Plain(m.View()), "\n")[0]

	assert.Equal(t, "legacy-billing  ▐private▌     ", title, "pills that do not fit are dropped from the end")
	assert.NotContains(t, title, "…")

	m.SetDimensions(10, 36)
	title = strings.Split(uitest.Plain(m.View()), "\n")[0]
	assert.Equal(t, "legacy-bi…", title, "only the name itself is ever truncated")
}

func TestModel_Keys(t *testing.T) {
	m := NewModel(80, 24)

	keys := m.Keys()

	assert.Equal(t, []string{"tab"}, keys.NextTab.Keys())
	assert.Equal(t, []string{"shift+tab"}, keys.PrevTab.Keys())
}
