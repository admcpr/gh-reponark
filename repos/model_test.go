package repos

import (
	"fmt"
	"strings"
	"testing"

	"gh-reponark/filter"
	"gh-reponark/filters"
	"gh-reponark/github"
	"gh-reponark/github/githubtest"
	"gh-reponark/repo"
	"gh-reponark/ui"
	"gh-reponark/ui/uitest"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func newOrgModel() *Model {
	return NewModel(&githubtest.Fake{}, ui.OrgKey{Name: "demo", IsUser: false}, 80, 24)
}

// itemNames returns the names of the repositories that pass the filters.
func itemNames(m *Model) []string {
	names := make([]string, len(m.visible))
	for i, c := range m.visible {
		names[i] = c.Name
	}
	return names
}

// repoNames returns the names of the loaded repositories in order.
func repoNames(m *Model) []string {
	names := make([]string, len(m.repos))
	for i, r := range m.repos {
		names[i] = r.Name
	}
	return names
}

// namesPage builds a page of repository names.
func namesPage(totalCount int, hasNextPage bool, endCursor string, names ...string) repositoryPageMsg {
	refs := make([]github.RepositoryRef, len(names))
	for i, name := range names {
		refs[i] = github.RepositoryRef{Name: name, Url: "https://github.com/demo/" + name}
	}
	return repositoryPageMsg(github.RepositoryPage{
		Repositories: refs,
		TotalCount:   totalCount,
		EndCursor:    endCursor,
		HasNextPage:  hasNextPage,
	})
}

// numberedRepos returns n repositories with distinct names.
func numberedRepos(n int) []repo.Repository {
	repos := make([]repo.Repository, n)
	for i, name := range numberedNames(n) {
		repos[i] = repo.Repository{Name: name}
	}
	return repos
}

// lastNamesPage builds a page that completes the listing.
func lastNamesPage(names ...string) repositoryPageMsg {
	return namesPage(len(names), false, "", names...)
}

// numberedNames returns n distinct repository names.
func numberedNames(n int) []string {
	names := make([]string, n)
	for i := range names {
		names[i] = fmt.Sprintf("repo-%02d", i)
	}
	return names
}

// newLoadedModel returns a model that has received all of its repositories.
func newLoadedModel(repositories ...repo.Repository) *Model {
	m := newOrgModel()
	names := make([]string, len(repositories))
	for i, r := range repositories {
		names[i] = r.Name
	}
	m.Update(lastNamesPage(names...))
	for _, batch := range batchNames(names) {
		var repos []repo.Repository
		for _, name := range batch {
			for _, r := range repositories {
				if r.Name == name {
					repos = append(repos, r)
				}
			}
		}
		m.Update(repositoryBatchMsg(repos))
	}
	return m
}

// drive runs commands and feeds the resulting messages back into the model
// until nothing is left to do, the way the Bubble Tea runtime would.
func drive(m *Model, cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	switch msg := cmd().(type) {
	case tea.BatchMsg:
		for _, c := range msg {
			drive(m, c)
		}
	case repositoryPageMsg, repositoryBatchMsg:
		_, next := m.Update(msg)
		drive(m, next)
	}
}

func TestNewModel(t *testing.T) {
	tests := []struct {
		name   string
		key    ui.OrgKey
		isUser bool
	}{
		{name: "organization", key: ui.OrgKey{Name: "acme", IsUser: false}, isUser: false},
		{name: "user", key: ui.OrgKey{Name: "octocat", IsUser: true}, isUser: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(&githubtest.Fake{}, tt.key, 80, 24)

			assert.Equal(t, tt.key.Name, m.Title)
			assert.Equal(t, tt.isUser, m.isUser)
			assert.Equal(t, 80, m.width)
			assert.Equal(t, 24, m.height)
			assert.Equal(t, 0, m.repoCount)
			assert.Empty(t, m.repos)
			assert.Nil(t, m.filters)
			assert.Equal(t, 0.0, m.progress.Percent())
		})
	}
}

