package shared

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// ErrorModel is a screen that shows an error and lets the user go back.
type ErrorModel struct {
	err    error
	keymap errorKeyMap
	width  int
	height int
}

func NewErrorModel(err error, width, height int) *ErrorModel {
	return &ErrorModel{
		err:    err,
		keymap: newErrorKeyMap(),
		width:  width,
		height: height,
	}
}

func (m *ErrorModel) Err() error {
	return m.err
}

func (m *ErrorModel) SetDimensions(width, height int) {
	m.width = width
	m.height = height
}

func (m *ErrorModel) Init() tea.Cmd {
	return nil
}

func (m *ErrorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok && key.Matches(msg, m.keymap.Back) {
		return m, func() tea.Msg { return PreviousMsg{} }
	}
	return m, nil
}

func (m *ErrorModel) View() tea.View {
	message := "Unknown error"
	if m.err != nil {
		message = m.err.Error()
	}

	// The same vocabulary as every other screen: a semantic pill for the
	// state, body text for what happened and a dim hint for the way out.
	width := Max(1, m.width)
	title := Pill("Error", AppColors.Bad) + " " + StrongStyle.Render("Something went wrong")
	body := TextBodyStyle.Width(width).Render(message)
	hint := DimStyle.Render("Press esc to go back.")

	return tea.NewView(fmt.Sprint(lipgloss.JoinVertical(lipgloss.Left, title, "", body, "", hint)))
}

func (m *ErrorModel) Breadcrumb() string {
	return "Error"
}

func (m *ErrorModel) Help() Help {
	return Help{Short: m.keymap.ShortHelp()}
}

type errorKeyMap struct {
	Back key.Binding
}

func newErrorKeyMap() errorKeyMap {
	return errorKeyMap{
		Back: key.NewBinding(
			key.WithKeys("esc", "enter"),
			key.WithHelp("esc", "back"),
		),
	}
}

func (k errorKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Back}
}

func (k errorKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}
