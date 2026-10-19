// Package overview summarises a set of repositories as a whole, as cards:
// each one question about the organisation answered by a small chart, with
// the repositories behind every slice of it and the filters that select
// them. It states what is set and which repositories differ; which setting
// is right is the organisation's call, so nothing here is scored or ranked
// by severity.
package overview

import (
	"sort"
	"time"

	"gh-reponark/filter"
	"gh-reponark/repo"
)

// now is the clock ages are measured against; tests replace it.
var now = time.Now

// Tone says how a slice is coloured. Neutral slices take the accent; Good
// is the green of a setting that is on; Warn and Bad are only used where
// the rest of the application already uses them as a legend, such as the
// private and archived dots beside repository names.
type Tone int

const (
	Neutral Tone = iota
	Good
	Warn
	Bad
)

// Kind is the chart a card is drawn as.
type Kind int

const (
	// Bars is a list of rows, each a label, a bar of its share and a count.
	Bars Kind = iota
	// Histogram is a row of columns, one per bucket, labelled underneath.
	Histogram
	// Stacked is one bar split between the slices, with a legend.
	Stacked
)

// Slice is one part of a card: a label, the repositories it counts and the
// filters that select them. A slice with no Filters is selected by name.
type Slice struct {
	Label string
	Count int
	Tone  Tone
	Names []string
	// Filters select the slice's repositories; nil means by name.
	Filters []filter.Filter
	// Majority, when set, is the value most repositories share; the slice
	// is then the rest, described by Minority, e.g. Majority "off" and
	// Minority "on" for a toggle most repositories leave off. Count is
	// then how many share the majority value, and Names the rest.
	Majority string
	Minority string
	// InBar says whether a Stacked card draws the slice in its bar and
	// legend; the rest are listed underneath, grouped by Group.
	InBar bool
	Group string
}

// Card is one question about the repositories, answered by its slices.
type Card struct {
	Title string
	// Noun names what the card's rows are, for the help, e.g. "language"
	// or "setting": what j/k steps through.
	Noun string
	// Headline is the figure beside the title, e.g. "40 repos".
	Headline string
	Kind     Kind
	Slices   []Slice
	Total    int
	// Note is the line under the chart, e.g. "median 3w ago".
	Note string
}

// Largest is the biggest count among the slices, for scaling bars.
func (c Card) Largest() int {
	largest := 0
	for _, s := range c.Slices {
		largest = max(largest, s.Count)
	}
	return largest
}

// MaxRows is how many lines a card's chart has. A Bars card has at most
// this many slices, so every slice is drawn.
const MaxRows = 5

// Cards builds every card over repos, in display order.
func Cards(repos []repo.RepoConfig) []Card {
	return []Card{
		RepoMix(repos),
		LastPush(repos),
		Languages(repos, MaxRows),
		Licenses(repos, MaxRows),
		Security(repos),
		Alerts(repos),
		BranchRules(repos),
		RuleSources(repos),
		MergeHabits(repos),
		Features(repos),
		OtherSettings(repos, MaxRows),
	}
}

// count is one slice in the making: a label, the repositories that match a
// predicate and the filters that select them.
type count struct {
	label   string
	tone    Tone
	match   func(c repo.RepoConfig) bool
	filters []filter.Filter
	inBar   bool
	group   string
}

// slices counts repos into the given slices, keeping every slice so a
// card has the same rows whatever the organisation.
func slices(repos []repo.RepoConfig, counts []count) []Slice {
	out := make([]Slice, 0, len(counts))
	for _, k := range counts {
		s := Slice{Label: k.label, Tone: k.tone, Filters: k.filters, InBar: k.inBar, Group: k.group}
		for _, c := range repos {
			if k.match(c) {
				s.Count++
				s.Names = append(s.Names, c.Name)
			}
		}
		out = append(out, s)
	}
	return out
}

func on(name string) filter.Filter  { return filter.NewBoolFilter(name, true) }
func off(name string) filter.Filter { return filter.NewBoolFilter(name, false) }

// ValueCount is one value of a text property and the repositories that
// have it.
type ValueCount struct {
	Value string
	Count int
	Names []string
}

// ValueCounts tallies a text property's values, most common first and
// alphabetical among equals.
func ValueCounts(repos []repo.RepoConfig, property string) []ValueCount {
	byValue := map[string]*ValueCount{}
	for _, c := range repos {
		v := c.Text(property)
		vc, ok := byValue[v]
		if !ok {
			vc = &ValueCount{Value: v}
			byValue[v] = vc
		}
		vc.Count++
		vc.Names = append(vc.Names, c.Name)
	}
	values := make([]ValueCount, 0, len(byValue))
	for _, vc := range byValue {
		values = append(values, *vc)
	}
	sort.Slice(values, func(i, j int) bool {
		if values[i].Count != values[j].Count {
			return values[i].Count > values[j].Count
		}
		return values[i].Value < values[j].Value
	})
	return values
}

// Sum adds up an integer property across the repositories and names the
// one with the largest value.
func Sum(repos []repo.RepoConfig, property string) (sum int, top string, largest int) {
	for _, c := range repos {
		v := c.Int(property)
		sum += v
		if v > largest || top == "" {
			top, largest = c.Name, v
		}
	}
	return sum, top, largest
}
