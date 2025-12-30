package repo

import (
	"fmt"
	"gh-reponark/shared"
	"sort"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Model struct {
	repoHeader   HeaderModel
	repository   RepoConfig
	settingsList list.Model
	activeTab    int
	width        int
	height       int
	help         help.Model
	keymap       KeyMap
}

func NewModel(width, height int) Model {
	return Model{
		repoHeader: NewHeaderModel(width, []string{}, 0),
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
	m.repoHeader.SetDimensions(width, height)
	m.help.SetWidth(width)
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) SelectRepo(repository RepoConfig) {
	m.repository = repository
	key := m.repository.GroupKeys[m.activeTab]
	m.repoHeader = NewHeaderModel(m.width, m.repository.GroupKeys, m.activeTab)
	m.settingsList = NewSettingsList(m.repository.PropertyGroups[key], key, m.width, m.height)
}

func (m *Model) SelectTab(index int) {
	m.activeTab = index
	key := m.repository.GroupKeys[index]
	m.settingsList = NewSettingsList(m.repository.PropertyGroups[key], key, m.width, m.height)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "tab":
			if m.activeTab < len(m.repository.PropertyGroups)-1 {
				m.SelectTab(m.activeTab + 1)
			} else {
				m.SelectTab(0)
			}
			repoHeader, _ := m.repoHeader.Update(TabSelectMessage{Index: m.activeTab})
			m.repoHeader = repoHeader.(HeaderModel)
		case "shift+tab":
			if m.activeTab > 0 {
				m.SelectTab(m.activeTab - 1)
			} else {
				m.SelectTab(len(m.repository.PropertyGroups) - 1)
			}
			repoHeader, _ := m.repoHeader.Update(TabSelectMessage{Index: m.activeTab})
			m.repoHeader = repoHeader.(HeaderModel)
		}
	}
	return m, cmd
}

func (m Model) View() tea.View {
	// frameWidth, frameHeight := style.Settings.GetFrameSize()
	// settings := appStyle.
	// 	Width(m.width).
	// 	Render(m.settingsTable.View())

	settings := m.settingsList.View()

	return tea.NewView(fmt.Sprint(lipgloss.JoinVertical(lipgloss.Left, settings)))
}

func (m Model) HeaderView() tea.View {
	return m.repoHeader.View()
}

func (m Model) HelpView() tea.View {
	return tea.NewView(m.help.View(m.keymap))
}

func NewSettingsList(activeSettings []RepoProperty, title string, width, height int) list.Model {
	sort.Slice(activeSettings, func(i, j int) bool {
		return activeSettings[i].Name < activeSettings[j].Name
	})

	items := make([]list.Item, len(activeSettings))

	for i, setting := range activeSettings {
		items[i] = shared.NewListItem(setting.Name, setting.String())
	}

	delegate := shared.DefaultDelegate
	delegate.Styles.SelectedDesc = delegate.Styles.NormalDesc
	delegate.Styles.SelectedTitle = delegate.Styles.NormalTitle
	delegate.SetSpacing(0)

	list := list.New(items, delegate, width, height)
	list.Title = title
	list.SetShowFilter(false)
	list.SetShowHelp(false)
	list.SetShowPagination(false)
	list.SetShowStatusBar(false)
	list.Styles.Title = shared.TitleStyle
	list.SetHeight(height)

	return list
}
