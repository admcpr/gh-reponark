package filters

import (
	"fmt"
	"image/color"
	"sort"
	"strings"
	"time"

	"gh-reponark/repo"
	"gh-reponark/shared"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// FiltersMsg carries the filters chosen on the filter screen back to the
// repository browser.
type FiltersMsg FilterMap

// OpenFiltersMsg asks the application to show the filter screen, seeded with
// the filters that are currently applied and the repositories they apply to.
type OpenFiltersMsg struct {
	Filters FilterMap
	Repos   []repo.RepoConfig
}

// Model is the filter screen: every filterable property down the left, with
// the filter on it, and an editor for the highlighted one on the right.
// Edits apply as they are made so the match count stays current.
type Model struct {
	filters    FilterMap
	repos      []repo.RepoConfig
	properties []repo.PropertySchema

	matches []int // indexes into properties that match the search
	cursor  int   // index into matches
	offset  int   // first visible line of the property list

	search    textinput.Model
	searching bool

	editor  editor
	editing bool
	before  Filter // the filter when editing began, restored on cancel

	keymap filterKeyMap
	help   help.Model
	width  int
	height int
}

// NewModel creates the filter screen. current holds the filters already
// applied; they are copied so edits only reach the caller via FiltersMsg.
// repos are used to count matches and describe the values on offer.
func NewModel(current FilterMap, repos []repo.RepoConfig, width, height int) *Model {
	selected := make(FilterMap, len(current))
	for name, filter := range current {
		selected[name] = filter
	}

	var properties []repo.PropertySchema
	for _, group := range repo.Groups() {
		for _, p := range group.Properties {
			if isSupportedPropertyType(p.Type) {
				properties = append(properties, p)
			}
		}
	}

	search := newInput("search properties", "")
	search.Prompt = "/ "
	styles := search.Styles()
	styles.Focused.Prompt = shared.AccentStyle
	styles.Blurred.Prompt = shared.AccentStyle
	search.SetStyles(styles)

	m := &Model{
		filters:    selected,
		repos:      repos,
		properties: properties,
		search:     search,
		keymap:     newFilterKeyMap(),
		help:       shared.NewHelpModel(width),
		width:      width,
		height:     height,
	}
	m.refreshMatches()
	return m
}

func (m *Model) SetDimensions(width, height int) {
	m.width = width
	m.height = height
	m.help.SetWidth(width)
}

func (m *Model) Init() tea.Cmd {
	return nil
}

// Filters returns the filters as currently edited.
func (m *Model) Filters() FilterMap {
	return m.filters
}

// selected returns the highlighted property, if any match the search.
func (m *Model) selected() (repo.PropertySchema, bool) {
	if m.cursor < 0 || m.cursor >= len(m.matches) {
		return repo.PropertySchema{}, false
	}
	return m.properties[m.matches[m.cursor]], true
}

// refreshMatches reapplies the search and keeps the highlight in range.
func (m *Model) refreshMatches() {
	query := strings.ToLower(strings.TrimSpace(m.search.Value()))
	m.matches = m.matches[:0]
	for i, p := range m.properties {
		haystack := strings.ToLower(p.Name + " " + repo.GroupTitle(p.Group))
		if strings.Contains(haystack, query) {
			m.matches = append(m.matches, i)
		}
	}
	m.cursor = shared.Max(0, shared.Min(m.cursor, len(m.matches)-1))
	m.selectionChanged()
}

// selectionChanged points the editor at the highlighted property.
func (m *Model) selectionChanged() {
	if p, ok := m.selected(); ok {
		m.editor = newEditor(p, m.filters[p.Name])
	} else {
		m.editor = nil
	}
}

func (m *Model) moveCursor(delta int) {
	if len(m.matches) == 0 {
		return
	}
	m.cursor = shared.Max(0, shared.Min(m.cursor+delta, len(m.matches)-1))
	m.selectionChanged()
}

// setFilter applies f to the highlighted property; nil removes its filter.
func (m *Model) setFilter(f Filter) {
	p, ok := m.selected()
	if !ok {
		return
	}
	if f == nil {
		delete(m.filters, p.Name)
	} else {
		m.filters[p.Name] = f
	}
}

// applyEditor applies whatever the editor describes, unless it cannot be
// read yet, in which case the last good filter stays in place.
func (m *Model) applyEditor() {
	if f, err := m.editor.Filter(); err == nil {
		m.setFilter(f)
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, isKey := msg.(tea.KeyPressMsg)
	switch {
	case m.searching:
		return m, m.updateSearch(msg)
	case m.editing:
		return m, m.updateEditor(msg)
	case isKey:
		return m, m.updateList(keyMsg)
	}
	return m, nil
}

func (m *Model) updateSearch(msg tea.Msg) tea.Cmd {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(msg, m.keymap.Cancel):
			m.search.SetValue("")
			fallthrough
		case key.Matches(msg, m.keymap.Done, m.keymap.Down):
			m.searching = false
			m.search.Blur()
			m.refreshMatches()
			return nil
		}
	}
	var cmd tea.Cmd
	m.search, cmd = m.search.Update(msg)
	m.cursor = 0
	m.refreshMatches()
	return cmd
}

