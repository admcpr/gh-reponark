package org

import (
	"fmt"
	"sort"

	"gh-reponark/filters"
	"gh-reponark/github"
	"gh-reponark/repo"
	"gh-reponark/shared"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// maxConcurrentBatches bounds how many configuration requests are in flight
// at once, so a large organization does not trip GitHub's secondary rate
// limits while still loading several batches in parallel.
const maxConcurrentBatches = 4

// repositoryPageMsg carries one page of repository names.
type repositoryPageMsg github.RepositoryPage

// repositoryBatchMsg carries the configuration of one batch of repositories.
type repositoryBatchMsg []repo.Repository

type Model struct {
	svc github.Service

	Title     string
	repoCount int
	repos     []repo.RepoConfig
	filters   filters.FilterMap
	isUser    bool

	// Loading state: names collected from the listing, batches of names not
	// yet requested, and how many batch requests are outstanding.
	names    []string
	pending  [][]string
	inFlight int

	help      help.Model
	keymap    orgKeyMap
	repoList  list.Model
	repoModel repo.Model

	width  int
	height int

	progress progress.Model
}

func NewModel(svc github.Service, orgKey shared.OrgKey, width, height int) *Model {
	help := shared.NewHelpModel(width)
	keymap := newOrgKeyMap()

	m := &Model{
		svc:       svc,
		Title:     orgKey.Name,
		isUser:    orgKey.IsUser,
		width:     width,
		height:    height,
		help:      help,
		keymap:    keymap,
		repoModel: repo.NewModel(width/2, height),
		progress:  progress.New(progress.WithoutPercentage()),
	}
	m.repoList = m.newRepoList(nil)

	return m
}

// newRepoList builds the repository list with the shared key bindings.
func (m *Model) newRepoList(items []list.Item) list.Model {
	repoList := list.New(items, shared.SimpleItemDelegate{}, m.width/2, m.height-2)
	repoList.Title = fmt.Sprintf("Organization: %s ", m.Title)
	repoList.Styles.Title = shared.TitleStyle
	repoList.SetStatusBarItemName("Repository", "Repositories")
	repoList.SetShowHelp(false)
	repoList.SetShowTitle(true)
	// The list moves with the same bindings the help footer advertises.
	repoList.KeyMap.CursorUp = m.keymap.Up
	repoList.KeyMap.CursorDown = m.keymap.Down
	return repoList
}

func (m *Model) SetDimensions(width, height int) {
	m.width = width
	m.height = height
	m.help.SetWidth(width)
}

// repoItem is a list entry that carries the repository it stands for, so the
// selection never has to be mapped back to an index in another slice.
type repoItem struct {
	config repo.RepoConfig
}

func (i repoItem) FilterValue() string { return "" }
func (i repoItem) String() string      { return i.config.Name }

// populateRepoList rebuilds the list from the repositories that pass the
// current filters and points the detail pane at the first of them.
func (m *Model) populateRepoList() {
	filteredRepositories := m.filters.FilterRepos(m.repos)
	items := make([]list.Item, len(filteredRepositories))
	for i, config := range filteredRepositories {
		items[i] = repoItem{config: config}
	}

	m.repoList = m.newRepoList(items)
	if selected, ok := m.selectedRepo(); ok {
		m.repoModel.SelectRepo(selected)
	}
}

// selectedRepo returns the repository highlighted in the list, if any.
func (m *Model) selectedRepo() (repo.RepoConfig, bool) {
	item, ok := m.repoList.SelectedItem().(repoItem)
	if !ok {
		return repo.RepoConfig{}, false
	}
	return item.config, true
}

func (m *Model) Init() tea.Cmd {
	return m.loadRepositoryPage("")
}

// loadRepositoryPage returns a command that fetches one page of repository
// names, starting after the given cursor.
func (m *Model) loadRepositoryPage(after string) tea.Cmd {
	return func() tea.Msg {
		page, err := m.svc.ListRepositories(m.Title, m.isUser, after)
		if err != nil {
			return shared.ErrorMsg{Err: err}
		}
		return repositoryPageMsg(page)
	}
}

// loadRepositoryBatch returns a command that fetches the configuration of
// one batch of repositories.
func (m *Model) loadRepositoryBatch(names []string) tea.Cmd {
	return func() tea.Msg {
		repositories, err := m.svc.GetRepositories(m.Title, names)
		if err != nil {
			return shared.ErrorMsg{Err: err}
		}
		return repositoryBatchMsg(repositories)
	}
}

// startBatches requests pending batches until maxConcurrentBatches are in
// flight and returns the commands that run them.
func (m *Model) startBatches() []tea.Cmd {
	var cmds []tea.Cmd
	for len(m.pending) > 0 && m.inFlight < maxConcurrentBatches {
		batch := m.pending[0]
		m.pending = m.pending[1:]
		m.inFlight++
		cmds = append(cmds, m.loadRepositoryBatch(batch))
	}
	return cmds
}

// batchNames splits names into batches of at most github.RepositoryBatchSize.
func batchNames(names []string) [][]string {
	var batches [][]string
	for start := 0; start < len(names); start += github.RepositoryBatchSize {
		end := start + github.RepositoryBatchSize
		if end > len(names) {
			end = len(names)
		}
		batches = append(batches, names[start:end])
	}
	return batches
}

// loadedFraction is how much of the listing has arrived, for the progress bar.
func (m *Model) loadedFraction() float64 {
	if m.repoCount <= 0 {
		return 1.0
	}
	return float64(len(m.repos)) / float64(m.repoCount)
}

// finishLoading sorts the repositories, shows them and completes the progress bar.
func (m *Model) finishLoading() tea.Cmd {
	sort.Slice(m.repos, func(i, j int) bool {
		return m.repos[i].Name < m.repos[j].Name
	})
	if len(m.repos) > 0 {
		m.populateRepoList()
	}
	return m.progress.SetPercent(1.0)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case repositoryPageMsg:
		m.repoCount = msg.TotalCount
		for _, ref := range msg.Repositories {
			m.names = append(m.names, ref.Name)
		}

		if msg.HasNextPage {
			return m, m.loadRepositoryPage(msg.EndCursor)
		}

		// The listing is complete; the names are the authoritative count.
		m.repoCount = len(m.names)
		if m.repoCount == 0 {
			return m, m.finishLoading()
		}
		m.pending = batchNames(m.names)
		return m, tea.Batch(m.startBatches()...)

	case repositoryBatchMsg:
		m.inFlight--
		for _, repository := range msg {
			m.repos = append(m.repos, repo.NewRepoConfig(repository))
		}

		if len(m.repos) >= m.repoCount && len(m.pending) == 0 && m.inFlight == 0 {
			return m, m.finishLoading()
		}

		cmds := append([]tea.Cmd{m.progress.SetPercent(m.loadedFraction())}, m.startBatches()...)
		return m, tea.Batch(cmds...)

	case filters.FiltersMsg:
		m.filters = filters.FilterMap(msg)
		m.populateRepoList()
		return m, nil

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel
		return m, cmd

	case tea.KeyPressMsg:
		repoKeys := m.repoModel.Keys()
		switch {
		case key.Matches(msg, m.keymap.Filters):
			return m, func() tea.Msg {
				return filters.OpenFiltersMsg{Filters: m.filters}
			}
		case key.Matches(msg, m.keymap.Back):
			return m, func() tea.Msg {
				return shared.PreviousMsg{}
			}
		case key.Matches(msg, repoKeys.NextTab, repoKeys.PrevTab):
			repoModel, cmd := m.repoModel.Update(msg)
			m.repoModel = repoModel.(repo.Model)
			return m, cmd
		default:
			m.repoList, cmd = m.repoList.Update(msg)
			return m, cmd
		}
	default:
		m.repoList, cmd = m.repoList.Update(msg)
	}

	return m, cmd
}