func TestModel_SetDimensions(t *testing.T) {
	m := newOrgModel()

	m.SetDimensions(120, 40)

	assert.Equal(t, 120, m.width)
	assert.Equal(t, 40, m.height)
}

func TestModel_Update_FiltersMsg(t *testing.T) {
	m := newLoadedModel(
		repo.Repository{Name: "archived", IsArchived: true},
		repo.Repository{Name: "active", IsArchived: false},
	)

	filterMap := filter.FilterMap{"Is Archived": filter.NewBoolFilter("Is Archived", true)}
	_, cmd := m.Update(filters.FiltersMsg(filterMap))

	assert.Nil(t, cmd)
	assert.Equal(t, filterMap, m.filters)
	assert.Equal(t, []string{"archived"}, itemNames(m))
}

func TestModel_SelectedRepo_FollowsFilteredList(t *testing.T) {
	m := newLoadedModel(
		repo.Repository{Name: "alpha", IsArchived: false},
		repo.Repository{Name: "bravo", IsArchived: true},
		repo.Repository{Name: "charlie", IsArchived: true},
	)
	m.Update(filters.FiltersMsg(filter.FilterMap{"Is Archived": filter.NewBoolFilter("Is Archived", true)}))
	assert.Equal(t, []string{"bravo", "charlie"}, itemNames(m))

	selected, ok := m.selectedRepo()
	assert.True(t, ok)
	assert.Equal(t, "bravo", selected.Name, "the first visible repo should be selected, not the first of all repos")

	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	selected, ok = m.selectedRepo()
	assert.True(t, ok)
	assert.Equal(t, "charlie", selected.Name)
	assert.Contains(t, uitest.Plain(m.View()), "charlie")
}

func TestModel_View_AllFilteredOutShowsEmptyState(t *testing.T) {
	m := newLoadedModel(repo.Repository{Name: "alpha"}, repo.Repository{Name: "bravo"})
	m.Update(filters.FiltersMsg(filter.FilterMap{"Is Archived": filter.NewBoolFilter("Is Archived", true)}))

	_, ok := m.selectedRepo()
	assert.False(t, ok)
	assert.Empty(t, itemNames(m))
	assert.Contains(t, uitest.Plain(m.View()), "No repositories found")
}

func TestModel_SelectedRepo_NoneBeforeLoad(t *testing.T) {
	m := newOrgModel()

	_, ok := m.selectedRepo()

	assert.False(t, ok)
}

func TestModel_Update_FiltersMsg_BeforeReposLoad(t *testing.T) {
	m := newOrgModel()
	filterMap := filter.FilterMap{"Is Archived": filter.NewBoolFilter("Is Archived", true)}

	assert.NotPanics(t, func() { m.Update(filters.FiltersMsg(filterMap)) })

	assert.Equal(t, filterMap, m.filters, "filters should be kept for when the repos arrive")
	assert.Empty(t, m.visible)
}

func TestModel_Update_FiltersMsg_ClearingFiltersRestoresRepos(t *testing.T) {
	m := newLoadedModel(
		repo.Repository{Name: "archived", IsArchived: true},
		repo.Repository{Name: "active", IsArchived: false},
	)
	m.Update(filters.FiltersMsg(filter.FilterMap{"Is Archived": filter.NewBoolFilter("Is Archived", true)}))

	m.Update(filters.FiltersMsg(filter.FilterMap{}))

	assert.Len(t, m.visible, 2)
}

func TestModel_Update_FilterKeyOpensFilters(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{name: "lowercase f", key: "f"},
		{name: "uppercase F", key: "F"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newOrgModel()
			m.filters = filter.FilterMap{"Is Fork": filter.NewBoolFilter("Is Fork", false)}

			code := []rune(tt.key)[0]
			_, cmd := m.Update(tea.KeyPressMsg{Code: code, Text: tt.key})

			if assert.NotNil(t, cmd) {
				open, ok := cmd().(filters.OpenFiltersMsg)
				assert.True(t, ok, "expected an OpenFiltersMsg")
				assert.Equal(t, m.filters, open.Filters)
			}
		})
	}
}

