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
	stack  shared.ModelStack
	width  int
	height int
}

func NewMainModel() MainModel {
	stack := shared.ModelStack{}
	stack.Push(user.NewModel(0, 0))
	// stack.Push(filters.NewBoolModel("Is something true", false, 0, 0))
	// stack.Push(NewDateModel("Date between", time.Now(), time.Now().Add(time.Hour*24*7), 0, 0))
	// stack.Push(NewIntModel("Number between", 0, 100, 0, 0))
	// stack.Push(filters.NewModel(0, 0))

	return MainModel{
		stack: stack,
	}
}

func (m *MainModel) SetDimensions(width, height int) {
	m.width = width
	m.height = height
	// 2 is subtracted from the width and height to account for the border
	m.stack.SetDimensions(width-2, height-2)
}

func (m MainModel) Init() tea.Cmd {
	child, _ := m.stack.Peek()
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
	currentModel, _ := m.stack.Pop()
	currentModel, cmd = currentModel.Update(msg)
	m.stack.Push(currentModel)
	return cmd
}

func (m MainModel) View() tea.View {
	child, _ := m.stack.Peek()
	borderStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(shared.AppColors.Green)
	v := tea.NewView(borderStyle.Render(lipgloss.PlaceHorizontal(m.width-2, lipgloss.Left, fmt.Sprint(child.View()))))
	v.AltScreen = true
	return v
}

func (m *MainModel) Next(message shared.NextMsg) tea.Cmd {
	var newModel tea.Model
	head, _ := m.stack.Peek()

	switch head.(type) {
	case *user.Model:
		newModel = org.NewModel(message.ModelData, m.width-2, m.height-2)
	case *org.Model:
		newModel = filters.NewModel(message.ModelData, m.width-2, m.height-2)
	case *filters.Model:
		newModel = filters.NewFilterModel(message.ModelData, m.width-2, m.height-2)
	}

	cmd := newModel.Init()
	m.stack.Push(newModel)

	return cmd
}

func (m *MainModel) Previous(message shared.PreviousMsg) tea.Cmd {
	_, err := m.stack.Pop()

	if err != nil {
		return tea.Quit
	}

	if message.Message != nil {
		return m.UpdateChild(message.Message)
	}

	return nil
}
