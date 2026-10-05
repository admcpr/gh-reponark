package filters

import (
	"fmt"
	"strings"

	"gh-reponark/filter"
	"gh-reponark/repo"
	"gh-reponark/ui"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// FiltersMsg carries the filters chosen on the filter screen back to the
// repository browser.
type FiltersMsg filter.FilterMap

// OpenFiltersMsg asks the application to show the filter screen, seeded with
// the filters that are currently applied and the repositories they apply to.
type OpenFiltersMsg struct {
	Filters filter.FilterMap
	Repos   []repo.RepoConfig
}

// Model is the filter screen: every filterable property down the left, with
// the filter on it, and an editor for the highlighted one on the right.
// Edits apply as they are made so the match count stays current.
type Model struct {
	filters    filter.FilterMap
	repos      []repo.RepoConfig
	properties []repo.PropertySchema

	matches []int // indexes into properties that match the search
	// cursor is the highlighted match. The list scrolls by rendered line,
	// headings included, rather than by match, so the cursor has no window
	// of its own and the first visible line is kept in offset.
	cursor ui.Cursor
	offset int

	search    textinput.Model
	searching bool

	editor  editor
	editing bool
	before  filter.Filter // the filter when editing began, restored on cancel

	keymap filterKeyMap
	width  int
	height int
}

// NewModel creates the filter screen. current holds the filters already
// applied; they are copied so edits only reach the caller via FiltersMsg.
// repos are used to count matches and describe the values on offer.
func NewModel(current filter.FilterMap, repos []repo.RepoConfig, width, height int) *Model {
	selected := make(filter.FilterMap, len(current))
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
	styles.Focused.Prompt = ui.AccentStyle
	styles.Blurred.Prompt = ui.AccentStyle
	search.SetStyles(styles)

	m := &Model{
		filters:    selected,
		repos:      repos,
		properties: properties,
		search:     search,
		keymap:     newFilterKeyMap(),
		width:      width,
		height:     height,
	}
	m.refreshMatches()
	return m
}

func (m *Model) SetDimensions(width, height int) {
	m.width = width
	m.height = height
}

func (m *Model) Init() tea.Cmd {
	return nil
}

// Filters returns the filters as currently edited.
func (m *Model) Filters() filter.FilterMap {
	return m.filters
}

// selected returns the highlighted property, if any match the search.
func (m *Model) selected() (repo.PropertySchema, bool) {
	if m.cursor.Index >= len(m.matches) {
		return repo.PropertySchema{}, false
	}
	return m.properties[m.matches[m.cursor.Index]], true
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
	m.cursor.Resize(0, len(m.matches))
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

// moveCursor highlights the match delta places away, stopping at either end.
func (m *Model) moveCursor(delta int) {
	m.cursor.Move(delta)
	m.selectionChanged()
}

// setFilter applies f to the highlighted property; nil removes its filter.
func (m *Model) setFilter(f filter.Filter) {
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
		case key.Matches(msg, m.keymap.EndSearch):
			m.searching = false
			m.search.Blur()
			m.refreshMatches()
			return nil
		}
	}
	var cmd tea.Cmd
	m.search, cmd = m.search.Update(msg)
	m.cursor.Set(0)
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
			return ui.PreviousMsg{Message: FiltersMsg(m.filters)}
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
		m.filters = filter.FilterMap{}
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

func (m *Model) View() tea.View {
	list := ui.Lines(m.listLines(), m.listWidth(), m.height)
	rule := ui.DimStyle.Render("│")
	divider := strings.TrimSuffix(strings.Repeat(rule+"\n", m.height), "\n")
	editor := ui.Lines(m.editorLines(), m.editorWidth(), m.height)
	gap := strings.TrimSuffix(strings.Repeat(" \n", m.height), "\n")
	return tea.NewView(lipgloss.JoinHorizontal(lipgloss.Top, list, gap, divider, gap, editor, gap))
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

// Typing reports whether a text field has focus: the search, or a number,
// date or text editor. The yes/no editor takes single keys, not text.
func (m Model) Typing() bool {
	if m.searching {
		return true
	}
	_, choosing := m.editor.(*boolEditor)
	return m.editing && !choosing
}

// Help lists the keys for whatever has focus.
func (m Model) Help() ui.Help {
	finish := []key.Binding{m.keymap.Done, m.keymap.Cancel}
	switch {
	case m.searching:
		return ui.Help{Short: []key.Binding{
			ui.Hint("type", "to search"),
			m.keymap.EndSearch,
			ui.WithHelp(m.keymap.Cancel, "clear"),
		}}
	case m.editing:
		keys := append(m.editor.Keys(), finish...)
		var full [][]key.Binding
		if e, ok := m.editor.(*boolEditor); ok {
			full = [][]key.Binding{e.FullKeys(), finish}
		}
		return ui.Help{Short: keys, Full: full}
	}

	short := []key.Binding{
		ui.Combine("j/k", "property", m.keymap.Down, m.keymap.Up),
		m.keymap.Edit,
	}
	if _, ok := m.editor.(*boolEditor); ok {
		short = append(short, ui.WithHelp(m.keymap.Toggle, "toggle"))
	}
	short = append(short, m.keymap.Search, m.keymap.Clear, m.keymap.Back)

	return ui.Help{
		Short: short,
		Full: [][]key.Binding{
			{
				ui.WithHelp(m.keymap.Up, "property up"),
				ui.WithHelp(m.keymap.Down, "property down"),
				m.keymap.Top,
				m.keymap.Bottom,
				m.keymap.Search,
			},
			{m.keymap.Edit, m.keymap.Toggle, m.keymap.Clear, m.keymap.ClearAll},
			{m.keymap.Back},
		},
	}
}