func TestModel_Update_EscGoesBack(t *testing.T) {
	m := newOrgModel()

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if assert.NotNil(t, cmd) {
		prev, ok := cmd().(ui.PreviousMsg)
		assert.True(t, ok, "expected a PreviousMsg")
		assert.Nil(t, prev.Message)
	}
}

func TestModel_Update_TabForwardsToRepoModel(t *testing.T) {
	m := newLoadedModel(repo.Repository{Name: "alpha"})

	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	assert.Equal(t, 1, m.repoModel.ActiveTab())

	m.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	assert.Equal(t, 0, m.repoModel.ActiveTab())
}

func TestModel_Update_InspectorFocus(t *testing.T) {
	enter := tea.KeyPressMsg{Code: tea.KeyEnter}
	right := tea.KeyPressMsg{Code: 'l', Text: "l"}
	left := tea.KeyPressMsg{Code: 'h', Text: "h"}
	esc := tea.KeyPressMsg{Code: tea.KeyEscape}

	for _, tt := range []struct {
		name     string
		in, out  tea.KeyPressMsg
		wantBack bool
	}{
		{name: "l in, h out", in: right, out: left},
		{name: "enter in, esc out", in: enter, out: esc},
		{name: "right arrow in, left arrow out", in: tea.KeyPressMsg{Code: tea.KeyRight}, out: tea.KeyPressMsg{Code: tea.KeyLeft}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := newLoadedModel(repo.Repository{Name: "alpha"}, repo.Repository{Name: "bravo"})

			m.Update(tt.in)
			assert.True(t, m.inspecting)
			assert.True(t, m.repoModel.Focused())

			_, cmd := m.Update(tt.out)
			assert.Nil(t, cmd, "leaving the inspector does not leave the screen")
			assert.False(t, m.inspecting)
			assert.False(t, m.repoModel.Focused())
		})
	}
}

func TestModel_Update_MovementFollowsFocus(t *testing.T) {
	m := newLoadedModel(repo.Repository{Name: "alpha"}, repo.Repository{Name: "bravo"})
	down := tea.KeyPressMsg{Code: 'j', Text: "j"}

	m.Update(down)
	assert.Equal(t, 1, m.list.Index, "j moves the repo cursor from the list")
	assert.Equal(t, 0, m.repoModel.ActiveProperty())

	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m.Update(down)
	m.Update(down)
	assert.Equal(t, 1, m.list.Index, "j moves the property cursor from the inspector")
	assert.Equal(t, 2, m.repoModel.ActiveProperty())

	m.Update(tea.KeyPressMsg{Code: 'G', Text: "G"})
	assert.Equal(t, len(m.repoModel.ActiveGroup().Properties)-1, m.repoModel.ActiveProperty())
	m.Update(tea.KeyPressMsg{Code: 'g', Text: "g"})
	assert.Equal(t, 0, m.repoModel.ActiveProperty())
	assert.Equal(t, 1, m.list.Index)
}

func TestModel_Update_TabKeepsInspectorFocus(t *testing.T) {
	m := newLoadedModel(repo.Repository{Name: "alpha"})
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})

	assert.True(t, m.inspecting)
	assert.Equal(t, 1, m.repoModel.ActiveTab())
}

