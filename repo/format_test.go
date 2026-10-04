package repo

import (
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
		property RepoProperty
		want     string
	}{
		{name: "true", property: RepoProperty{Value: true}, want: "✓ yes"},
		{name: "false", property: RepoProperty{Value: false}, want: "✗ no"},
		{name: "int", property: RepoProperty{Value: 1204}, want: "1204"},
		{name: "disk usage", property: RepoProperty{Name: "Disk Usage", Unit: "kb", Value: 2048}, want: "2.0 MB"},
		{name: "string", property: RepoProperty{Value: "main"}, want: "main"},
		{name: "empty string", property: RepoProperty{Value: ""}, want: "—"},
		{name: "zero time", property: RepoProperty{Value: time.Time{}}, want: "—"},
		{name: "time", property: RepoProperty{Value: today.AddDate(0, 0, -3)}, want: "3d ago  2026/09/22"},
		{name: "unsupported", property: RepoProperty{Value: 1.5}, want: "—"},
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

	plainCount := FormatValue(RepoProperty{Name: "Open Issues", Value: 3})
	alerts := FormatValue(RepoProperty{Name: "Vulnerability Alerts", Value: 3})
	assert.NotEqual(t, plainCount, alerts, "open alerts are coloured differently from other counts")

	recent := FormatCell(RepoProperty{Value: today.AddDate(0, -1, 0)})
	stale := FormatCell(RepoProperty{Value: today.AddDate(-1, 0, 0)})
	assert.NotEqual(t, recent[:12], stale[:12], "stale dates are coloured differently")
}

func TestFormatCell(t *testing.T) {
	withNow(t, time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name     string
		property RepoProperty
		want     string
	}{
		{name: "true", property: RepoProperty{Value: true}, want: "●"},
		{name: "false", property: RepoProperty{Value: false}, want: "○"},
		{name: "int", property: RepoProperty{Value: 1204}, want: "1.2k"},
		{name: "disk usage", property: RepoProperty{Unit: "kb", Value: 1536}, want: "1.5 MB"},
		{name: "time", property: RepoProperty{Value: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)}, want: "2mo ago"},
		{name: "string", property: RepoProperty{Value: "main"}, want: "main"},
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
		property RepoProperty
		max      int
		want     string
	}{
		{name: "true", property: RepoProperty{Value: true}, want: "[ ● ] yes"},
		{name: "false", property: RepoProperty{Value: false}, want: "[ ○ ] no"},
		{name: "int is right-aligned in eight cells", property: RepoProperty{Name: "Open Issues", Value: 100}, want: "     100"},
		{name: "int with bar", property: RepoProperty{Name: "Open Issues", Value: 100}, max: 400, want: "     100  ━━━━━╌╌╌╌╌"},
		{name: "zero keeps an empty bar", property: RepoProperty{Name: "Open Issues", Value: 0}, max: 400, want: "       0  ╌╌╌╌╌╌╌╌╌╌"},
		{name: "no max, no bar", property: RepoProperty{Name: "Open Issues", Value: 100}, want: "     100"},
		{name: "identifiers get no bar", property: RepoProperty{Name: "Database ID", Identifier: true, Value: 100}, max: 400, want: "     100"},
		{name: "disk usage keeps its unit", property: RepoProperty{Name: "Disk Usage", Unit: "kb", Value: 1536}, want: "  1.5 MB"},
		{name: "date", property: RepoProperty{Value: today.AddDate(0, 0, -3)}, want: "3d ago  22 Sep 2026"},
		{name: "zero date", property: RepoProperty{Value: time.Time{}}, want: "—"},
		{name: "text", property: RepoProperty{Value: "main"}, want: "main"},
		{name: "empty text", property: RepoProperty{Value: ""}, want: "—"},
		{name: "url", property: RepoProperty{Value: "https://example.com"}, want: "https://example.com"},
		{name: "unsupported", property: RepoProperty{Value: 1.5}, want: "—"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ansi.Strip(FormatRow(tt.property, tt.max)))
		})
	}

	url := FormatRow(RepoProperty{Value: "https://example.com"}, 0)
	text := FormatRow(RepoProperty{Value: "example"}, 0)
	assert.Contains(t, url, "\x1b[4;", "URLs are underlined")
	assert.NotContains(t, text, "\x1b[4;")

	date := FormatRow(RepoProperty{Value: today.AddDate(0, 0, -3)}, 0)
	assert.Greater(t, strings.Count(date, "38;2;"), 1, "the relative and absolute dates have their own colours")
}

