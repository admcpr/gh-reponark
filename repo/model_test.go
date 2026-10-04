package repo

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"gh-reponark/shared"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	sizeForRows(t, &m, 80, 3)
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
	m.SetDimensions(80, 36)
	m.SelectProperty(2)
	m.SetFocused(true)

	content := plain(m.View())

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

	assert.Contains(t, plain(unfocused), "▸ Id", "the focused property is marked even without focus")
	assert.NotEqual(t, fmt.Sprint(unfocused.Content), fmt.Sprint(focused.Content), "focus changes the highlight")
	assert.Equal(t, plain(unfocused), plain(focused), "but not the text")
}

func TestModel_PageSize(t *testing.T) {
	// The full layout at seventy columns: name, description, facts and two
	// chip rows; six tiles; a gap; the segmented control; the footer card.
	full := layout{description: 1, facts: true, chipRows: 2, tiles: 6, bars: true, card: true}
	assert.Equal(t, 16, full.chrome())

	// At 36 rows nothing is shed: the full budget applies.
	m := NewModel(70, 36)
	assert.Equal(t, full.chrome(), m.chrome())
	assert.Equal(t, 36-full.chrome(), m.PageSize())

	m = NewModel(70, full.chrome()+minRows)
	assert.Equal(t, full.chrome(), m.chrome(), "nothing is shed while twelve rows remain")
	assert.Equal(t, minRows, m.PageSize())
}

func TestModel_Layout_KeepsTwelveRows(t *testing.T) {
	m := newSelectedModel()
	for _, size := range [][2]int{{57, 33}, {77, 41}, {57, 27}, {70, 24}, {70, 20}, {57, 19}} {
		m.SetDimensions(size[0], size[1])
		assert.GreaterOrEqual(t, m.PageSize(), minRows, "%v", size)
		lines := strings.Split(plain(m.View()), "\n")
		assert.Len(t, lines, size[1], "%v", size)
	}

	m.SetDimensions(57, 33) // the inspector at 120x36
	assert.Equal(t, 17, m.PageSize())
	l := m.layout()
	assert.Equal(t, layout{description: 1, facts: true, chipRows: 2, tiles: 5, bars: true, card: true}, l)

	m.SetDimensions(77, 41) // the inspector at 160x44
	assert.Equal(t, 25, m.PageSize())
	assert.Equal(t, layout{description: 2, facts: true, chipRows: 1, tiles: 6, bars: true, card: true}, m.layout())

	m.SetDimensions(57, 27)
	assert.Equal(t, layout{description: 0, facts: true, chipRows: 2, tiles: 5, bars: true, card: true}, m.layout(), "the description goes first")

	m.SetDimensions(70, 21)
	assert.Equal(t, layout{description: 0, facts: true, chipRows: 1, tiles: 0, bars: true, card: false}, m.layout(), "then the second chip row, then the tiles; a short pane has the plain footer")
	assert.NotContains(t, plain(m.View()), "╭", "no footer card below 24 rows")
	assert.Contains(t, plain(m.View()), "The Node ID of the Repository.", "the description is still shown")
}

func TestModel_View_LongDescription(t *testing.T) {
	m := NewModel(49, 36)
	m.SelectRepo(NewRepoConfig(Repository{Name: "long", Description: "Command line for the Cobalt platform that is rather long so it wraps"}))

	line := strings.Split(plain(m.View()), "\n")[1]
	assert.Equal(t, "Command line for the Cobalt platform that is rat…", line, "one line, cut with an ellipsis and no padding before it")

	m.SetDimensions(57, 44)
	lines := strings.Split(plain(m.View()), "\n")
	assert.Equal(t, "Command line for the Cobalt platform that is rather long ", lines[1], "two lines at forty rows or more")
	assert.Equal(t, "so it wraps                                              ", lines[2])
}