func TestModel_Update_CursorMovement(t *testing.T) {
	m := newLoadedModel(numberedRepos(40)...)
	m.SetDimensions(80, 12)

	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	assert.Equal(t, 1, m.list.Index)

	m.Update(tea.KeyPressMsg{Code: tea.KeyPgDown})
	assert.Equal(t, 1+m.listRows(), m.list.Index, "page down moves a screenful")

	m.Update(tea.KeyPressMsg{Code: 'G', Text: "G"})
	assert.Equal(t, 39, m.list.Index)
	assert.Equal(t, 40-m.listRows(), m.list.Offset, "the list scrolls to keep the selection on screen")

	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	assert.Equal(t, 39, m.list.Index, "the cursor stops at the last repository")

	m.Update(tea.KeyPressMsg{Code: 'g', Text: "g"})
	assert.Equal(t, 0, m.list.Index)
	assert.Equal(t, 0, m.list.Offset)

	m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	assert.Equal(t, 0, m.list.Index, "the cursor stops at the first repository")
}

func TestModel_Update_CursorFollowsInspector(t *testing.T) {
	m := newLoadedModel(repo.Repository{Name: "alpha"}, repo.Repository{Name: "bravo"})

	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})

	assert.Contains(t, uitest.Plain(m.repoModel.View()), "bravo")
}

func TestModel_Update_ToggleView(t *testing.T) {
	m := newLoadedModel(repo.Repository{Name: "alpha"}, repo.Repository{Name: "bravo"})
	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	assert.Equal(t, listView, m.mode)

	m.Update(tea.KeyPressMsg{Code: 'v', Text: "v"})
	assert.Equal(t, matrixView, m.mode)
	assert.Equal(t, 1, m.list.Index, "switching views keeps the selection")

	m.Update(tea.KeyPressMsg{Code: 'v', Text: "v"})
	assert.Equal(t, listView, m.mode)
}

func TestModel_Update_EnterInspectsFromMatrix(t *testing.T) {
	m := newLoadedModel(repo.Repository{Name: "alpha"})
	m.Update(tea.KeyPressMsg{Code: 'v', Text: "v"})
	m.Update(tea.KeyPressMsg{Code: 'l', Text: "l"})

	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.Equal(t, listView, m.mode)
	assert.True(t, m.inspecting, "the inspector opens focused")
	assert.Equal(t, 1, m.repoModel.ActiveProperty(), "on the column that was focused")
}

func TestModel_Update_ToggleViewReturnsFocusToList(t *testing.T) {
	m := newLoadedModel(repo.Repository{Name: "alpha"})
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	m.Update(tea.KeyPressMsg{Code: 'v', Text: "v"})
	m.Update(tea.KeyPressMsg{Code: 'v', Text: "v"})

	assert.False(t, m.inspecting)
}

func TestModel_Update_ColumnKeysOnlyMoveInMatrix(t *testing.T) {
	m := newLoadedModel(repo.Repository{Name: "alpha"})

	m.Update(tea.KeyPressMsg{Code: 'l', Text: "l"})
	assert.Equal(t, 0, m.repoModel.ActiveProperty(), "l focuses the inspector in the list view")
	m.Update(tea.KeyPressMsg{Code: 'h', Text: "h"})

	m.mode = matrixView
	m.Update(tea.KeyPressMsg{Code: 'l', Text: "l"})
	m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	assert.Equal(t, 2, m.repoModel.ActiveProperty())

	m.Update(tea.KeyPressMsg{Code: 'h', Text: "h"})
	assert.Equal(t, 1, m.repoModel.ActiveProperty())
}

func TestModel_View_Loaded(t *testing.T) {
	m := newLoadedModel(repo.Repository{Name: "alpha"}, repo.Repository{Name: "bravo"})

	content := uitest.Plain(m.View())

	assert.NotContains(t, content, "Getting repositories")
	assert.Contains(t, content, "alpha")
	assert.Contains(t, content, "bravo")
	assert.Contains(t, content, "NAME", "the list has column headings")
	assert.Contains(t, content, "1/2", "the list shows the selected position")
	assert.Contains(t, content, "Overview", "the inspector shows the group tabs")
}

