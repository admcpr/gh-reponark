package filters

import (
	"gh-reponark/repo"
	"testing"

	"github.com/stretchr/testify/suite"
)

type BoolFilterSuite struct {
	suite.Suite
	filter BoolFilter
}

func (s *BoolFilterSuite) SetupTest() {
	s.filter = NewBoolFilter("test", true)
}

func (s *BoolFilterSuite) TestNewBoolFilter() {
	filter := NewBoolFilter("example", true)
	s.Equal("example", filter.Name())
	s.True(filter.Value)
}

func (s *BoolFilterSuite) TestGetName() {
	s.Equal("test", s.filter.Name())
}

func (s *BoolFilterSuite) TestMatches() {
	tests := []struct {
		name     string
		property repo.RepoProperty
		want     bool
	}{
		{
			name:     "matching bool value",
			property: repo.RepoProperty{Type: "bool", Value: true},
			want:     true,
		},
		{
			name:     "non-matching bool value",
			property: repo.RepoProperty{Type: "bool", Value: false},
			want:     false,
		},
		{
			name:     "non-bool property type",
			property: repo.RepoProperty{Type: "string", Value: "true"},
			want:     false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.want, s.filter.Matches(tt.property))
		})
	}
}

func (s *BoolFilterSuite) TestString() {
	expected := "test = Yes"
	s.Equal(expected, s.filter.String())
}

func TestBoolFilterSuite(t *testing.T) {
	suite.Run(t, new(BoolFilterSuite))
}
