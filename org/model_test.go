package org

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"gh-reponark/filters"
	"gh-reponark/github"
	"gh-reponark/github/githubtest"
	"gh-reponark/repo"
	"gh-reponark/shared"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

// plain renders a view to a string with all ANSI styling removed so tests can
// assert on the visible text.
func plain(v tea.View) string {
	return ansi.Strip(fmt.Sprint(v.Content))
}

func newOrgModel() *Model {
	return NewModel(&githubtest.Fake{}, shared.OrgKey{Name: "demo", IsUser: false}, 80, 24)
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
		key    shared.OrgKey
		isUser bool
	}{
		{name: "organization", key: shared.OrgKey{Name: "acme", IsUser: false}, isUser: false},
		{name: "user", key: shared.OrgKey{Name: "octocat", IsUser: true}, isUser: true},
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
	assert.Equal(t, 120, m.help.Width())
}

func TestModel_Init_LoadsFirstPage(t *testing.T) {
	tests := []struct {
		name string
		key  shared.OrgKey
	}{
		{name: "organization", key: shared.OrgKey{Name: "acme", IsUser: false}},
		{name: "user", key: shared.OrgKey{Name: "octocat", IsUser: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &githubtest.Fake{Repositories: []repo.Repository{{Name: "widgets"}, {Name: "gadgets"}}}
			m := NewModel(fake, tt.key, 80, 24)

			cmd := m.Init()

			if assert.NotNil(t, cmd) {
				msg, ok := cmd().(repositoryPageMsg)
				assert.True(t, ok, "expected a repositoryPageMsg")
				assert.Equal(t, []github.RepositoryRef{{Name: "widgets"}, {Name: "gadgets"}}, msg.Repositories)
				assert.Equal(t, 2, msg.TotalCount)
				assert.False(t, msg.HasNextPage)
			}
			assert.Equal(t, []githubtest.ListCall{{Login: tt.key.Name, IsUser: tt.key.IsUser, After: ""}}, fake.ListCalls)
		})
	}
}

func TestModel_Init_ReportsErrors(t *testing.T) {
	fake := &githubtest.Fake{ListErr: errors.New("no such org")}
	m := NewModel(fake, shared.OrgKey{Name: "acme"}, 80, 24)

	msg, ok := m.Init()().(shared.ErrorMsg)

	assert.True(t, ok, "expected an ErrorMsg")
	assert.EqualError(t, msg.Err, "no such org")
}

func TestModel_LoadRepositoryBatch(t *testing.T) {
	fake := &githubtest.Fake{Repositories: []repo.Repository{{Name: "widgets", IsArchived: true}}}
	m := NewModel(fake, shared.OrgKey{Name: "acme"}, 80, 24)

	msg, ok := m.loadRepositoryBatch([]string{"widgets", "gadgets"})().(repositoryBatchMsg)

	assert.True(t, ok, "expected a repositoryBatchMsg")
	assert.Len(t, msg, 2)
	assert.Equal(t, "widgets", msg[0].Name)
	assert.True(t, msg[0].IsArchived)
	assert.Equal(t, "gadgets", msg[1].Name)
	assert.Equal(t, []githubtest.GetCall{{Owner: "acme", Names: []string{"widgets", "gadgets"}}}, fake.GetCalls)
}

func TestModel_LoadRepositoryBatch_ReportsErrors(t *testing.T) {
	fake := &githubtest.Fake{GetErr: errors.New("rate limited")}
	m := NewModel(fake, shared.OrgKey{Name: "acme"}, 80, 24)

	msg, ok := m.loadRepositoryBatch([]string{"widgets"})().(shared.ErrorMsg)

	assert.True(t, ok, "expected an ErrorMsg")
	assert.EqualError(t, msg.Err, "rate limited")
}

func TestBatchNames(t *testing.T) {
	tests := []struct {
		count     int
		wantSizes []int
	}{
		{count: 0, wantSizes: nil},
		{count: 1, wantSizes: []int{1}},
		{count: github.RepositoryBatchSize, wantSizes: []int{github.RepositoryBatchSize}},
		{count: github.RepositoryBatchSize + 1, wantSizes: []int{github.RepositoryBatchSize, 1}},
		{count: 25, wantSizes: []int{10, 10, 5}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d names", tt.count), func(t *testing.T) {
			batches := batchNames(numberedNames(tt.count))

			sizes := make([]int, 0, len(batches))
			flattened := []string{}
			for _, batch := range batches {
				sizes = append(sizes, len(batch))
				flattened = append(flattened, batch...)
			}
			if tt.wantSizes == nil {
				assert.Empty(t, sizes)
			} else {
				assert.Equal(t, tt.wantSizes, sizes)
			}
			assert.Equal(t, numberedNames(tt.count), flattened, "batches should preserve order and lose nothing")
		})
	}
}

