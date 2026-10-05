// Package org is the repository browser for one organization or user.
package repos

import (
	"fmt"
	"sort"
	"strings"

	"gh-reponark/github"
	"gh-reponark/repo"
	"gh-reponark/ui"

	tea "charm.land/bubbletea/v2"
)

// maxConcurrentBatches bounds how many configuration requests are in flight
// at once, so a large organization does not trip GitHub's secondary rate
// limits while still loading several batches in parallel.
const maxConcurrentBatches = 4

// repositoryPageMsg carries one page of repository names.
type repositoryPageMsg github.RepositoryPage

// repositoryBatchMsg carries the configuration of one batch of repositories.
type repositoryBatchMsg []repo.Repository

// loading reports whether repositories are still arriving.
func (m *Model) loading() bool { return m.progress.Percent() < 1 }

// loadRepositoryPage returns a command that fetches one page of repository
// names, starting after the given cursor.
func (m *Model) loadRepositoryPage(after string) tea.Cmd {
	return func() tea.Msg {
		page, err := m.svc.ListRepositories(m.Title, m.isUser, after)
		if err != nil {
			return ui.ErrorMsg{Err: err}
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
			return ui.ErrorMsg{Err: err}
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
	// The inspector draws each repository's counts against the whole org.
	m.repoModel.SetOrgRepos(m.repos)
	if len(m.repos) > 0 {
		m.applyFilters()
	}
	return m.progress.SetPercent(1.0)
}

// ProgressView is the loading screen: the account being opened, a spinner
// with the current phase and count, the gradient bar and a ticker of the
// repositories that have just arrived.
func (m *Model) ProgressView() tea.View {
	width := ui.LoadingWidth(m.width)
	m.progress.SetWidth(width)

	title := ui.StrongStyle.Render(m.Title) + "  " + ui.TintedPill("loading", ui.AppColors.Accent, ui.AppColors.AccentTint)

	phase := "Listing repositories…"
	if m.repoCount > 0 && (len(m.pending) > 0 || m.inFlight > 0 || len(m.repos) > 0) {
		phase = fmt.Sprintf("Fetching settings  %s of %s",
			ui.ValueStyle.Render(fmt.Sprint(len(m.repos))), ui.ValueStyle.Render(fmt.Sprint(m.repoCount)))
	}
	status := m.spinner.View() + " " + ui.TextBodyStyle.Render(phase)

	return tea.NewView(ui.Loading(m.width, m.height, title, status, m.progress.View(), m.ticker(width)))
}

// ticker is a line of the repository names listed so far, in the order they
// arrived, scrolling right to left once there are more than fit. It draws on
// the names rather than the loaded settings, so it starts moving as soon as
// the first page of the listing lands.
func (m *Model) ticker(width int) string {
	if len(m.names) == 0 {
		return ""
	}
	return ui.DimStyle.Render(ui.Marquee(strings.Join(m.names, "  ·  "), width, m.tickerOffset))
}
