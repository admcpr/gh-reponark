package filters

import (
	"testing"

	"gh-reponark/repo"
	"gh-reponark/shared"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

func newSearchModelWithProperties(t *testing.T) FilterSearchModel {
	t.Helper()

	m := NewFilterSearchModel()
	msg := filtersListMsg(repo.RepoConfig{
		Properties: map[string]repo.RepoProperty{
			"Name":      {Name: "Name", Type: "string", Description: "The name of the repository."},
			"Url":       {Name: "Url", Type: "string", Description: "The HTTP URL."},
			"Languages": {Name: "Languages", Type: "[]string", Description: "Unsupported"},
		},
	})

	updated, _ := m.Update(msg)
	return updated.(FilterSearchModel)
}

func typeInto(m FilterSearchModel, text string) FilterSearchModel {
	for _, r := range text {
		updated, _ := m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = updated.(FilterSearchModel)
	}
	return m
}

func TestNewFilterSearchModel(t *testing.T) {
	m := NewFilterSearchModel()

	assert.True(t, m.textinput.Focused())
	assert.True(t, m.textinput.ShowSuggestions)
	assert.Equal(t, "Type to search", m.textinput.Placeholder)
	assert.Equal(t, "Add filter: ", m.textinput.Prompt)
	assert.Empty(t, m.properties)
}

func TestFilterSearchModel_Init(t *testing.T) {
	m := NewFilterSearchModel()
	assert.NotNil(t, m.Init())
}

func TestGetFilters(t *testing.T) {
	msg, ok := getFilters().(filtersListMsg)

	assert.True(t, ok, "getFilters should return a filtersListMsg")
	assert.Len(t, msg.Properties, 49)
}

func TestFilterSearchModel_Update_FiltersListMsg(t *testing.T) {
	m := newSearchModelWithProperties(t)

	assert.Len(t, m.properties, 2, "only supported property types should be registered")
	assert.Contains(t, m.properties, "Name")
	assert.Contains(t, m.properties, "Url")
	assert.NotContains(t, m.properties, "Languages")
	assert.ElementsMatch(t, []string{"Name", "Url"}, m.textinput.AvailableSuggestions())
	assert.Equal(t, Property{Name: "Name", Description: "The name of the repository.", Type: "string"}, m.properties["Name"])
}

func TestFilterSearchModel_SuggestionLookups(t *testing.T) {
	m := newSearchModelWithProperties(t)

	// Nothing typed yet so there is no current suggestion.
	_, exists := m.CurrentPropertySuggestion()
	assert.False(t, exists)
	assert.Equal(t, "", m.LookupDescription())

	m = typeInto(m, "N")

	prop, exists := m.CurrentPropertySuggestion()
	assert.True(t, exists)
	assert.Equal(t, "Name", prop.Name)
	assert.Equal(t, "The name of the repository.", m.LookupDescription())

	// Typing something that matches nothing clears the suggestion.
	m = typeInto(m, "zzz")
	_, exists = m.CurrentPropertySuggestion()
	assert.False(t, exists)
	assert.Equal(t, "", m.LookupDescription())
}

func TestFilterSearchModel_Update_EnterWithoutSuggestion(t *testing.T) {
	m := newSearchModelWithProperties(t)

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.Nil(t, cmd)
}

func TestFilterSearchModel_Update_EnterWithSuggestion(t *testing.T) {
	m := typeInto(newSearchModelWithProperties(t), "U")

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.NotNil(t, cmd)

	next, ok := cmd().(shared.NextMsg)
	assert.True(t, ok, "expected a NextMsg")
	assert.Equal(t, Property{Name: "Url", Description: "The HTTP URL.", Type: "string"}, next.ModelData)
}

func TestFilterSearchModel_SendNextMsg(t *testing.T) {
	m := typeInto(newSearchModelWithProperties(t), "Na")

	next, ok := m.SendNextMsg().(shared.NextMsg)
	assert.True(t, ok)
	assert.Equal(t, "Name", next.ModelData.(Property).Name)
}

func TestFilterSearchModel_View(t *testing.T) {
	m := newSearchModelWithProperties(t)
	assert.NotEmpty(t, plain(m.View()))

	m = typeInto(m, "N")
	assert.Contains(t, plain(m.View()), "The name of the repository.")
}
