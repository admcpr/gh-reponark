package filters

import (
	"fmt"
	"gh-reponark/repo"
	"gh-reponark/shared"
	"sort"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type AddFilterMsg Filter
type FiltersMsg FilterMap

// OpenFiltersMsg asks the application to show the filter screen, seeded with
// the filters that are currently applied.
type OpenFiltersMsg struct{ Filters FilterMap }

type Model struct {
	filterSearch tea.Model
	filtersList  list.Model
	help         help.Model
	keymap       filterKeyMap
	filters      FilterMap
	width        int
	height       int
}

func (m *Model) SetDimensions(width, height int) {
	m.width = width
	m.height = height
	m.help.SetWidth(width)
}

// NewModel creates the filter screen. current holds the filters already
// applied; they are copied so edits only reach the caller via FiltersMsg.
func NewModel(current FilterMap, width, height int) *Model {
	fsm := NewFilterSearchModel()
	list := list.New([]list.Item{}, shared.SimpleItemDelegate{}, width, height-4)

	help := shared.NewHelpModel(width)
	keymap := newFilterKeyMap()

	selected := make(FilterMap, len(current))
	for name, filter := range current {
		selected[name] = filter
	}

	return &Model{
		filterSearch: fsm,
		filtersList:  list,
		help:         help,
		keymap:       keymap,
		filters:      selected,
		width:        width,
		height:       height,
	}
}

func (m *Model) Init() tea.Cmd {
	return m.filterSearch.Init()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, m.keymap.Back) {
			return m, func() tea.Msg {
				return shared.PreviousMsg{Message: FiltersMsg(m.filters)}
			}
		}

	case AddFilterMsg:
		m.filters[msg.Name()] = Filter(msg)
		m.filterSearch = NewFilterSearchModel()
		cmd := m.filterSearch.Init()
		return m, cmd
	}

	m.filterSearch, cmd = m.filterSearch.Update(msg)

	return m, cmd
}

// NewFilterModel returns the editor for a property, or nil when the property
// type cannot be filtered on.
func NewFilterModel(property repo.PropertySchema, width, height int) tea.Model {
	switch property.Type {
	case "bool":
		return NewBoolModel(property.Name, false, width, height)
	case "int":
		return NewIntModel(property.Name, 0, 100000, width, height)
	case "time.Time":
		return NewDateModel(property.Name, time.Time{}, time.Now(), width, height)
	case "string":
		return NewStringModel(property.Name, "", width, height)
	default:
		return nil
	}
}

func isSupportedPropertyType(t string) bool {
	switch t {
	case "bool", "int", "time.Time", "string":
		return true
	default:
		return false
	}
}

func (m Model) View() tea.View {
	m.filtersList = NewFiltersList(m.filters, m.width, m.height)
	filtersListView := m.filtersList.View()

	search := fmt.Sprint(m.filterSearch.View().Content)
	return tea.NewView(fmt.Sprint(lipgloss.JoinVertical(lipgloss.Left, search, filtersListView)))
	// }
}

func (m Model) HeaderView() tea.View {
	return tea.NewView(shared.TitleStyle.Render("Filters"))
}

func (m Model) HelpView() tea.View {
	return tea.NewView(m.help.View(m.keymap))
}

func NewFiltersList(filters map[string]Filter, width, height int) list.Model {
	items := make([]list.Item, len(filters))
	i := 0
	for _, filter := range filters {
		items[i] = shared.SimpleItem(filter.Name())
		i++
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].(shared.SimpleItem) < items[j].(shared.SimpleItem)
	})

	list := list.New(items, shared.SimpleItemDelegate{}, width, height-8)
	list.Styles.Title = shared.TitleStyle
	list.Title = "Selected Filters"
	list.SetShowHelp(false)
	list.SetShowStatusBar(false)
	list.SetShowTitle(true)

	return list
}