func (m *Model) updateEditor(msg tea.Msg) tea.Cmd {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(msg, m.keymap.Done):
			m.stopEditing()
			return nil
		case key.Matches(msg, m.keymap.Cancel):
			m.setFilter(m.before)
			m.stopEditing()
			m.selectionChanged()
			return nil
		}
	}
	cmd := m.editor.Update(msg)
	m.applyEditor()
	return cmd
}

func (m *Model) stopEditing() {
	m.editing = false
	m.editor.Blur()
}

func (m *Model) updateList(msg tea.KeyPressMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keymap.Back):
		return func() tea.Msg {
			return shared.PreviousMsg{Message: FiltersMsg(m.filters)}
		}
	case key.Matches(msg, m.keymap.Up):
		m.moveCursor(-1)
	case key.Matches(msg, m.keymap.Down):
		m.moveCursor(1)
	case key.Matches(msg, m.keymap.Top):
		m.moveCursor(-len(m.matches))
	case key.Matches(msg, m.keymap.Bottom):
		m.moveCursor(len(m.matches))
	case key.Matches(msg, m.keymap.Search):
		m.searching = true
		return m.search.Focus()
	case key.Matches(msg, m.keymap.ClearAll):
		m.filters = FilterMap{}
		m.selectionChanged()
	case m.editor == nil:
		return nil
	case key.Matches(msg, m.keymap.Edit):
		p, _ := m.selected()
		m.before = m.filters[p.Name]
		m.editing = true
		return m.editor.Focus()
	case key.Matches(msg, m.keymap.Toggle):
		if e, ok := m.editor.(*boolEditor); ok {
			e.Cycle()
			m.applyEditor()
		}
	case key.Matches(msg, m.keymap.Clear):
		m.setFilter(nil)
		m.selectionChanged()
	}
	return nil
}

// ---- view

// propertyType is how each kind of property is marked: a glyph beside it in
// the list and a coloured label in the editor.
type propertyType struct {
	glyph string
	label string
	color color.Color
}

func typeOf(t string) propertyType {
	switch t {
	case "bool":
		return propertyType{"●", "yes / no", shared.AppColors.Green}
	case "int":
		return propertyType{"#", "number", shared.AppColors.BrightYellow}
	case "time.Time":
		return propertyType{"◆", "date", shared.AppColors.BrightPurple}
	default:
		return propertyType{"¶", "text", shared.AppColors.BrightBlue}
	}
}

func (t propertyType) Glyph() string {
	return lipgloss.NewStyle().Foreground(t.color).Render(t.glyph)
}

// listChrome is the number of lines above the property list: the search
// line and the active filters.
const listChrome = 2

func (m *Model) listWidth() int {
	return shared.Max(30, shared.Min(44, m.width*2/5))
}

