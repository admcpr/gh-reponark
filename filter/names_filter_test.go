package filter

import (
	"testing"

	"gh-reponark/repo"

	"github.com/stretchr/testify/assert"
)

func TestNamesFilter(t *testing.T) {
	f := NewNamesFilter([]string{"bravo", "alpha"})

	assert.Equal(t, "Name", f.Name())
	assert.True(t, f.Matches(repo.RepoProperty{Value: "alpha"}))
	assert.False(t, f.Matches(repo.RepoProperty{Value: "charlie"}))
	assert.False(t, f.Matches(repo.RepoProperty{Value: 3}))
	assert.Equal(t, []string{"alpha", "bravo"}, f.Names())
	assert.Equal(t, "alpha, bravo", f.String())
	assert.Equal(t, "one of 2 repos", f.Condition())
	assert.Equal(t, "is alpha", NewNamesFilter([]string{"alpha"}).Condition())
	assert.Equal(t, "Languages › other (2 repos)", NewLabelledNamesFilter("Languages › other", []string{"a", "b"}).Condition())
	assert.Equal(t, "Last push › never (1 repo)", NewLabelledNamesFilter("Last push › never", []string{"a"}).Condition())

	repos := []repo.RepoConfig{
		repo.NewRepoConfig(repo.Repository{Name: "alpha"}),
		repo.NewRepoConfig(repo.Repository{Name: "charlie"}),
	}
	kept := FilterMap{"Name": f}.FilterRepos(repos)
	assert.Len(t, kept, 1)
	assert.Equal(t, "alpha", kept[0].Name)
}
