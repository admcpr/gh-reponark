package filters

import (
	"fmt"
	"gh-reponark/repo"
	"math"
)

// NoMin and NoMax leave one end of an IntFilter's range open.
const (
	NoMin = math.MinInt
	NoMax = math.MaxInt
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

func (f IntFilter) Condition() string {
	switch {
	case f.From == NoMin && f.To == NoMax:
		return "any"
	case f.From == NoMin:
		return fmt.Sprintf("≤ %d", f.To)
	case f.To == NoMax:
		return fmt.Sprintf("≥ %d", f.From)
	case f.From == f.To:
		return fmt.Sprintf("= %d", f.From)
	default:
		return fmt.Sprintf("%d – %d", f.From, f.To)
	}
}