func TestIntText(t *testing.T) {
	assert.Equal(t, "1204", intText(RepoProperty{Value: 1204}, false))
	assert.Equal(t, "1.2k", intText(RepoProperty{Value: 1204}, true))
	assert.Equal(t, "1.5 MB", intText(RepoProperty{Unit: "kb", Value: 1536}, false), "a unit wins over compaction")
	assert.Equal(t, "1.5 MB", intText(RepoProperty{Unit: "kb", Value: 1536}, true))
	assert.Equal(t, "0", intText(RepoProperty{Value: "not an int"}, false))
}

func TestKeyFacts(t *testing.T) {
	today := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	withNow(t, today)
	c := NewRepoConfig(Repository{StargazerCount: 2100, ForkCount: 38, DiskUsage: 49152, PushedAt: today.AddDate(0, 0, -3)})
	c.Properties["Primary Language"] = RepoProperty{Name: "Primary Language", Value: "Go"}
	c.Properties["License Name"] = RepoProperty{Name: "License Name", Value: "MIT License"}

	assert.Equal(t, "Go · MIT · 2.1k ★ · 38 forks · pushed 3d ago · 48.0 MB", ansi.Strip(KeyFacts(c, 80)))
	assert.Equal(t, "Go · MIT · 2.1k ★ · 38 forks", ansi.Strip(KeyFacts(c, 30)), "facts are dropped from the end to fit")
	assert.Equal(t, "0 ★ · 0 forks", ansi.Strip(KeyFacts(NewRepoConfig(Repository{}), 80)), "unknown facts are skipped")
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

func TestCompactNumber(t *testing.T) {
	assert.Equal(t, "0", CompactNumber(0))
	assert.Equal(t, "999", CompactNumber(999))
	assert.Equal(t, "1k", CompactNumber(1000))
	assert.Equal(t, "1.2k", CompactNumber(1204))
	assert.Equal(t, "15.3k", CompactNumber(15300))
}

func TestKilobytes(t *testing.T) {
	assert.Equal(t, "512 KB", Kilobytes(512))
	assert.Equal(t, "1.5 MB", Kilobytes(1536))
	assert.Equal(t, "2.0 GB", Kilobytes(2*1024*1024))
}

func TestShortName(t *testing.T) {
	assert.Equal(t, "Wiki", ShortName("Has Wiki Enabled"))
	assert.Equal(t, "Archived", ShortName("Is Archived"))
	assert.Equal(t, "Administer", ShortName("Viewer Can Administer"))
	assert.Equal(t, "Stargazer", ShortName("Stargazer Count"))
	assert.Equal(t, "Default Branch", ShortName("Default Branch"))
}

func TestVisibilityMarkAndBadges(t *testing.T) {
	public := NewRepoConfig(Repository{})
	private := NewRepoConfig(Repository{IsPrivate: true})
	archived := NewRepoConfig(Repository{IsPrivate: true, IsArchived: true, IsFork: true, IsTemplate: true})

	assert.Equal(t, "●", ansi.Strip(VisibilityMark(public)))
	assert.NotEqual(t, VisibilityMark(public), VisibilityMark(private))
	assert.NotEqual(t, VisibilityMark(private), VisibilityMark(archived))

	// Badges are pills: half-block caps either side of the label.
	assert.Equal(t, []string{"▐public▌"}, stripAll(BadgeList(public)))
	assert.Equal(t, []string{"▐private▌", "▐archived▌", "▐fork▌", "▐template▌"}, stripAll(BadgeList(archived)))
	withLanguage := NewRepoConfig(Repository{})
	withLanguage.Properties["Primary Language"] = NewRepoProperty("Primary Language", "1⟭ Overview", "Go", "string", "")
	assert.Equal(t, []string{"▐public▌", "▐Go▌"}, stripAll(BadgeList(withLanguage)), "the language is the last pill")
	assert.Equal(t, "● public ● private ● archived", ansi.Strip(VisibilityLegend()))
}

func stripAll(lines []string) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = ansi.Strip(l)
	}
	return out
}
