package main

import (
	"fmt"
	"gh-reponark/filters"
	"gh-reponark/github"
	"gh-reponark/org"
	"gh-reponark/shared"
	"gh-reponark/user"
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
		return m, m.open(filters.NewModel(msg.Filters, msg.Repos, contentWidth, contentHeight))

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

	// The top edge is drawn separately so it can carry the breadcrumb.
	body := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, true, true, true).
		BorderForeground(shared.AppColors.Blue).
		Width(layout.bodyWidth).
		Height(layout.bodyHeight - 1).
		Render(lipgloss.Place(layout.interiorWidth, layout.interiorHeight, lipgloss.Left, lipgloss.Top, cappedChild))

	stacked := lipgloss.JoinVertical(lipgloss.Left, layout.header, body, layout.footer)
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
	height := shared.Max(1, m.height-3) // top and bottom edges, and the footer
	return width, height
}

// breadcrumb names every open screen that has a title, from the first to the
// current one.
func (m MainModel) breadcrumb() []string {
	crumbs := []string{"reponark"}
	for _, screen := range m.nav.Screens() {
		if titled, ok := screen.(shared.Titled); ok && titled.Breadcrumb() != "" {
			crumbs = append(crumbs, titled.Breadcrumb())
		}
	}
	return crumbs
}

// renderTopEdge draws the frame's top border with the breadcrumb on the left
// and the current screen's status on the right, e.g.
//
//	┌─ reponark › acme-corp › Filters ──────────── 37 of 148 repos ─┐
//
// When space runs short the status goes first, then the oldest crumbs.
func (m MainModel) renderTopEdge(model tea.Model, width int) string {
	border := lipgloss.NewStyle().Foreground(shared.AppColors.Blue)
	separator := shared.DimStyle.Render(" › ")

	status := ""
	if sp, ok := model.(shared.StatusProvider); ok && sp.Status() != "" {
		status = " " + shared.DimStyle.Render(sp.Status()) + " "
	}

	crumbs := m.breadcrumb()
	render := func(crumbs []string) string {
		parts := make([]string, len(crumbs))
		for i, crumb := range crumbs {
			if i == len(crumbs)-1 {
				parts[i] = shared.StrongStyle.Render(crumb)
			} else {
				parts[i] = shared.DimStyle.Render(crumb)
			}
		}
		return " " + strings.Join(parts, separator) + " "
	}

	// Corners and the rule either side of the title and status.
	const chrome = 6
	title := render(crumbs)
	if lipgloss.Width(title)+lipgloss.Width(status)+chrome > width {
		status = ""
	}
	for len(crumbs) > 1 && lipgloss.Width(title)+chrome > width {
		crumbs = crumbs[1:]
		title = render(append([]string{"…"}, crumbs...))
	}

	fill := shared.Max(0, width-lipgloss.Width(title)-lipgloss.Width(status)-4)
	edge := border.Render("┌─") + title + border.Render(strings.Repeat("─", fill)) + status + border.Render("─┐")
	return shared.Fit(edge, width)
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
	footerHeight   int
	bodyWidth      int
	bodyHeight     int
	interiorWidth  int
	interiorHeight int
}

// computeLayout sizes the frame: the body, whose top edge is the header,
// fills the height left over by the footer.
func (m MainModel) computeLayout(child tea.Model) layoutParts {
	bodyWidth := shared.Max(4, m.width)

	footerRaw := m.renderFooter(child)
	footerHeight := lipgloss.Height(footerRaw)
	if footerHeight < 1 {
		footerHeight = 1
		if footerRaw == "" {
			footerRaw = " "
		}
	}
	footer := lipgloss.NewStyle().Width(bodyWidth).Height(footerHeight).Render(footerRaw)

	bodyHeight := shared.Max(3, m.height-footerHeight)

	return layoutParts{
		header:         m.renderTopEdge(child, bodyWidth),
		footer:         footer,
		footerHeight:   footerHeight,
		bodyWidth:      bodyWidth,
		bodyHeight:     bodyHeight,
		interiorWidth:  shared.Max(1, bodyWidth-2),
		interiorHeight: shared.Max(1, bodyHeight-2),
	}
}