// editorWidth leaves a gap either side of the divider and one before the
// frame, so rules and counts do not run into the border.
func (m *Model) editorWidth() int { return shared.Max(1, m.width-m.listWidth()-4) }

func (m *Model) View() tea.View {
	list := shared.Lines(m.listLines(), m.listWidth(), m.height)
	rule := lipgloss.NewStyle().Foreground(shared.AppColors.Blue).Render("│")
	divider := strings.TrimSuffix(strings.Repeat(rule+"\n", m.height), "\n")
	editor := shared.Lines(m.editorLines(), m.editorWidth(), m.height)
	gap := strings.TrimSuffix(strings.Repeat(" \n", m.height), "\n")
	return tea.NewView(lipgloss.JoinHorizontal(lipgloss.Top, list, gap, divider, gap, editor, gap))
}

func (m *Model) listLines() []string {
	width := m.listWidth()

	var search string
	switch {
	case m.searching:
		search = m.search.View()
	case m.search.Value() != "":
		search = shared.AccentStyle.Render("/ ") + shared.ValueStyle.Render(m.search.Value())
	default:
		search = shared.AccentStyle.Render("/") + shared.DimStyle.Render(" search properties")
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
			rows = append(rows, heading(repo.GroupTitle(group), "", width))
		}
		if i == m.cursor {
			cursorLine = len(rows)
		}
		rows = append(rows, m.propertyRow(p, i == m.cursor, width))
	}
	if len(rows) == 0 {
		rows = append(rows, shared.TextBodyStyle.Render("No properties match your search"))
	}

	visible := shared.Max(1, m.height-listChrome)
	// Keep the group heading above the first property in view.
	target := cursorLine
	if m.cursor == 0 {
		target = 0
	}
	m.offset = shared.ScrollOffset(m.offset, target, visible, len(rows))
	m.offset = shared.ScrollOffset(m.offset, cursorLine, visible, len(rows))
	end := shared.Min(len(rows), m.offset+visible)
	return append(lines, rows[m.offset:end]...)
}

// heading is a section title followed by a rule to the edge, with optional
// detail on the right, e.g. "Matching ──────── 3 of 9".
func heading(title, detail string, width int) string {
	left := shared.HeadingStyle.Render(title) + " "
	right := ""
	if detail != "" {
		right = " " + shared.AccentStyle.Render(detail)
	}
	fill := shared.Max(0, width-lipgloss.Width(left)-lipgloss.Width(right))
	return left + lipgloss.NewStyle().Foreground(shared.AppColors.Blue).Render(strings.Repeat("─", fill)) + right
}

func (m *Model) propertyRow(p repo.PropertySchema, highlighted bool, width int) string {
	f, active := m.filters[p.Name]

	marker := "  "
	if highlighted {
		marker = shared.AccentStyle.Render("▌ ")
		if m.searching || m.editing {
			marker = shared.DimStyle.Render("▌ ")
		}
	}
	name := shared.TextBodyStyle
	if highlighted || active {
		name = shared.StrongStyle
	}

	// The condition of an active filter sits at the right edge.
	condition := ""
	if active {
		condition = f.Condition()
		if lipgloss.Width(condition) > shared.Half(width) {
			condition = shared.Fit(condition, shared.Half(width))
		}
	}
	nameWidth := width - 4 - lipgloss.Width(condition) - 1
	return marker + typeOf(p.Type).Glyph() + " " + name.Render(shared.Fit(p.Name, nameWidth)) + " " + shared.AccentStyle.Render(condition)
}

// chips shows each active filter as a chip, with a count of any that do not
// fit on the line.
func (m *Model) chips(width int) string {
	if len(m.filters) == 0 {
		return shared.DimStyle.Render("No filters yet: every repo is shown")
	}
	names := make([]string, 0, len(m.filters))
	for name := range m.filters {
		names = append(names, name)
	}
	sort.Strings(names)

	line := ""
	for i, name := range names {
		chip := shared.ChipStyle.Render(" " + name + ": " + m.filters[name].Condition() + " ")
		more := ""
		if rest := len(names) - i - 1; rest > 0 {
			more = shared.AccentStyle.Render(fmt.Sprintf(" +%d", rest))
		}
		if lipgloss.Width(line)+lipgloss.Width(chip)+lipgloss.Width(more) > width {
			if line == "" {
				return shared.Fit(chip, width)
			}
			return line + shared.AccentStyle.Render(fmt.Sprintf("+%d", len(names)-i))
		}
		line += chip + " "
	}
	return line
}

