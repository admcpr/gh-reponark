package filters

import (
	"fmt"
	"gh-reponark/repo"
	"gh-reponark/shared"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// FilterSearchModel lets the user pick a repository property to filter on by
// typing its name with autocompletion. The candidates come straight from
// repo.Schema.
type FilterSearchModel struct {
	textinput  textinput.Model
	keymap     filterKeyMap
	properties map[string]repo.PropertySchema
}

// EditFilterMsg asks the application to open the editor for a property so a
// filter on it can be added.
type EditFilterMsg struct{ Property repo.PropertySchema }

func NewFilterSearchModel() FilterSearchModel {
	keymap := newFilterKeyMap()

	ti := textinput.New()
	ti.Placeholder = "Type to search"
	ti.Prompt = "Add filter: "
	// Set styles using the new textinput Styles API
	styles := textinput.DefaultStyles(false)
	styles.Focused.Prompt = shared.PromptStyle.Width(len(ti.Prompt))
	styles.Blurred.Prompt = shared.PromptStyle.Width(len(ti.Prompt))
	styles.Focused.Text = shared.TextStyle
	styles.Blurred.Text = shared.TextStyle
	styles.Cursor.Color = shared.AppColors.Foreground
	ti.SetStyles(styles)
	ti.Focus()
	ti.CharLimit = 50
	ti.SetWidth(20)
	ti.ShowSuggestions = true
	// Autocompletion uses the same bindings the help footer advertises.
	ti.KeyMap.AcceptSuggestion = keymap.Complete
	ti.KeyMap.NextSuggestion = keymap.NextSuggestion
	ti.KeyMap.PrevSuggestion = keymap.PrevSuggestion

	properties := make(map[string]repo.PropertySchema)
	var suggestions []string
	for _, property := range repo.Schema() {
		if !isSupportedPropertyType(property.Type) {
			continue
		}
		suggestions = append(suggestions, property.Name)
		properties[property.Name] = property
	}
	ti.SetSuggestions(suggestions)

	return FilterSearchModel{
		textinput:  ti,
		keymap:     keymap,
		properties: properties,
	}
}

func (m FilterSearchModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m FilterSearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok && key.Matches(msg, m.keymap.Select) {
		if _, exists := m.CurrentPropertySuggestion(); exists {
			return m, m.SendEditFilterMsg
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.textinput, cmd = m.textinput.Update(msg)

	return m, cmd
}

func (m FilterSearchModel) View() tea.View {
	style := lipgloss.NewStyle().Margin(0, 0, 1, 2)
	search := lipgloss.JoinVertical(lipgloss.Left, m.textinput.View(), m.LookupDescription())
	return tea.NewView(fmt.Sprint(style.Render(search)))
}

func (m FilterSearchModel) LookupDescription() string {
	prop, exists := m.properties[m.textinput.CurrentSuggestion()]
	if exists {
		return prop.Description
	}
	return ""
}

func (m FilterSearchModel) CurrentPropertySuggestion() (repo.PropertySchema, bool) {
	prop, exists := m.properties[m.textinput.CurrentSuggestion()]
	return prop, exists
}

func (m FilterSearchModel) SendEditFilterMsg() tea.Msg {
	property, _ := m.CurrentPropertySuggestion()
	return EditFilterMsg{Property: property}
}
