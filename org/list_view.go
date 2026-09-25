package org

import (
	"fmt"
	"strings"

	"gh-reponark/repo"
	"gh-reponark/shared"

	"charm.land/lipgloss/v2"
)

// Column widths in the repository list. The name column takes what is left.
const (
	languageWidth = 11
	starsWidth    = 6
	pushedWidth   = 9
)

// listWidth is the width of the repository list; the inspector gets the rest
// after the scrollbar and a gap.
func (m *Model) listWidth() int      { return shared.Half(m.width) }
func (m *Model) inspectorWidth() int { return shared.Max(1, m.width-m.listWidth()-2) }

// listRows is how many repositories the list shows, below its column
// headings and above its status line.
func (m *Model) listRows() int { return shared.Max(1, m.height-2) }

// listColumns decides which optional columns fit beside a usable name column.
func (m *Model) listColumns() (nameWidth int, language, stars, pushed bool) {
	nameWidth = m.listWidth() - 4 // selection marker and visibility dot
	fits := func(w int) bool {
		if nameWidth-w >= 16 {
			nameWidth -= w
			return true
		}
		return false
	}
	pushed = fits(pushedWidth)
	stars = fits(starsWidth)
	language = fits(languageWidth)
	return nameWidth, language, stars, pushed
}

// listView lays the repository list beside the inspector for the selected
// repository.
func (m *Model) listView() string {
	width, rows := m.listWidth(), m.listRows()
	nameWidth, language, stars, pushed := m.listColumns()

	heading := "    " + shared.Fit("NAME", nameWidth)
	if language {
		heading += shared.Fit("LANGUAGE", languageWidth)
	}
	if stars {
		heading += shared.FitRight("★", starsWidth-1) + " "
	}
	if pushed {
		heading += shared.FitRight("PUSHED", pushedWidth)
	}
	lines := []string{shared.ColumnHeading.Render(shared.Fit(heading, width))}

	for i := m.listOffset; i < len(m.visible) && i < m.listOffset+rows; i++ {
		lines = append(lines, m.listRow(m.visible[i], i == m.cursor, nameWidth, language, stars, pushed))
	}
	for len(lines) < rows+1 {
		lines = append(lines, "")
	}
	lines = append(lines, m.listStatus())

	list := shared.Lines(lines, width, m.height)
	scrollbar := strings.Join(scrollbar(m.listOffset, rows, len(m.visible)), "\n")

	m.repoModel.SetDimensions(m.inspectorWidth(), m.height)
	inspector := fmt.Sprint(m.repoModel.View().Content)

	return lipgloss.JoinHorizontal(lipgloss.Top, list, scrollbar, " ", inspector)
}

func (m *Model) listRow(c repo.RepoConfig, selected bool, nameWidth int, language, stars, pushed bool) string {
	marker, name := "  ", shared.Fit(c.Name, nameWidth)
	switch {
	case selected && m.inspecting:
		marker = shared.DimStyle.Render("▌ ")
		name = shared.StrongStyle.Render(name)
	case selected:
		marker = shared.AccentStyle.Render("▌ ")
		name = shared.StrongStyle.Render(name)
	}

	row := marker + repo.VisibilityMark(c) + " " + name
	if language {
		row += shared.DimStyle.Render(shared.Fit(c.Text("Primary Language"), languageWidth))
	}
	if stars {
		row += shared.StarStyle.Render(shared.FitRight(repo.CompactNumber(c.Int("Stargazer Count")), starsWidth-1)) + " "
	}
	if pushed {
		row += shared.FitRight(repo.FormatCell(c.Properties["Pushed At"]), pushedWidth)
	}
	return row
}

// listStatus shows where the selection is and explains the visibility dots.
// How many repos the filters hide is in the frame's status.
func (m *Model) listStatus() string {
	position := fmt.Sprintf("%d/%d", m.cursor+1, len(m.visible))
	return shared.AccentStyle.Render(position) + "  " + repo.VisibilityLegend()
}

// scrollbar draws a one-column track the height of the window with a thumb
// sized and placed to show which part of total the window covers. A heading
// line above and a status line below keep it level with the list rows.
func scrollbar(offset, rows, total int) []string {
	track := shared.DimStyle.Render("│")
	bar := []string{track}
	thumbSize, thumbStart := 0, 0
	if total > rows {
		thumbSize = shared.Max(1, rows*rows/total)
		thumbStart = (rows - thumbSize) * offset / (total - rows)
	}
	for i := 0; i < rows; i++ {
		if i >= thumbStart && i < thumbStart+thumbSize {
			bar = append(bar, shared.AccentStyle.Render("┃"))
		} else {
			bar = append(bar, track)
		}
	}
	return append(bar, track)
}
