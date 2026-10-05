package repos

import (
	"fmt"
	"strings"

	"gh-reponark/inspector"
	"gh-reponark/repo"
	"gh-reponark/ui"

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
func (m *Model) listWidth() int      { return ui.Half(m.width) }
func (m *Model) inspectorWidth() int { return ui.Max(1, m.width-m.listWidth()-2) }

// listRows is how many repositories the list shows, below its column
// headings and above its status line.
func (m *Model) listRows() int { return ui.Max(1, m.height-2) }

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

	heading := "    " + ui.Fit("NAME", nameWidth)
	if language {
		heading += ui.Fit("LANGUAGE", languageWidth)
	}
	if stars {
		heading += ui.FitRight("★", starsWidth-1) + " "
	}
	if pushed {
		heading += ui.FitRight("PUSHED", pushedWidth)
	}
	lines := []string{ui.ColumnHeading.Render(ui.Fit(heading, width))}

	from, to := m.list.Visible()
	for i := from; i < to; i++ {
		lines = append(lines, m.listRow(m.visible[i], i == m.list.Index, nameWidth, language, stars, pushed))
	}
	for len(lines) < rows+1 {
		lines = append(lines, "")
	}
	lines = append(lines, m.listStatus())

	list := ui.Lines(lines, width, m.height)
	scrollbar := strings.Join(scrollbar(from, rows, len(m.visible)), "\n")

	m.repoModel.SetDimensions(m.inspectorWidth(), m.height)
	inspector := fmt.Sprint(m.repoModel.View().Content)

	return lipgloss.JoinHorizontal(lipgloss.Top, list, scrollbar, " ", inspector)
}

func (m *Model) listRow(c repo.RepoConfig, selected bool, nameWidth int, language, stars, pushed bool) string {
	marker, name := "  ", ui.Fit(c.Name, nameWidth)
	switch {
	case selected && m.inspecting:
		marker = ui.Marker(ui.AppColors.Dim)
		name = ui.StrongStyle.Render(name)
	case selected:
		marker = ui.Marker(ui.AppColors.Accent)
		name = ui.StrongStyle.Render(name)
	}

	row := marker + inspector.VisibilityMark(c) + " " + name
	if language {
		row += ui.DimStyle.Render(ui.Fit(c.Text("Primary Language"), languageWidth))
	}
	if stars {
		row += ui.StarStyle.Render(ui.FitRight(repo.CompactNumber(c.Int("Stargazer Count")), starsWidth-1)) + " "
	}
	if pushed {
		row += ui.FitRight(inspector.FormatCell(c.Properties["Pushed At"]), pushedWidth)
	}
	if selected {
		return ui.HighlightRow(row, m.listWidth(), !m.inspecting)
	}
	return row
}

// listStatus shows where the selection is and explains the visibility dots.
// How many repos the filters hide is in the frame's status.
func (m *Model) listStatus() string {
	position := fmt.Sprintf("%d/%d", m.list.Index+1, len(m.visible))
	return ui.AccentStyle.Render(position) + "  " + inspector.VisibilityLegend()
}

// scrollbar draws a one-column track the height of the window with a thumb
// sized and placed to show which part of total the window covers. A heading
// line above and a status line below keep it level with the list rows.
func scrollbar(offset, rows, total int) []string {
	track := ui.DimStyle.Render("│")
	bar := []string{track}
	thumbSize, thumbStart := 0, 0
	if total > rows {
		thumbSize = ui.Max(1, rows*rows/total)
		thumbStart = (rows - thumbSize) * offset / (total - rows)
	}
	for i := 0; i < rows; i++ {
		if i >= thumbStart && i < thumbStart+thumbSize {
			bar = append(bar, ui.AccentStyle.Render("┃"))
		} else {
			bar = append(bar, track)
		}
	}
	return append(bar, track)
}
