package filters

import (
	"fmt"
	"gh-reponark/repo"
)

type IntFilter struct {
	name string
	From int
	To   int
}

func NewIntFilter(name string, from, to int) IntFilter {
	return IntFilter{name: name, From: from, To: to}
}

func (f IntFilter) Name() string {
	return f.name
}

func (f IntFilter) Matches(property repo.RepoProperty) bool {
	if property.Type != "int" {
		return false
	}

	value := property.Value.(int)

	return value >= f.From && value <= f.To
}

func (f IntFilter) String() string {
	return fmt.Sprintf("%s between %d and %d", f.name, f.From, f.To)
}
