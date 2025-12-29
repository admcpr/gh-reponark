package filters

import (
	"fmt"
	"gh-reponark/repo"
	"time"
)

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

	return (date.After(f.From) || date.Equal(f.From)) && (date.Before(f.To) || date.Equal(f.To))
}

func (f DateFilter) String() string {
	return fmt.Sprintf("%s between %s and %s", f.name, f.From.Format("2006-01-02"), f.To.Format("2006-01-02"))
}
