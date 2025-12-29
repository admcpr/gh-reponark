package main

import (
	"fmt"
	"gh-reponark/filters"
	"gh-reponark/org"
	"gh-reponark/shared"
	"gh-reponark/user"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type MainModel struct {
	nav    shared.Navigator
	width  int
	height int
}

func NewMainModel() MainModel {
	nav := shared.NewNavigator()
	nav.SetValidator(validateTransition)
	_ = nav.Push(user.NewModel(0, 0))
	// stack.Push(filters.NewBoolModel("Is something true", false, 0, 0))
	// stack.Push(NewDateModel("Date between", time.Now(), time.Now().Add(time.Hour*24*7), 0, 0))
	// stack.Push(NewIntModel("Number between", 0, 100, 0, 0))
	// stack.Push(filters.NewModel(0, 0))

	return MainModel{
		nav: nav,
	}
}

func (m *MainModel) SetDimensions(width, height int) {
	m.width = width
	m.height = height
	// 2 is subtracted from the width and height to account for the border
	m.nav.SetDimensions(width-2, height-2)
}

func (m MainModel) Init() tea.Cmd {
	child, _ := m.nav.Current()
	return child.Init()
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// This should only do a couple of things
	// 1. Handle ctrl+c to quit ✔️
	// 2. Handle window sizing ✔️
	// 3. Handle Forward & Back navigation (creating models as needed) and updating state ✔️
	// 4. Call Update on the active model ✔️

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetDimensions(msg.Width, msg.Height)
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		default:
			cmd = m.UpdateChild(msg)
		}
	case shared.NextMsg:
		cmd = m.Next(msg)
		return m, cmd
	case shared.PreviousMsg:
		cmd = m.Previous(msg)
		return m, cmd
	default:
		cmd = m.UpdateChild(msg)
	}

	return m, cmd
}

func (m *MainModel) UpdateChild(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	currentModel, _ := m.nav.Pop()
	currentModel, cmd = currentModel.Update(msg)
	_ = m.nav.Push(currentModel)
	return cmd
}

func (m MainModel) View() tea.View {
	child, _ := m.nav.Current()
	borderStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(shared.AppColors.Green)
	childView := fmt.Sprint(child.View().Content)
	v := tea.NewView(borderStyle.Render(lipgloss.PlaceHorizontal(m.width-2, lipgloss.Left, childView)))
	v.AltScreen = true
	return v
}

func (m *MainModel) Next(message shared.NextMsg) tea.Cmd {
	var newModel tea.Model
	head, _ := m.nav.Current()

	switch head.(type) {
	case *user.Model:
		newModel = org.NewModel(message.ModelData, m.width-2, m.height-2)
	case *org.Model:
		newModel = filters.NewModel(message.ModelData, m.width-2, m.height-2)
	case *filters.Model:
		newModel = filters.NewFilterModel(message.ModelData, m.width-2, m.height-2)
	}

	if newModel == nil {
		return nil
	}

	cmd := newModel.Init()
	if err := m.nav.Push(newModel); err != nil {
		// Transition not allowed; ignore the navigation request.
		return nil
	}

	return cmd
}

func (m *MainModel) Previous(message shared.PreviousMsg) tea.Cmd {
	_, err := m.nav.Pop()

	if err != nil {
		return tea.Quit
	}

	if message.Message != nil {
		return m.UpdateChild(message.Message)
	}

	return nil
}

// validateTransition restricts navigation order between screens.
func validateTransition(current, next tea.Model) error {
	switch current.(type) {
	case *user.Model:
		if _, ok := next.(*org.Model); ok {
			return nil
		}
	case *org.Model:
		if _, ok := next.(*filters.Model); ok {
			return nil
		}
	case *filters.Model:
		if isFilterDetail(next) {
			return nil
		}
	case *filters.BoolModel, *filters.IntModel, *filters.DateModel, *filters.StringModel:
		// From detail models, allow any next (they should navigate back via Previous)
		return nil
	}

	return fmt.Errorf("invalid transition %T -> %T", current, next)
}

func isFilterDetail(m tea.Model) bool {
	switch m.(type) {
	case *filters.BoolModel, *filters.IntModel, *filters.DateModel, *filters.StringModel:
		return true
	default:
		return false
	}
}
