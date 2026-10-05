package filters

import (
	"gh-reponark/ui/uitest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModel_View_WithoutRepos(t *testing.T) {
	content := uitest.Plain(NewModel(nil, nil, 100, 30).View())

	assert.Contains(t, content, "FILTER ")
	assert.NotContains(t, content, "ACROSS YOUR REPOS", "there is nothing to chart")
	assert.NotContains(t, content, "MATCHING")
}
