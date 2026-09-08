package repo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchema(t *testing.T) {
	schema := Schema()

	assert.Len(t, schema, 49)
	assert.Equal(t, "Id", schema[0].Name, "declaration order should be preserved")
	assert.Equal(t, "Viewer Has Starred", schema[len(schema)-1].Name)

	byName := map[string]PropertySchema{}
	for _, p := range schema {
		byName[p.Name] = p
		assert.NotEmpty(t, p.Group, "%s should have a group", p.Name)
		assert.NotEmpty(t, p.Description, "%s should have a description", p.Name)
		assert.Contains(t, []string{"bool", "int", "string", "time.Time"}, p.Type, "%s has an unexpected type", p.Name)
	}

	assert.Equal(t, PropertySchema{
		Name:        "Primary Language",
		Group:       "1⟭ Overview",
		Type:        "string",
		Description: "The primary programming language of the repository.",
	}, byName["Primary Language"])
	assert.Equal(t, "time.Time", byName["Created At"].Type)
	assert.Equal(t, "bool", byName["Is Archived"].Type)
	assert.Equal(t, "int", byName["Open Pull Requests"].Type)
}

func TestSchema_MatchesToProperties(t *testing.T) {
	properties := ToProperties(Repository{})

	for _, p := range Schema() {
		actual, ok := properties[p.Name]
		if !assert.True(t, ok, "property %q missing from ToProperties", p.Name) {
			continue
		}
		assert.Equal(t, p.Group, actual.Group)
		assert.Equal(t, p.Type, actual.Type)
		assert.Equal(t, p.Description, actual.Description)
	}
}

func TestSchema_ReturnsCopy(t *testing.T) {
	first := Schema()
	first[0].Name = "mutated"

	assert.Equal(t, "Id", Schema()[0].Name)
}

func TestRepositoryProperties_Order(t *testing.T) {
	properties := repositoryProperties(Repository{})

	names := make([]string, 4)
	for i := range names {
		names[i] = properties[i].Name
	}

	assert.Equal(t, []string{"Id", "Database ID", "Name", "Name With Owner"}, names)
	assert.Len(t, properties, 49)
}
