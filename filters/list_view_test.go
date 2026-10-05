package filters

import (
	"gh-reponark/ui/uitest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModel_ListScrollsToCursor(t *testing.T) {
	m := NewModel(nil, nil, 100, 10)

	press(m, "G")

	content := uitest.Plain(m.View())
	assert.Contains(t, content, m.properties[len(m.properties)-1].Name)
	assert.NotContains(t, content, "Database ID")
}

func TestModel_View_NoFilters(t *testing.T) {
	assert.Contains(t, uitest.Plain(newTestModel().View()), "No filters yet: every repo is shown")
}

func TestMatchingNames(t *testing.T) {
	repos := testRepos()

	assert.Equal(t, []string{"● old  ● new  ● busy"}, uitest.StripAll(matchingNames(repos, 40, 5)))
	assert.Equal(t, []string{"● old  ● new", "● busy"}, uitest.StripAll(matchingNames(repos, 14, 5)), "names wrap")
	assert.Equal(t, []string{"● old  +2 more"}, uitest.StripAll(matchingNames(repos, 6, 1)), "extra names are counted")
	assert.Equal(t, []string{"No repos match these filters"}, uitest.StripAll(matchingNames(nil, 40, 5)))
}
