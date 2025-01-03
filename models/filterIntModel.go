package models

import (
	"fmt"
	"gh-hubbub/structs"
	"strconv"

	"github.com/charmbracelet/bubbles/v2/textinput"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type IntModel struct {
	Name      string
	fromInput textinput.Model
	toInput   textinput.Model
	width     int
	height    int
}

func (m *IntModel) SetDimensions(width, height int) {
	m.width = width
	m.height = height
}

func intValidator(s, prompt string) error {
	_, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("please enter an integer for the `%s` value", prompt)
	}

	return nil
}

func NewIntInputModel(prompt string, value int) textinput.Model {
	m := textinput.New()
	m.Placeholder = fmt.Sprint(value)
	m.Prompt = prompt
	m.CharLimit = 5
	m.Validate = func(s string) error { return intValidator(s, prompt) }
	m.PromptStyle = promptStyle
	m.TextStyle = textStyle

	return m
}

func NewIntModel(title string, from, to, width, height int) IntModel {
	m := IntModel{
		Name:      title,
		fromInput: NewIntInputModel("From", from),
		toInput:   NewIntInputModel("To", to),
	}

	m.fromInput.Focus()

	m.width = width
	m.height = height

	return m
}

func (m IntModel) Init() (tea.Model, tea.Cmd) {
	return m, textinput.Blink
}

func (m IntModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			return m, m.SendAddFilterMsg
		case "tab", "shift+tab":
			if m.fromInput.Focused() {
				m.fromInput.Blur()
				m.toInput.Focus()
			} else {
				m.toInput.Blur()
				m.fromInput.Focus()
			}
		default:
			if m.fromInput.Focused() {
				m.fromInput, cmd = m.fromInput.Update(msg)
			} else {
				m.toInput, cmd = m.toInput.Update(msg)
			}
		}
	}

	return m, cmd
}

func (m IntModel) View() string {
	errorText := ""
	if m.fromInput.Err != nil {
		errorText = "\n" + errorStyle.Render(m.fromInput.Err.Error())
	}
	if m.toInput.Err != nil {
		errorText = "\n" + errorStyle.Render(m.toInput.Err.Error())
	}
	title := fmt.Sprintf("%s - w: %d h: %d", m.Name, m.width, m.height)
	inputs := lipgloss.JoinVertical(lipgloss.Left, m.fromInput.View(), m.toInput.View())
	contents := lipgloss.JoinVertical(lipgloss.Center, modalTitleStyle.Render(title), inputs, errorText)
	return lipgloss.PlaceHorizontal(m.width, lipgloss.Center, modalStyle.Render(contents))
}

func (m *IntModel) GetValue() (int, int) {
	from, _ := strconv.Atoi(m.fromInput.Value())
	to, _ := strconv.Atoi(m.toInput.Value())
	return from, to
}

func (m IntModel) SendAddFilterMsg() tea.Msg {
	from, to := m.GetValue()
	return structs.NewFilterInt(m.Name, from, to)
}
