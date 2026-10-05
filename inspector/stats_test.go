package inspector

import (
	"strings"
	"testing"
	"time"

	"gh-reponark/repo"
	"gh-reponark/ui"
	"gh-reponark/ui/uitest"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func statsRepos() []repo.RepoConfig {
	day := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	return []repo.RepoConfig{
		repo.NewRepoConfig(repo.Repository{Name: "a", StargazerCount: 0, HasWikiEnabled: true, PushedAt: day}),
		repo.NewRepoConfig(repo.Repository{Name: "b", StargazerCount: 50, HasWikiEnabled: true, PushedAt: day.AddDate(0, 0, 1)}),
		repo.NewRepoConfig(repo.Repository{Name: "c", StargazerCount: 100, PushedAt: day.AddDate(0, 0, 2)}),
		repo.NewRepoConfig(repo.Repository{Name: "d", StargazerCount: 1000, PushedAt: day.AddDate(0, 0, 3)}),
	}
}

func TestOrgStats(t *testing.T) {
	stats := OrgStats([]repo.RepoConfig{
		repo.NewRepoConfig(repo.Repository{StargazerCount: 10, ForkCount: 3, HasWikiEnabled: true}),
		repo.NewRepoConfig(repo.Repository{StargazerCount: 4, ForkCount: 7}),
	})

	assert.Equal(t, Stats{Total: 2, Max: 10, Values: []int{4, 10}}, stats["Stargazer Count"], "int values are sorted")
	assert.Equal(t, 7, stats["Fork Count"].Max)
	assert.Equal(t, 0, stats["Open Issues"].Max)
	assert.Equal(t, Stats{Total: 2, On: 1}, stats["Has Wiki Enabled"])
	assert.Empty(t, OrgStats(nil))
}

func TestStatsFor(t *testing.T) {
	repos := statsRepos()
	current := repos[2]
	current.Properties["Default Branch"] = repo.RepoProperty{Name: "Default Branch", Value: "main"}
	repos[0].Properties["Default Branch"] = repo.RepoProperty{Name: "Default Branch", Value: "main"}
	org := OrgStats(repos)

	stars := StatsFor(org, "Stargazer Count", repos, current)
	assert.Equal(t, 4, stars.Total)
	assert.Equal(t, 1000, stars.Max)
	assert.Equal(t, []int{0, 50, 100, 1000}, stars.Values)
	assert.Equal(t, "more than 2 of 4 repos · max 1k", stars.Context(current.Properties["Stargazer Count"]))

	wiki := StatsFor(org, "Has Wiki Enabled", repos, current)
	assert.Equal(t, 2, wiki.On)
	assert.Equal(t, "on in 2 of 4 repos", wiki.Context(current.Properties["Has Wiki Enabled"]))

	branch := StatsFor(org, "Default Branch", repos, current)
	assert.Equal(t, 2, branch.Same)
	assert.Equal(t, "shared by 2 of 4 repos", branch.Context(current.Properties["Default Branch"]))

	pushed := StatsFor(org, "Pushed At", repos, current)
	assert.Equal(t, 2, pushed.Older)
	assert.Equal(t, "newer than 2 of 4 repos", pushed.Context(current.Properties["Pushed At"]))

	id := StatsFor(org, "Database ID", repos, current)
	assert.Equal(t, "", id.Context(current.Properties["Database ID"]), "an identifier has no rank")

	size := StatsFor(org, "Disk Usage", repos, current)
	assert.Equal(t, "more than 0 of 4 repos · max 0 KB", size.Context(current.Properties["Disk Usage"]), "the largest keeps its unit")

	assert.Equal(t, "", StatsFor(nil, "Stargazer Count", nil, current).Context(current.Properties["Stargazer Count"]), "nothing before the org loads")
}

func TestBar(t *testing.T) {
	assert.Equal(t, "━━━━━━━━━━", ansi.Strip(bar(100, 100, 10, ui.ValueStyle)))
	assert.Equal(t, "━━━╌╌╌╌╌╌╌", ansi.Strip(bar(10, 100, 10, ui.ValueStyle)), "square-root scale")
	assert.Equal(t, "━╌╌╌╌╌╌╌╌╌", ansi.Strip(bar(1, 100, 10, ui.ValueStyle)), "anything non-zero shows")
	assert.Equal(t, "╌╌╌╌╌╌╌╌╌╌", ansi.Strip(bar(0, 100, 10, ui.ValueStyle)))
	assert.Equal(t, "╌╌╌╌╌╌╌╌╌╌", ansi.Strip(bar(5, 0, 10, ui.ValueStyle)), "no max, no fill")
	assert.Equal(t, "", bar(5, 100, 0, ui.ValueStyle))
}

func TestSqrtShare(t *testing.T) {
	assert.Equal(t, 1.0, sqrtShare(100, 100))
	assert.Equal(t, 0.5, sqrtShare(25, 100))
	assert.Equal(t, 0.0, sqrtShare(0, 100))
	assert.Equal(t, 0.0, sqrtShare(5, 0))
}

func TestStats_Histogram(t *testing.T) {
	repos := statsRepos()
	s := StatsFor(OrgStats(repos), "Stargazer Count", repos, repos[1])

	histogram := s.Histogram(50)
	assert.Equal(t, histogramBuckets, len([]rune(ansi.Strip(histogram))))
	assert.True(t, strings.HasPrefix(ansi.Strip(histogram), "█"), "three of four repos fall in the first bucket: %q", ansi.Strip(histogram))
	assert.Equal(t, "", Stats{}.Histogram(1), "nothing to draw without values")
}

func TestSparkline(t *testing.T) {
	assert.Equal(t, "█▁▄▁", ansi.Strip(Sparkline([]int{8, 0, 4, 1}, -1, sparkStyle)), "empty buckets keep a baseline")
	assert.Equal(t, "▁▁", ansi.Strip(Sparkline([]int{0, 0}, -1, sparkStyle)))

	plain := Sparkline([]int{8, 0, 4, 1}, -1, sparkStyle)
	highlighted := Sparkline([]int{8, 0, 4, 1}, 2, sparkStyle)
	assert.NotEqual(t, plain, highlighted, "the highlighted bucket has its own colour")
	assert.Equal(t, ansi.Strip(plain), ansi.Strip(highlighted))
	assert.NotEqual(t, plain, Sparkline([]int{8, 0, 4, 1}, -1, ui.AccentStyle), "the bar colour is the caller's")

	assert.Equal(t, 1, strings.Count(Sparkline([]int{3, 3, 3}, -1, sparkStyle), "\x1b[m"), "a run of one style is rendered once")
}

func TestFractionBar(t *testing.T) {
	assert.Equal(t, "━━━━━╌╌╌╌╌", ansi.Strip(FractionBar(2, 4, 10)))
	assert.Equal(t, "━╌╌╌╌╌╌╌╌╌", ansi.Strip(FractionBar(1, 100, 10)), "one is never nothing")
	assert.Equal(t, "╌╌╌╌╌╌╌╌╌╌", ansi.Strip(FractionBar(0, 4, 10)))
	assert.Equal(t, "", FractionBar(1, 0, 10))
}

func TestTileCount(t *testing.T) {
	assert.Equal(t, 6, tileCount(70), "SIZE joins from seventy columns")
	assert.Equal(t, 5, tileCount(69))
	assert.Equal(t, 5, tileCount(57))
	assert.Equal(t, 2, tileCount(22))
	assert.Equal(t, 0, tileCount(21), "nothing when fewer than two tiles fit")
}

func TestRenderStats(t *testing.T) {
	c := repo.NewRepoConfig(repo.Repository{StargazerCount: 25, ForkCount: 100})
	stats := map[string]Stats{"Stargazer Count": {Max: 100}, "Fork Count": {Max: 100}}

	lines := RenderStats(c, stats, 70, tileCount(70)) // six tiles of eleven
	assert.Len(t, lines, 3)
	for _, line := range lines {
		assert.Equal(t, 70, lipgloss.Width(line))
	}
	plainLines := uitest.StripAll(lines)
	assert.Contains(t, plainLines[0], "★ STARS")
	assert.Contains(t, plainLines[0], "▣ SIZE")
	assert.Contains(t, plainLines[1], "25")
	bars := []rune(plainLines[2])
	assert.Equal(t, 5, strings.Count(string(bars[:10]), "█"), "a quarter of the max fills half the bar on the square-root scale")
	assert.Equal(t, 10, strings.Count(string(bars[11:21]), "█"), "the max fills the bar")

	narrow := uitest.StripAll(RenderStats(c, stats, 57, tileCount(57)))
	assert.Contains(t, narrow[0], "⇄ PRS")
	assert.NotContains(t, narrow[0], "SIZE", "SIZE needs seventy columns")

	assert.Nil(t, RenderStats(c, stats, 20, 0), "nothing without tiles")
}

func TestTile_Value(t *testing.T) {
	c := repo.NewRepoConfig(repo.Repository{StargazerCount: 25, DiskUsage: 2048})
	stats := map[string]Stats{"Stargazer Count": {Max: 100}}

	text, fraction := tiles[0].value(c, stats)
	assert.Equal(t, "25", text)
	assert.Equal(t, 0.5, fraction)

	text, fraction = tiles[5].value(c, nil)
	assert.Equal(t, "2.0 MB", text, "SIZE keeps its unit")
	assert.Equal(t, 1.0, fraction, "without an organization the value is its own max")
}