func TestModel_Update_NamesPage_RequestsNextPage(t *testing.T) {
	fake := &githubtest.Fake{}
	m := NewModel(fake, shared.OrgKey{Name: "acme"}, 80, 24)

	_, cmd := m.Update(namesPage(4, true, "cursor-2", "alpha", "bravo"))

	assert.Equal(t, 4, m.repoCount)
	assert.Equal(t, []string{"alpha", "bravo"}, m.names)
	assert.Empty(t, m.pending, "no configuration is fetched until every name is known")
	if assert.NotNil(t, cmd) {
		cmd()
		assert.Equal(t, []githubtest.ListCall{{Login: "acme", After: "cursor-2"}}, fake.ListCalls)
	}
}

func TestModel_Update_NamesComplete_StartsBatches(t *testing.T) {
	tests := []struct {
		name         string
		count        int
		wantStarted  int
		wantPending  int
		wantInFlight int
	}{
		{name: "fewer batches than the limit", count: 25, wantStarted: 3, wantPending: 0, wantInFlight: 3},
		{name: "more batches than the limit", count: 50, wantStarted: maxConcurrentBatches, wantPending: 1, wantInFlight: maxConcurrentBatches},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &githubtest.Fake{}
			m := NewModel(fake, shared.OrgKey{Name: "acme"}, 80, 24)

			_, cmd := m.Update(lastNamesPage(numberedNames(tt.count)...))

			assert.Equal(t, tt.count, m.repoCount)
			assert.Len(t, m.pending, tt.wantPending)
			assert.Equal(t, tt.wantInFlight, m.inFlight)
			if assert.NotNil(t, cmd) {
				var started int
				if batch, ok := cmd().(tea.BatchMsg); ok {
					started = len(batch)
					for _, c := range batch {
						c()
					}
				} else {
					started = 1
				}
				assert.Equal(t, tt.wantStarted, started)
				assert.Len(t, fake.GetCalls, tt.wantStarted)
				assert.Equal(t, numberedNames(10), fake.GetCalls[0].Names)
			}
		})
	}
}

func TestModel_Update_BatchMsg_StartsNextPendingBatch(t *testing.T) {
	fake := &githubtest.Fake{}
	m := NewModel(fake, shared.OrgKey{Name: "acme"}, 80, 24)
	m.Update(lastNamesPage(numberedNames(50)...))
	assert.Len(t, m.pending, 1)

	repos := make([]repo.Repository, 10)
	for i, name := range numberedNames(10) {
		repos[i] = repo.Repository{Name: name}
	}
	_, cmd := m.Update(repositoryBatchMsg(repos))

	assert.Len(t, m.repos, 10)
	assert.Empty(t, m.pending, "a finished batch frees a slot for the last pending one")
	assert.Equal(t, maxConcurrentBatches, m.inFlight)
	assert.InDelta(t, 0.2, m.progress.Percent(), 0.0001)
	if assert.NotNil(t, cmd) {
		batch, ok := cmd().(tea.BatchMsg)
		assert.True(t, ok, "expected the progress update plus the next fetch")
		assert.Len(t, batch, 2)
	}
}

func TestModel_Update_AllBatchesLoaded(t *testing.T) {
	m := newOrgModel()
	m.Update(lastNamesPage("zulu", "alpha"))

	updated, cmd := m.Update(repositoryBatchMsg{{Name: "zulu"}, {Name: "alpha"}})

	assert.Same(t, m, updated)
	assert.NotNil(t, cmd)
	assert.Equal(t, 0, m.inFlight)
	assert.Equal(t, 1.0, m.progress.Percent())
	assert.Equal(t, []string{"alpha", "zulu"}, repoNames(m), "repos should be sorted by name")
	assert.Equal(t, []string{"alpha", "zulu"}, itemNames(m))

	selected, ok := m.selectedRepo()
	assert.True(t, ok)
	assert.Equal(t, "alpha", selected.Name)
}

