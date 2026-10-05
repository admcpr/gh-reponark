package inspector

import (
	"gh-reponark/repo"
	"gh-reponark/ui/uitest"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
)

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
		lines := strings.Split(uitest.Plain(m.View()), "\n")
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
	assert.NotContains(t, uitest.Plain(m.View()), "╭", "no footer card below 24 rows")
	assert.Contains(t, uitest.Plain(m.View()), "The Node ID of the Repository.", "the description is still shown")
}

func TestModel_View_Narrow(t *testing.T) {
	m := newSelectedModel()
	m.SetDimensions(minMediumWidth-1, 20)
	m.SetFocused(true)

	content := uitest.Plain(m.View())
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
		m.SetOrgRepos([]repo.RepoConfig{newTestRepoConfig(), repo.NewRepoConfig(repo.Repository{StargazerCount: 84})})
		content := uitest.Plain(m.View())
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
