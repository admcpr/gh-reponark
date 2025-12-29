package filters

import (
	"fmt"
	"gh-reponark/shared"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// StringModel captures a free-form string filter value.
type StringModel struct {
	name   string
	input  textinput.Model
	width  int
	height int
}

func NewStringModel(name string, value string, width, height int) *StringModel {
	ti := textinput.New()
	ti.Prompt = "Value: "
	ti.Placeholder = value
	styles := textinput.DefaultStyles(false)
	styles.Focused.Prompt = shared.PromptStyle
	styles.Blurred.Prompt = shared.PromptStyle
	styles.Focused.Text = shared.TextStyle
	styles.Blurred.Text = shared.TextStyle
	styles.Cursor.Color = shared.AppColors.Foreground
	ti.SetStyles(styles)
	ti.SetValue(value)
	ti.Focus()

	return &StringModel{
		name:   name,
		input:  ti,
		width:  width,
		height: height,
	}
}

func (m *StringModel) SetDimensions(width, height int) {
	m.width = width
	m.height = height
}

func (m *StringModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *StringModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			return m, m.SendAddFilterMsg
		case "esc":
			return m, func() tea.Msg { return shared.PreviousMsg{} }
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *StringModel) View() tea.View {
	contents := lipgloss.JoinVertical(
		lipgloss.Center,
		shared.ModalTitleStyle.Render(m.name),
		m.input.View(),
	)
	return tea.NewView(fmt.Sprint(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, shared.ModalStyle.Render(contents))))
}

func (m *StringModel) Value() string {
	return m.input.Value()
}

func (m *StringModel) Name() string {
	return m.name
}

func (m *StringModel) SendAddFilterMsg() tea.Msg {
	return shared.PreviousMsg{Message: AddFilterMsg(NewStringFilter(m.name, m.Value()))}
}
