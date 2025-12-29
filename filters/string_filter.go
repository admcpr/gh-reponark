package filters

import "gh-reponark/repo"

// StringFilter matches string properties via substring (case-insensitive).
type StringFilter struct {
	name  string
	value string
}

func NewStringFilter(name, value string) StringFilter {
	return StringFilter{name: name, value: value}
}

func (f StringFilter) Name() string  { return f.name }
func (f StringFilter) Value() string { return f.value }

// Matches implements property-level filtering.
func (f StringFilter) Matches(property repo.RepoProperty) bool {
	s, ok := property.Value.(string)
	if !ok {
		return false
	}
	return containsCaseInsensitive(s, f.value)
}

func (f StringFilter) String() string {
	return f.value
}

// FilterRepos retains repos whose property value matches the filter value (substring match).
func (f StringFilter) FilterRepos(repos []repo.RepoConfig) []repo.RepoConfig {
	filtered := make([]repo.RepoConfig, 0, len(repos))
	for _, repoCfg := range repos {
		prop, ok := repoCfg.Properties[f.name]
		if !ok {
			continue
		}
		if s, ok := prop.Value.(string); ok && containsCaseInsensitive(s, f.value) {
			filtered = append(filtered, repoCfg)
		}
	}
	return filtered
}
