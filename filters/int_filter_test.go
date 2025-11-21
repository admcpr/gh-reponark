package filters

import (
	"gh-reponark/repo"
	"testing"

	"github.com/stretchr/testify/suite"
)

type IntFilterSuite struct {
	suite.Suite
	filter IntFilter
}

func (s *IntFilterSuite) SetupTest() {
	s.filter = NewIntFilter("test", 1, 100)
}

func (s *IntFilterSuite) TestNewIntFilter() {
	filter := NewIntFilter("stars", 0, 1000)
	s.Equal("stars", filter.Name())
	s.Equal(0, filter.From)
	s.Equal(1000, filter.To)
}

func (s *IntFilterSuite) TestGetName() {
	s.Equal("test", s.filter.Name())
}

func (s *IntFilterSuite) TestMatches() {
	tests := []struct {
		name     string
		property repo.RepoProperty
		want     bool
	}{
		{
			name: "value within range",
			property: repo.RepoProperty{
				Type:  "int",
				Value: 50,
			},
			want: true,
		},
		{
			name: "value at lower bound",
			property: repo.RepoProperty{
				Type:  "int",
				Value: 1,
			},
			want: true,
		},
		{
			name: "value at upper bound",
			property: repo.RepoProperty{
				Type:  "int",
				Value: 100,
			},
			want: true,
		},
		{
			name: "value below range",
			property: repo.RepoProperty{
				Type:  "int",
				Value: 0,
			},
			want: false,
		},
		{
			name: "value above range",
			property: repo.RepoProperty{
				Type:  "int",
				Value: 101,
			},
			want: false,
		},
		{
			name: "invalid property type",
			property: repo.RepoProperty{
				Type:  "string",
				Value: "50",
			},
			want: false,
		},
		{
			name: "negative value",
			property: repo.RepoProperty{
				Type:  "int",
				Value: -1,
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

func TestIntFilterSuite(t *testing.T) {
	suite.Run(t, new(IntFilterSuite))
}
