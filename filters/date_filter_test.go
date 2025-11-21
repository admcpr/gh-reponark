package filters

import (
	"gh-reponark/repo"
	"testing"
	"time"

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
