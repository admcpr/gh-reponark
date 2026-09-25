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

// viewMode is how the repositories are laid out.
type viewMode int

const (
	// listView shows one repository per row beside an inspector for the
	// selected one.
	listView viewMode = iota
	// matrixView shows every repository against every property in the
	// active group, for spotting the ones set up differently.
	matrixView
)

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
	repoModel repo.Model

	// The repositories that pass the filters, which one is selected, and the
	// first row and column each view has scrolled to. Both views share the
	// selection, so switching between them keeps your place.
	visible      []repo.RepoConfig
	cursor       int
	mode         viewMode
	inspecting   bool // the list view's keys move through properties, not repos
	listOffset   int
	matrixOffset int
	columnOffset int

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
	m.repoModel.SetDimensions(m.inspectorWidth(), height)

	return m
}

func (m *Model) SetDimensions(width, height int) {
	m.width = width
	m.height = height
	m.help.SetWidth(width)
	m.repoModel.SetDimensions(m.inspectorWidth(), height)
	m.scrollToCursor()
}

// applyFilters rebuilds the visible repositories from the ones that pass the
// current filters and selects the first of them.
func (m *Model) applyFilters() {
	m.visible = m.filters.FilterRepos(m.repos)
	m.cursor, m.listOffset, m.matrixOffset = 0, 0, 0
	m.selectionChanged()
}

// selectedRepo returns the highlighted repository, if any.
func (m *Model) selectedRepo() (repo.RepoConfig, bool) {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return repo.RepoConfig{}, false
	}
	return m.visible[m.cursor], true
}

// moveCursor selects the repository delta rows away, stopping at either end.
func (m *Model) moveCursor(delta int) {
	if len(m.visible) == 0 {
		return
	}
	m.cursor = shared.Max(0, shared.Min(m.cursor+delta, len(m.visible)-1))
	m.selectionChanged()
}

// selectionChanged points the inspector at the selected repository and
// scrolls both views so it stays on screen.
func (m *Model) selectionChanged() {
	if selected, ok := m.selectedRepo(); ok {
		m.repoModel.SelectRepo(selected)
	}
	m.scrollToCursor()
}

func (m *Model) scrollToCursor() {
	m.listOffset = shared.ScrollOffset(m.listOffset, m.cursor, m.listRows(), len(m.visible))
	m.matrixOffset = shared.ScrollOffset(m.matrixOffset, m.cursor, m.matrixRows(), len(m.visible))
}

// pageSize is how many rows the current view shows at once.
func (m *Model) pageSize() int {
	if m.mode == matrixView {
		return m.matrixRows()
	}
	return m.listRows()
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
		m.applyFilters()
	}
	return m.progress.SetPercent(1.0)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		m.applyFilters()
		return m, nil

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel
		return m, cmd

	case tea.KeyPressMsg:
		return m, m.handleKey(msg)
	}

	return m, nil
}

func (m *Model) handleKey(msg tea.KeyPressMsg) tea.Cmd {
	repoKeys := m.repoModel.Keys()
	listMode := m.mode == listView
	switch {
	case key.Matches(msg, m.keymap.Filters):
		return func() tea.Msg {
			return filters.OpenFiltersMsg{Filters: m.filters, Repos: m.repos}
		}
	case m.inspecting && key.Matches(msg, m.keymap.Back, m.keymap.Left):
		m.setInspecting(false)
	case key.Matches(msg, m.keymap.Back):
		return func() tea.Msg {
			return shared.PreviousMsg{}
		}
	case key.Matches(msg, m.keymap.ToggleView):
		m.toggleView()
	case key.Matches(msg, m.keymap.Inspect), listMode && key.Matches(msg, m.keymap.Right):
		m.mode = listView
		m.setInspecting(true)
		m.scrollToCursor()
	case !listMode && key.Matches(msg, m.keymap.Left):
		m.repoModel.SelectProperty(m.repoModel.ActiveProperty() - 1)
	case !listMode && key.Matches(msg, m.keymap.Right):
		m.repoModel.SelectProperty(m.repoModel.ActiveProperty() + 1)
	case key.Matches(msg, repoKeys.NextTab, repoKeys.PrevTab):
		repoModel, cmd := m.repoModel.Update(msg)
		m.repoModel = repoModel.(repo.Model)
		return cmd
	default:
		if delta, ok := m.movement(msg); ok {
			if m.inspecting {
				m.repoModel.SelectProperty(m.repoModel.ActiveProperty() + delta)
			} else {
				m.moveCursor(delta)
			}
		}
	}
	return nil
}

// movement is how far a vertical movement key moves the focused cursor.
// Moves past either end are clamped by whoever applies them.
func (m *Model) movement(msg tea.KeyPressMsg) (int, bool) {
	page, all := m.pageSize(), len(m.visible)
	if m.inspecting {
		page, all = m.repoModel.PageSize(), len(m.repoModel.ActiveGroup().Properties)
	}
	switch {
	case key.Matches(msg, m.keymap.Up):
		return -1, true
	case key.Matches(msg, m.keymap.Down):
		return 1, true
	case key.Matches(msg, m.keymap.PageUp):
		return -page, true
	case key.Matches(msg, m.keymap.PageDown):
		return page, true
	case key.Matches(msg, m.keymap.Top):
		return -all, true
	case key.Matches(msg, m.keymap.Bottom):
		return all, true
	}
	return 0, false
}