func TestModel_LoadsEverythingFromService(t *testing.T) {
	names := numberedNames(25)
	repositories := make([]repo.Repository, len(names))
	for i, name := range names {
		repositories[len(names)-1-i] = repo.Repository{Name: name} // reverse order to exercise sorting
	}
	fake := &githubtest.Fake{Repositories: repositories, PageSize: 10}
	m := NewModel(fake, shared.OrgKey{Name: "acme"}, 80, 24)

	drive(m, m.Init())

	assert.Equal(t, 25, m.repoCount)
	assert.Equal(t, 1.0, m.progress.Percent())
	assert.Equal(t, 0, m.inFlight)
	assert.Empty(t, m.pending)
	assert.Equal(t, names, repoNames(m))
	assert.Len(t, m.visible, 25)

	assert.Equal(t, []githubtest.ListCall{
		{Login: "acme", After: ""},
		{Login: "acme", After: "10"},
		{Login: "acme", After: "20"},
	}, fake.ListCalls, "names should be listed page by page")

	assert.Len(t, fake.GetCalls, 3, "configurations should be fetched in batches")
	for _, call := range fake.GetCalls {
		assert.Equal(t, "acme", call.Owner)
		assert.LessOrEqual(t, len(call.Names), github.RepositoryBatchSize)
	}
}

func TestModel_Update_NoRepositories(t *testing.T) {
	m := newOrgModel()

	_, cmd := m.Update(lastNamesPage())

	assert.Equal(t, 0, m.repoCount)
	assert.NotNil(t, cmd)
	assert.Equal(t, 1.0, m.progress.Percent(), "loading should finish so the empty state is shown")
	assert.Contains(t, plain(m.View()), "No repositories found")
}

func TestModel_Update_FiltersMsg(t *testing.T) {
	m := newLoadedModel(
		repo.Repository{Name: "archived", IsArchived: true},
		repo.Repository{Name: "active", IsArchived: false},
	)

	filterMap := filters.FilterMap{"Is Archived": filters.NewBoolFilter("Is Archived", true)}
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
	m.Update(filters.FiltersMsg(filters.FilterMap{"Is Archived": filters.NewBoolFilter("Is Archived", true)}))
	assert.Equal(t, []string{"bravo", "charlie"}, itemNames(m))

	selected, ok := m.selectedRepo()
	assert.True(t, ok)
	assert.Equal(t, "bravo", selected.Name, "the first visible repo should be selected, not the first of all repos")

	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	selected, ok = m.selectedRepo()
	assert.True(t, ok)
	assert.Equal(t, "charlie", selected.Name)
	assert.Contains(t, plain(m.View()), "charlie")
}

func TestModel_View_AllFilteredOutShowsEmptyState(t *testing.T) {
	m := newLoadedModel(repo.Repository{Name: "alpha"}, repo.Repository{Name: "bravo"})
	m.Update(filters.FiltersMsg(filters.FilterMap{"Is Archived": filters.NewBoolFilter("Is Archived", true)}))

	_, ok := m.selectedRepo()
	assert.False(t, ok)
	assert.Empty(t, itemNames(m))
	assert.Contains(t, plain(m.View()), "No repositories found")
}

func TestModel_SelectedRepo_NoneBeforeLoad(t *testing.T) {
	m := newOrgModel()

	_, ok := m.selectedRepo()

	assert.False(t, ok)
}

func TestModel_Update_FiltersMsg_BeforeReposLoad(t *testing.T) {
	m := newOrgModel()
	filterMap := filters.FilterMap{"Is Archived": filters.NewBoolFilter("Is Archived", true)}

	assert.NotPanics(t, func() { m.Update(filters.FiltersMsg(filterMap)) })

	assert.Equal(t, filterMap, m.filters, "filters should be kept for when the repos arrive")
	assert.Empty(t, m.visible)
}

func TestModel_Update_FiltersMsg_ClearingFiltersRestoresRepos(t *testing.T) {
	m := newLoadedModel(
		repo.Repository{Name: "archived", IsArchived: true},
		repo.Repository{Name: "active", IsArchived: false},
	)
	m.Update(filters.FiltersMsg(filters.FilterMap{"Is Archived": filters.NewBoolFilter("Is Archived", true)}))

	m.Update(filters.FiltersMsg(filters.FilterMap{}))

	assert.Len(t, m.visible, 2)
}

