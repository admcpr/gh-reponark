package filters

import (
	"fmt"
	"sort"
	"strings"

	"gh-reponark/inspector"
	"gh-reponark/repo"
	"gh-reponark/ui"

	"charm.land/lipgloss/v2"
)

// listChrome is the number of lines above the property list: the search
// line and the active filters.
const listChrome = 2

func (m *Model) listWidth() int {
	return ui.Max(30, ui.Min(44, m.width*2/5))
}

func (m *Model) listLines() []string {
	width := m.listWidth()

	var search string
	switch {
	case m.searching:
		search = m.search.View()
	case m.search.Value() != "":
		search = ui.AccentStyle.Render("/ ") + ui.ValueStyle.Render(m.search.Value())
	default:
		search = ui.AccentStyle.Render("/") + ui.DimStyle.Render(" search properties")
	}
	lines := []string{search, m.chips(width)}

	// Lay out the matching properties under their group headings, noting
	// which line the highlight is on so the list can scroll to it.
	var rows []string
	cursorLine, group := 0, ""
	for i, index := range m.matches {
		p := m.properties[index]
		if p.Group != group {
			group = p.Group
			rows = append(rows, headingLabel(repo.GroupTitle(group)))
		}
		if i == m.cursor.Index {
			cursorLine = len(rows)
		}
		rows = append(rows, m.propertyRow(p, i == m.cursor.Index, width))
	}
	if len(rows) == 0 {
		rows = append(rows, ui.TextBodyStyle.Render("No properties match your search"))
	}

	visible := ui.Max(1, m.height-listChrome)
	// Keep the group heading above the first property in view.
	target := cursorLine
	if m.cursor.Index == 0 {
		target = 0
	}
	m.offset = ui.ScrollOffset(m.offset, target, visible, len(rows))
	m.offset = ui.ScrollOffset(m.offset, cursorLine, visible, len(rows))
	end := ui.Min(len(rows), m.offset+visible)
	return append(lines, rows[m.offset:end]...)
}

// headingLabel is a section title in the column-heading style: upper case
// and dim, the way the list's group headings and the editor's sections are
// both set.
func headingLabel(title string) string {
	return ui.ColumnHeading.Render(strings.ToUpper(title))
}

// heading is a headingLabel followed by a dim rule to the edge that keeps
// the editor's sections apart, with optional detail on the right, e.g.
// "MATCHING ──────── 3 of 9".
func heading(title, detail string, width int) string {
	left := headingLabel(title) + " "
	right := ""
	if detail != "" {
		right = " " + ui.AccentStyle.Render(detail)
	}
	fill := ui.Max(0, width-lipgloss.Width(left)-lipgloss.Width(right))
	return left + ui.DimStyle.Render(strings.Repeat("─", fill)) + right
}

func (m *Model) propertyRow(p repo.PropertySchema, highlighted bool, width int) string {
	f, active := m.filters[p.Name]

	marker := "  "
	if highlighted {
		marker = ui.Marker(ui.AppColors.Accent)
		if m.searching || m.editing {
			marker = ui.Marker(ui.AppColors.Dim)
		}
	}
	name := ui.TextBodyStyle
	if highlighted || active {
		name = ui.StrongStyle
	}

	// The condition of an active filter sits at the right edge.
	condition := ""
	if active {
		condition = f.Condition()
		if lipgloss.Width(condition) > ui.Half(width) {
			condition = ui.Fit(condition, ui.Half(width))
		}
	}
	nameWidth := width - 4 - lipgloss.Width(condition) - 1
	row := marker + inspector.TypeOf(p.Type).Glyph() + " " + name.Render(ui.Fit(p.Name, nameWidth)) + " " + ui.AccentStyle.Render(condition)
	if highlighted {
		return ui.HighlightRow(row, width, !m.searching && !m.editing)
	}
	return row
}

// chips shows each active filter as a chip in the accent tint, the way the
// inspector shows posture, with a count of any that do not fit on the line.
func (m *Model) chips(width int) string {
	if len(m.filters) == 0 {
		return ui.DimStyle.Render("No filters yet: every repo is shown")
	}
	names := make([]string, 0, len(m.filters))
	for name := range m.filters {
		names = append(names, name)
	}
	sort.Strings(names)

	line := ""
	for i, name := range names {
		chip := ui.TintedPill(name+": "+m.filters[name].Condition(), ui.AppColors.Accent, ui.AppColors.AccentTint)
		more := ""
		if rest := len(names) - i - 1; rest > 0 {
			more = ui.AccentStyle.Render(fmt.Sprintf(" +%d", rest))
		}
		if lipgloss.Width(line)+lipgloss.Width(chip)+lipgloss.Width(more) > width {
			if line == "" {
				return ui.Fit(chip, width)
			}
			return line + ui.AccentStyle.Render(fmt.Sprintf("+%d", len(names)-i))
		}
		line += chip + " "
	}
	return line
}

// matchingNames lists repos a line at a time, each with its visibility dot,
// using at most room lines and ending with a count of any left over.
func matchingNames(repos []repo.RepoConfig, width, room int) []string {
	if len(repos) == 0 {
		return []string{ui.WarnStyle.Render("No repos match these filters")}
	}
	var lines []string
	line := ""
	for i, c := range repos {
		entry := inspector.VisibilityMark(c) + " " + ui.ValueStyle.Render(c.Name)
		if line != "" && lipgloss.Width(line)+2+lipgloss.Width(entry) > width {
			if len(lines) == room-1 {
				return append(lines, line+ui.AccentStyle.Render(fmt.Sprintf("  +%d more", len(repos)-i)))
			}
			lines = append(lines, line)
			line = ""
		}
		if line != "" {
			line += "  "
		}
		line += entry
	}
	return append(lines, line)
}
