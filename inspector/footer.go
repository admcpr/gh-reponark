package inspector

import (
	"fmt"
	"strings"

	"gh-reponark/repo"
	"gh-reponark/ui"

	"charm.land/lipgloss/v2"
)

// scrollHint says how many properties are scrolled out of view on either
// side, e.g. "▴ 1 · 6 ▾", or nothing when all are shown.
func scrollHint(above, below int) string {
	var hints []string
	if above > 0 {
		hints = append(hints, fmt.Sprintf("▴ %d", above))
	}
	if below > 0 {
		hints = append(hints, fmt.Sprintf("%d ▾", below))
	}
	return strings.Join(hints, " · ")
}

// footerCard describes the focused property in a bordered card: its name
// and type, its place in the group, its description and where this
// repository's value sits in the organization. The gradient runs around the
// card either way; it is bright while the pane has focus and sinks towards
// the background when it does not.
func (m Model) footerCard(group repo.Group, width, above, below int) []string {
	style := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1).
		BorderForegroundBlend(ui.FrameGradient(m.focused)...)
	inner := ui.Max(1, width-style.GetHorizontalFrameSize())
	rows := []string{"", "", "", ""}
	if m.property.Index < len(group.Properties) {
		p := group.Properties[m.property.Index]
		right := fmt.Sprintf("%d/%d", m.property.Index+1, len(group.Properties))
		if hint := scrollHint(above, below); hint != "" {
			right = hint + "  " + right
		}
		title := ui.StrongStyle.Render(p.Name) + " " + TypePill(p.Type)
		rows[0] = title + ui.FitRight(ui.DimStyle.Render(right), ui.Max(0, inner-lipgloss.Width(title)))
		wrapped := ui.Wrap(ui.TextBodyStyle, p.Description, inner)
		copy(rows[1:3], wrapped[:ui.Min(2, len(wrapped))])
		rows[3] = m.context(p, inner)
	}
	for i := range rows {
		rows[i] = ui.Fit(rows[i], inner)
	}
	return strings.Split(style.Render(strings.Join(rows, "\n")), "\n")
}

// context is the focused property's place in the organization: a sentence
// such as "on in 14 of 40 repos" with a bar for the share, or "more than 39
// of 40 repos · max 12k" with a histogram of the distribution. It is empty
// until the organization has loaded.
func (m Model) context(p repo.PropertySchema, width int) string {
	prop := m.repository.Properties[p.Name]
	stats := StatsFor(m.stats, p.Name, m.repos, m.repository)
	text := stats.Context(prop)
	if text == "" {
		return ""
	}
	var figure string
	switch value := prop.Value.(type) {
	case bool:
		figure = FractionBar(stats.On, stats.Total, barWidth)
	case int:
		figure = stats.Histogram(value)
	}
	line := ui.DimStyle.Render(text)
	if figure != "" && lipgloss.Width(text)+2+lipgloss.Width(figure) <= width {
		line += "  " + figure
	}
	return line
}

// footerPlain is the compact footer: a rule carrying the scroll hint, then
// the focused property's description.
func (m Model) footerPlain(group repo.Group, width, above, below int) []string {
	rule := strings.Repeat("─", width)
	if hint := scrollHint(above, below); hint != "" && width > lipgloss.Width(hint)+4 {
		rule = strings.Repeat("─", width-lipgloss.Width(hint)-2) + " " + hint + " "
	}
	lines := []string{ui.DimStyle.Render(rule)}
	if m.property.Index < len(group.Properties) {
		wrapped := ui.Wrap(ui.DimStyle, group.Properties[m.property.Index].Description, width)
		lines = append(lines, wrapped[:ui.Min(2, len(wrapped))]...)
	}
	return lines
}
