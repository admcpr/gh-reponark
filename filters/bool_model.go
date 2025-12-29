package filters

import (
	"fmt"
	"gh-reponark/shared"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type BoolModel struct {
	name   string
	value  bool
	width  int
	height int
}

func NewBoolModel(name string, value bool, width, height int) *BoolModel {
	m := &BoolModel{
		name:  name,
		value: value,
	}

	m.width = width
	m.height = height

	return m
}

type BoolFilterMessage struct {
	Name  string
	Value bool
}

func (m *BoolModel) SetDimensions(width, height int) {
	m.width = width
	m.height = height
}

func (m *BoolModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *BoolModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			return m, m.SendAddFilterMsg
		case "esc":
			return m, func() tea.Msg {
				return shared.PreviousMsg{}
			}
		case "y", "Y":
			m.value = true
		case "n", "N":
			m.value = false
		case "right", "left":
			m.value = !m.value
		}
	}

	return m, cmd
}

func (m *BoolModel) View() tea.View {
	yesButtonStyle := shared.ButtonStyle
	noButtonStyle := shared.ButtonStyle
	if m.value {
		yesButtonStyle = shared.ActiveButtonStyle
	} else {
		noButtonStyle = shared.ActiveButtonStyle
	}
	buttons := lipgloss.JoinHorizontal(lipgloss.Left, yesButtonStyle.Render("Yes"), noButtonStyle.Render("No"))
	contents := lipgloss.JoinVertical(lipgloss.Center, shared.ModalTitleStyle.Render(m.name), buttons)

	return tea.NewView(fmt.Sprint(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, shared.ModalStyle.Render(contents))))
}

func (m *BoolModel) Value() bool {
	return m.value
}

func (m *BoolModel) Name() string {
	return m.name
}

func (m *BoolModel) SendAddFilterMsg() tea.Msg {
	return shared.PreviousMsg{Message: AddFilterMsg(NewBoolFilter(m.name, m.Value()))}
}
