package repo

import (
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
		{name: "disk usage", property: RepoProperty{Name: "Disk Usage", Value: 2048}, want: "2.0 MB"},
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
		{name: "time", property: RepoProperty{Value: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)}, want: "2mo ago"},
		{name: "string", property: RepoProperty{Value: "main"}, want: "main"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ansi.Strip(FormatCell(tt.property)))
		})
	}
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

func TestVisibilityMarkAndBadges(t *testing.T) {
	public := NewRepoConfig(Repository{})
	private := NewRepoConfig(Repository{IsPrivate: true})
	archived := NewRepoConfig(Repository{IsPrivate: true, IsArchived: true, IsFork: true, IsTemplate: true})

	assert.Equal(t, "●", ansi.Strip(VisibilityMark(public)))
	assert.NotEqual(t, VisibilityMark(public), VisibilityMark(private))
	assert.NotEqual(t, VisibilityMark(private), VisibilityMark(archived))

	assert.Equal(t, "public", ansi.Strip(Badges(public)))
	assert.Equal(t, "private archived fork template", ansi.Strip(Badges(archived)))
	assert.Equal(t, "● public ● private ● archived", ansi.Strip(VisibilityLegend()))
}
