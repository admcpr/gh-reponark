package repo

import (
	"fmt"
	"strings"

	"gh-reponark/shared"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Model is the inspector for one repository: a header with the name, badge
// pills, description, key facts and posture chips; a strip of stat tiles
// scaled against the rest of the organization; a segmented control for the
// property groups; the active group as rows; and a card describing the
// focused property with its place in the organization. The parent screen moves between properties
// with SelectProperty while the pane is focused.
type Model struct {
	repository RepoConfig
	repos      []RepoConfig
	// stats is every property summarised across repos, computed once when
	// they load; the bars and tiles are scaled against it.
	stats          map[string]Stats
	activeTab      int
	activeProperty int
	propertyOffset int
	focused        bool
	width          int
	height         int
	keymap         KeyMap
}

func NewModel(width, height int) Model {
	return Model{
		repository: RepoConfig{Properties: map[string]RepoProperty{}, PropertyGroups: map[string][]RepoProperty{}},
		width:      width,
		height:     height,
		keymap:     NewRepoKeyMap(),
	}
}

func (m *Model) SetDimensions(width, height int) {
	m.width = width
	m.height = height
	m.SelectProperty(m.activeProperty)
}

func (m Model) Init() tea.Cmd {
	return nil
}

// Keys returns the bindings this pane responds to, so a parent screen can
// forward exactly those keys and advertise them in its own help.
func (m Model) Keys() KeyMap {
	return m.keymap
}

// SelectRepo shows repository, keeping the current tab and property so the
// same setting can be compared from one repository to the next.
func (m *Model) SelectRepo(repository RepoConfig) {
	m.repository = repository
}

// SetOrgRepos tells the inspector about every loaded repository, so counts
// can be drawn against the largest in the organization and the focused
// property put in context ("on in 14 of 40 repos").
func (m *Model) SetOrgRepos(repos []RepoConfig) {
	m.repos = repos
	m.stats = OrgStats(repos)
}

// SelectTab switches to the group at index, wrapping around at either end,
// and focuses its first property.
func (m *Model) SelectTab(index int) {
	count := len(Groups())
	m.activeTab = ((index % count) + count) % count
	m.activeProperty = 0
	m.propertyOffset = 0
}

// SelectProperty focuses the property at index within the active group.
func (m *Model) SelectProperty(index int) {
	total := len(m.ActiveGroup().Properties)
	m.activeProperty = shared.Max(0, shared.Min(index, total-1))
	m.propertyOffset = shared.ScrollOffset(m.propertyOffset, m.activeProperty, m.propertyRows(), total)
}

// SetFocused says whether keys are moving between this pane's properties,
// which decides how strongly the focused property is highlighted.
func (m *Model) SetFocused(focused bool) { m.focused = focused }

// Focused reports whether the pane has focus.
func (m Model) Focused() bool { return m.focused }

// PageSize is how many properties the grid shows at once.
func (m Model) PageSize() int { return m.propertyRows() }

// propertyRows is how many lines the property area has.
func (m Model) propertyRows() int {
	return shared.Max(1, m.height-m.chrome())
}

// ActiveTab is the index of the group being shown.
func (m Model) ActiveTab() int { return m.activeTab }

// ActiveProperty is the index of the focused property within the active group.
func (m Model) ActiveProperty() int { return m.activeProperty }

// ActiveGroup is the group being shown.
func (m Model) ActiveGroup() Group { return Groups()[m.activeTab] }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(msg, m.keymap.NextTab):
			m.SelectTab(m.activeTab + 1)
		case key.Matches(msg, m.keymap.PrevTab):
			m.SelectTab(m.activeTab - 1)
		}
	}
	return m, nil
}

// Size tiers. Width decides what the header carries; height decides how the
// footer is drawn and how much of the header survives.
const (
	// minMediumWidth is the full layout: header, tiles, bars and the footer
	// card. Below it the pane is a plain list with a compact header. SIZE
	// joins the tiles from sizeTileWidth.
	minMediumWidth = 50
	// oneRowChipWidth is where all six chips fit a single row even with
	// their longest labels, so the second row is not reserved.
	oneRowChipWidth = 76
	// minListWidth is the narrowest pane with a name column; below it each
	// property is one plain line.
	minListWidth = 12
	// minCardHeight is the shortest pane that draws the footer card; below
	// it the footer is a rule and two plain description lines.
	minCardHeight = 24
	// tallHeight is where the description gets a second line.
	tallHeight = 40
	// minRows is how many property rows the header gives way to keep.
	minRows = 12

	footerCardLines  = 6
	footerPlainLines = 3
	segmentLines     = 1
	gapLines         = 1
)

// layout is what the pane draws at its current size.
type layout struct {
	// plain is the sub-minListWidth pane: the name, the groups and one line
	// per property, nothing else.
	plain       bool
	description int
	facts       bool
	chipRows    int
	// tiles is how many stat tiles the strip shows; zero for no strip.
	tiles int
	// bars draws a bar beside each count; card draws the footer card.
	bars bool
	card bool
}