func (m *Model) editorLines() []string {
	p, ok := m.selected()
	if !ok {
		return nil
	}
	width := m.editorWidth()
	kind := typeOf(p.Type)

	lines := []string{
		kind.Glyph() + " " + shared.StrongStyle.Render(p.Name) + "  " +
			shared.PillStyle(kind.color).Render(" "+kind.label+" "),
		lipgloss.NewStyle().Foreground(shared.AppColors.Blue).Render("in " + repo.GroupTitle(p.Group)),
	}
	lines = append(lines, wrap(shared.TextBodyStyle, p.Description, width)...)

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

// wrap renders text in style, wrapped to width, one string per line.
func wrap(style lipgloss.Style, text string, width int) []string {
	wrapped := lipgloss.NewStyle().Width(width).Render(text)
	lines := strings.Split(wrapped, "\n")
	for i, line := range lines {
		lines[i] = style.Render(line)
	}
	return lines
}

// chart shows how the loaded repositories' values of p are spread, so a
// filter can be chosen without guessing.
func (m *Model) chart(p repo.PropertySchema, width int) []string {
	switch p.Type {
	case "bool":
		return boolChart(m.repos, p.Name, width)
	case "int":
		values := make([]float64, len(m.repos))
		for i, c := range m.repos {
			values[i] = float64(c.Int(p.Name))
		}
		return rangeChart(values, width, shared.StarStyle, func(v float64) string { return fmt.Sprint(int(v)) })
	case "time.Time":
		var values []float64
		for _, c := range m.repos {
			if t := c.Time(p.Name); !t.IsZero() {
				values = append(values, float64(t.Unix()))
			}
		}
		if len(values) == 0 {
			return []string{shared.TextBodyStyle.Render("No repos have this date set")}
		}
		lines := rangeChart(values, width, lipgloss.NewStyle().Foreground(shared.AppColors.BrightPurple),
			func(v float64) string { return time.Unix(int64(v), 0).UTC().Format("2006-01-02") })
		if missing := len(m.repos) - len(values); missing > 0 {
			lines = append(lines, shared.DimStyle.Render(fmt.Sprintf("%d with no date", missing)))
		}
		return lines
	default:
		return textChart(m.repos, p.Name, width)
	}
}

// boolChart is one bar split into the repos with the property on and off.
func boolChart(repos []repo.RepoConfig, name string, width int) []string {
	yes := 0
	for _, c := range repos {
		if c.Bool(name) {
			yes++
		}
	}
	on, off := blocks(float64(yes)/float64(len(repos)), shared.Min(width, 40))
	return []string{
		shared.GoodStyle.Render(on) + shared.BadStyle.Render(off),
		shared.GoodStyle.Render("● yes ") + shared.ValueStyle.Render(fmt.Sprint(yes)) + "   " +
			shared.BadStyle.Render("● no ") + shared.ValueStyle.Render(fmt.Sprint(len(repos)-yes)),
	}
}

// rangeChart is a sparkline of how values spread between the smallest and
// largest, labelled at both ends, with the median below.
func rangeChart(values []float64, width int, style lipgloss.Style, format func(float64) string) []string {
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	lo, hi, median := sorted[0], sorted[len(sorted)-1], sorted[len(sorted)/2]

	span := shared.Min(width, 40)
	labels := shared.Fit(shared.ValueStyle.Render(format(lo)), span/2) + shared.FitRight(shared.ValueStyle.Render(format(hi)), span-span/2)
	return []string{
		style.Render(sparkline(histogram(values, span))),
		labels,
		shared.TextBodyStyle.Render("median ") + shared.ValueStyle.Render(format(median)),
	}
}

// textChart bars the most common values of a text property.
func textChart(repos []repo.RepoConfig, name string, width int) []string {
	counts := map[string]int{}
	for _, c := range repos {
		counts[c.Text(name)]++
	}
	if counts[""] == len(repos) {
		return []string{shared.TextBodyStyle.Render("No repos have a value set")}
	}
	if len(counts) == len(repos) && len(repos) > 1 {
		return []string{shared.TextBodyStyle.Render("Every repo has a different value")}
	}

	values := make([]string, 0, len(counts))
	for v := range counts {
		values = append(values, v)
	}
	sort.Slice(values, func(i, j int) bool {
		if counts[values[i]] != counts[values[j]] {
			return counts[values[i]] > counts[values[j]]
		}
		return values[i] < values[j]
	})
	if len(values) > 5 {
		values = values[:5]
	}

	labelWidth := shared.Min(16, width/3)
	barWidth := shared.Max(1, shared.Min(24, width-labelWidth-5))
	bar := lipgloss.NewStyle().Foreground(shared.AppColors.BrightBlue)
	lines := make([]string, len(values))
	for i, v := range values {
		label := shared.ValueStyle.Render(shared.Fit(v, labelWidth))
		if v == "" {
			label = shared.DimStyle.Render(shared.Fit("(none)", labelWidth))
		}
		filled, _ := blocks(float64(counts[v])/float64(counts[values[0]]), barWidth)
		lines[i] = label + " " + bar.Render(filled) + " " + shared.ValueStyle.Render(fmt.Sprint(counts[v]))
	}
	return lines
}

// matchingNames lists repos a line at a time, each with its visibility dot,
// using at most room lines and ending with a count of any left over.
func matchingNames(repos []repo.RepoConfig, width, room int) []string {
	if len(repos) == 0 {
		return []string{shared.WarnStyle.Render("No repos match these filters")}
	}
	var lines []string
	line := ""
	for i, c := range repos {
		entry := repo.VisibilityMark(c) + " " + shared.ValueStyle.Render(c.Name)
		if line != "" && lipgloss.Width(line)+2+lipgloss.Width(entry) > width {
			if len(lines) == room-1 {
				return append(lines, line+shared.AccentStyle.Render(fmt.Sprintf("  +%d more", len(repos)-i)))
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

func (m Model) Breadcrumb() string {
	return "Filters"
}

// Status counts how many repositories the filters let through.
func (m Model) Status() string {
	if m.repos == nil {
		return fmt.Sprintf("%d active", len(m.filters))
	}
	return fmt.Sprintf("%d of %d repos match", len(m.filters.FilterRepos(m.repos)), len(m.repos))
}

func (m Model) HelpView() tea.View {
	return tea.NewView(m.help.View(m.helpKeys()))
}

// helpKeys lists the bindings for whatever has focus.
func (m Model) helpKeys() shared.KeyBindings {
	switch {
	case m.searching:
		return shared.KeyBindings{
			key.NewBinding(key.WithHelp("type", "to search")),
			m.keymap.Done,
			withHelp(m.keymap.Cancel, "clear"),
		}
	case m.editing:
		keys := shared.KeyBindings(m.editor.Keys())
		return append(keys, m.keymap.Done, m.keymap.Cancel)
	}
	keys := shared.KeyBindings{
		key.NewBinding(key.WithKeys("j", "k"), key.WithHelp("j/k", "move")),
		m.keymap.Edit,
	}
	if _, ok := m.editor.(*boolEditor); ok {
		keys = append(keys, m.keymap.Toggle)
	}
	return append(keys, m.keymap.Search, m.keymap.Clear, m.keymap.ClearAll, m.keymap.Back)
}

// withHelp returns a copy of binding with a different description.
func withHelp(binding key.Binding, desc string) key.Binding {
	binding.SetHelp(binding.Help().Key, desc)
	return binding
}
