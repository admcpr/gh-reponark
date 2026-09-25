package org

import (
	"fmt"
	"strings"

	"gh-reponark/repo"
	"gh-reponark/shared"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// matrixChrome is the number of lines around the matrix rows: tabs, two
// heading lines and a rule above; a rule, totals and the focused cell below.
const matrixChrome = 7

// matrixRows is how many repositories the matrix shows.
func (m *Model) matrixRows() int { return shared.Max(1, m.height-matrixChrome) }

// matrixColumn is one property shown as a column of the matrix.
type matrixColumn struct {
	property repo.PropertySchema
	heading  [2]string
	width    int
}

func newMatrixColumn(p repo.PropertySchema) matrixColumn {
	heading := splitHeading(shortName(p.Name))
	content := shared.Max(lipgloss.Width(heading[0]), lipgloss.Width(heading[1]))
	switch p.Type {
	case "int":
		content = shared.Max(content, 6)
	case "time.Time":
		content = shared.Max(content, 8)
	case "string":
		content = shared.Max(content, 14)
	}
	return matrixColumn{property: p, heading: heading, width: shared.Min(content, 18) + 2}
}

// shortName drops the words every property in a group shares, so "Has Wiki
// Enabled" is headed "Wiki".
func shortName(name string) string {
	for _, prefix := range []string{"Viewer Can ", "Viewer ", "Is ", "Has "} {
		name = strings.TrimPrefix(name, prefix)
	}
	for _, suffix := range []string{" Enabled", " Allowed", " Count"} {
		name = strings.TrimSuffix(name, suffix)
	}
	return name
}

// splitHeading breaks a heading over two lines where it makes the wider line
// narrowest, bottom-aligned so one-word headings sit on the rule.
func splitHeading(heading string) [2]string {
	words := strings.Fields(heading)
	if len(words) < 2 {
		return [2]string{"", heading}
	}
	best, bestWidth := 1, len(heading)
	for i := 1; i < len(words); i++ {
		top, bottom := strings.Join(words[:i], " "), strings.Join(words[i:], " ")
		if w := shared.Max(len(top), len(bottom)); w < bestWidth {
			best, bestWidth = i, w
		}
	}
	return [2]string{strings.Join(words[:best], " "), strings.Join(words[best:], " ")}
}

// visibleColumns returns the columns that fit in width, scrolled so the
// focused one is among them, and whether any are hidden on either side.
func (m *Model) visibleColumns(columns []matrixColumn, width int) (shown []matrixColumn, first int, more bool) {
	active := m.repoModel.ActiveProperty()
	m.columnOffset = shared.Min(m.columnOffset, active)
	for {
		used := 0
		for i := m.columnOffset; i <= active; i++ {
			used += columns[i].width
		}
		if used <= width || m.columnOffset == active {
			break
		}
		m.columnOffset++
	}

	used := 0
	for i := m.columnOffset; i < len(columns) && used+columns[i].width <= width; i++ {
		shown = append(shown, columns[i])
		used += columns[i].width
	}
	return shown, m.columnOffset, m.columnOffset+len(shown) < len(columns)
}

// matrixView shows every visible repository against every property in the
// active group, so a repository set up differently from the rest stands out.
func (m *Model) matrixView() string {
	group := m.repoModel.ActiveGroup()
	active := m.repoModel.ActiveProperty()
	rows := m.matrixRows()

	nameWidth := 10
	for _, c := range m.visible {
		nameWidth = shared.Max(nameWidth, lipgloss.Width(c.Name))
	}
	nameWidth = shared.Min(nameWidth, shared.Max(10, m.width/3))
	prefixWidth := nameWidth + 7 // marker, dot and the divider around the name

	columns := make([]matrixColumn, len(group.Properties))
	for i, p := range group.Properties {
		columns[i] = newMatrixColumn(p)
	}
	shown, first, more := m.visibleColumns(columns, m.width-prefixWidth-2)

	lines := []string{repo.RenderTabs(repo.GroupTitles(), m.width, m.repoModel.ActiveTab())}

	for line := 0; line < 2; line++ {
		row := strings.Repeat(" ", prefixWidth)
		if line == 1 && first > 0 {
			row = shared.Fit("", prefixWidth-2) + shared.DimStyle.Render("‹ ")
		}
		for i, col := range shown {
			style := shared.ColumnHeading
			if first+i == active {
				style = shared.AccentStyle.Bold(true)
			}
			row += style.Render(shared.Fit(col.heading[line], col.width))
		}
		if line == 1 && more {
			row += shared.DimStyle.Render("›")
		}
		lines = append(lines, row)
	}

	rule := func(joint string) string {
		return shared.DimStyle.Render(strings.Repeat("─", prefixWidth-2) + joint + strings.Repeat("─", shared.Max(0, m.width-prefixWidth+1)))
	}
	lines = append(lines, rule("┼"))

	for r := m.matrixOffset; r < len(m.visible) && r < m.matrixOffset+rows; r++ {
		lines = append(lines, m.matrixRow(m.visible[r], r == m.cursor, shown, first, nameWidth))
	}
	for len(lines) < rows+4 {
		lines = append(lines, strings.Repeat(" ", prefixWidth-2)+shared.DimStyle.Render("│"))
	}

	lines = append(lines, rule("┴"), m.matrixTotals(shown, prefixWidth), m.focusedCell(group.Properties[active]))
	return shared.Lines(lines, m.width, m.height)
}

func (m *Model) matrixRow(c repo.RepoConfig, selected bool, shown []matrixColumn, first, nameWidth int) string {
	marker, name := "  ", shared.Fit(c.Name, nameWidth)
	if selected {
		marker = shared.AccentStyle.Render("▌ ")
		name = shared.StrongStyle.Render(name)
	}
	row := marker + repo.VisibilityMark(c) + " " + name + shared.DimStyle.Render(" │ ")

	for i, col := range shown {
		cell := repo.FormatCell(c.Properties[col.property.Name])
		if col.property.Type == "int" {
			cell = shared.FitRight(cell, col.width-2) + "  "
		} else {
			cell = shared.Fit(cell, col.width)
		}
		if selected && first+i == m.repoModel.ActiveProperty() {
			cell = focusCell(cell)
		}
		row += cell
	}
	return row
}

// focusCell highlights a cell's content with one space either side, leaving
// the rest of the column as padding so the focus reads as one cell.
func focusCell(cell string) string {
	content := []rune(ansi.Strip(cell))
	start, end := 0, len(content)
	for start < end && content[start] == ' ' {
		start++
	}
	for end > start && content[end-1] == ' ' {
		end--
	}
	start, end = shared.Max(0, start-1), shared.Min(len(content), end+1)
	return string(content[:start]) + shared.FocusCellStyle.Render(string(content[start:end])) + string(content[end:])
}

// matrixTotals counts, under each yes/no column, how many visible
// repositories have the setting on. Mixed columns are highlighted because
// they are the ones worth a closer look.
func (m *Model) matrixTotals(shown []matrixColumn, prefixWidth int) string {
	row := shared.DimStyle.Render(shared.Fit("  on", prefixWidth-2) + "│ ")
	for _, col := range shown {
		if col.property.Type != "bool" {
			row += strings.Repeat(" ", col.width)
			continue
		}
		on := 0
		for _, c := range m.visible {
			if c.Bool(col.property.Name) {
				on++
			}
		}
		style := shared.DimStyle
		if on > 0 && on < len(m.visible) {
			style = shared.AccentStyle
		}
		row += style.Render(shared.Fit(fmt.Sprintf("%d/%d", on, len(m.visible)), col.width))
	}
	return row
}

// focusedCell spells out the focused cell's repository, property, full value
// and what the property means.
func (m *Model) focusedCell(p repo.PropertySchema) string {
	selected, _ := m.selectedRepo()
	return shared.StrongStyle.Render(selected.Name) +
		shared.DimStyle.Render(" · "+p.Name+" = ") +
		repo.FormatValue(selected.Properties[p.Name]) +
		shared.DimStyle.Render("   "+p.Description)
}
