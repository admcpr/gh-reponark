package filters

import (
	"fmt"
	"gh-reponark/shared"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type BoolModel struct {
	name   string
	value  bool
	keymap editorKeyMap
	help   help.Model
	width  int
	height int
}

func NewBoolModel(name string, value bool, width, height int) *BoolModel {
	m := &BoolModel{
		name:   name,
		value:  value,
		keymap: newEditorKeyMap(),
		help:   shared.NewHelpModel(width),
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
	m.help.SetWidth(width)
}

func (m *BoolModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *BoolModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(msg, m.keymap.Apply):
			return m, m.SendAddFilterMsg
		case key.Matches(msg, m.keymap.Back):
			return m, func() tea.Msg {
				return shared.PreviousMsg{}
			}
		case key.Matches(msg, m.keymap.Yes):
			m.value = true
		case key.Matches(msg, m.keymap.No):
			m.value = false
		case key.Matches(msg, m.keymap.Toggle):
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

func (m *BoolModel) HelpView() tea.View {
	return tea.NewView(m.help.View(shared.KeyBindings{
		m.keymap.Yes,
		m.keymap.No,
		m.keymap.Toggle,
		m.keymap.Apply,
		m.keymap.Back,
	}))
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
