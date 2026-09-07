package org

import (
	"fmt"
	"testing"

	"gh-reponark/filters"
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
	return NewModel(shared.OrgKey{Name: "demo", IsUser: false}, 80, 24)
}

func newOrgQueryMsg(names ...string) orgQueryMsg {
	query := OrgQuery{}
	for _, name := range names {
		query.Organization.Repositories.Nodes = append(query.Organization.Repositories.Nodes,
			struct {
				Name string
				Url  string
			}{Name: name})
	}
	return orgQueryMsg(query)
}

func newRepoQueryMsg(repository repo.Repository) repoQueryMsg {
	return repoQueryMsg(repo.Query{Repository: repository})
}

// newLoadedModel returns a model that has received all of its repositories.
func newLoadedModel(repositories ...repo.Repository) *Model {
	m := newOrgModel()
	names := make([]string, len(repositories))
	for i, r := range repositories {
		names[i] = r.Name
	}
	m.Update(newOrgQueryMsg(names...))
	for _, r := range repositories {
		m.Update(newRepoQueryMsg(r))
	}
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
			m := NewModel(tt.key, 80, 24)

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

func TestNewModel_PanicsOnWrongData(t *testing.T) {
	assert.Panics(t, func() { NewModel("not an org key", 80, 24) })
}

func TestModel_SetDimensions(t *testing.T) {
	m := newOrgModel()

	m.SetDimensions(120, 40)

	assert.Equal(t, 120, m.width)
	assert.Equal(t, 40, m.height)
	assert.Equal(t, 120, m.help.Width())
}

func TestModel_Init(t *testing.T) {
	m := newOrgModel()
	assert.NotNil(t, m.Init())
}

func TestModel_Update_OrgQueryMsg(t *testing.T) {
	m := newOrgModel()

	updated, cmd := m.Update(newOrgQueryMsg("alpha", "bravo"))

	assert.Same(t, m, updated)
	assert.Equal(t, 2, m.repoCount)
	assert.InDelta(t, 0.1, m.progress.Percent(), 0.0001)
	if assert.NotNil(t, cmd) {
		batch, ok := cmd().(tea.BatchMsg)
		assert.True(t, ok, "expected a batch of commands")
		assert.Len(t, batch, 3, "one progress command plus one fetch per repository")
	}
}

func TestModel_Update_OrgQueryMsg_NoRepositories(t *testing.T) {
	m := newOrgModel()

	_, cmd := m.Update(newOrgQueryMsg())

	assert.Equal(t, 0, m.repoCount)
	assert.NotNil(t, cmd, "the progress command should still be returned")
}

func TestModel_Update_RepoQueryMsg_Partial(t *testing.T) {
	m := newOrgModel()
	m.Update(newOrgQueryMsg("alpha", "bravo"))

	_, cmd := m.Update(newRepoQueryMsg(repo.Repository{Name: "bravo"}))

	assert.NotNil(t, cmd)
	assert.Len(t, m.repos, 1)
	assert.Less(t, m.progress.Percent(), 1.0)
	assert.InDelta(t, 0.55, m.progress.Percent(), 0.0001)
	assert.Empty(t, m.repoList.Items(), "list should not be populated until every repo arrives")
}

func TestModel_Update_RepoQueryMsg_Complete(t *testing.T) {
	m := newLoadedModel(repo.Repository{Name: "zulu"}, repo.Repository{Name: "alpha"})

	assert.Len(t, m.repos, 2)
	assert.Equal(t, "alpha", m.repos[0].Name, "repos should be sorted by name")
	assert.Equal(t, "zulu", m.repos[1].Name)
	assert.Equal(t, 1.0, m.progress.Percent())

	items := m.repoList.Items()
	assert.Len(t, items, 2)
	assert.Equal(t, shared.SimpleItem("alpha"), items[0])
	assert.Equal(t, shared.SimpleItem("zulu"), items[1])
	assert.Equal(t, "Organization: demo ", m.repoList.Title)
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
				next, ok := cmd().(shared.NextMsg)
				assert.True(t, ok, "expected a NextMsg")
				assert.Equal(t, m.filters, next.ModelData)
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

func TestModel_View_WhileLoading(t *testing.T) {
	m := newOrgModel()
	m.Update(newOrgQueryMsg("alpha", "bravo"))
	m.Update(newRepoQueryMsg(repo.Repository{Name: "alpha"}))

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
			m := NewModel(tt.key, 80, 24)
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