func (m *Model) View() tea.View {
	if m.progress.Percent() < 1 {
		return m.ProgressView()
	}

	selected, ok := m.selectedRepo()
	if !ok {
		repoList := shared.AppStyle.Width(shared.Half(m.width)).Render(m.repoList.View())
		empty := shared.AppStyle.Width(shared.Half(m.width)).Render("No repositories found")
		return tea.NewView(fmt.Sprint(lipgloss.JoinHorizontal(lipgloss.Top, repoList, empty)))
	}

	m.repoModel.SelectRepo(selected)

	var repoList = shared.AppStyle.Width(shared.Half(m.width)).Render(m.repoList.View())
	var settings = shared.AppStyle.Width(shared.Half(m.width)).Render(fmt.Sprint(m.repoModel.View().Content))
	var rightPanel = lipgloss.JoinVertical(lipgloss.Center, settings)

	var views = []string{repoList, rightPanel}

	return tea.NewView(fmt.Sprint(lipgloss.JoinHorizontal(lipgloss.Top, views...)))
}

func (m Model) HeaderView() tea.View {
	label := m.Title
	if label == "" {
		label = "Repositories"
	}

	if m.repoCount > 0 {
		label = fmt.Sprintf("%s (%d repos)", label, m.repoCount)
	}

	prefix := "Organization"
	if m.isUser {
		prefix = "User"
	}

	title := fmt.Sprintf("%s: %s", prefix, label)
	return tea.NewView(shared.TitleStyle.Render(title))
}

func (m Model) HelpView() tea.View {
	return tea.NewView(m.help.View(m.helpKeys()))
}

// helpKeys lists every binding this screen handles, including the ones it
// forwards to the detail pane, in the order they appear in the footer.
func (m Model) helpKeys() shared.KeyBindings {
	repoKeys := m.repoModel.Keys()
	return shared.KeyBindings{
		m.keymap.Up,
		m.keymap.Down,
		repoKeys.NextTab,
		repoKeys.PrevTab,
		m.keymap.Filters,
		m.keymap.Back,
	}
}

func (m *Model) ProgressView() tea.View {
	m.progress.SetWidth(m.width)
	text := fmt.Sprintf("Getting repositories ... %d of %d\n", len(m.repos), m.repoCount)
	return tea.NewView(fmt.Sprint(lipgloss.JoinVertical(lipgloss.Center, text, m.progress.View())))
}

// orgKeyMap holds the bindings the repository list screen handles itself. Tab
// switching is owned by the detail pane's repo.KeyMap.
type orgKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Filters key.Binding
	Back    key.Binding
}

func newOrgKeyMap() orgKeyMap {
	return orgKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Filters: key.NewBinding(
			key.WithKeys("f", "F"),
			key.WithHelp("f", "filters"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
	}
}
