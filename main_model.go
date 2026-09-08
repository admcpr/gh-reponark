package main

import (
	"fmt"
	"gh-reponark/filters"
	"gh-reponark/github"
	"gh-reponark/org"
	"gh-reponark/shared"
	"gh-reponark/user"
	"reflect"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// quitKey exits the application from any screen.
var quitKey = key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit"))

type MainModel struct {
	svc    github.Service
	nav    shared.Navigator
	width  int
	height int
}

func NewMainModel(svc github.Service) MainModel {
	nav := shared.NewNavigator()
	nav.Push(user.NewModel(svc, 0, 0))

	return MainModel{
		svc: svc,
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

// Update routes messages. Navigation requests are typed messages emitted by
// the screens themselves, so every transition is visible in this one switch:
//
//	user  --OpenOrgMsg-->      org
//	org   --OpenFiltersMsg-->  filters
//	filters --EditFilterMsg--> filter editor (bool/int/date/string)
//	any   --ErrorMsg-->        error screen
//	any   --PreviousMsg-->     back
func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	contentWidth, contentHeight := m.contentDimensions()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetDimensions(msg.Width, msg.Height)
		return m, nil

	case tea.KeyPressMsg:
		if key.Matches(msg, quitKey) {
			return m, tea.Quit
		}
		return m, m.UpdateChild(msg)

	case shared.OpenOrgMsg:
		return m, m.open(org.NewModel(m.svc, msg.Key, contentWidth, contentHeight))

	case filters.OpenFiltersMsg:
		return m, m.open(filters.NewModel(msg.Filters, contentWidth, contentHeight))

	case filters.EditFilterMsg:
		editor := filters.NewFilterModel(msg.Property, contentWidth, contentHeight)
		if editor == nil {
			// The property type has no editor; stay where we are.
			return m, nil
		}
		return m, m.open(editor)

	case shared.PreviousMsg:
		return m, m.Previous(msg)

	case shared.ErrorMsg:
		return m, m.ShowError(msg.Err)

	default:
		return m, m.UpdateChild(msg)
	}
}

func (m *MainModel) UpdateChild(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	currentModel, _ := m.nav.Pop()
	currentModel, cmd = currentModel.Update(msg)
	m.nav.Push(currentModel)
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

// open initialises a new screen and makes it the current one.
func (m *MainModel) open(screen tea.Model) tea.Cmd {
	cmd := screen.Init()
	m.nav.Push(screen)
	return cmd
}

// Previous returns to the screen below the current one. Going back from the
// first screen quits, so the navigator is never left empty.
func (m *MainModel) Previous(message shared.PreviousMsg) tea.Cmd {
	if m.nav.Len() <= 1 {
		return tea.Quit
	}
	_, _ = m.nav.Pop()

	if message.Message != nil {
		return m.UpdateChild(message.Message)
	}

	return nil
}

// ShowError pushes an error screen on top of the current one. Going back from
// it returns to the screen that reported the error.
func (m *MainModel) ShowError(err error) tea.Cmd {
	contentWidth, contentHeight := m.contentDimensions()
	return m.open(shared.NewErrorModel(err, contentWidth, contentHeight))
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