func TestModel_Update_ProgressFrameMsg(t *testing.T) {
	m := newOrgModel()

	updated, _ := m.Update(progress.FrameMsg{})

	assert.Same(t, m, updated)
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
			m.filters = filters.FilterMap{"Is Fork": filters.NewBoolFilter("Is Fork", false)}

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
		prev, ok := cmd().(shared.PreviousMsg)
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
	assert.Equal(t, 1, m.cursor, "j moves the repo cursor from the list")
	assert.Equal(t, 0, m.repoModel.ActiveProperty())

	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m.Update(down)
	m.Update(down)
	assert.Equal(t, 1, m.cursor, "j moves the property cursor from the inspector")
	assert.Equal(t, 2, m.repoModel.ActiveProperty())

	m.Update(tea.KeyPressMsg{Code: 'G', Text: "G"})
	assert.Equal(t, len(m.repoModel.ActiveGroup().Properties)-1, m.repoModel.ActiveProperty())
	m.Update(tea.KeyPressMsg{Code: 'g', Text: "g"})
	assert.Equal(t, 0, m.repoModel.ActiveProperty())
	assert.Equal(t, 1, m.cursor)
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
	assert.Equal(t, 1, m.cursor)

	m.Update(tea.KeyPressMsg{Code: tea.KeyPgDown})
	assert.Equal(t, 1+m.listRows(), m.cursor, "page down moves a screenful")

	m.Update(tea.KeyPressMsg{Code: 'G', Text: "G"})
	assert.Equal(t, 39, m.cursor)
	assert.Equal(t, 40-m.listRows(), m.listOffset, "the list scrolls to keep the selection on screen")

	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	assert.Equal(t, 39, m.cursor, "the cursor stops at the last repository")

	m.Update(tea.KeyPressMsg{Code: 'g', Text: "g"})
	assert.Equal(t, 0, m.cursor)
	assert.Equal(t, 0, m.listOffset)

	m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	assert.Equal(t, 0, m.cursor, "the cursor stops at the first repository")
}

func TestModel_Update_CursorFollowsInspector(t *testing.T) {
	m := newLoadedModel(repo.Repository{Name: "alpha"}, repo.Repository{Name: "bravo"})

	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})

	assert.Contains(t, plain(m.repoModel.View()), "bravo")
}

