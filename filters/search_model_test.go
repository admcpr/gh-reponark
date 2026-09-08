package filters

import (
	"testing"

	"gh-reponark/repo"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

func typeInto(m FilterSearchModel, text string) FilterSearchModel {
	for _, r := range text {
		updated, _ := m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = updated.(FilterSearchModel)
	}
	return m
}

func schemaNames() []string {
	schema := repo.Schema()
	names := make([]string, len(schema))
	for i, p := range schema {
		names[i] = p.Name
	}
	return names
}

func TestNewFilterSearchModel(t *testing.T) {
	m := NewFilterSearchModel()

	assert.True(t, m.textinput.Focused())
	assert.True(t, m.textinput.ShowSuggestions)
	assert.Equal(t, "Type to search", m.textinput.Placeholder)
	assert.Equal(t, "Add filter: ", m.textinput.Prompt)
}

func TestNewFilterSearchModel_SuggestsEverySchemaProperty(t *testing.T) {
	m := NewFilterSearchModel()

	assert.Equal(t, schemaNames(), m.textinput.AvailableSuggestions(), "suggestions should follow the schema order")
	assert.Len(t, m.properties, len(repo.Schema()))
	for name, property := range m.properties {
		assert.Equal(t, name, property.Name)
		assert.True(t, isSupportedPropertyType(property.Type), "%s should have an editor", name)
	}

	assert.Equal(t, repo.PropertySchema{
		Name:        "Is Archived",
		Group:       "2⟭ Status",
		Type:        "bool",
		Description: "Indicates if the repository is archived.",
	}, m.properties["Is Archived"])
}

func TestFilterSearchModel_Init(t *testing.T) {
	m := NewFilterSearchModel()
	assert.NotNil(t, m.Init())
}

func TestFilterSearchModel_SuggestionLookups(t *testing.T) {
	m := NewFilterSearchModel()

	// Nothing typed yet so there is no current suggestion.
	_, exists := m.CurrentPropertySuggestion()
	assert.False(t, exists)
	assert.Equal(t, "", m.LookupDescription())

	m = typeInto(m, "Is A")

	prop, exists := m.CurrentPropertySuggestion()
	assert.True(t, exists)
	assert.Equal(t, "Is Archived", prop.Name)
	assert.Equal(t, "Indicates if the repository is archived.", m.LookupDescription())

	// Typing something that matches nothing clears the suggestion.
	m = typeInto(m, "zzz")
	_, exists = m.CurrentPropertySuggestion()
	assert.False(t, exists)
	assert.Equal(t, "", m.LookupDescription())
}

func TestFilterSearchModel_SuggestionIsCaseInsensitive(t *testing.T) {
	m := typeInto(NewFilterSearchModel(), "stargazer")

	prop, exists := m.CurrentPropertySuggestion()
	assert.True(t, exists)
	assert.Equal(t, "Stargazer Count", prop.Name)
}

func TestFilterSearchModel_Update_EnterWithoutSuggestion(t *testing.T) {
	m := NewFilterSearchModel()

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.Nil(t, cmd)
}

func TestFilterSearchModel_Update_EnterWithSuggestion(t *testing.T) {
	m := typeInto(NewFilterSearchModel(), "Stargazer")

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.NotNil(t, cmd)

	edit, ok := cmd().(EditFilterMsg)
	assert.True(t, ok, "expected an EditFilterMsg")
	assert.Equal(t, "Stargazer Count", edit.Property.Name)
	assert.Equal(t, "int", edit.Property.Type)
}

func TestFilterSearchModel_SendEditFilterMsg(t *testing.T) {
	m := typeInto(NewFilterSearchModel(), "Na")

	edit, ok := m.SendEditFilterMsg().(EditFilterMsg)
	assert.True(t, ok)
	assert.Equal(t, "Name", edit.Property.Name, "the first matching property in schema order wins")
}

func TestFilterSearchModel_View(t *testing.T) {
	m := NewFilterSearchModel()
	assert.NotEmpty(t, plain(m.View()))

	m = typeInto(m, "Is A")
	assert.Contains(t, plain(m.View()), "Indicates if the repository is archived.")
}
