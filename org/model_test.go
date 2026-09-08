package org

import (
	"errors"
	"fmt"
	"testing"

	"gh-reponark/filters"
	"gh-reponark/github"
	"gh-reponark/github/githubtest"
	"gh-reponark/repo"
	"gh-reponark/shared"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
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

// itemNames returns the repository names currently shown in the list.
func itemNames(m *Model) []string {
	items := m.repoList.Items()
	names := make([]string, len(items))
	for i, item := range items {
		names[i] = item.(repoItem).config.Name
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
	assert.Equal(t, "Organization: demo ", m.repoList.Title)

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
	assert.Len(t, m.repoList.Items(), 25)

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

func TestRepoItem(t *testing.T) {
	item := repoItem{config: repo.RepoConfig{Name: "widgets"}}

	assert.Equal(t, "widgets", item.String())
	assert.Equal(t, "", item.FilterValue(), "built-in list filtering stays disabled")
}

func TestModel_Update_FiltersMsg_BeforeReposLoad(t *testing.T) {
	m := newOrgModel()
	filterMap := filters.FilterMap{"Is Archived": filters.NewBoolFilter("Is Archived", true)}

	assert.NotPanics(t, func() { m.Update(filters.FiltersMsg(filterMap)) })

	assert.Equal(t, filterMap, m.filters, "filters should be kept for when the repos arrive")
	assert.Empty(t, m.repoList.Items())
}

func TestModel_Update_FiltersMsg_ClearingFiltersRestoresRepos(t *testing.T) {
	m := newLoadedModel(
		repo.Repository{Name: "archived", IsArchived: true},
		repo.Repository{Name: "active", IsArchived: false},
	)
	m.Update(filters.FiltersMsg(filters.FilterMap{"Is Archived": filters.NewBoolFilter("Is Archived", true)}))

	m.Update(filters.FiltersMsg(filters.FilterMap{}))

	assert.Len(t, m.repoList.Items(), 2)
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
	groupKeys := m.repos[0].GroupKeys

	m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	assert.Contains(t, plain(m.repoModel.HeaderView()), groupKeys[1])

	m.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	assert.Contains(t, plain(m.repoModel.HeaderView()), groupKeys[0])
}

func TestModel_Update_OtherKeysGoToRepoList(t *testing.T) {
	m := newLoadedModel(repo.Repository{Name: "alpha"}, repo.Repository{Name: "bravo"})
	assert.Equal(t, 0, m.repoList.Index())

	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})

	assert.Equal(t, 1, m.repoList.Index())
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
	assert.Contains(t, content, "Organization: demo")
}

func TestModel_View_NoRepositories(t *testing.T) {
	m := newOrgModel()
	m.progress.SetPercent(1.0)

	content := plain(m.View())

	assert.Contains(t, content, "No repositories found")
}

func TestModel_HeaderView(t *testing.T) {
	tests := []struct {
		name      string
		key       shared.OrgKey
		repoCount int
		want      string
	}{
		{name: "organization", key: shared.OrgKey{Name: "acme"}, want: "Organization: acme"},
		{name: "user", key: shared.OrgKey{Name: "octocat", IsUser: true}, want: "User: octocat"},
		{name: "organization with repo count", key: shared.OrgKey{Name: "acme"}, repoCount: 3, want: "Organization: acme (3 repos)"},
		{name: "empty title falls back", key: shared.OrgKey{Name: ""}, want: "Organization: Repositories"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(&githubtest.Fake{}, tt.key, 80, 24)
			m.repoCount = tt.repoCount

			assert.Contains(t, plain(m.HeaderView()), tt.want)
		})
	}
}

func TestModel_HelpView(t *testing.T) {
	m := newOrgModel()

	content := plain(m.HelpView())

	assert.Contains(t, content, "filters")
	assert.Contains(t, content, "next tab")
	assert.Contains(t, content, "prev tab")
	assert.Contains(t, content, "back")
}

func TestModel_HelpKeys_IncludeDetailPaneBindings(t *testing.T) {
	m := newOrgModel()

	keys := m.helpKeys()

	assert.Len(t, keys, 6)
	assert.Equal(t, m.repoModel.Keys().NextTab.Keys(), keys[2].Keys())
	assert.Equal(t, m.repoModel.Keys().PrevTab.Keys(), keys[3].Keys())
	assert.Equal(t, m.keymap.Filters.Keys(), keys[4].Keys())
}

func TestModel_ListUsesKeyMapBindings(t *testing.T) {
	m := newLoadedModel(repo.Repository{Name: "alpha"}, repo.Repository{Name: "bravo"})

	m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	assert.Equal(t, 1, m.repoList.Index())
	m.Update(tea.KeyPressMsg{Code: 'k', Text: "k"})
	assert.Equal(t, 0, m.repoList.Index())

	assert.Equal(t, m.keymap.Up.Keys(), m.repoList.KeyMap.CursorUp.Keys())
	assert.Equal(t, m.keymap.Down.Keys(), m.repoList.KeyMap.CursorDown.Keys())
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
}
