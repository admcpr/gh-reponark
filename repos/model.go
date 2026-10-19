package repos

import (
	"fmt"

	"gh-reponark/filter"
	"gh-reponark/filters"
	"gh-reponark/github"
	"gh-reponark/inspector"
	"gh-reponark/overview"
	"gh-reponark/repo"
	"gh-reponark/ui"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

// viewMode is how the repositories are laid out.
type viewMode int

const (
	// listView shows one repository per row beside an inspector for the
	// selected one.
	listView viewMode = iota
	// matrixView shows every repository against every property in the
	// active group, for spotting the ones set up differently.
	matrixView
	// overviewView shows the owner's profile and the repositories as a
	// whole: totals, breakdowns and every property of the active group
	// aggregated across them.
	overviewView
)

type Model struct {
	svc github.Service

	Title     string
	repoCount int
	repos     []repo.RepoConfig
	filters   filter.FilterMap
	isUser    bool

	// Loading state: names collected from the listing, batches of names not
	// yet requested, and how many batch requests are outstanding.
	names    []string
	pending  [][]string
	inFlight int

	// The owner's profile for the overview, loaded alongside the
	// repositories; ownerErr is kept rather than shown as an error screen,
	// since the repositories are still worth browsing without it.
	owner       github.Owner
	ownerLoaded bool
	ownerErr    error

	keymap    orgKeyMap
	repoModel inspector.Model

	// The repositories that pass the filters and each view's cursor over
	// them. Both views share the selection, so switching between them keeps
	// your place: list.Index is the selected repository and matrix.Index
	// mirrors it, while each keeps its own window since the views show a
	// different number of rows. columnOffset is the matrix's first column.
	visible      []repo.RepoConfig
	list         ui.Cursor
	matrix       ui.Cursor
	mode         viewMode
	inspecting   bool // the list view's keys move through properties, not repos
	columnOffset int
	// The overview's cards over the visible repositories, built when first
	// drawn after the filters change, and its cursors: the focused card,
	// the focused slice of it and the first row of cards in view.
	cards   []overview.Card
	card    int
	slice   int
	cardRow int
	// fromOverview is set while the list shows a slice the overview was
	// filtered to; esc then returns there with overviewFilters, the
	// filters from before, put back.
	fromOverview    bool
	overviewFilters filter.FilterMap

	width  int
	height int

	progress progress.Model
	spinner  spinner.Model
	// tickerOffset is how far the loading screen's ticker of repository
	// names has scrolled; it advances with the spinner.
	tickerOffset int
}

func NewModel(svc github.Service, orgKey ui.OrgKey, width, height int) *Model {
	keymap := newOrgKeyMap()

	m := &Model{
		svc:       svc,
		Title:     orgKey.Name,
		isUser:    orgKey.IsUser,
		width:     width,
		height:    height,
		keymap:    keymap,
		repoModel: inspector.NewModel(width/2, height),
		progress:  ui.NewProgress(),
		spinner:   ui.NewSpinner(),
	}
	m.repoModel.SetDimensions(m.inspectorWidth(), height)

	return m
}

func (m *Model) SetDimensions(width, height int) {
	m.width = width
	m.height = height
	m.repoModel.SetDimensions(m.inspectorWidth(), height)
	m.resizeCursors()
}

// resizeCursors fits each view's window to the pane and the visible
// repositories, keeping the selection on screen.
func (m *Model) resizeCursors() {
	m.list.Resize(m.listRows(), len(m.visible))
	m.matrix.Resize(m.matrixRows(), len(m.visible))
}

// applyFilters rebuilds the visible repositories from the ones that pass the
// current filters and selects the first of them.
func (m *Model) applyFilters() {
	m.visible = m.filters.FilterRepos(m.repos)
	m.cards = nil
	m.list, m.matrix = ui.Cursor{}, ui.Cursor{}
	m.resizeCursors()
	m.selectionChanged()
}

// selectedRepo returns the highlighted repository, if any.
func (m *Model) selectedRepo() (repo.RepoConfig, bool) {
	if m.list.Index >= len(m.visible) {
		return repo.RepoConfig{}, false
	}
	return m.visible[m.list.Index], true
}

// moveCursor selects the repository delta rows away, stopping at either end,
// in both views.
func (m *Model) moveCursor(delta int) {
	m.list.Move(delta)
	m.matrix.Set(m.list.Index)
	m.selectionChanged()
}

// selectionChanged points the inspector at the selected repository.
func (m *Model) selectionChanged() {
	if selected, ok := m.selectedRepo(); ok {
		m.repoModel.SelectRepo(selected)
	}
}

// pageSize is how many rows the current view shows at once.
func (m *Model) pageSize() int {
	switch m.mode {
	case matrixView:
		return m.matrixRows()
	case overviewView:
		return m.overviewLayout().rows * m.overviewLayout().columns
	default:
		return m.listRows()
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.loadRepositoryPage(""), m.loadOwner, m.spinner.Tick)
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

	case ownerMsg:
		m.owner, m.ownerLoaded = github.Owner(msg), true
		return m, nil

	case ownerErrMsg:
		m.ownerErr = msg.err
		return m, nil

	case filters.FiltersMsg:
		// Filters edited by hand are kept: esc no longer undoes them.
		m.filters = filter.FilterMap(msg)
		m.fromOverview = false
		m.applyFilters()
		return m, nil

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel
		return m, cmd

	case spinner.TickMsg:
		// The spinner only turns while there is something to wait for.
		if !m.loading() {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		m.tickerOffset++
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
	case m.fromOverview && m.mode == listView && key.Matches(msg, m.keymap.Back):
		m.backToOverview()
	case key.Matches(msg, m.keymap.Back):
		return func() tea.Msg {
			return ui.PreviousMsg{}
		}
	case key.Matches(msg, m.keymap.ToggleView):
		m.toggleView()
	case m.mode == overviewView:
		m.handleOverviewKey(msg)
	case key.Matches(msg, m.keymap.Inspect), listMode && key.Matches(msg, m.keymap.Right):
		if _, ok := m.selectedRepo(); !ok {
			// Nothing is loaded or everything is filtered out.
			return nil
		}
		m.mode = listView
		m.setInspecting(true)
	case m.mode == matrixView && key.Matches(msg, m.keymap.Left):
		m.repoModel.SelectProperty(m.repoModel.ActiveProperty() - 1)
	case m.mode == matrixView && key.Matches(msg, m.keymap.Right):
		m.repoModel.SelectProperty(m.repoModel.ActiveProperty() + 1)
	case key.Matches(msg, repoKeys.NextTab, repoKeys.PrevTab):
		repoModel, cmd := m.repoModel.Update(msg)
		m.repoModel = repoModel.(inspector.Model)
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

// handleOverviewKey moves around the overview: tab from card to card, as
// it moves between groups elsewhere, j/k through the focused card's
// slices, h/l to the card either side, and enter to the list filtered to
// the focused slice.
func (m *Model) handleOverviewKey(msg tea.KeyPressMsg) {
	repoKeys := m.repoModel.Keys()
	switch {
	case key.Matches(msg, m.keymap.Inspect):
		m.filterToSlice()
	case key.Matches(msg, repoKeys.NextTab):
		m.cycleCard(1)
	case key.Matches(msg, repoKeys.PrevTab):
		m.cycleCard(-1)
	case key.Matches(msg, m.keymap.Down):
		m.moveSlice(1)
	case key.Matches(msg, m.keymap.Up):
		m.moveSlice(-1)
	case key.Matches(msg, m.keymap.Left):
		m.moveCard(-1)
	case key.Matches(msg, m.keymap.Right):
		m.moveCard(1)
	case key.Matches(msg, m.keymap.PageUp):
		m.moveCard(-m.pageSize())
	case key.Matches(msg, m.keymap.PageDown):
		m.moveCard(m.pageSize())
	case key.Matches(msg, m.keymap.Top):
		m.moveCard(-len(m.overviewCards()))
	case key.Matches(msg, m.keymap.Bottom):
		m.moveCard(len(m.overviewCards()))
	}
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

// toggleView moves to the next view: the list, the matrix, the overview
// and round again. All three share the selected repository and property.
// Leaving the list this way keeps a slice's filters, so esc no longer
// undoes them.
func (m *Model) toggleView() {
	m.setInspecting(false)
	m.fromOverview = false
	m.mode = (m.mode + 1) % 3
}

// backToOverview leaves a slice the overview was filtered to, putting the
// filters from before back.
func (m *Model) backToOverview() {
	m.filters = m.overviewFilters
	m.fromOverview, m.overviewFilters = false, nil
	m.applyFilters()
	m.mode = overviewView
}

func (m *Model) View() tea.View {
	if m.progress.Percent() < 1 {
		return m.ProgressView()
	}

	if len(m.visible) == 0 {
		if len(m.filters) > 0 {
			return tea.NewView(ui.DimStyle.Render(fmt.Sprintf("No repositories match the %d %s · f to change them",
				len(m.filters), ui.Plural(len(m.filters), "filter", "filters"))))
		}
		return tea.NewView(ui.DimStyle.Render("No repositories found"))
	}

	switch m.mode {
	case matrixView:
		return tea.NewView(m.matrixView())
	case overviewView:
		return tea.NewView(m.overviewView())
	default:
		return tea.NewView(m.listView())
	}
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
		return fmt.Sprintf("%d of %d repos · %d %s", len(m.visible), len(m.repos), len(m.filters), ui.Plural(len(m.filters), "filter", "filters"))
	default:
		return fmt.Sprintf("%d %s", len(m.repos), ui.Plural(len(m.repos), "repo", "repos"))
	}
}
