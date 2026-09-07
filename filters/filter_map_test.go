package filters

import (
	"gh-reponark/repo"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestRepo(name string, archived bool, stars int, language string) repo.RepoConfig {
	return repo.RepoConfig{
		Name: name,
		Properties: map[string]repo.RepoProperty{
			"Is Archived":      {Name: "Is Archived", Type: "bool", Value: archived},
			"Stargazer Count":  {Name: "Stargazer Count", Type: "int", Value: stars},
			"Primary Language": {Name: "Primary Language", Type: "string", Value: language},
		},
	}
}

func repoNames(repos []repo.RepoConfig) []string {
	names := make([]string, len(repos))
	for i, r := range repos {
		names[i] = r.Name
	}
	return names
}

func TestFilterMap_FilterRepos(t *testing.T) {
	repos := []repo.RepoConfig{
		newTestRepo("alpha", true, 10, "Go"),
		newTestRepo("bravo", false, 50, "Rust"),
		newTestRepo("charlie", true, 100, "golang"),
	}

	tests := []struct {
		name    string
		filters FilterMap
		want    []string
	}{
		{
			name:    "nil filter map returns every repo",
			filters: nil,
			want:    []string{"alpha", "bravo", "charlie"},
		},
		{
			name:    "empty filter map returns every repo",
			filters: FilterMap{},
			want:    []string{"alpha", "bravo", "charlie"},
		},
		{
			name:    "bool filter",
			filters: FilterMap{"Is Archived": NewBoolFilter("Is Archived", true)},
			want:    []string{"alpha", "charlie"},
		},
		{
			name:    "int filter",
			filters: FilterMap{"Stargazer Count": NewIntFilter("Stargazer Count", 20, 100)},
			want:    []string{"bravo", "charlie"},
		},
		{
			name: "multiple filters are combined with AND",
			filters: FilterMap{
				"Is Archived":     NewBoolFilter("Is Archived", true),
				"Stargazer Count": NewIntFilter("Stargazer Count", 20, 100),
			},
			want: []string{"charlie"},
		},
		{
			name:    "string filter is applied as a repo level filter",
			filters: FilterMap{"Primary Language": NewStringFilter("Primary Language", "go")},
			want:    []string{"alpha", "charlie"},
		},
		{
			name:    "filter on a missing property excludes every repo",
			filters: FilterMap{"Does Not Exist": NewBoolFilter("Does Not Exist", true)},
			want:    []string{},
		},
		{
			name:    "no matches returns an empty slice",
			filters: FilterMap{"Stargazer Count": NewIntFilter("Stargazer Count", 1000, 2000)},
			want:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filters.FilterRepos(repos)
			assert.NotNil(t, got)
			assert.Equal(t, tt.want, repoNames(got))
		})
	}
}

func TestFilterMap_FilterRepos_DoesNotModifyInput(t *testing.T) {
	repos := []repo.RepoConfig{
		newTestRepo("alpha", true, 10, "Go"),
		newTestRepo("bravo", false, 50, "Rust"),
	}
	filters := FilterMap{"Is Archived": NewBoolFilter("Is Archived", false)}

	filtered := filters.FilterRepos(repos)

	assert.Equal(t, []string{"bravo"}, repoNames(filtered))
	assert.Equal(t, []string{"alpha", "bravo"}, repoNames(repos))
}
