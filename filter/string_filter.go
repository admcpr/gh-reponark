package filter

import (
	"fmt"

	"gh-reponark/repo"
)

// StringFilter matches string properties via substring (case-insensitive),
// or, when built with NewExactStringFilter, only the exact value.
type StringFilter struct {
	name  string
	value string
	exact bool
}

func NewStringFilter(name, value string) StringFilter {
	return StringFilter{name: name, value: value}
}

// NewExactStringFilter matches only properties whose value is exactly
// value, so a filter for "Java" does not also keep "JavaScript". It is what
// the overview's slices use, since they count repositories by exact value.
func NewExactStringFilter(name, value string) StringFilter {
	return StringFilter{name: name, value: value, exact: true}
}

func (f StringFilter) Name() string  { return f.name }
func (f StringFilter) Value() string { return f.value }

// Matches implements property-level filtering.
func (f StringFilter) Matches(property repo.RepoProperty) bool {
	s, ok := property.Value.(string)
	if !ok {
		return false
	}
	if f.exact {
		return s == f.value
	}
	return containsCaseInsensitive(s, f.value)
}

func (f StringFilter) String() string {
	return f.value
}

// FilterRepos retains repos whose property value matches the filter.
func (f StringFilter) FilterRepos(repos []repo.RepoConfig) []repo.RepoConfig {
	filtered := make([]repo.RepoConfig, 0, len(repos))
	for _, repoCfg := range repos {
		if prop, ok := repoCfg.Properties[f.name]; ok && f.Matches(prop) {
			filtered = append(filtered, repoCfg)
		}
	}
	return filtered
}

func (f StringFilter) Condition() string {
	if f.exact {
		return fmt.Sprintf("is %q", f.value)
	}
	return fmt.Sprintf("contains %q", f.value)
}