// headerLines is how many lines the header takes.
func (l layout) headerLines() int {
	return 1 + l.description + boolInt(l.facts) + l.chipRows
}

// footerLines is how many lines sit below the property rows.
func (l layout) footerLines() int {
	switch {
	case l.plain:
		return 0
	case l.card:
		return footerCardLines
	default:
		return footerPlainLines
	}
}

// chrome is how many lines the layout spends around the property rows.
func (l layout) chrome() int {
	if l.plain {
		return 1 + segmentLines
	}
	return l.headerLines() + boolInt(l.tiles > 0)*statsLines + gapLines + segmentLines + l.footerLines()
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// layout decides what fits at the pane's size. Width sets the tier; then
// the header sheds, in order, the description, the second chip row, the
// tiles, the facts and the chips until at least minRows property rows are
// left.
func (m Model) layout() layout {
	w, h := m.width, m.height
	switch {
	case w < minListWidth:
		return layout{plain: true}
	case w < minMediumWidth:
		return layout{description: 1}
	}
	l := layout{description: 1, facts: true, chipRows: 1, tiles: tileCount(w), bars: true, card: h >= minCardHeight}
	if h >= tallHeight {
		l.description = 2
	}
	if w < oneRowChipWidth {
		l.chipRows = 2
	}
	for h-l.chrome() < minRows {
		switch {
		case l.description > 0:
			l.description--
		case l.chipRows > 1:
			l.chipRows = 1
		case l.tiles > 0:
			l.tiles = 0
		case l.facts:
			l.facts = false
		case l.chipRows > 0:
			l.chipRows = 0
		default:
			return l
		}
	}
	return l
}

// chrome is the number of lines around the property rows at this size.
func (m Model) chrome() int { return m.layout().chrome() }

func (m Model) View() tea.View {
	width := shared.Max(1, m.width)
	height := shared.Max(1, m.height)
	group := m.ActiveGroup()
	l := m.layout()
	rows := shared.Max(1, m.height-l.chrome())
	// The pane may have been resized since the focus last moved.
	offset := shared.ScrollOffset(m.propertyOffset, m.activeProperty, rows, len(group.Properties))
	end := shared.Min(len(group.Properties), offset+rows)

	if l.plain {
		lines := []string{m.titleLine(width), RenderSegments(GroupTitles(), width, m.activeTab, m.focused)}
		for _, p := range group.Properties[offset:end] {
			lines = append(lines, shared.Fit(p.Name+" "+FormatValue(m.repository.Properties[p.Name]), width))
		}
		return tea.NewView(shared.Lines(lines, width, height))
	}

	lines := m.header(l, width)
	if l.tiles > 0 {
		lines = append(lines, RenderStats(m.repository, m.stats, width, l.tiles)...)
	}
	lines = append(lines, "")
	lines = append(lines, RenderSegments(GroupTitles(), width, m.activeTab, m.focused))

	// The properties in view, one per line, padded to the full row budget.
	nameWidth := m.nameWidth(group)
	body := make([]string, 0, rows)
	for i := offset; i < end; i++ {
		body = append(body, m.propertyRow(group.Properties[i], i == m.activeProperty, nameWidth, width, l.bars))
	}
	for len(body) < rows {
		body = append(body, "")
	}
	lines = append(lines, body...)

	above, below := offset, len(group.Properties)-end
	if l.card {
		lines = append(lines, m.footerCard(group, width, above, below)...)
	} else {
		lines = append(lines, m.footerPlain(group, width, above, below)...)
	}
	return tea.NewView(shared.Lines(lines, width, height))
}

// header is the name and badges, the description, the key facts and the
// posture chips, as many of each as the layout allows.
func (m Model) header(l layout, width int) []string {
	lines := []string{m.titleLine(width)}
	if l.description > 0 {
		lines = append(lines, m.summary(width, l.description)...)
	}
	if l.facts {
		lines = append(lines, KeyFacts(m.repository, width))
	}
	if l.chipRows > 0 {
		lines = append(lines, PostureRows(Posture(m.repository), width, l.chipRows)...)
	}
	return lines
}

// titleLine is the repository's name with its badge pills. Pills are shed
// from the least important end rather than cut, and only the name itself is
// ever truncated.
func (m Model) titleLine(width int) string {
	name := shared.StrongStyle.Render(m.repository.Name)
	badges, _ := shared.JoinFit(BadgeList(m.repository), " ", width-lipgloss.Width(name)-2)
	if badges == "" {
		return shared.Fit(name, width)
	}
	return shared.Fit(name+"  "+badges, width)
}

// summary is the repository's description wrapped to at most n lines, the
// last ellipsised if it runs on, or a note that it has none.
func (m Model) summary(width, n int) []string {
	var lines []string
	if description := m.repository.Text("Description"); description == "" {
		lines = []string{shared.DimStyle.Italic(true).Render("No description")}
	} else {
		wrapped := shared.Wrap(shared.TextBodyStyle, description, width)
		for i := 0; i < n && i < len(wrapped); i++ {
			line := wrapped[i]
			if i == n-1 && len(wrapped) > n {
				// The rest runs on past the last line, so it is cut with an ellipsis.
				line = strings.Join(wrapped[i:], " ")
			}
			lines = append(lines, shared.Fit(line, width))
		}
	}
	for len(lines) < n {
		lines = append(lines, "")
	}
	return lines
}

// nameWidth fits the longest property name in group, up to half the pane.
func (m Model) nameWidth(group Group) int {
	longest := 0
	for _, p := range group.Properties {
		longest = shared.Max(longest, lipgloss.Width(p.Name))
	}
	return shared.Min(longest, shared.Half(m.width))
}

// propertyRow is one property as a line: a marker, the name and the value,
// with the focused property as a full-width band. Counts get a bar against
// the organization when bars is set and the organization is known.
func (m Model) propertyRow(p PropertySchema, active bool, nameWidth, width int, bars bool) string {
	marker, name := "  ", shared.TextBodyStyle.Render(shared.Fit(p.Name, nameWidth))
	switch {
	case active && m.focused:
		marker = shared.AccentStyle.Render("▸ ")
		name = shared.AccentStyle.Bold(true).Render(shared.Fit(p.Name, nameWidth))
	case active:
		// Show where focus will land without competing with the repo list.
		marker = shared.DimStyle.Render("▸ ")
		name = shared.StrongStyle.Render(shared.Fit(p.Name, nameWidth))
	}
	max := 0
	if bars {
		max = m.stats[p.Name].Max
	}
	row := marker + name + "  " + FormatRow(m.repository.Properties[p.Name], max)
	if active {
		return shared.HighlightRow(row, width, m.focused)
	}
	return shared.Fit(row, width)
}

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
func (m Model) footerCard(group Group, width, above, below int) []string {
	style := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1).
		BorderForegroundBlend(shared.FrameGradient(m.focused)...)
	inner := shared.Max(1, width-style.GetHorizontalFrameSize())
	rows := []string{"", "", "", ""}
	if m.activeProperty < len(group.Properties) {
		p := group.Properties[m.activeProperty]
		right := fmt.Sprintf("%d/%d", m.activeProperty+1, len(group.Properties))
		if hint := scrollHint(above, below); hint != "" {
			right = hint + "  " + right
		}
		title := shared.StrongStyle.Render(p.Name) + " " + TypePill(p.Type)
		rows[0] = title + shared.FitRight(shared.DimStyle.Render(right), shared.Max(0, inner-lipgloss.Width(title)))
		wrapped := shared.Wrap(shared.TextBodyStyle, p.Description, inner)
		copy(rows[1:3], wrapped[:shared.Min(2, len(wrapped))])
		rows[3] = m.context(p, inner)
	}
	for i := range rows {
		rows[i] = shared.Fit(rows[i], inner)
	}
	return strings.Split(style.Render(strings.Join(rows, "\n")), "\n")
}

