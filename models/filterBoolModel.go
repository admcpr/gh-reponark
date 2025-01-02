package models

import (
	"gh-hubbub/structs"
	"gh-hubbub/style"

	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

var (
	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#888B7E")).
			Padding(0, 3).
			Margin(2)

	activeButtonStyle = buttonStyle.
				Foreground(lipgloss.Color("#FFF7DB")).
				Background(lipgloss.Color("#F25D94")).
				Underline(true)
)

type BoolModel struct {
	Name  string
	Value bool
}

func NewBoolModel(name string, value bool) BoolModel {
	m := BoolModel{
		Name:  name,
		Value: value,
	}

	return m
}

type BoolFilterMessage struct {
	Name  string
	Value bool
}

func (m BoolModel) Init() (tea.Model, tea.Cmd) {
	return m, textinput.Blink
}

func (m BoolModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			return m, m.SendFilterMsg
		case "y", "Y":
			m.Value = true
		case "n", "N":
			m.Value = false
		case "right", "left":
			m.Value = !m.Value
		}
	}

	return m, cmd
}

func (m BoolModel) View() string {
	yesButtonStyle := buttonStyle
	noButtonStyle := buttonStyle
	if m.Value {
		yesButtonStyle = activeButtonStyle
	} else {
		noButtonStyle = activeButtonStyle
	}
	buttons := lipgloss.JoinHorizontal(lipgloss.Left, yesButtonStyle.Render("Yes"), noButtonStyle.Render("No"))
	return lipgloss.JoinVertical(lipgloss.Center, style.Title.Render(m.Name), buttons)
}

func (m *BoolModel) GetValue() bool {
	return m.Value
}

func (m BoolModel) SendFilterMsg() tea.Msg {
	return structs.NewFilterBool(m.Name, m.GetValue())
}
