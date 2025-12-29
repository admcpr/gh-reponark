package filters

import (
	"gh-reponark/repo"
)

type Filter interface {
	Name() string
	Matches(property repo.RepoProperty) bool
	String() string
}

type RepoFilter interface {
	FilterRepos([]repo.RepoConfig) []repo.RepoConfig
}