func TestModel_View_NoDescription(t *testing.T) {
	m := NewModel(80, 36)
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

func TestModel_View_Narrow(t *testing.T) {
	m := newSelectedModel()
	m.SetDimensions(minMediumWidth-1, 20)
	m.SetFocused(true)

	content := plain(m.View())
	lines := strings.Split(content, "\n")

	assert.Len(t, lines, 20)
	for _, line := range lines {
		assert.Equal(t, minMediumWidth-1, lipgloss.Width(line), "%q", line)
	}
	assert.Contains(t, content, "test-repo  ▐private▌ ▐archived▌")
	assert.Contains(t, content, "Widgets for everyone")
	assert.NotContains(t, content, "★ STARS", "no tiles")
	assert.NotContains(t, content, "▐Archived▌", "no posture chips")
	assert.NotContains(t, content, "╭", "no footer card")
	assert.Contains(t, content, "The Node ID of the Repository.", "the description is still shown")
	assert.Equal(t, 20-m.chrome(), m.PageSize())
	assert.Equal(t, 7, m.chrome(), "name, description, gap, groups, rule and two description lines")
	assert.Equal(t, layout{description: 1}, m.layout(), "no tiles, bars, chips or card")
}

func TestModel_View_NarrowPanesShedDecoration(t *testing.T) {
	m := newSelectedModel()
	m.SetFocused(true)

	for _, tt := range []struct {
		width                    int
		tiles, size, chips, card bool
		bars                     bool
	}{
		{width: 80, tiles: true, size: true, chips: true, card: true, bars: true},
		{width: 64, tiles: true, size: false, chips: true, card: true, bars: true},
		{width: 50, tiles: true, size: false, chips: true, card: true, bars: true},
		{width: 49, tiles: false, size: false, chips: false, card: false, bars: false},
		{width: 12, tiles: false, size: false, chips: false, card: false, bars: false},
		{width: 11, tiles: false, size: false, chips: false, card: false, bars: false},
	} {
		m.SetDimensions(tt.width, 36)
		m.SelectTab(2) // Metrics: Stargazer Count has a bar once the org is known
		m.SetOrgRepos([]RepoConfig{newTestRepoConfig(), NewRepoConfig(Repository{StargazerCount: 84})})
		content := plain(m.View())
		lines := strings.Split(content, "\n")

		assert.Len(t, lines, 36, "width %d", tt.width)
		for _, line := range lines {
			assert.Equal(t, tt.width, lipgloss.Width(line), "width %d: %q", tt.width, line)
		}
		assert.Equal(t, tt.tiles, strings.Contains(content, "★ STARS"), "width %d: tiles", tt.width)
		assert.Equal(t, tt.size, strings.Contains(content, "▣ SIZE"), "width %d: SIZE tile", tt.width)
		assert.Equal(t, tt.chips, strings.Contains(content, "▐Archived▌"), "width %d: chips", tt.width)
		assert.Equal(t, tt.card, strings.Contains(content, "╭"), "width %d: footer card", tt.width)
		assert.Equal(t, tt.bars, strings.Contains(content, "━━━━━━━╌╌╌"), "width %d: bars", tt.width)
		assert.Contains(t, content, "test-repo", "width %d", tt.width)
		if tt.width >= minListWidth {
			assert.Contains(t, content, "▸ Disk", "width %d: the focused property keeps its marker", tt.width)
		}
		assert.Equal(t, 36-m.chrome(), m.PageSize(), "width %d: rows left for properties", tt.width)
	}
}

func TestModel_View_TinyPanes(t *testing.T) {
	m := newSelectedModel()
	for _, size := range [][2]int{{1, 1}, {8, 3}, {20, 5}, {11, 14}} {
		m.SetDimensions(size[0], size[1])
		lines := strings.Split(plain(m.View()), "\n")
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
			m.SetOrgRepos([]RepoConfig{newTestRepoConfig()})
			m.SetFocused(focused)
			m.SetDimensions(size[0], size[1])
			for tab := range Groups() {
				m.SelectTab(tab)
				m.SelectProperty(1)
				require.NotPanics(t, func() { m.View() }, "%v focused=%v tab=%d", size, focused, tab)
				lines := strings.Split(plain(m.View()), "\n")
				assert.Len(t, lines, shared.Max(1, size[1]), "%v", size)
				for _, line := range lines {
					assert.Equal(t, shared.Max(1, size[0]), lipgloss.Width(line), "%v: %q", size, line)
				}
			}
		}
	}
}

func TestModel_View_PillsAreShedNotCut(t *testing.T) {
	m := NewModel(30, 36)
	r := Repository{Name: "legacy-billing", IsPrivate: true, IsArchived: true, IsFork: true, IsTemplate: true}
	r.PrimaryLanguage.Name = "PHP"
	m.SelectRepo(NewRepoConfig(r))

	title := strings.Split(plain(m.View()), "\n")[0]

	assert.Equal(t, "legacy-billing  ▐private▌     ", title, "pills that do not fit are dropped from the end")
	assert.NotContains(t, title, "…")

	m.SetDimensions(10, 36)
	title = strings.Split(plain(m.View()), "\n")[0]
	assert.Equal(t, "legacy-bi…", title, "only the name itself is ever truncated")
}

func TestModel_SetOrgRepos_AddsBarsAndContext(t *testing.T) {
	m := newSelectedModel()
	m.SetDimensions(80, 36)
	m.SelectTab(2) // Metrics
	m.SelectProperty(2)
	before := plain(m.View())

	m.SetOrgRepos([]RepoConfig{
		newTestRepoConfig(),
		NewRepoConfig(Repository{Name: "popular", StargazerCount: 84}),
	})
	after := plain(m.View())

	assert.NotContains(t, before, "━", "no bars before the org is known")
	assert.NotContains(t, before, "of 2 repos")
	assert.Contains(t, after, "      42  ━━━━━━━╌╌╌", "42 stars is half the org max: seven of ten cells on the square-root scale")
	assert.Contains(t, after, "more than 0 of 2 repos · max 84", "the footer puts the count in context")
	assert.NotContains(t, after, "p50", "no percentiles")

	m.SelectTab(1) // Status: Is Archived
	content := plain(m.View())
	assert.Contains(t, content, "on in 1 of 2 repos  ━━━━━╌╌╌╌╌", "booleans get the share of repos with the setting on")

	m.SelectTab(0)
	m.SelectProperty(1) // Database ID
	content = plain(m.View())
	assert.NotContains(t, content, "Database ID ▐count▌ ▴", "sanity: the title row is intact")
	assert.NotContains(t, content, "more than", "an identifier has no rank")
	row := strings.Split(content, "\n")
	var idRow string
	for _, line := range row {
		if strings.Contains(line, "▸ Database ID") {
			idRow = line
		}
	}
	assert.NotContains(t, idRow, "━", "and no bar")
}

func TestModel_Keys(t *testing.T) {
	m := NewModel(80, 24)

	keys := m.Keys()

	assert.Equal(t, []string{"tab"}, keys.NextTab.Keys())
	assert.Equal(t, []string{"shift+tab"}, keys.PrevTab.Keys())
}
