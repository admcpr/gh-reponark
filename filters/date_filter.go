package filters

import (
	"fmt"
	"gh-reponark/repo"
	"time"
)

// DateFilter matches dates between From and To inclusive. A zero From or To
// leaves that end of the range open.
type DateFilter struct {
	name string
	From time.Time
	To   time.Time
}

func NewDateFilter(name string, from, to time.Time) DateFilter {
	return DateFilter{name: name, From: from, To: to}
}

func (f DateFilter) Name() string {
	return f.name
}

func (f DateFilter) Matches(property repo.RepoProperty) bool {
	if property.Type != "time.Time" {
		return false
	}

	date := property.Value.(time.Time)

	afterFrom := f.From.IsZero() || !date.Before(f.From)
	beforeTo := f.To.IsZero() || !date.After(f.To)
	return afterFrom && beforeTo
}

func (f DateFilter) String() string {
	return fmt.Sprintf("%s between %s and %s", f.name, f.From.Format("2006-01-02"), f.To.Format("2006-01-02"))
}

func (f DateFilter) Condition() string {
	const layout = "2006-01-02"
	switch {
	case f.From.IsZero() && f.To.IsZero():
		return "any"
	case f.From.IsZero():
		return "before " + f.To.Format(layout)
	case f.To.IsZero():
		return "after " + f.From.Format(layout)
	default:
		return f.From.Format(layout) + " – " + f.To.Format(layout)
	}
}
