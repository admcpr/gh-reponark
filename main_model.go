package main

import (
	"fmt"
	"gh-reponark/filters"
	"gh-reponark/org"
	"gh-reponark/shared"
	"gh-reponark/user"
	"reflect"
	"strings"

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

	layout := m.computeLayout(child)

	m.nav.SetDimensions(layout.interiorWidth, layout.interiorHeight)
	child, _ = m.nav.Current()
	childView := viewContent(child.View())
	cappedChild := lipgloss.NewStyle().
		Width(layout.interiorWidth).
		MaxHeight(layout.interiorHeight).
		Render(childView)

	body := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(shared.AppColors.Blue).
		Width(layout.bodyWidth).
		Height(layout.bodyHeight).
		Render(lipgloss.Place(layout.interiorWidth, layout.interiorHeight, lipgloss.Left, lipgloss.Top, cappedChild))

	sections := []string{}
	if layout.header != "" {
		sections = append(sections, layout.header)
	}
	sections = append(sections, body)
	sections = append(sections, layout.footer)

	stacked := lipgloss.JoinVertical(lipgloss.Left, sections...)
	framed := lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, stacked)
	v := tea.NewView(framed)
	v.AltScreen = true
	return v
}

func (m *MainModel) Next(message shared.NextMsg) tea.Cmd {
	contentWidth, contentHeight := m.contentDimensions()
	var newModel tea.Model
	head, _ := m.nav.Current()

	switch head.(type) {
	case *user.Model:
		newModel = org.NewModel(message.ModelData, contentWidth, contentHeight)
	case *org.Model:
		newModel = filters.NewModel(message.ModelData, contentWidth, contentHeight)
	case *filters.Model:
		newModel = filters.NewFilterModel(message.ModelData, contentWidth, contentHeight)
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

func viewContent(v tea.View) string {
	return fmt.Sprint(v.Content)
}

func (m MainModel) contentDimensions() (int, int) {
	width := shared.Max(1, m.width-2)
	height := shared.Max(1, m.height-4)
	return width, height
}

func (m MainModel) renderHeader(model tea.Model) string {
	if hp, ok := model.(shared.HeaderProvider); ok {
		return viewContent(hp.HeaderView())
	}

	typeOf := reflect.TypeOf(model)
	if typeOf.Kind() == reflect.Ptr {
		typeOf = typeOf.Elem()
	}

	if typeOf.Name() == "" {
		return ""
	}

	return shared.LayoutHeaderStyle.Render(typeOf.Name())
}

func (m MainModel) renderFooter(model tea.Model) string {
	footerStyle := shared.LayoutFooterStyle

	if hp, ok := model.(shared.HelpProvider); ok {
		content := viewContent(hp.HelpView())
		if strings.TrimSpace(content) != "" {
			return footerStyle.Render(content)
		}
	}
	return footerStyle.Foreground(shared.AppColors.BrightBlack).
		Render("esc: back | ctrl+c: quit")
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

type layoutParts struct {
	header         string
	footer         string
	headerHeight   int
	footerHeight   int
	bodyWidth      int
	bodyHeight     int
	interiorWidth  int
	interiorHeight int
}

func (m MainModel) computeLayout(child tea.Model) layoutParts {
	bodyWidth := shared.Max(4, m.width)

	header := strings.TrimSpace(m.renderHeader(child))
	if header != "" {
		header = lipgloss.NewStyle().Width(bodyWidth).Render(header)
	}
	headerHeight := lipgloss.Height(header)

	footerRaw := m.renderFooter(child)
	footerHeight := lipgloss.Height(footerRaw)
	if footerHeight < 1 {
		footerHeight = 1
		if footerRaw == "" {
			footerRaw = " "
		}
	}
	footer := lipgloss.NewStyle().Width(bodyWidth).Height(footerHeight).Render(footerRaw)

	availableHeight := m.height - headerHeight - footerHeight
	if availableHeight < 1 {
		availableHeight = 1
	}
	bodyHeight := shared.Max(3, availableHeight)
	// Ensure we never exceed the terminal height so footer stays visible.
	maxBody := m.height - headerHeight - footerHeight
	if maxBody < 1 {
		maxBody = 1
	}
	if bodyHeight > maxBody {
		bodyHeight = maxBody
	}
	interiorWidth := shared.Max(1, bodyWidth-2)
	interiorHeight := shared.Max(1, bodyHeight-2)

	return layoutParts{
		header:         header,
		footer:         footer,
		headerHeight:   headerHeight,
		footerHeight:   footerHeight,
		bodyWidth:      bodyWidth,
		bodyHeight:     bodyHeight,
		interiorWidth:  interiorWidth,
		interiorHeight: interiorHeight,
	}
}
