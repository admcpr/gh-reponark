package repo

import (
	"strings"

	"gh-reponark/shared"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Model is the inspector for one repository: a tab per property group, a
// grid of the group's properties and the description of the focused one.
// The parent screen moves between properties with SelectProperty while the
// pane is focused.
type Model struct {
	repository     RepoConfig
	activeTab      int
	activeProperty int
	propertyOffset int
	focused        bool
	width          int
	height         int
	help           help.Model
	keymap         KeyMap
}

func NewModel(width, height int) Model {
	return Model{
		repository: RepoConfig{Properties: map[string]RepoProperty{}, PropertyGroups: map[string][]RepoProperty{}},
		width:      width,
		height:     height,
		help:       shared.NewHelpModel(width),
		keymap:     NewRepoKeyMap(),
	}
}

func (m *Model) SetDimensions(width, height int) {
	m.width = width
	m.height = height
	m.help.SetWidth(width)
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

// propertyRows is how many properties fit in the grid.
func (m Model) propertyRows() int {
	return shared.Max(1, m.height-inspectorChrome)
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

// inspectorChrome is the number of lines around the property grid: name,
// description, tabs and a rule above it; a rule and two description lines below.
const inspectorChrome = 7

func (m Model) View() tea.View {
	width := shared.Max(1, m.width)
	group := m.ActiveGroup()

	lines := []string{
		shared.StrongStyle.Render(m.repository.Name) + "  " + Badges(m.repository),
		m.summary(),
		RenderTabs(GroupTitles(), width, m.activeTab),
		shared.DimStyle.Render(strings.Repeat("─", width)),
	}

	rows := m.propertyRows()
	// The pane may have been resized since the focus last moved.
	offset := shared.ScrollOffset(m.propertyOffset, m.activeProperty, rows, len(group.Properties))
	nameWidth := m.nameWidth(group)
	for i := offset; i < len(group.Properties) && i < offset+rows; i++ {
		lines = append(lines, m.propertyRow(group.Properties[i], i == m.activeProperty, nameWidth, width))
	}
	for len(lines) < rows+4 {
		lines = append(lines, "")
	}

	lines = append(lines, shared.DimStyle.Render(strings.Repeat("─", width)))
	if m.activeProperty < len(group.Properties) {
		description := group.Properties[m.activeProperty].Description
		wrapped := lipgloss.NewStyle().Width(width).Render(description)
		for _, line := range strings.Split(wrapped, "\n")[:shared.Min(2, strings.Count(wrapped, "\n")+1)] {
			lines = append(lines, shared.DimStyle.Render(line))
		}
	}

	return tea.NewView(shared.Lines(lines, width, shared.Max(1, m.height)))
}

// summary is the repository's description, or a note that it has none.
func (m Model) summary() string {
	if description := m.repository.Text("Description"); description != "" {
		return shared.DimStyle.Render(description)
	}
	return shared.DimStyle.Italic(true).Render("No description")
}

// nameWidth fits the longest property name in group, up to half the pane.
func (m Model) nameWidth(group Group) int {
	longest := 0
	for _, p := range group.Properties {
		longest = shared.Max(longest, lipgloss.Width(p.Name))
	}
	return shared.Min(longest, shared.Half(m.width))
}

func (m Model) propertyRow(p PropertySchema, active bool, nameWidth, width int) string {
	marker, name := "  ", shared.Fit(p.Name, nameWidth)
	switch {
	case active && m.focused:
		marker = shared.AccentStyle.Render("▸ ")
		name = shared.AccentStyle.Bold(true).Render(name)
	case active:
		// Show where focus will land without competing with the repo list.
		marker = shared.DimStyle.Render("▸ ")
	}
	value := FormatValue(m.repository.Properties[p.Name])
	return shared.Fit(marker+name+"  "+value, width)
}

func (m Model) HelpView() tea.View {
	return tea.NewView(m.help.View(m.keymap))
}
