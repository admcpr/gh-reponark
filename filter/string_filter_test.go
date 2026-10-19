package filter

import (
	"gh-reponark/repo"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type StringFilterSuite struct {
	suite.Suite
	filter StringFilter
}

func (s *StringFilterSuite) SetupTest() {
	s.filter = NewStringFilter("test", "Go")
}

func (s *StringFilterSuite) TestNewStringFilter() {
	filter := NewStringFilter("language", "rust")
	s.Equal("language", filter.Name())
	s.Equal("rust", filter.Value())
}

func (s *StringFilterSuite) TestGetName() {
	s.Equal("test", s.filter.Name())
}

func (s *StringFilterSuite) TestMatches() {
	tests := []struct {
		name     string
		property repo.RepoProperty
		want     bool
	}{
		{
			name:     "exact match",
			property: repo.RepoProperty{Type: "string", Value: "Go"},
			want:     true,
		},
		{
			name:     "case insensitive match",
			property: repo.RepoProperty{Type: "string", Value: "gOLANG"},
			want:     true,
		},
		{
			name:     "substring match",
			property: repo.RepoProperty{Type: "string", Value: "Django"},
			want:     true,
		},
		{
			name:     "no match",
			property: repo.RepoProperty{Type: "string", Value: "Rust"},
			want:     false,
		},
		{
			name:     "empty value",
			property: repo.RepoProperty{Type: "string", Value: ""},
			want:     false,
		},
		{
			name:     "non-string property value",
			property: repo.RepoProperty{Type: "int", Value: 42},
			want:     false,
		},
		{
			name:     "nil property value",
			property: repo.RepoProperty{},
			want:     false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.want, s.filter.Matches(tt.property))
		})
	}
}

func (s *StringFilterSuite) TestEmptyFilterValueMatchesAnyString() {
	filter := NewStringFilter("test", "")
	s.True(filter.Matches(repo.RepoProperty{Type: "string", Value: "anything"}))
	s.True(filter.Matches(repo.RepoProperty{Type: "string", Value: ""}))
}

func (s *StringFilterSuite) TestString() {
	s.Equal("Go", s.filter.String())
}

func (s *StringFilterSuite) TestFilterRepos() {
	repos := []repo.RepoConfig{
		{Name: "matches", Properties: map[string]repo.RepoProperty{"test": {Value: "golang"}}},
		{Name: "no match", Properties: map[string]repo.RepoProperty{"test": {Value: "rust"}}},
		{Name: "missing property", Properties: map[string]repo.RepoProperty{}},
		{Name: "non-string property", Properties: map[string]repo.RepoProperty{"test": {Value: 1}}},
		{Name: "also matches", Properties: map[string]repo.RepoProperty{"test": {Value: "GO"}}},
	}

	filtered := s.filter.FilterRepos(repos)

	s.Len(filtered, 2)
	s.Equal("matches", filtered[0].Name)
	s.Equal("also matches", filtered[1].Name)
}

func (s *StringFilterSuite) TestFilterReposEmptyInput() {
	filtered := s.filter.FilterRepos([]repo.RepoConfig{})
	s.NotNil(filtered)
	s.Empty(filtered)
}

func TestStringFilterSuite(t *testing.T) {
	suite.Run(t, new(StringFilterSuite))
}

func TestStringFilter_Condition(t *testing.T) {
	assert.Equal(t, `contains "go"`, NewStringFilter("Primary Language", "go").Condition())
}

func TestExactStringFilter(t *testing.T) {
	f := NewExactStringFilter("Primary Language", "Java")

	assert.Equal(t, "Primary Language", f.Name())
	assert.Equal(t, "Java", f.Value())
	assert.True(t, f.Matches(repo.RepoProperty{Value: "Java"}))
	assert.False(t, f.Matches(repo.RepoProperty{Value: "JavaScript"}), "no substring match")
	assert.False(t, f.Matches(repo.RepoProperty{Value: "java"}), "case matters")
	assert.Equal(t, `is "Java"`, f.Condition())

	var repos []repo.RepoConfig
	for name, language := range map[string]string{"a": "Java", "b": "JavaScript"} {
		r := repo.Repository{Name: name}
		r.PrimaryLanguage.Name = language
		repos = append(repos, repo.NewRepoConfig(r))
	}
	kept := FilterMap{f.Name(): f}.FilterRepos(repos)
	assert.Len(t, kept, 1)
	assert.Equal(t, "a", kept[0].Name)
	assert.Len(t, FilterMap{f.Name(): NewStringFilter("Primary Language", "java")}.FilterRepos(repos), 2,
		"the substring filter is unchanged")
}
