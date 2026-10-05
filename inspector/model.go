package inspector

import (
	"strings"

	"gh-reponark/repo"
	"gh-reponark/ui"

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
	repository repo.RepoConfig
	repos      []repo.RepoConfig
	// stats is every property summarised across repos, computed once when
	// they load; the bars and tiles are scaled against it.
	stats     map[string]Stats
	activeTab int
	// property is the focused property within the active group and the
	// rows of properties in view.
	property ui.Cursor
	focused  bool
	width    int
	height   int
	keymap   KeyMap
}

func NewModel(width, height int) Model {
	m := Model{
		repository: repo.RepoConfig{Properties: map[string]repo.RepoProperty{}, PropertyGroups: map[string][]repo.RepoProperty{}},
		width:      width,
		height:     height,
		keymap:     NewRepoKeyMap(),
	}
	m.SelectProperty(0)
	return m
}

func (m *Model) SetDimensions(width, height int) {
	m.width = width
	m.height = height
	m.SelectProperty(m.property.Index)
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
func (m *Model) SelectRepo(repository repo.RepoConfig) {
	m.repository = repository
}

// SetOrgRepos tells the inspector about every loaded repository, so counts
// can be drawn against the largest in the organization and the focused
// property put in context ("on in 14 of 40 repos").
func (m *Model) SetOrgRepos(repos []repo.RepoConfig) {
	m.repos = repos
	m.stats = OrgStats(repos)
}

// SelectTab switches to the group at index, wrapping around at either end,
// and focuses its first property.
func (m *Model) SelectTab(index int) {
	count := len(repo.Groups())
	m.activeTab = ((index % count) + count) % count
	m.SelectProperty(0)
}

// SelectProperty focuses the property at index within the active group,
// stopping at either end, and scrolls it into view.
func (m *Model) SelectProperty(index int) {
	m.property.Resize(m.propertyRows(), len(m.ActiveGroup().Properties))
	m.property.Set(index)
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
	return ui.Max(1, m.height-m.chrome())
}

// ActiveTab is the index of the group being shown.
func (m Model) ActiveTab() int { return m.activeTab }

// ActiveProperty is the index of the focused property within the active group.
func (m Model) ActiveProperty() int { return m.property.Index }

// ActiveGroup is the group being shown.
func (m Model) ActiveGroup() repo.Group { return repo.Groups()[m.activeTab] }

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

func (m Model) View() tea.View {
	width := ui.Max(1, m.width)
	height := ui.Max(1, m.height)
	group := m.ActiveGroup()
	l := m.layout()
	rows := m.propertyRows()
	offset, end := m.property.Visible()

	if l.plain {
		lines := []string{m.titleLine(width), RenderSegments(repo.GroupTitles(), width, m.activeTab, m.focused)}
		for _, p := range group.Properties[offset:end] {
			lines = append(lines, ui.Fit(p.Name+" "+FormatValue(m.repository.Properties[p.Name]), width))
		}
		return tea.NewView(ui.Lines(lines, width, height))
	}

	lines := m.header(l, width)
	if l.tiles > 0 {
		lines = append(lines, RenderStats(m.repository, m.stats, width, l.tiles)...)
	}
	lines = append(lines, "")
	lines = append(lines, RenderSegments(repo.GroupTitles(), width, m.activeTab, m.focused))

	// The properties in view, one per line, padded to the full row budget.
	nameWidth := m.nameWidth(group)
	body := make([]string, 0, rows)
	for i := offset; i < end; i++ {
		body = append(body, m.propertyRow(group.Properties[i], i == m.property.Index, nameWidth, width, l.bars))
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
	return tea.NewView(ui.Lines(lines, width, height))
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
	name := ui.StrongStyle.Render(m.repository.Name)
	badges, _ := ui.JoinFit(BadgeList(m.repository), " ", width-lipgloss.Width(name)-2)
	if badges == "" {
		return ui.Fit(name, width)
	}
	return ui.Fit(name+"  "+badges, width)
}

// summary is the repository's description wrapped to at most n lines, the
// last ellipsised if it runs on, or a note that it has none.
func (m Model) summary(width, n int) []string {
	var lines []string
	if description := m.repository.Text("Description"); description == "" {
		lines = []string{ui.DimStyle.Italic(true).Render("No description")}
	} else {
		wrapped := ui.Wrap(ui.TextBodyStyle, description, width)
		for i := 0; i < n && i < len(wrapped); i++ {
			line := wrapped[i]
			if i == n-1 && len(wrapped) > n {
				// The rest runs on past the last line, so it is cut with an ellipsis.
				line = strings.Join(wrapped[i:], " ")
			}
			lines = append(lines, ui.Fit(line, width))
		}
	}
	for len(lines) < n {
		lines = append(lines, "")
	}
	return lines
}

// nameWidth fits the longest property name in group, up to half the pane.
func (m Model) nameWidth(group repo.Group) int {
	longest := 0
	for _, p := range group.Properties {
		longest = ui.Max(longest, lipgloss.Width(p.Name))
	}
	return ui.Min(longest, ui.Half(m.width))
}

// propertyRow is one property as a line: a marker, the name and the value,
// with the focused property as a full-width band. Counts get a bar against
// the organization when bars is set and the organization is known.
func (m Model) propertyRow(p repo.PropertySchema, active bool, nameWidth, width int, bars bool) string {
	marker, name := "  ", ui.TextBodyStyle.Render(ui.Fit(p.Name, nameWidth))
	switch {
	case active && m.focused:
		marker = ui.AccentStyle.Render("▸ ")
		name = ui.AccentStyle.Bold(true).Render(ui.Fit(p.Name, nameWidth))
	case active:
		// Show where focus will land without competing with the repo list.
		marker = ui.DimStyle.Render("▸ ")
		name = ui.StrongStyle.Render(ui.Fit(p.Name, nameWidth))
	}
	max := 0
	if bars {
		max = m.stats[p.Name].Max
	}
	row := marker + name + "  " + FormatRow(m.repository.Properties[p.Name], max)
	if active {
		return ui.HighlightRow(row, width, m.focused)
	}
	return ui.Fit(row, width)
}