func TestModel_Update_ToggleView(t *testing.T) {
	m := newLoadedModel(repo.Repository{Name: "alpha"}, repo.Repository{Name: "bravo"})
	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	assert.Equal(t, listView, m.mode)

	m.Update(tea.KeyPressMsg{Code: 'v', Text: "v"})
	assert.Equal(t, matrixView, m.mode)
	assert.Equal(t, 1, m.cursor, "switching views keeps the selection")

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

func TestModel_LoadedFraction(t *testing.T) {
	m := newOrgModel()
	assert.Equal(t, 1.0, m.loadedFraction(), "nothing to load counts as complete")

	m.repoCount = 4
	m.repos = []repo.RepoConfig{{Name: "a"}}
	assert.InDelta(t, 0.25, m.loadedFraction(), 0.0001)
}

func TestModel_View_WhileLoading(t *testing.T) {
	m := newOrgModel()
	m.Update(lastNamesPage(numberedNames(25)...))
	repos := make([]repo.Repository, 10)
	for i, name := range numberedNames(10) {
		repos[i] = repo.Repository{Name: name}
	}
	m.Update(repositoryBatchMsg(repos))

	content := plain(m.View())

	assert.Contains(t, content, "Getting repositories")
	assert.Contains(t, content, "10 of 25")
}

func TestModel_View_Loaded(t *testing.T) {
	m := newLoadedModel(repo.Repository{Name: "alpha"}, repo.Repository{Name: "bravo"})

	content := plain(m.View())

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
		lines := strings.Split(plain(m.View()), "\n")

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
	m.Update(filters.FiltersMsg(filters.FilterMap{"Is Archived": filters.NewBoolFilter("Is Archived", true)}))

	assert.Contains(t, plain(m.View()), "1/1")
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

	content := plain(m.View())

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

	assert.Contains(t, plain(m.View()), "›", "hidden columns to the right are marked")
	assert.NotContains(t, plain(m.View()), "‹")

	m.repoModel.SelectProperty(len(m.repoModel.ActiveGroup().Properties) - 1)
	content := plain(m.View())
	assert.Contains(t, content, "Template", "the focused column is scrolled into view")
	assert.Contains(t, content, "‹", "hidden columns to the left are marked")
}

func TestModel_View_NoRepositories(t *testing.T) {
	m := newOrgModel()
	m.progress.SetPercent(1.0)

	content := plain(m.View())

	assert.Contains(t, content, "No repositories found")
}

func TestModel_Breadcrumb(t *testing.T) {
	assert.Equal(t, "acme", NewModel(&githubtest.Fake{}, shared.OrgKey{Name: "acme"}, 80, 24).Breadcrumb())
	assert.Equal(t, "octocat", NewModel(&githubtest.Fake{}, shared.OrgKey{Name: "octocat", IsUser: true}, 80, 24).Breadcrumb())
	assert.Equal(t, "Repositories", NewModel(&githubtest.Fake{}, shared.OrgKey{}, 80, 24).Breadcrumb())
}

func TestModel_Status(t *testing.T) {
	loading := newOrgModel()
	assert.Equal(t, "loading repositories", loading.Status())

	one := newLoadedModel(repo.Repository{Name: "alpha"})
	assert.Equal(t, "1 repo", one.Status())

	m := newLoadedModel(repo.Repository{Name: "alpha", IsArchived: true}, repo.Repository{Name: "bravo"})
	assert.Equal(t, "2 repos", m.Status())

	m.Update(filters.FiltersMsg(filters.FilterMap{"Is Archived": filters.NewBoolFilter("Is Archived", true)}))
	assert.Equal(t, "1 of 2 repos · 1 filter", m.Status())
}

func TestModel_FilterKeySendsRepos(t *testing.T) {
	m := newLoadedModel(repo.Repository{Name: "alpha"})

	_, cmd := m.Update(tea.KeyPressMsg{Code: 'f', Text: "f"})

	open := cmd().(filters.OpenFiltersMsg)
	assert.Len(t, open.Repos, 1, "the filter screen counts matches against the loaded repos")
}

func TestModel_HelpView(t *testing.T) {
	m := newOrgModel()

	content := plain(m.HelpView())
	assert.Contains(t, content, "j/k repo")
	assert.Contains(t, content, "l/enter inspect")
	assert.Contains(t, content, "group")
	assert.Contains(t, content, "v matrix")
	assert.Contains(t, content, "filters")
	assert.Contains(t, content, "back")

	m.setInspecting(true)
	content = plain(m.HelpView())
	assert.Contains(t, content, "j/k property")
	assert.Contains(t, content, "h/esc repos")
	m.setInspecting(false)

	m.mode = matrixView
	content = plain(m.HelpView())
	assert.Contains(t, content, "column")
	assert.Contains(t, content, "enter open")
	assert.Contains(t, content, "v list")
	assert.Contains(t, content, "esc back", "the matrix footer fits in 80 columns")
}

func TestModel_HelpKeys_IncludeDetailPaneBindings(t *testing.T) {
	m := newOrgModel()
	repoKeys := m.repoModel.Keys()

	keys := m.helpKeys()

	assert.Len(t, keys, 6)
	assert.Equal(t, append(m.keymap.Right.Keys(), m.keymap.Inspect.Keys()...), keys[1].Keys())
	assert.Equal(t, append(repoKeys.NextTab.Keys(), repoKeys.PrevTab.Keys()...), keys[2].Keys())
	assert.Equal(t, m.keymap.Filters.Keys(), keys[4].Keys())
}

func TestModel_ListUsesKeyMapBindings(t *testing.T) {
	m := newLoadedModel(repo.Repository{Name: "alpha"}, repo.Repository{Name: "bravo"})

	m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	assert.Equal(t, 1, m.cursor)
	m.Update(tea.KeyPressMsg{Code: 'k', Text: "k"})
	assert.Equal(t, 0, m.cursor)
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

func TestShortName(t *testing.T) {
	assert.Equal(t, "Wiki", shortName("Has Wiki Enabled"))
	assert.Equal(t, "Archived", shortName("Is Archived"))
	assert.Equal(t, "Administer", shortName("Viewer Can Administer"))
	assert.Equal(t, "Stargazer", shortName("Stargazer Count"))
	assert.Equal(t, "Default Branch", shortName("Default Branch"))
}

func TestModel_ProgressView(t *testing.T) {
	m := newOrgModel()

	content := plain(m.ProgressView())

	assert.Contains(t, content, "Getting repositories ... 0 of 0")
}

func TestNewOrgKeyMap(t *testing.T) {
	keymap := newOrgKeyMap()

	assert.Equal(t, []string{"up", "k"}, keymap.Up.Keys())
	assert.Equal(t, []string{"down", "j"}, keymap.Down.Keys())
	assert.Equal(t, []string{"f", "F"}, keymap.Filters.Keys())
	assert.Equal(t, []string{"esc"}, keymap.Back.Keys())
	assert.Equal(t, []string{"v"}, keymap.ToggleView.Keys())
	assert.Equal(t, []string{"enter"}, keymap.Inspect.Keys())
	assert.Equal(t, []string{"left", "h"}, keymap.Left.Keys())
	assert.Equal(t, []string{"right", "l"}, keymap.Right.Keys())
}
