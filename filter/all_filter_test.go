package filter

import (
	"testing"

	"gh-reponark/repo"

	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
	atLeastOne := NewIntFilter("Vulnerability Alerts", 1, NoMax)
	none := NewIntFilter("Vulnerability Alerts", 0, 0)
	upToThree := NewIntFilter("Vulnerability Alerts", NoMin, 3)

	assert.Equal(t, atLeastOne, All(atLeastOne), "one filter is returned as is")
	assert.Equal(t, atLeastOne, All(nil, atLeastOne))

	f := All(atLeastOne, upToThree)
	assert.Equal(t, "Vulnerability Alerts", f.Name())
	assert.True(t, f.Matches(repo.RepoProperty{Type: "int", Value: 2}))
	assert.False(t, f.Matches(repo.RepoProperty{Type: "int", Value: 5}))
	assert.False(t, f.Matches(repo.RepoProperty{Type: "int", Value: 0}))
	assert.Equal(t, "≥ 1 and ≤ 3", f.Condition())

	assert.False(t, All(atLeastOne, none).Matches(repo.RepoProperty{Type: "int", Value: 0}), "contradictory filters match nothing")
	assert.Len(t, All(f, none).(AllFilter).Filters(), 3, "nested combinations flatten")
}
