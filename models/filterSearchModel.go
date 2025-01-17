package models

import (
	"gh-hubbub/queries"
	"gh-hubbub/structs"

	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type FilterSearchModel struct {
	textinput  textinput.Model
	repository queries.Repository
	properties map[string]property
}

func NewFilterSearchModel() FilterSearchModel {
	ti := textinput.New()
	ti.Placeholder = "Type to search"
	ti.Prompt = "Add filter: "
	ti.PromptStyle = promptStyle.Width(len(ti.Prompt))
	ti.Cursor.Style = cursorStyle
	ti.Focus()
	ti.CharLimit = 50
	ti.SetWidth(20)
	ti.ShowSuggestions = true

	repository := queries.Repository{}

	return FilterSearchModel{
		textinput:  ti,
		repository: repository,
		properties: make(map[string]property),
	}
}

type PropertySelectedMsg property

func (m FilterSearchModel) Init() (tea.Model, tea.Cmd) {
	return m, tea.Batch(getFilters, textinput.Blink)
}

func (m FilterSearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			_, exists := m.CurrentPropertySuggestion()
			if exists {
				return m, m.SendNextMsg
			}
			return m, nil
		}
	case filtersListMsg:
		var suggestions []string
		for _, r := range msg.Properties {
			suggestions = append(suggestions, r.Name)
			m.properties[r.Name] = property{r.Name, r.Description, r.Type}
		}
		m.textinput.SetSuggestions(suggestions)
	}

	var cmd tea.Cmd
	m.textinput, cmd = m.textinput.Update(msg)

	return m, cmd
}

func (m FilterSearchModel) View() string {
	return lipgloss.JoinVertical(lipgloss.Left, m.textinput.View(), "\n", m.LookupDescription())
}

func (m FilterSearchModel) LookupDescription() string {
	prop, exists := m.properties[m.textinput.CurrentSuggestion()]
	if exists {
		return prop.Description
	} else {
		return ""
	}
}

func (m FilterSearchModel) CurrentPropertySuggestion() (property, bool) {
	prop, exists := m.properties[m.textinput.CurrentSuggestion()]
	return prop, exists
}

func (m FilterSearchModel) SendNextMsg() tea.Msg {
	property, _ := m.CurrentPropertySuggestion()
	return NextMessage{ModelData: property}
}

func getFilters() tea.Msg {
	rq := queries.Repository{}
	rp := structs.NewRepoProperties(rq)

	return filtersListMsg(rp)
}
