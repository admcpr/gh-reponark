package filters

import (
	"fmt"
	"gh-reponark/shared"
	"strconv"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type IntModel struct {
	name      string
	fromInput textinput.Model
	toInput   textinput.Model
	keymap    editorKeyMap
	help      help.Model
	width     int
	height    int
}

func (m *IntModel) SetDimensions(width, height int) {
	m.width = width
	m.height = height
	m.help.SetWidth(width)
}

func intValidator(s, prompt string) error {
	_, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("please enter an integer for the `%s` value", prompt)
	}

	return nil
}

func newIntInputModel(prompt string, value int) textinput.Model {
	m := textinput.New()
	m.Placeholder = fmt.Sprint(value)
	m.Prompt = prompt
	m.CharLimit = 5
	m.Validate = func(s string) error { return intValidator(s, prompt) }
	styles := textinput.DefaultStyles(false)
	styles.Focused.Prompt = shared.PromptStyle
	styles.Blurred.Prompt = shared.PromptStyle
	styles.Focused.Text = shared.TextStyle
	styles.Blurred.Text = shared.TextStyle
	styles.Cursor.Color = shared.AppColors.Foreground
	m.SetStyles(styles)

	return m
}

func NewIntModel(title string, from, to, width, height int) *IntModel {
	m := &IntModel{
		name:      title,
		fromInput: newIntInputModel("From", from),
		toInput:   newIntInputModel("To", to),
		keymap:    newEditorKeyMap(),
		help:      shared.NewHelpModel(width),
	}

	m.fromInput.Focus()

	m.width = width
	m.height = height

	return m
}

func (m *IntModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *IntModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(msg, m.keymap.Apply):
			return m, m.SendAddFilterMsg
		case key.Matches(msg, m.keymap.Back):
			return m, func() tea.Msg {
				return shared.PreviousMsg{}
			}
		case key.Matches(msg, m.keymap.NextField):
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

func (m *IntModel) View() tea.View {
	errorText := ""
	if m.fromInput.Err != nil {
		errorText = "\n" + shared.ErrorStyle.Render(m.fromInput.Err.Error())
	}
	if m.toInput.Err != nil {
		errorText = "\n" + shared.ErrorStyle.Render(m.toInput.Err.Error())
	}
	inputs := lipgloss.JoinVertical(lipgloss.Left, m.fromInput.View(), m.toInput.View())
	contents := lipgloss.JoinVertical(lipgloss.Center, shared.ModalTitleStyle.Render(m.name), inputs, errorText)
	return tea.NewView(fmt.Sprint(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, shared.ModalStyle.Render(contents))))
}

func (m *IntModel) HelpView() tea.View {
	return tea.NewView(m.help.View(shared.KeyBindings{
		m.keymap.NextField,
		m.keymap.Apply,
		m.keymap.Back,
	}))
}

func (m *IntModel) Value() (int, int) {
	from, _ := strconv.Atoi(m.fromInput.Value())
	to, _ := strconv.Atoi(m.toInput.Value())
	return from, to
}

func (m *IntModel) SendAddFilterMsg() tea.Msg {
	from, to := m.Value()
	return shared.PreviousMsg{Message: AddFilterMsg(NewIntFilter(m.name, from, to))}
}

func (m *IntModel) Name() string {
	return m.name
}
