package inspector

import (
	"fmt"
	"strings"
	"time"

	"gh-reponark/repo"
	"gh-reponark/ui"

	"charm.land/lipgloss/v2"
)

// now is the clock relative dates are measured against; tests replace it.
var now = time.Now

// staleAfter is how long since a push before a repository is flagged as stale.
const staleAfter = 180 * 24 * time.Hour

// FormatValue renders a property's value compactly, styled by type. The
// matrix uses it to spell out the focused cell.
func FormatValue(p repo.RepoProperty) string {
	switch value := p.Value.(type) {
	case bool:
		if value {
			return ui.GoodStyle.Render("✓ yes")
		}
		return ui.DimStyle.Render("✗ no")
	case int:
		return countStyle(p, value).Render(intText(p, false))
	case time.Time:
		if value.IsZero() {
			return ui.DimStyle.Render("—")
		}
		return timeStyle(value).Render(Ago(value)) + ui.DimStyle.Render("  "+value.Format("2006/01/02"))
	case string:
		if value == "" {
			return ui.DimStyle.Render("—")
		}
		return ui.ValueStyle.Render(value)
	default:
		return ui.DimStyle.Render("—")
	}
}

// FormatCell renders a property's value compactly for a matrix column.
func FormatCell(p repo.RepoProperty) string {
	switch value := p.Value.(type) {
	case bool:
		if value {
			return ui.GoodStyle.Render("●")
		}
		return ui.DimStyle.Render("○")
	case int:
		return countStyle(p, value).Render(intText(p, true))
	case time.Time:
		if value.IsZero() {
			return ui.DimStyle.Render("—")
		}
		return timeStyle(value).Render(Ago(value))
	default:
		return FormatValue(p)
	}
}

// intText is an integer property's value as text: in its unit when it has
// one, otherwise as a plain or, when compact, shortened count.
func intText(p repo.RepoProperty, compact bool) string {
	value, _ := p.Value.(int)
	switch {
	case p.Unit == "kb":
		return repo.Kilobytes(value)
	case compact:
		return repo.CompactNumber(value)
	default:
		return fmt.Sprint(value)
	}
}

// valueWidth is the field integers and dates are right-aligned in on an
// inspector row: wide enough for "02 Jan 2006", so numbers and dates share a
// right edge and the bars and relative ages beside them start in the same
// column.
const valueWidth = 11

// FormatRow renders a property's value for an inspector row. Booleans are a
// toggle cell, integers are right-aligned with a bar against max, the
// largest in the organization, when it is known (identifiers get no bar),
// dates lead with how long ago, and URLs are links.
func FormatRow(p repo.RepoProperty, max int) string {
	switch value := p.Value.(type) {
	case bool:
		if value {
			return ui.ToggleOnStyle.Render("[ ● ]") + ui.GoodStyle.Render(" yes")
		}
		return ui.ToggleOffStyle.Render("[ ○ ]") + ui.DimStyle.Render(" no")
	case int:
		style := countStyle(p, value)
		number := style.Render(ui.FitRight(intText(p, false), valueWidth))
		if max <= 0 || p.Identifier {
			return number
		}
		return number + "  " + bar(value, max, barWidth, style)
	case time.Time:
		if value.IsZero() {
			return ui.DimStyle.Render("—")
		}
		// The date is the fact and leads; how long ago it was follows in a
		// quieter voice, or in the warning colour when it is stale.
		ago := ui.DimStyle
		if now().Sub(value) > staleAfter {
			ago = ui.WarnStyle
		}
		return ui.ValueStyle.Render(ui.FitRight(value.Format("02 Jan 2006"), valueWidth)) + "  " + ago.Render(Ago(value))
	case string:
		switch {
		case value == "":
			return ui.DimStyle.Render("—")
		case isLink(value):
			return ui.LinkStyle.Render(value)
		default:
			return ui.ValueStyle.Render(value)
		}
	default:
		return ui.DimStyle.Render("—")
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
func KeyFacts(c repo.RepoConfig, width int) string {
	var facts []string
	if language := c.Text("Primary Language"); language != "" {
		facts = append(facts, ui.LanguageStyle.Render(language))
	}
	if license := c.Text("License Name"); license != "" {
		facts = append(facts, ui.TextBodyStyle.Render(strings.TrimSuffix(license, " License")))
	}
	facts = append(facts, ui.StarStyle.Render(repo.CompactNumber(c.Int("Stargazer Count"))+" ★"))
	forks, forkWord := c.Int("Fork Count"), " forks"
	if forks == 1 {
		forkWord = " fork"
	}
	facts = append(facts, ui.ForkStyle.Render(repo.CompactNumber(forks))+ui.DimStyle.Render(forkWord))
	if pushed := c.Time("Pushed At"); !pushed.IsZero() {
		facts = append(facts, ui.DimStyle.Render("pushed ")+timeStyle(pushed).Render(Ago(pushed)))
	}
	if size := c.Int("Disk Usage"); size > 0 {
		facts = append(facts, ui.ValueStyle.Render(repo.Kilobytes(size)))
	}

	line, _ := ui.JoinFit(facts, ui.DimStyle.Render(" · "), width)
	return line
}

// countStyle highlights counts that need attention.
func countStyle(p repo.RepoProperty, value int) lipgloss.Style {
	if p.Name == "Vulnerability Alerts" && value > 0 {
		return ui.BadStyle
	}
	return ui.ValueStyle
}

func timeStyle(t time.Time) lipgloss.Style {
	if now().Sub(t) > staleAfter {
		return ui.WarnStyle
	}
	return ui.ValueStyle
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

// VisibilityMark is a coloured dot for a repository: red when archived,
// yellow when private and green when public.
func VisibilityMark(c repo.RepoConfig) string {
	switch {
	case c.Bool("Is Archived"):
		return ui.BadStyle.Render("●")
	case c.Bool("Is Private"):
		return ui.WarnStyle.Render("●")
	default:
		return ui.GoodStyle.Render("●")
	}
}

// VisibilityLegend explains the colours used by VisibilityMark.
func VisibilityLegend() string {
	return ui.GoodStyle.Render("●") + ui.DimStyle.Render(" public ") +
		ui.WarnStyle.Render("●") + ui.DimStyle.Render(" private ") +
		ui.BadStyle.Render("●") + ui.DimStyle.Render(" archived")
}

// BadgeList labels what kind of repository c is as pills: its visibility,
// whether it is archived, a fork or a template, and its primary language.
// They come most important first, so a narrow line can drop pills from the
// end without cutting through one.
func BadgeList(c repo.RepoConfig) []string {
	var badges []string
	if c.Bool("Is Private") {
		badges = append(badges, ui.Pill("private", ui.AppColors.Warn))
	} else {
		badges = append(badges, ui.Pill("public", ui.AppColors.Good))
	}
	if c.Bool("Is Archived") {
		badges = append(badges, ui.Pill("archived", ui.AppColors.Bad))
	}
	if c.Bool("Is Fork") {
		badges = append(badges, ui.Pill("fork", ui.ForkStyle.GetForeground()))
	}
	if c.Bool("Is Template") {
		badges = append(badges, ui.Pill("template", ui.AppColors.Accent))
	}
	if language := c.Text("Primary Language"); language != "" {
		badges = append(badges, ui.Pill(language, ui.LanguageStyle.GetForeground()))
	}
	return badges
}
