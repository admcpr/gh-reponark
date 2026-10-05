package inspector

import (
	"gh-reponark/repo"
	"gh-reponark/ui/uitest"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

// withNow fixes the clock for relative dates for the rest of the test.
func withNow(t *testing.T, at time.Time) {
	t.Helper()
	original := now
	now = func() time.Time { return at }
	t.Cleanup(func() { now = original })
}

func TestFormatValue(t *testing.T) {
	today := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	withNow(t, today)

	tests := []struct {
		name     string
		property repo.RepoProperty
		want     string
	}{
		{name: "true", property: repo.RepoProperty{Value: true}, want: "✓ yes"},
		{name: "false", property: repo.RepoProperty{Value: false}, want: "✗ no"},
		{name: "int", property: repo.RepoProperty{Value: 1204}, want: "1204"},
		{name: "disk usage", property: repo.RepoProperty{Name: "Disk Usage", Unit: "kb", Value: 2048}, want: "2.0 MB"},
		{name: "string", property: repo.RepoProperty{Value: "main"}, want: "main"},
		{name: "empty string", property: repo.RepoProperty{Value: ""}, want: "—"},
		{name: "zero time", property: repo.RepoProperty{Value: time.Time{}}, want: "—"},
		{name: "time", property: repo.RepoProperty{Value: today.AddDate(0, 0, -3)}, want: "3d ago  2026/09/22"},
		{name: "unsupported", property: repo.RepoProperty{Value: 1.5}, want: "—"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ansi.Strip(FormatValue(tt.property)))
		})
	}
}

func TestFormatValue_HighlightsWhatNeedsAttention(t *testing.T) {
	today := time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)
	withNow(t, today)

	plainCount := FormatValue(repo.RepoProperty{Name: "Open Issues", Value: 3})
	alerts := FormatValue(repo.RepoProperty{Name: "Vulnerability Alerts", Value: 3})
	assert.NotEqual(t, plainCount, alerts, "open alerts are coloured differently from other counts")

	recent := FormatCell(repo.RepoProperty{Value: today.AddDate(0, -1, 0)})
	stale := FormatCell(repo.RepoProperty{Value: today.AddDate(-1, 0, 0)})
	assert.NotEqual(t, recent[:12], stale[:12], "stale dates are coloured differently")
}

func TestFormatCell(t *testing.T) {
	withNow(t, time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name     string
		property repo.RepoProperty
		want     string
	}{
		{name: "true", property: repo.RepoProperty{Value: true}, want: "●"},
		{name: "false", property: repo.RepoProperty{Value: false}, want: "○"},
		{name: "int", property: repo.RepoProperty{Value: 1204}, want: "1.2k"},
		{name: "disk usage", property: repo.RepoProperty{Unit: "kb", Value: 1536}, want: "1.5 MB"},
		{name: "time", property: repo.RepoProperty{Value: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)}, want: "2mo ago"},
		{name: "string", property: repo.RepoProperty{Value: "main"}, want: "main"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ansi.Strip(FormatCell(tt.property)))
		})
	}
}

func TestFormatRow(t *testing.T) {
	today := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	withNow(t, today)

	tests := []struct {
		name     string
		property repo.RepoProperty
		max      int
		want     string
	}{
		{name: "true", property: repo.RepoProperty{Value: true}, want: "[ ● ] yes"},
		{name: "false", property: repo.RepoProperty{Value: false}, want: "[ ○ ] no"},
		{name: "int is right-aligned in eleven cells", property: repo.RepoProperty{Name: "Open Issues", Value: 100}, want: "        100"},
		{name: "int with bar", property: repo.RepoProperty{Name: "Open Issues", Value: 100}, max: 400, want: "        100  ━━━━━╌╌╌╌╌"},
		{name: "zero keeps an empty bar", property: repo.RepoProperty{Name: "Open Issues", Value: 0}, max: 400, want: "          0  ╌╌╌╌╌╌╌╌╌╌"},
		{name: "no max, no bar", property: repo.RepoProperty{Name: "Open Issues", Value: 100}, want: "        100"},
		{name: "identifiers get no bar", property: repo.RepoProperty{Name: "Database ID", Identifier: true, Value: 100}, max: 400, want: "        100"},
		{name: "disk usage keeps its unit", property: repo.RepoProperty{Name: "Disk Usage", Unit: "kb", Value: 1536}, want: "     1.5 MB"},
		{name: "date", property: repo.RepoProperty{Value: today.AddDate(0, 0, -3)}, want: "22 Sep 2026  3d ago"},
		{name: "zero date", property: repo.RepoProperty{Value: time.Time{}}, want: "—"},
		{name: "text", property: repo.RepoProperty{Value: "main"}, want: "main"},
		{name: "empty text", property: repo.RepoProperty{Value: ""}, want: "—"},
		{name: "url", property: repo.RepoProperty{Value: "https://example.com"}, want: "https://example.com"},
		{name: "unsupported", property: repo.RepoProperty{Value: 1.5}, want: "—"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ansi.Strip(FormatRow(tt.property, tt.max)))
		})
	}

	url := FormatRow(repo.RepoProperty{Value: "https://example.com"}, 0)
	text := FormatRow(repo.RepoProperty{Value: "example"}, 0)
	assert.Contains(t, url, "\x1b[4;", "URLs are underlined")
	assert.NotContains(t, text, "\x1b[4;")

	date := FormatRow(repo.RepoProperty{Value: today.AddDate(0, 0, -3)}, 0)
	assert.Greater(t, strings.Count(date, "38;2;"), 1, "the relative and absolute dates have their own colours")
	assert.Equal(t, today.AddDate(0, 0, -3).Format("02 Jan 2006")+"  3d ago", ansi.Strip(date), "the date leads and the relative part follows")
}

