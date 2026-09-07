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

// lastPage builds a page message that completes the listing.
func lastPage(repositories ...repo.Repository) repositoryPageMsg {
	return repositoryPageMsg(github.RepositoryPage{
		Repositories: repositories,
		TotalCount:   len(repositories),
		HasNextPage:  false,
	})
}

// newLoadedModel returns a model that has received all of its repositories.
func newLoadedModel(repositories ...repo.Repository) *Model {
	m := newOrgModel()
	m.Update(lastPage(repositories...))
	return m
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
				assert.Equal(t, fake.Repositories, msg.Repositories)
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

func TestModel_Update_SinglePage(t *testing.T) {
	m := newOrgModel()

	updated, cmd := m.Update(lastPage(repo.Repository{Name: "zulu"}, repo.Repository{Name: "alpha"}))

	assert.Same(t, m, updated)
	assert.NotNil(t, cmd)
	assert.Equal(t, 2, m.repoCount)
	assert.Equal(t, 1.0, m.progress.Percent())
	assert.Len(t, m.repos, 2)
	assert.Equal(t, "alpha", m.repos[0].Name, "repos should be sorted by name")
	assert.Equal(t, "zulu", m.repos[1].Name)

	items := m.repoList.Items()
	assert.Len(t, items, 2)
	assert.Equal(t, shared.SimpleItem("alpha"), items[0])
	assert.Equal(t, shared.SimpleItem("zulu"), items[1])
	assert.Equal(t, "Organization: demo ", m.repoList.Title)
}

func TestModel_Update_PartialPageRequestsNext(t *testing.T) {
	fake := &githubtest.Fake{}
	m := NewModel(fake, shared.OrgKey{Name: "acme"}, 80, 24)

	_, cmd := m.Update(repositoryPageMsg(github.RepositoryPage{
		Repositories: []repo.Repository{{Name: "bravo"}},
		TotalCount:   4,
		EndCursor:    "cursor-1",
		HasNextPage:  true,
	}))

	assert.Equal(t, 4, m.repoCount)
	assert.Len(t, m.repos, 1)
	assert.InDelta(t, 0.25, m.progress.Percent(), 0.0001)
	assert.Empty(t, m.repoList.Items(), "list should not be populated until every page arrives")

	if assert.NotNil(t, cmd) {
		batch, ok := cmd().(tea.BatchMsg)
		assert.True(t, ok, "expected a batch with the progress update and the next page fetch")
		assert.Len(t, batch, 2)
		// The second command fetches the next page using the returned cursor.
		batch[1]()
		assert.Equal(t, []githubtest.ListCall{{Login: "acme", IsUser: false, After: "cursor-1"}}, fake.ListCalls)
	}
}

func TestModel_LoadsAllPagesFromService(t *testing.T) {
	fake := &githubtest.Fake{
		Repositories: []repo.Repository{{Name: "delta"}, {Name: "alpha"}, {Name: "charlie"}, {Name: "bravo"}, {Name: "echo"}},
		PageSize:     2,
	}
	m := NewModel(fake, shared.OrgKey{Name: "acme"}, 80, 24)

	// Drive the command loop by hand: run each returned command and feed the
	// resulting message back until nothing else needs fetching.
	var run func(cmd tea.Cmd)
	run = func(cmd tea.Cmd) {
		if cmd == nil {
			return
		}
		switch msg := cmd().(type) {
		case tea.BatchMsg:
			for _, c := range msg {
				run(c)
			}
		case repositoryPageMsg:
			_, next := m.Update(msg)
			run(next)
		}
	}
	run(m.Init())

	assert.Equal(t, 5, m.repoCount)
	assert.Len(t, m.repos, 5)
	assert.Equal(t, 1.0, m.progress.Percent())
	assert.Equal(t, []githubtest.ListCall{
		{Login: "acme", After: ""},
		{Login: "acme", After: "2"},
		{Login: "acme", After: "4"},
	}, fake.ListCalls)

	names := make([]string, len(m.repos))
	for i, r := range m.repos {
		names[i] = r.Name
	}
	assert.Equal(t, []string{"alpha", "bravo", "charlie", "delta", "echo"}, names)
	assert.Len(t, m.repoList.Items(), 5)
}

func TestModel_Update_NoRepositories(t *testing.T) {
	m := newOrgModel()

	_, cmd := m.Update(lastPage())

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
	items := m.repoList.Items()
	assert.Len(t, items, 1)
	assert.Equal(t, shared.SimpleItem("archived"), items[0])
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
	m.Update(repositoryPageMsg(github.RepositoryPage{
		Repositories: []repo.Repository{{Name: "alpha"}},
		TotalCount:   2,
		EndCursor:    "cursor-1",
		HasNextPage:  true,
	}))

	content := plain(m.View())

	assert.Contains(t, content, "Getting repositories")
	assert.Contains(t, content, "1 of 2")
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
	assert.Contains(t, content, "next pane")
	assert.Contains(t, content, "back")
}

func TestModel_ProgressView(t *testing.T) {
	m := newOrgModel()

	content := plain(m.ProgressView())

	assert.Contains(t, content, "Getting repositories ... 0 of 0")
}

func TestInternalOrgKeyMap(t *testing.T) {
	keymap := orgKeyMap{}

	short := keymap.ShortHelp()
	assert.Len(t, short, 6)

	full := keymap.FullHelp()
	assert.Len(t, full, 1)
	assert.Equal(t, short, full[0])
}
