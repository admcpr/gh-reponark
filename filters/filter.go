package filters

import (
	"gh-reponark/repo"
)

type Filter interface {
	Name() string
	Matches(property repo.RepoProperty) bool
	String() string
	// Condition describes what the filter requires without naming the
	// property, e.g. "no", "10 – 100" or "contains go".
	Condition() string
}

type RepoFilter interface {
	FilterRepos([]repo.RepoConfig) []repo.RepoConfig
}
