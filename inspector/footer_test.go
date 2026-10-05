package inspector

import (
	"gh-reponark/repo"
	"gh-reponark/ui/uitest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModel_SetOrgRepos_AddsBarsAndContext(t *testing.T) {
	m := newSelectedModel()
	m.SetDimensions(80, 36)
	m.SelectTab(2) // Metrics
	m.SelectProperty(2)
	before := uitest.Plain(m.View())

	m.SetOrgRepos([]repo.RepoConfig{
		newTestRepoConfig(),
		repo.NewRepoConfig(repo.Repository{Name: "popular", StargazerCount: 84}),
	})
	after := uitest.Plain(m.View())

	assert.NotContains(t, before, "━", "no bars before the org is known")
	assert.NotContains(t, before, "of 2 repos")
	assert.Contains(t, after, "      42  ━━━━━━━╌╌╌", "42 stars is half the org max: seven of ten cells on the square-root scale")
	assert.Contains(t, after, "more than 0 of 2 repos · max 84", "the footer puts the count in context")
	assert.NotContains(t, after, "p50", "no percentiles")

	m.SelectTab(1) // Status: Is Archived
	content := uitest.Plain(m.View())
	assert.Contains(t, content, "on in 1 of 2 repos  ━━━━━╌╌╌╌╌", "booleans get the share of repos with the setting on")

	m.SelectTab(0)
	m.SelectProperty(1) // Database ID
	content = uitest.Plain(m.View())
	assert.NotContains(t, content, "Database ID ▐count▌ ▴", "sanity: the title row is intact")
	assert.NotContains(t, content, "more than", "an identifier has no rank")
	row := strings.Split(content, "\n")
	var idRow string
	for _, line := range row {
		if strings.Contains(line, "▸ Database ID") {
			idRow = line
		}
	}
	assert.NotContains(t, idRow, "━", "and no bar")
}
