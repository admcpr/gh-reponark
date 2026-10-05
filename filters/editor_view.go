package filters

import (
	"fmt"

	"gh-reponark/inspector"
	"gh-reponark/repo"
	"gh-reponark/ui"
)

// editorWidth leaves a gap either side of the divider and one before the
// frame, so rules and counts do not run into the border.
func (m *Model) editorWidth() int { return ui.Max(1, m.width-m.listWidth()-4) }

func (m *Model) editorLines() []string {
	p, ok := m.selected()
	if !ok {
		return nil
	}
	width := m.editorWidth()
	kind := inspector.TypeOf(p.Type)

	lines := []string{
		kind.Glyph() + " " + ui.StrongStyle.Render(p.Name) + " " + inspector.TypePill(p.Type),
		ui.DimStyle.Render("in " + repo.GroupTitle(p.Group)),
	}
	lines = append(lines, ui.Wrap(ui.TextBodyStyle, p.Description, width)...)

	lines = append(lines, "", heading("Filter", "", width))
	lines = append(lines, m.editor.View(m.editing, width)...)

	if len(m.repos) == 0 {
		return lines
	}

	lines = append(lines, "", heading("Across your repos", "", width))
	lines = append(lines, m.chart(p, width)...)

	matching := m.filters.FilterRepos(m.repos)
	lines = append(lines, "", heading("Matching", fmt.Sprintf("%d of %d", len(matching), len(m.repos)), width))
	room := m.height - len(lines)
	return append(lines, matchingNames(matching, width, room)...)
}