func TestIntText(t *testing.T) {
	assert.Equal(t, "1204", intText(repo.RepoProperty{Value: 1204}, false))
	assert.Equal(t, "1.2k", intText(repo.RepoProperty{Value: 1204}, true))
	assert.Equal(t, "1.5 MB", intText(repo.RepoProperty{Unit: "kb", Value: 1536}, false), "a unit wins over compaction")
	assert.Equal(t, "1.5 MB", intText(repo.RepoProperty{Unit: "kb", Value: 1536}, true))
	assert.Equal(t, "0", intText(repo.RepoProperty{Value: "not an int"}, false))
}

func TestKeyFacts(t *testing.T) {
	today := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	withNow(t, today)
	c := repo.NewRepoConfig(repo.Repository{StargazerCount: 2100, ForkCount: 38, DiskUsage: 49152, PushedAt: today.AddDate(0, 0, -3)})
	c.Properties["Primary Language"] = repo.RepoProperty{Name: "Primary Language", Value: "Go"}
	c.Properties["License Name"] = repo.RepoProperty{Name: "License Name", Value: "MIT License"}

	assert.Equal(t, "Go · MIT · 2.1k ★ · 38 forks · pushed 3d ago · 48.0 MB", ansi.Strip(KeyFacts(c, 80)))
	assert.Equal(t, "Go · MIT · 2.1k ★ · 38 forks", ansi.Strip(KeyFacts(c, 30)), "facts are dropped from the end to fit")
	assert.Equal(t, "0 ★ · 0 forks", ansi.Strip(KeyFacts(repo.NewRepoConfig(repo.Repository{}), 80)), "unknown facts are skipped")
}

func TestAgo(t *testing.T) {
	today := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	withNow(t, today)

	assert.Equal(t, "today", Ago(today.Add(-time.Hour)))
	assert.Equal(t, "today", Ago(today.Add(time.Hour)), "future times count as today")
	assert.Equal(t, "1d ago", Ago(today.AddDate(0, 0, -1)))
	assert.Equal(t, "29d ago", Ago(today.AddDate(0, 0, -29)))
	assert.Equal(t, "3mo ago", Ago(today.AddDate(0, 0, -95)))
	assert.Equal(t, "2y ago", Ago(today.AddDate(-2, 0, -1)))
}

func TestVisibilityMarkAndBadges(t *testing.T) {
	public := repo.NewRepoConfig(repo.Repository{})
	private := repo.NewRepoConfig(repo.Repository{IsPrivate: true})
	archived := repo.NewRepoConfig(repo.Repository{IsPrivate: true, IsArchived: true, IsFork: true, IsTemplate: true})

	assert.Equal(t, "●", ansi.Strip(VisibilityMark(public)))
	assert.NotEqual(t, VisibilityMark(public), VisibilityMark(private))
	assert.NotEqual(t, VisibilityMark(private), VisibilityMark(archived))

	// Badges are pills: half-block caps either side of the label.
	assert.Equal(t, []string{"▐public▌"}, uitest.StripAll(BadgeList(public)))
	assert.Equal(t, []string{"▐private▌", "▐archived▌", "▐fork▌", "▐template▌"}, uitest.StripAll(BadgeList(archived)))
	withLanguage := repo.NewRepoConfig(repo.Repository{})
	withLanguage.Properties["Primary Language"] = repo.NewRepoProperty("Primary Language", "1⟭ Overview", "Go", "string", "")
	assert.Equal(t, []string{"▐public▌", "▐Go▌"}, uitest.StripAll(BadgeList(withLanguage)), "the language is the last pill")
	assert.Equal(t, "● public ● private ● archived", ansi.Strip(VisibilityLegend()))
}
