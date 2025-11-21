package filters

import (
	"gh-reponark/repo"
)

type Filter interface {
	Name() string
	Matches(property repo.RepoProperty) bool
	String() string
}