func TestModel_View_FitsDimensions(t *testing.T) {
	m := newLoadedModel(numberedRepos(40)...)
	m.SetDimensions(90, 20)

	for _, mode := range []viewMode{listView, matrixView} {
		m.mode = mode
		lines := strings.Split(uitest.Plain(m.View()), "\n")

		assert.Len(t, lines, 20, "mode %d", mode)
		for _, line := range lines {
			assert.LessOrEqual(t, lipgloss.Width(line), 90, "mode %d: %q", mode, line)
		}
	}
}

func TestModel_View_ListShowsFilteredCount(t *testing.T) {
	m := newLoadedModel(
		repo.Repository{Name: "alpha", IsArchived: true},
		repo.Repository{Name: "bravo"},
	)
	m.Update(filters.FiltersMsg(filter.FilterMap{"Is Archived": filter.NewBoolFilter("Is Archived", true)}))

	assert.Contains(t, uitest.Plain(m.View()), "1/1")
	assert.Equal(t, "1 of 2 repos · 1 filter", m.Status())
}

func TestModel_View_Matrix(t *testing.T) {
	m := newLoadedModel(
		repo.Repository{Name: "alpha", HasWikiEnabled: true},
		repo.Repository{Name: "bravo"},
	)
	m.SetDimensions(100, 20)
	m.mode = matrixView
	m.repoModel.SelectTab(3) // Features

	content := uitest.Plain(m.View())

	assert.Contains(t, content, "Wiki", "properties are columns, without the shared prefix")
	assert.Contains(t, content, "Discussions")
	assert.NotContains(t, content, "Has Wiki Enabled  ", "headings are shortened")
	assert.Contains(t, content, "1/2", "the wiki column counts how many repos have it on")
	assert.Contains(t, content, "alpha · Has Issues Enabled = ✗ no", "the focused cell is spelled out")
}

func TestModel_View_MatrixScrollsColumns(t *testing.T) {
	m := newLoadedModel(repo.Repository{Name: "alpha"})
	m.SetDimensions(50, 20)
	m.mode = matrixView
	m.repoModel.SelectTab(1) // Status has more columns than fit in 50 cells

	assert.Contains(t, uitest.Plain(m.View()), "›", "hidden columns to the right are marked")
	assert.NotContains(t, uitest.Plain(m.View()), "‹")

	m.repoModel.SelectProperty(len(m.repoModel.ActiveGroup().Properties) - 1)
	content := uitest.Plain(m.View())
	assert.Contains(t, content, "Reason", "the focused column (Lock Reason) is scrolled into view")
	assert.Contains(t, content, "‹", "hidden columns to the left are marked")
}

func TestModel_View_NoRepositories(t *testing.T) {
	m := newOrgModel()
	m.progress.SetPercent(1.0)

	content := uitest.Plain(m.View())

	assert.Contains(t, content, "No repositories found")
}

func TestModel_Breadcrumb(t *testing.T) {
	assert.Equal(t, "acme", NewModel(&githubtest.Fake{}, ui.OrgKey{Name: "acme"}, 80, 24).Breadcrumb())
	assert.Equal(t, "octocat", NewModel(&githubtest.Fake{}, ui.OrgKey{Name: "octocat", IsUser: true}, 80, 24).Breadcrumb())
	assert.Equal(t, "Repositories", NewModel(&githubtest.Fake{}, ui.OrgKey{}, 80, 24).Breadcrumb())
}

func TestModel_Status(t *testing.T) {
	loading := newOrgModel()
	assert.Equal(t, "loading repositories", loading.Status())

	one := newLoadedModel(repo.Repository{Name: "alpha"})
	assert.Equal(t, "1 repo", one.Status())

	m := newLoadedModel(repo.Repository{Name: "alpha", IsArchived: true}, repo.Repository{Name: "bravo"})
	assert.Equal(t, "2 repos", m.Status())

	m.Update(filters.FiltersMsg(filter.FilterMap{"Is Archived": filter.NewBoolFilter("Is Archived", true)}))
	assert.Equal(t, "1 of 2 repos · 1 filter", m.Status())
}

