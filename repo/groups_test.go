package repo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroups(t *testing.T) {
	groups := Groups()

	assert.Equal(t, []string{"Overview", "Status", "Metrics", "Features", "Merge", "Permissions", "Security"}, GroupTitles())
	assert.Equal(t, NewRepoConfig(Repository{}).GroupKeys, keysOf(groups), "same order as a repository's groups")

	total := 0
	for _, g := range groups {
		total += len(g.Properties)
		for _, p := range g.Properties {
			assert.Equal(t, g.Key, p.Group)
		}
	}
	assert.Equal(t, len(Schema()), total, "every property is in exactly one group")
	assert.Equal(t, "Id", groups[0].Properties[0].Name, "properties keep their declaration order")
}

func TestGroupTitle(t *testing.T) {
	assert.Equal(t, "Metrics", GroupTitle("3⟭ Metrics"))
	assert.Equal(t, "plain", GroupTitle("plain"))
}

func TestRepoConfig_Accessors(t *testing.T) {
	r := Repository{IsFork: true, StargazerCount: 7, Description: "hi"}
	c := NewRepoConfig(r)

	assert.True(t, c.Bool("Is Fork"))
	assert.Equal(t, 7, c.Int("Stargazer Count"))
	assert.Equal(t, "hi", c.Text("Description"))
	assert.True(t, c.Time("Pushed At").IsZero())

	assert.False(t, c.Bool("Stargazer Count"), "a property of another type reads as the zero value")
	assert.Equal(t, 0, c.Int("Missing"))
}

func TestNewRepoConfig_GroupsKeepDeclarationOrder(t *testing.T) {
	c := NewRepoConfig(Repository{})

	names := []string{}
	for _, p := range c.PropertyGroups["5⟭ Merge"] {
		names = append(names, p.Name)
	}

	assert.Equal(t, []string{"Merge Commit Allowed", "Rebase Merge Allowed", "Squash Merge Allowed", "Auto Merge Allowed", "Delete Branch On Merge", "Default Branch"}, names)
}

func keysOf(groups []Group) []string {
	keys := make([]string, len(groups))
	for i, g := range groups {
		keys[i] = g.Key
	}
	return keys
}
