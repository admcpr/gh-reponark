package accounts

import (
	"fmt"
	"strings"

	"gh-reponark/repo"
	"gh-reponark/ui"

	"charm.land/lipgloss/v2"
)

// Layout thresholds. Below shortHeight the cards shrink to two lines and
// below narrowWidth they lose the monogram and the URL line.
const (
	shortHeight = 16
	narrowWidth = 50
)

// short is true when the pane is too low for four-line cards.
func (m Model) short() bool { return m.height < shortHeight }

// narrow is true when the pane is too narrow for the monogram and URL.
func (m Model) narrow() bool { return m.width < narrowWidth }

// cardLines is how many lines one card takes at the current size.
func (m Model) cardLines() int {
	switch {
	case m.short():
		return 2
	case m.narrow():
		return 3
	default:
		return 4
	}
}

// heading is the one-line strip above the cards: the column heading and a
// chip counting the organizations.
func (m Model) heading(width int) string {
	count := len(m.items) - 1
	var chip string
	switch count {
	case 0:
		chip = ui.GhostPill("no organisations")
	case 1:
		chip = ui.TintedPill("1 organisation", ui.AppColors.Accent, ui.AppColors.AccentTint)
	default:
		chip = ui.TintedPill(fmt.Sprintf("%d organisations", count), ui.AppColors.Accent, ui.AppColors.AccentTint)
	}
	return ui.Fit(ui.ColumnHeading.Render("  ACCOUNTS")+"  "+chip, width)
}

// card renders one account. The selected card carries the accent marker on
// every line and sits on the selection band; the others are quieter.
func (m Model) card(a account, selected bool, width int) []string {
	marker := "  "
	if selected {
		marker = ui.Marker(ui.AppColors.Accent)
	}
	// Text lines start under the name, which sits after the monogram.
	indent := marker
	if !m.narrow() {
		indent += "   "
	}
	inner := width - lipgloss.Width(indent)

	lines := []string{marker + m.nameLine(a, selected, width-lipgloss.Width(marker))}
	if !m.short() {
		lines = append(lines, indent+m.descriptionLine(a, inner))
	}
	lines = append(lines, indent+m.factsLine(a, inner))
	if !m.short() && !m.narrow() {
		url := ui.DimStyle
		if selected {
			url = ui.LinkTextStyle
		}
		lines = append(lines, indent+url.Render(ui.Fit(a.url, inner)))
	}

	for i, line := range lines {
		if selected {
			lines[i] = ui.HighlightRow(line, width, true)
		} else {
			lines[i] = ui.Fit(line, width)
		}
	}
	return lines
}

// nameLine is the card's first line: the monogram, the display name with
// the login beside it, and the account's pills.
func (m Model) nameLine(a account, selected bool, width int) string {
	var pills []string
	if a.isUser {
		pills = append(pills, ui.Pill("you", ui.AppColors.Accent))
	} else {
		pills = append(pills, ui.GhostPill("org"))
	}
	if a.admin {
		pills = append(pills, ui.TintedPill("admin", ui.AppColors.Good, ui.AppColors.GoodTint))
	}
	if a.verified {
		pills = append(pills, ui.TintedPill("verified", ui.AppColors.Link, verifiedTint))
	}
	tags := strings.Join(pills, " ")

	prefix := ""
	if !m.narrow() {
		prefix = monogram(a.login) + " "
	}

	name, login := a.name, ""
	if name == "" {
		name = a.login
	} else if a.name != a.login {
		login = a.login
	}
	nameStyle := ui.TextBodyStyle.Bold(true)
	if selected {
		nameStyle = ui.StrongStyle
	}
	title := nameStyle.Render(name)
	if login != "" {
		title += "  " + ui.DimStyle.Render(login)
	}

	// The pills are never cut; the title gives way first, losing the login
	// whole before the name is truncated.
	room := width - lipgloss.Width(prefix) - lipgloss.Width(tags) - 2
	if room < 1 {
		return ui.Fit(prefix+title, width)
	}
	if lipgloss.Width(title) > room && login != "" {
		title = nameStyle.Render(name)
	}
	if lipgloss.Width(title) > room {
		title = ui.Fit(title, room)
	}
	return prefix + title + "  " + tags
}

// descriptionLine is the bio or description on one line, or a quiet note
// that there is none.
func (m Model) descriptionLine(a account, width int) string {
	if strings.TrimSpace(a.description) == "" {
		return ui.DimStyle.Italic(true).Render("No description")
	}
	return ui.TextBodyStyle.Render(ui.Fit(strings.Join(strings.Fields(a.description), " "), width))
}

// factsLine is the chips line: counts in the foreground colour with what
// they count in dim, shed from the end when the pane is narrow.
func (m Model) factsLine(a account, width int) string {
	fact := func(n int, word string) string {
		return ui.TextBodyStyle.Render(repo.CompactNumber(n)) + ui.DimStyle.Render(" "+word)
	}
	facts := []string{fact(a.repos, ui.Plural(a.repos, "repo", "repos"))}
	if a.repos > 0 {
		facts = append(facts, fact(a.public, "public"))
		if private := a.repos - a.public; private > 0 {
			facts = append(facts, fact(private, "private"))
		}
	}
	if a.members > 0 {
		word := ui.Plural(a.members, "member", "members")
		if a.isUser {
			word = ui.Plural(a.members, "follower", "followers")
		}
		facts = append(facts, fact(a.members, word))
	}
	if !a.created.IsZero() {
		facts = append(facts, ui.DimStyle.Render("since ")+ui.TextBodyStyle.Render(fmt.Sprint(a.created.Year())))
	}
	line, _ := ui.JoinFit(facts, ui.DimStyle.Render(" · "), width)
	return line
}

// verifiedTint is the dark version of the link colour, the background of
// the verified pill.
var verifiedTint = lipgloss.Darken(ui.AppColors.Link, 0.7)