func TestModel_FilterKeySendsRepos(t *testing.T) {
	m := newLoadedModel(repo.Repository{Name: "alpha"})

	_, cmd := m.Update(tea.KeyPressMsg{Code: 'f', Text: "f"})

	open := cmd().(filters.OpenFiltersMsg)
	assert.Len(t, open.Repos, 1, "the filter screen counts matches against the loaded repos")
}

func TestModel_Help(t *testing.T) {
	m := newLoadedModel(repo.Repository{Name: "alpha"})

	assert.Equal(t, "j/k repo  enter inspect  tab group  v matrix  f filters  esc back", m.Help().String())

	m.setInspecting(true)
	assert.Equal(t, "j/k property  tab group  v matrix  f filters  h/esc repos", m.Help().String(),
		"the way back out of the inspector comes last")
	m.setInspecting(false)

	m.mode = matrixView
	assert.Equal(t, "hjkl move  enter inspect  tab group  v list  f filters  esc back", m.Help().String())
}

func TestModel_Help_NothingToBrowse(t *testing.T) {
	loading := newOrgModel()
	assert.Equal(t, "f filters  esc back", loading.Help().String())

	m := newLoadedModel(repo.Repository{Name: "alpha"})
	m.Update(filters.FiltersMsg(filter.FilterMap{"Is Archived": filter.NewBoolFilter("Is Archived", true)}))
	assert.Equal(t, "f filters  esc back", m.Help().String(), "everything is filtered out")
}

func TestModel_Help_FullListsEveryKey(t *testing.T) {
	m := newLoadedModel(repo.Repository{Name: "alpha"})

	for _, mode := range []viewMode{listView, matrixView} {
		m.mode = mode
		var keys []string
		for _, column := range m.Help().FullHelp() {
			for _, b := range column {
				keys = append(keys, b.Keys()...)
			}
		}
		for _, want := range []string{"pgup", "pgdown", "g", "G", "tab", "shift+tab", "v", "f", "esc"} {
			assert.Contains(t, keys, want, "mode %d", mode)
		}
	}
}

func TestModel_InspectNeedsARepo(t *testing.T) {
	m := newLoadedModel(repo.Repository{Name: "alpha"})
	m.Update(filters.FiltersMsg(filter.FilterMap{"Is Archived": filter.NewBoolFilter("Is Archived", true)}))

	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m.Update(tea.KeyPressMsg{Code: 'l', Text: "l"})

	assert.False(t, m.inspecting, "there is no repo to inspect")
}

func TestModel_ListUsesKeyMapBindings(t *testing.T) {
	m := newLoadedModel(repo.Repository{Name: "alpha"}, repo.Repository{Name: "bravo"})

	m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	assert.Equal(t, 1, m.list.Index)
	m.Update(tea.KeyPressMsg{Code: 'k', Text: "k"})
	assert.Equal(t, 0, m.list.Index)
}

func TestScrollbar(t *testing.T) {
	strip := func(lines []string) string { return ansi.Strip(strings.Join(lines, "")) }

	assert.Equal(t, "│││││", strip(scrollbar(0, 3, 2)), "everything fits, so there is no thumb")
	assert.Equal(t, "│┃┃│││││││││", strip(scrollbar(0, 10, 50)))
	assert.Equal(t, "│││││││││┃┃│", strip(scrollbar(40, 10, 50)), "scrolled to the end")
}

func TestSplitHeading(t *testing.T) {
	assert.Equal(t, [2]string{"", "Wiki"}, splitHeading("Wiki"))
	assert.Equal(t, [2]string{"Vulnerability", "Alerts"}, splitHeading("Vulnerability Alerts"))
	assert.Equal(t, [2]string{"Delete Branch", "On Merge"}, splitHeading("Delete Branch On Merge"))
}
