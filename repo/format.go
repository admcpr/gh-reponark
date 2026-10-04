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

// FormatValue renders a property's value compactly, styled by type. The
// matrix uses it to spell out the focused cell.
func FormatValue(p RepoProperty) string {
	switch value := p.Value.(type) {
	case bool:
		if value {
			return shared.GoodStyle.Render("✓ yes")
		}
		return shared.DimStyle.Render("✗ no")
	case int:
		return countStyle(p, value).Render(intText(p, false))
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
		return countStyle(p, value).Render(intText(p, true))
	case time.Time:
		if value.IsZero() {
			return shared.DimStyle.Render("—")
		}
		return timeStyle(value).Render(Ago(value))
	default:
		return FormatValue(p)
	}
}

// intText is an integer property's value as text: in its unit when it has
// one, otherwise as a plain or, when compact, shortened count.
func intText(p RepoProperty, compact bool) string {
	value, _ := p.Value.(int)
	switch {
	case p.Unit == "kb":
		return Kilobytes(value)
	case compact:
		return CompactNumber(value)
	default:
		return fmt.Sprint(value)
	}
}

// numberWidth is the field integers are right-aligned in on an inspector
// row, so the bars beside them line up.
const numberWidth = 8

// FormatRow renders a property's value for an inspector row. Booleans are a
// toggle cell, integers are right-aligned with a bar against max, the
// largest in the organization, when it is known (identifiers get no bar),
// dates lead with how long ago, and URLs are links.
func FormatRow(p RepoProperty, max int) string {
	switch value := p.Value.(type) {
	case bool:
		if value {
			return shared.ToggleOnStyle.Render("[ ● ]") + shared.GoodStyle.Render(" yes")
		}
		return shared.ToggleOffStyle.Render("[ ○ ]") + shared.DimStyle.Render(" no")
	case int:
		style := countStyle(p, value)
		number := style.Render(shared.FitRight(intText(p, false), numberWidth))
		if max <= 0 || p.Identifier {
			return number
		}
		return number + "  " + bar(value, max, barWidth, style)
	case time.Time:
		if value.IsZero() {
			return shared.DimStyle.Render("—")
		}
		return timeStyle(value).Bold(true).Render(Ago(value)) + "  " + shared.DimStyle.Render(value.Format("2 Jan 2006"))
	case string:
		switch {
		case value == "":
			return shared.DimStyle.Render("—")
		case isLink(value):
			return shared.LinkStyle.Render(value)
		default:
			return shared.ValueStyle.Render(value)
		}
	default:
		return shared.DimStyle.Render("—")
	}
}

// isLink says whether a value is a URL or clone address worth styling as one.
func isLink(s string) bool {
	for _, prefix := range []string{"https://", "http://", "git@", "ssh://"} {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

// KeyFacts is the inspector's one-line summary of a repository, in the dim
// voice of a byline with each figure in its own colour, e.g.
//
//	Go · MIT · 2.1k ★ · 38 forks · pushed 3d ago · 48 MB
//
// Facts are added in that order for as long as they fit in width.
func KeyFacts(c RepoConfig, width int) string {
	var facts []string
	if language := c.Text("Primary Language"); language != "" {
		facts = append(facts, shared.LanguageStyle.Render(language))
	}
	if license := c.Text("License Name"); license != "" {
		facts = append(facts, shared.TextBodyStyle.Render(strings.TrimSuffix(license, " License")))
	}
	facts = append(facts, shared.StarStyle.Render(CompactNumber(c.Int("Stargazer Count"))+" ★"))
	forks, forkWord := c.Int("Fork Count"), " forks"
	if forks == 1 {
		forkWord = " fork"
	}
	facts = append(facts, shared.ForkStyle.Render(CompactNumber(forks))+shared.DimStyle.Render(forkWord))
	if pushed := c.Time("Pushed At"); !pushed.IsZero() {
		facts = append(facts, shared.DimStyle.Render("pushed ")+timeStyle(pushed).Render(Ago(pushed)))
	}
	if size := c.Int("Disk Usage"); size > 0 {
		facts = append(facts, shared.ValueStyle.Render(Kilobytes(size)))
	}

	line, _ := shared.JoinFit(facts, shared.DimStyle.Render(" · "), width)
	return line
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

// BadgeList labels what kind of repository c is as pills: its visibility,
// whether it is archived, a fork or a template, and its primary language.
// They come most important first, so a narrow line can drop pills from the
// end without cutting through one.
func BadgeList(c RepoConfig) []string {
	var badges []string
	if c.Bool("Is Private") {
		badges = append(badges, shared.Pill("private", shared.AppColors.Warn))
	} else {
		badges = append(badges, shared.Pill("public", shared.AppColors.Good))
	}
	if c.Bool("Is Archived") {
		badges = append(badges, shared.Pill("archived", shared.AppColors.Bad))
	}
	if c.Bool("Is Fork") {
		badges = append(badges, shared.Pill("fork", shared.ForkStyle.GetForeground()))
	}
	if c.Bool("Is Template") {
		badges = append(badges, shared.Pill("template", shared.AppColors.Accent))
	}
	if language := c.Text("Primary Language"); language != "" {
		badges = append(badges, shared.Pill(language, shared.LanguageStyle.GetForeground()))
	}
	return badges
}

// ShortName drops the words every property in a group shares, so "Has Wiki
// Enabled" reads as "Wiki" where the group already says what it is.
func ShortName(name string) string {
	for _, prefix := range []string{"Viewer Can ", "Viewer ", "Is ", "Has "} {
		name = strings.TrimPrefix(name, prefix)
	}
	for _, suffix := range []string{" Enabled", " Allowed", " Count"} {
		name = strings.TrimSuffix(name, suffix)
	}
	return name
}
