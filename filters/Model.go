package filters

import (
	"gh-reponark/repo"
	"gh-reponark/shared"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/v2/help"
	"github.com/charmbracelet/bubbles/v2/key"
	"github.com/charmbracelet/bubbles/v2/list"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type AddFilterMsg Filter
type FiltersMsg FilterMap

type Model struct {
	filterSearch tea.Model
	filtersList  list.Model
	repository   repo.Repository
	help         help.Model
	keymap       filterKeyMap
	properties   map[string]Property
	filters      FilterMap
	width        int
	height       int
}

func (m *Model) SetDimensions(width, height int) {
	m.width = width
	m.height = height
}

type Property struct {
	Name        string
	Description string
	Type        string
}

func NewModel(modelData interface{}, width, height int) *Model {
	fsm := NewFilterSearchModel()
	list := list.New([]list.Item{}, shared.SimpleItemDelegate{}, width, height-4)
	repository := repo.Repository{}

	help := help.New()
	keymap := filterKeyMap{}

	return &Model{
		filterSearch: fsm,
		filtersList:  list,
		repository:   repository,
		help:         help,
		keymap:       keymap,
		properties:   make(map[string]Property),
		filters:      make(map[string]Filter),
		width:        width,
		height:       height,
	}
}

func (m Model) Init() (tea.Model, tea.Cmd) {
	_, cmd := m.filterSearch.Init()
	return m, cmd
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "ctrl+enter":
			return m, func() tea.Msg {
				return shared.PreviousMsg{Message: FiltersMsg(m.filters)}
			}
		}

	case AddFilterMsg:
		m.filters[msg.GetName()] = Filter(msg)
		m.filterSearch = NewFilterSearchModel()
		m.filterSearch, cmd = m.filterSearch.Init()
		return m, cmd
	}

	m.filterSearch, cmd = m.filterSearch.Update(msg)

	return m, cmd
}

func NewFilterModel(modelData interface{}, width, height int) tea.Model {
	property := modelData.(Property)

	switch property.Type {
	case "bool":
		return NewBoolModel(property.Name, false, width, height)
	case "int":
		return NewIntModel(property.Name, 0, 100000, width, height)
	case "time.Time":
		return NewDateModel(property.Name, time.Time{}, time.Now(), width, height)
	default:
		return nil
	}
}

func (m Model) View() string {
	m.filtersList = NewFiltersList(m.filters, m.width, m.height)
	filtersListView := m.filtersList.View()

	search := m.filterSearch.View()
	help := m.help.View(m.keymap)
	return lipgloss.JoinVertical(lipgloss.Left, filtersListView, search, help)
	// }
}

type filtersListMsg repo.RepoConfig

func NewFiltersList(filters map[string]Filter, width, height int) list.Model {
	items := make([]list.Item, len(filters))
	i := 0
	for _, filter := range filters {
		items[i] = shared.SimpleItem(filter.GetName())
		i++
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].(shared.SimpleItem) < items[j].(shared.SimpleItem)
	})

	list := list.New(items, shared.SimpleItemDelegate{}, width, height-8)
	list.Styles.Title = shared.TitleStyle
	list.Title = "Filters"
	list.SetShowHelp(false)
	list.SetShowStatusBar(false)
	list.SetShowTitle(true)

	return list
}

type filterKeyMap struct{}

func (k filterKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "complete")),
		key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "next suggestion")),
		key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "prev suggestion")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	}
}
func (k filterKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}
