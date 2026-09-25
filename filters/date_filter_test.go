package filters

import (
	"gh-reponark/repo"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DateFilterSuite struct {
	suite.Suite
	filter   DateFilter
	baseTime time.Time
	from     time.Time
	to       time.Time
}

func (s *DateFilterSuite) SetupTest() {
	s.baseTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	s.from = s.baseTime
	s.to = s.baseTime.AddDate(0, 1, 0) // one month later
	s.filter = NewDateFilter("test", s.from, s.to)
}

func (s *DateFilterSuite) TestNewDateFilter() {
	filter := NewDateFilter("example", s.from, s.to)
	s.Equal("example", filter.Name())
	s.Equal(s.from, filter.From)
	s.Equal(s.to, filter.To)
}

func (s *DateFilterSuite) TestGetName() {
	s.Equal("test", s.filter.Name())
}

func (s *DateFilterSuite) TestMatches() {
	tests := []struct {
		name     string
		property repo.RepoProperty
		want     bool
	}{
		{
			name: "date within range",
			property: repo.RepoProperty{
				Type:  "time.Time",
				Value: s.baseTime.AddDate(0, 0, 15),
			},
			want: true,
		},
		{
			name: "date before range",
			property: repo.RepoProperty{
				Type:  "time.Time",
				Value: s.baseTime.AddDate(0, 0, -1),
			},
			want: false,
		},
		{
			name: "date after range",
			property: repo.RepoProperty{
				Type:  "time.Time",
				Value: s.baseTime.AddDate(0, 2, 0),
			},
			want: false,
		},
		{
			name: "date on start boundary",
			property: repo.RepoProperty{
				Type:  "time.Time",
				Value: s.from,
			},
			want: true,
		},
		{
			name: "date on end boundary",
			property: repo.RepoProperty{
				Type:  "time.Time",
				Value: s.to,
			},
			want: true,
		},
		{
			name: "invalid property type",
			property: repo.RepoProperty{
				Type:  "string",
				Value: "2024-01-01",
			},
			want: false,
		},
		{
			name: "zero time",
			property: repo.RepoProperty{
				Type:  "time.Time",
				Value: time.Time{},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.want, s.filter.Matches(tt.property))
		})
	}
}

func TestDateFilterSuite(t *testing.T) {
	suite.Run(t, new(DateFilterSuite))
}

func (s *DateFilterSuite) TestString() {
	s.Equal("test between 2024-01-01 and 2024-02-01", s.filter.String())
}

func TestDateFilter_OpenBounds(t *testing.T) {
	cutoff := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	before := NewDateFilter("d", time.Time{}, cutoff)
	after := NewDateFilter("d", cutoff, time.Time{})
	early := repo.RepoProperty{Type: "time.Time", Value: cutoff.AddDate(-1, 0, 0)}
	late := repo.RepoProperty{Type: "time.Time", Value: cutoff.AddDate(1, 0, 0)}

	assert.True(t, before.Matches(early))
	assert.False(t, before.Matches(late))
	assert.False(t, after.Matches(early))
	assert.True(t, after.Matches(late))

	assert.Equal(t, "before 2024-01-01", before.Condition())
	assert.Equal(t, "after 2024-01-01", after.Condition())
	assert.Equal(t, "2024-01-01 – 2025-01-01", NewDateFilter("d", cutoff, cutoff.AddDate(1, 0, 0)).Condition())
	assert.Equal(t, "any", NewDateFilter("d", time.Time{}, time.Time{}).Condition())
}
