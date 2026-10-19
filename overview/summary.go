package overview

import (
	"gh-reponark/repo"
)

// Summary is a toggle or text property aggregated across the
// repositories: for toggles, how many are on with the names either side;
// for text, its values tallied, how many are distinct and how many unset.
type Summary struct {
	Name  string
	Type  string
	Total int

	On       int
	OnNames  []string
	OffNames []string

	Texts    []ValueCount
	Distinct int
	Unset    int
}

// Summarize aggregates property p, a toggle or text, across repos.
func Summarize(repos []repo.RepoConfig, p repo.PropertySchema) Summary {
	s := Summary{Name: p.Name, Type: p.Type, Total: len(repos)}
	switch p.Type {
	case "bool":
		for _, c := range repos {
			if c.Bool(p.Name) {
				s.On++
				s.OnNames = append(s.OnNames, c.Name)
			} else {
				s.OffNames = append(s.OffNames, c.Name)
			}
		}
	case "string":
		s.Texts = ValueCounts(repos, p.Name)
		for _, v := range s.Texts {
			if v.Value == "" {
				s.Unset = v.Count
			} else {
				s.Distinct++
			}
		}
	}
	return s
}

// Top is the most common set value of a text property, or nothing when
// every repository leaves it unset.
func (s Summary) Top() (ValueCount, bool) {
	for _, v := range s.Texts {
		if v.Value != "" {
			return v, true
		}
	}
	return ValueCount{}, false
}

// AllDistinct says whether no two repositories share a set value, which
// marks a text property as a name or address rather than a setting.
func (s Summary) AllDistinct() bool {
	return s.Distinct > 1 && s.Distinct == s.Total-s.Unset
}

// Minority is the shorter of a toggle's two sides, with the word for it, so
// a row can list the repositories that differ from the rest.
func (s Summary) Minority() (side string, names []string) {
	if s.On*2 > s.Total {
		return "off", s.OffNames
	}
	return "on", s.OnNames
}