// setInspecting moves focus between the repo list and the inspector.
func (m *Model) setInspecting(inspecting bool) {
	m.inspecting = inspecting
	m.repoModel.SetFocused(inspecting)
}

func (m *Model) toggleView() {
	m.setInspecting(false)
	if m.mode == listView {
		m.mode = matrixView
	} else {
		m.mode = listView
	}
	m.scrollToCursor()
}

func (m *Model) View() tea.View {
	if m.progress.Percent() < 1 {
		return m.ProgressView()
	}

	if len(m.visible) == 0 {
		return tea.NewView(shared.DimStyle.Render("No repositories found"))
	}

	if m.mode == matrixView {
		return tea.NewView(m.matrixView())
	}
	return tea.NewView(m.listView())
}

// Breadcrumb names the organization or user whose repositories are shown.
func (m Model) Breadcrumb() string {
	if m.Title == "" {
		return "Repositories"
	}
	return m.Title
}

// Status counts the repositories, and how many pass the filters when any are
// applied.
func (m Model) Status() string {
	switch {
	case m.progress.Percent() < 1:
		return "loading repositories"
	case len(m.filters) > 0:
		return fmt.Sprintf("%d of %d repos · %s", len(m.visible), len(m.repos), plural(len(m.filters), "filter"))
	default:
		return plural(len(m.repos), "repo")
	}
}

// plural formats a count with its noun, e.g. "1 repo" or "3 repos".
func plural(n int, noun string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", noun)
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

func (m Model) HelpView() tea.View {
	return tea.NewView(m.help.View(m.helpKeys()))
}

// helpKeys lists the bindings for the current view in the order they appear
// in the footer. Pairs of keys that do the same thing in opposite directions
// share one entry so the footer fits on a line.
func (m Model) helpKeys() shared.KeyBindings {
	repoKeys := m.repoModel.Keys()
	tabs := helpOnly(repoKeys.NextTab, repoKeys.PrevTab, "tab", "group")
	if m.mode == matrixView {
		return shared.KeyBindings{
			helpOnly(m.keymap.Down, m.keymap.Up, "j/k", "repo"),
			helpOnly(m.keymap.Left, m.keymap.Right, "h/l", "column"),
			tabs,
			withHelp(m.keymap.Inspect, "open"),
			withHelp(m.keymap.ToggleView, "list"),
			m.keymap.Filters,
			m.keymap.Back,
		}
	}
	if m.inspecting {
		return shared.KeyBindings{
			helpOnly(m.keymap.Down, m.keymap.Up, "j/k", "property"),
			helpOnly(m.keymap.Left, m.keymap.Back, "h/esc", "repos"),
			tabs,
			withHelp(m.keymap.ToggleView, "matrix"),
			m.keymap.Filters,
		}
	}
	return shared.KeyBindings{
		helpOnly(m.keymap.Down, m.keymap.Up, "j/k", "repo"),
		helpOnly(m.keymap.Right, m.keymap.Inspect, "l/enter", "inspect"),
		tabs,
		withHelp(m.keymap.ToggleView, "matrix"),
		m.keymap.Filters,
		m.keymap.Back,
	}
}

// helpOnly combines two bindings into one footer entry.
func helpOnly(a, b key.Binding, keys, desc string) key.Binding {
	return key.NewBinding(key.WithKeys(append(a.Keys(), b.Keys()...)...), key.WithHelp(keys, desc))
}

// withHelp returns a copy of binding with a different description.
func withHelp(binding key.Binding, desc string) key.Binding {
	binding.SetHelp(binding.Help().Key, desc)
	return binding
}

func (m *Model) ProgressView() tea.View {
	m.progress.SetWidth(m.width)
	text := fmt.Sprintf("Getting repositories ... %d of %d\n", len(m.repos), m.repoCount)
	return tea.NewView(fmt.Sprint(lipgloss.JoinVertical(lipgloss.Center, text, m.progress.View())))
}

// orgKeyMap holds the bindings the repository screen handles itself. Tab and
// property switching are owned by the inspector's repo.KeyMap.
type orgKeyMap struct {
	Up         key.Binding
	Down       key.Binding
	PageUp     key.Binding
	PageDown   key.Binding
	Top        key.Binding
	Bottom     key.Binding
	Left       key.Binding
	Right      key.Binding
	ToggleView key.Binding
	Inspect    key.Binding
	Filters    key.Binding
	Back       key.Binding
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
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+u"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+d"),
			key.WithHelp("pgdn", "page down"),
		),
		Top: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g", "first"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G", "last"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "column left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "column right"),
		),
		ToggleView: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "switch view"),
		),
		Inspect: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "inspect"),
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
