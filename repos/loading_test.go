package repos

import (
	"errors"
	"fmt"
	"testing"

	"gh-reponark/github"
	"gh-reponark/github/githubtest"
	"gh-reponark/repo"
	"gh-reponark/ui"
	"gh-reponark/ui/uitest"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

func TestModel_Init_LoadsFirstPage(t *testing.T) {
	tests := []struct {
		name string
		key  ui.OrgKey
	}{
		{name: "organization", key: ui.OrgKey{Name: "acme", IsUser: false}},
		{name: "user", key: ui.OrgKey{Name: "octocat", IsUser: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &githubtest.Fake{Repositories: []repo.Repository{{Name: "widgets"}, {Name: "gadgets"}}}
			m := NewModel(fake, tt.key, 80, 24)

			cmd := m.Init()

			if assert.NotNil(t, cmd) {
				msg, ok := uitest.LoadMsg(cmd).(repositoryPageMsg)
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
	m := NewModel(fake, ui.OrgKey{Name: "acme"}, 80, 24)

	msg, ok := uitest.LoadMsg(m.Init()).(ui.ErrorMsg)

	assert.True(t, ok, "expected an ErrorMsg")
	assert.EqualError(t, msg.Err, "no such org")
}

func TestModel_LoadRepositoryBatch(t *testing.T) {
	fake := &githubtest.Fake{Repositories: []repo.Repository{{Name: "widgets", IsArchived: true}}}
	m := NewModel(fake, ui.OrgKey{Name: "acme"}, 80, 24)

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
	m := NewModel(fake, ui.OrgKey{Name: "acme"}, 80, 24)

	msg, ok := m.loadRepositoryBatch([]string{"widgets"})().(ui.ErrorMsg)

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
	m := NewModel(fake, ui.OrgKey{Name: "acme"}, 80, 24)

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
			m := NewModel(fake, ui.OrgKey{Name: "acme"}, 80, 24)

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
	m := NewModel(fake, ui.OrgKey{Name: "acme"}, 80, 24)
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
	m := NewModel(fake, ui.OrgKey{Name: "acme"}, 80, 24)

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
	assert.Contains(t, uitest.Plain(m.View()), "No repositories found")
}

func TestModel_Update_ProgressFrameMsg(t *testing.T) {
	m := newOrgModel()

	updated, _ := m.Update(progress.FrameMsg{})

	assert.Same(t, m, updated)
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

	content := uitest.Plain(m.View())

	assert.Contains(t, content, "Fetching settings")
	assert.Contains(t, content, "10 of 25")
}

func TestModel_ProgressView(t *testing.T) {
	m := newOrgModel()

	content := uitest.Plain(m.ProgressView())

	assert.Contains(t, content, "Listing repositories…")
}
