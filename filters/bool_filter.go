package filters

import (
	"fmt"
	"gh-reponark/repo"
	"gh-reponark/shared"
)

type BoolFilter struct {
	name  string
	Value bool
}

func NewBoolFilter(name string, value bool) BoolFilter {
	return BoolFilter{name: name, Value: value}
}

func (f BoolFilter) Name() string {
	return f.name
}

func (f BoolFilter) Matches(property repo.RepoProperty) bool {
	if property.Type != "bool" {
		return false
	}

	return property.Value.(bool) == f.Value
}

func (f BoolFilter) String() string {
	return fmt.Sprintf("%s = %s", f.name, shared.YesNo(f.Value))
}
