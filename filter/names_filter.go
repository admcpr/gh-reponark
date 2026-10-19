package filter

import (
	"fmt"
	"sort"
	"strings"

	"gh-reponark/repo"
)

// NamesFilter keeps the repositories with the given names. It is how a
// selection that no property filter can express, such as "the repositories
// in this histogram bucket", is carried into the filter map: keyed on the
// Name property, so the filter screen shows it as a chip there. A label
// says where the selection came from, e.g. "Languages › other".
type NamesFilter struct {
	names map[string]bool
	label string
}

// NewNamesFilter builds a filter keeping exactly names.
func NewNamesFilter(names []string) NamesFilter {
	return NewLabelledNamesFilter("", names)
}

// NewLabelledNamesFilter builds a filter keeping exactly names, described by
// label rather than only by how many names it keeps.
func NewLabelledNamesFilter(label string, names []string) NamesFilter {
	set := make(map[string]bool, len(names))
	for _, name := range names {
		set[name] = true
	}
	return NamesFilter{names: set, label: label}
}

func (f NamesFilter) Name() string { return "Name" }

func (f NamesFilter) Matches(property repo.RepoProperty) bool {
	name, ok := property.Value.(string)
	return ok && f.names[name]
}

// Names lists the names kept, sorted.
func (f NamesFilter) Names() []string {
	names := make([]string, 0, len(f.names))
	for name := range f.names {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (f NamesFilter) String() string {
	return strings.Join(f.Names(), ", ")
}

func (f NamesFilter) Condition() string {
	n := len(f.names)
	switch {
	case f.label != "":
		return fmt.Sprintf("%s (%d %s)", f.label, n, plural(n, "repo", "repos"))
	case n == 1:
		return "is " + f.Names()[0]
	default:
		return fmt.Sprintf("one of %d repos", n)
	}
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