// context is the focused property's place in the organization: a sentence
// such as "on in 14 of 40 repos" with a bar for the share, or "more than 39
// of 40 repos · max 12k" with a histogram of the distribution. It is empty
// until the organization has loaded.
func (m Model) context(p PropertySchema, width int) string {
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
	line := shared.DimStyle.Render(text)
	if figure != "" && lipgloss.Width(text)+2+lipgloss.Width(figure) <= width {
		line += "  " + figure
	}
	return line
}

// footerPlain is the compact footer: a rule carrying the scroll hint, then
// the focused property's description.
func (m Model) footerPlain(group Group, width, above, below int) []string {
	rule := strings.Repeat("─", width)
	if hint := scrollHint(above, below); hint != "" && width > lipgloss.Width(hint)+4 {
		rule = strings.Repeat("─", width-lipgloss.Width(hint)-2) + " " + hint + " "
	}
	lines := []string{shared.DimStyle.Render(rule)}
	if m.activeProperty < len(group.Properties) {
		wrapped := shared.Wrap(shared.DimStyle, group.Properties[m.activeProperty].Description, width)
		lines = append(lines, wrapped[:shared.Min(2, len(wrapped))]...)
	}
	return lines
}

// typeLabel names a property's Go type the way a reader would.
func typeLabel(t string) string {
	switch t {
	case "bool":
		return "toggle"
	case "int":
		return "count"
	case "time.Time":
		return "date"
	default:
		return "text"
	}
}

// TypePill is the quiet capsule every screen tags a property's type with:
// the type label in dim text on the surface tint.
func TypePill(t string) string {
	return shared.GhostPill(typeLabel(t))
}
