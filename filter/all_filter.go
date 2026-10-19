package filter

import (
	"strings"

	"gh-reponark/repo"
)

// AllFilter keeps a repository only when every one of its filters, all on
// the same property, matches. It is how a new filter is added on top of one
// already applied to that property, rather than replacing it, so a
// selection can only narrow.
type AllFilter struct {
	name    string
	filters []Filter
}

// All combines filters on one property so that all of them must match.
// Nested combinations are flattened, and a single filter is returned as is.
func All(filters ...Filter) Filter {
	var flat []Filter
	for _, f := range filters {
		if all, ok := f.(AllFilter); ok {
			flat = append(flat, all.filters...)
		} else if f != nil {
			flat = append(flat, f)
		}
	}
	if len(flat) == 1 {
		return flat[0]
	}
	name := ""
	if len(flat) > 0 {
		name = flat[0].Name()
	}
	return AllFilter{name: name, filters: flat}
}

func (f AllFilter) Name() string { return f.name }

// Filters are the combined filters, in the order they were added.
func (f AllFilter) Filters() []Filter { return f.filters }

func (f AllFilter) Matches(property repo.RepoProperty) bool {
	for _, filter := range f.filters {
		if !filter.Matches(property) {
			return false
		}
	}
	return true
}

func (f AllFilter) String() string {
	parts := make([]string, len(f.filters))
	for i, filter := range f.filters {
		parts[i] = filter.String()
	}
	return strings.Join(parts, " and ")
}

func (f AllFilter) Condition() string {
	parts := make([]string, len(f.filters))
	for i, filter := range f.filters {
		parts[i] = filter.Condition()
	}
	return strings.Join(parts, " and ")
}
