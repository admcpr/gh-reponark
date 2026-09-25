package repo

import (
	"fmt"
	"strings"
	"time"

	"gh-reponark/shared"

	"charm.land/lipgloss/v2"
)

// now is the clock relative dates are measured against; tests replace it.
var now = time.Now

// staleAfter is how long since a push before a repository is flagged as stale.
const staleAfter = 180 * 24 * time.Hour

// FormatValue renders a property's value for the inspector, styled by type.
func FormatValue(p RepoProperty) string {
	switch value := p.Value.(type) {
	case bool:
		if value {
			return shared.GoodStyle.Render("✓ yes")
		}
		return shared.DimStyle.Render("✗ no")
	case int:
		if p.Name == "Disk Usage" {
			return shared.ValueStyle.Render(Kilobytes(value))
		}
		return countStyle(p, value).Render(fmt.Sprint(value))
	case time.Time:
		if value.IsZero() {
			return shared.DimStyle.Render("—")
		}
		return timeStyle(value).Render(Ago(value)) + shared.DimStyle.Render("  "+value.Format("2006/01/02"))
	case string:
		if value == "" {
			return shared.DimStyle.Render("—")
		}
		return shared.ValueStyle.Render(value)
	default:
		return shared.DimStyle.Render("—")
	}
}

// FormatCell renders a property's value compactly for a matrix column.
func FormatCell(p RepoProperty) string {
	switch value := p.Value.(type) {
	case bool:
		if value {
			return shared.GoodStyle.Render("●")
		}
		return shared.DimStyle.Render("○")
	case int:
		if p.Name == "Disk Usage" {
			return shared.ValueStyle.Render(Kilobytes(value))
		}
		return countStyle(p, value).Render(CompactNumber(value))
	case time.Time:
		if value.IsZero() {
			return shared.DimStyle.Render("—")
		}
		return timeStyle(value).Render(Ago(value))
	default:
		return FormatValue(p)
	}
}

// countStyle highlights counts that need attention.
func countStyle(p RepoProperty, value int) lipgloss.Style {
	if p.Name == "Vulnerability Alerts" && value > 0 {
		return shared.BadStyle
	}
	return shared.ValueStyle
}

func timeStyle(t time.Time) lipgloss.Style {
	if now().Sub(t) > staleAfter {
		return shared.WarnStyle
	}
	return shared.ValueStyle
}

// Ago describes how long before now t was, e.g. "3d ago" or "2y ago".
func Ago(t time.Time) string {
	days := int(now().Sub(t).Hours() / 24)
	switch {
	case days < 1:
		return "today"
	case days < 30:
		return fmt.Sprintf("%dd ago", days)
	case days < 365:
		return fmt.Sprintf("%dmo ago", days/30)
	default:
		return fmt.Sprintf("%dy ago", days/365)
	}
}

// CompactNumber shortens large counts, e.g. 1204 becomes "1.2k".
func CompactNumber(n int) string {
	if n < 1000 {
		return fmt.Sprint(n)
	}
	s := fmt.Sprintf("%.1f", float64(n)/1000)
	return strings.TrimSuffix(s, ".0") + "k"
}

// Kilobytes formats a size given in kilobytes, as GitHub reports disk usage.
func Kilobytes(kb int) string {
	switch {
	case kb < 1024:
		return fmt.Sprintf("%d KB", kb)
	case kb < 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(kb)/1024)
	default:
		return fmt.Sprintf("%.1f GB", float64(kb)/(1024*1024))
	}
}

// VisibilityMark is a coloured dot for a repository: red when archived,
// yellow when private and green when public.
func VisibilityMark(c RepoConfig) string {
	switch {
	case c.Bool("Is Archived"):
		return shared.BadStyle.Render("●")
	case c.Bool("Is Private"):
		return shared.WarnStyle.Render("●")
	default:
		return shared.GoodStyle.Render("●")
	}
}

// VisibilityLegend explains the colours used by VisibilityMark.
func VisibilityLegend() string {
	return shared.GoodStyle.Render("●") + shared.DimStyle.Render(" public ") +
		shared.WarnStyle.Render("●") + shared.DimStyle.Render(" private ") +
		shared.BadStyle.Render("●") + shared.DimStyle.Render(" archived")
}

// Badges labels what kind of repository c is: its visibility and whether it
// is archived, a fork or a template.
func Badges(c RepoConfig) string {
	var badges []string
	if c.Bool("Is Private") {
		badges = append(badges, shared.WarnStyle.Render("private"))
	} else {
		badges = append(badges, shared.GoodStyle.Render("public"))
	}
	if c.Bool("Is Archived") {
		badges = append(badges, shared.BadStyle.Render("archived"))
	}
	if c.Bool("Is Fork") {
		badges = append(badges, shared.ForkStyle.Render("fork"))
	}
	if c.Bool("Is Template") {
		badges = append(badges, shared.AccentStyle.Render("template"))
	}
	return strings.Join(badges, " ")
}
