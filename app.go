package main

import (
	"fmt"
	"gh-reponark/accounts"
	"gh-reponark/filters"
	"gh-reponark/github"
	"gh-reponark/repos"
	"gh-reponark/ui"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// quitKey exits the application from any screen.
var quitKey = key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit"))

// helpKey shows or hides every key the current screen handles.
var helpKey = key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help"))

// backKey is advertised for screens that describe no keys of their own.
var backKey = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back"))

type MainModel struct {
	svc    github.Service
	nav    Navigator
	width  int
	height int

	fullHelp bool // the footer lists every key, not just the short line
}

func NewMainModel(svc github.Service) MainModel {
	nav := NewNavigator()
	nav.Push(accounts.NewModel(svc, 0, 0))

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
//	accounts  --OpenOrgMsg-->      repos
//	repos     --OpenFiltersMsg-->  filters
//	any       --ErrorMsg-->        error screen
//	any       --PreviousMsg-->     back
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
		if key.Matches(msg, helpKey) && !m.typing() {
			m.fullHelp = !m.fullHelp
			return m, nil
		}
		return m, m.UpdateChild(msg)

	case ui.OpenOrgMsg:
		return m, m.open(repos.NewModel(m.svc, msg.Key, contentWidth, contentHeight))

	case filters.OpenFiltersMsg:
		return m, m.open(filters.NewModel(msg.Filters, msg.Repos, contentWidth, contentHeight))

	case ui.PreviousMsg:
		return m, m.Previous(msg)

	case ui.ErrorMsg:
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

	// The top edge is drawn separately so it can carry the breadcrumb. The
	// gradient travels around the whole frame, so the top edge is coloured
	// from the same ramp Lip Gloss uses for the other three sides. Width
	// and Height here include the border, so the body is bodyHeight-1 lines
	// (no top edge) around interiorHeight lines of content. Lip Gloss lays
	// its perimeter gradient out from the frame width and that content
	// height whether or not the top edge is drawn, so renderTopEdge takes
	// the top edge's share from a gradient of the same size.
	body := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, true, true, true).
		BorderForegroundBlend(ui.FrameGradient(true)...).
		Width(layout.bodyWidth).
		Height(layout.bodyHeight - 1).
		Render(lipgloss.Place(layout.interiorWidth, layout.interiorHeight, lipgloss.Left, lipgloss.Top, cappedChild))
	header := m.renderTopEdge(child, layout.bodyWidth, layout.interiorHeight)

	stacked := lipgloss.JoinVertical(lipgloss.Left, header, body, layout.footer)
	framed := lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, stacked)
	v := tea.NewView(framed)
	v.AltScreen = true
	// Paint the terminal in the palette's own base so the colours sit on the
	// background they were chosen for rather than whatever theme is running.
	v.BackgroundColor = ui.AppColors.Background
	v.ForegroundColor = ui.AppColors.Foreground
	return v
}

// open initialises a new screen and makes it the current one.
func (m *MainModel) open(screen tea.Model) tea.Cmd {
	m.fullHelp = false
	cmd := screen.Init()
	m.nav.Push(screen)
	return cmd
}

// Previous returns to the screen below the current one. Going back from the
// first screen quits, so the navigator is never left empty.
func (m *MainModel) Previous(message ui.PreviousMsg) tea.Cmd {
	if m.nav.Len() <= 1 {
		return tea.Quit
	}
	_, _ = m.nav.Pop()
	m.fullHelp = false

	if message.Message != nil {
		return m.UpdateChild(message.Message)
	}

	return nil
}

// ShowError pushes an error screen on top of the current one. Going back from
// it returns to the screen that reported the error.
func (m *MainModel) ShowError(err error) tea.Cmd {
	contentWidth, contentHeight := m.contentDimensions()
	return m.open(NewErrorModel(err, contentWidth, contentHeight))
}

func viewContent(v tea.View) string {
	return fmt.Sprint(v.Content)
}

func (m MainModel) contentDimensions() (int, int) {
	width := ui.Max(1, m.width-2)
	height := ui.Max(1, m.height-3) // top and bottom edges, and the footer
	return width, height
}

// breadcrumb names every open screen that has a title, from the first to the
// current one.
func (m MainModel) breadcrumb() []string {
	crumbs := []string{"reponark"}
	for _, screen := range m.nav.Screens() {
		if titled, ok := screen.(ui.Titled); ok && titled.Breadcrumb() != "" {
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
// The rule is coloured cell by cell from the perimeter gradient of a frame
// with interiorHeight lines of content, so it continues the blend Lip Gloss
// draws around the rest of the frame.
func (m MainModel) renderTopEdge(model tea.Model, width, interiorHeight int) string {
	separator := ui.DimStyle.Render(" › ")

	status := ""
	if sp, ok := model.(ui.StatusProvider); ok && sp.Status() != "" {
		status = " " + ui.DimStyle.Render(sp.Status()) + " "
	}

	crumbs := m.breadcrumb()
	render := func(crumbs []string) string {
		parts := make([]string, len(crumbs))
		for i, crumb := range crumbs {
			if i == len(crumbs)-1 {
				parts[i] = ui.StrongStyle.Render(crumb)
			} else {
				parts[i] = ui.DimStyle.Render(crumb)
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

	fill := ui.Max(0, width-lipgloss.Width(title)-lipgloss.Width(status)-4)
	gradient := ui.PerimeterGradient(width, interiorHeight)
	at := func(s string, from int) string {
		return ui.GradientRun(s, gradient[ui.Min(from, len(gradient)-1):])
	}
	ruleAt := 2 + lipgloss.Width(title)
	edge := at("┌─", 0) + title + at(strings.Repeat("─", fill), ruleAt) + status + at("─┐", width-2)
	return ui.Fit(edge, width)
}

// typing reports whether the current screen has a text field focused, in
// which case "?" is typed rather than toggling help.
func (m MainModel) typing() bool {
	child, err := m.nav.Current()
	if err != nil {
		return false
	}
	t, ok := child.(ui.Typing)
	return ok && t.Typing()
}

// renderFooter draws the current screen's keys. The short line gains "? help"
// just before its last entry, which is how to leave; the full view adds a
// column of keys that work everywhere.
func (m MainModel) renderFooter(model tea.Model) string {
	footerStyle := ui.FooterStyle

	// A screen with nothing to say about its keys still gets the two that
	// work everywhere, drawn the same way as every other footer.
	keys := ui.Help{Short: []key.Binding{backKey, quitKey}}
	hp, ok := model.(ui.HelpProvider)
	if ok {
		keys = hp.Help()
	}
	typing := m.typing()
	if !typing && len(keys.Short) > 0 {
		last := len(keys.Short) - 1
		keys.Short = append(append(append([]key.Binding{}, keys.Short[:last]...), helpKey), keys.Short[last])
	}

	help := ui.NewHelpModel(ui.Max(1, m.width-footerStyle.GetHorizontalFrameSize()))
	help.ShowAll = m.fullHelp && !typing
	if help.ShowAll {
		// The keys that work everywhere join the last column rather than
		// adding one, so the full view fits the same width as the short line.
		columns := append([][]key.Binding{}, keys.FullHelp()...)
		last := len(columns) - 1
		columns[last] = append(append([]key.Binding{}, columns[last]...), ui.WithHelp(helpKey, "hide help"), quitKey)
		keys.Full = columns
	}
	return footerStyle.Render(help.View(keys))
}

type layoutParts struct {
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
	bodyWidth := ui.Max(4, m.width)

	footerRaw := m.renderFooter(child)
	footerHeight := lipgloss.Height(footerRaw)
	if footerHeight < 1 {
		footerHeight = 1
		if footerRaw == "" {
			footerRaw = " "
		}
	}
	footer := lipgloss.NewStyle().Width(bodyWidth).Height(footerHeight).Render(footerRaw)

	bodyHeight := ui.Max(3, m.height-footerHeight)

	return layoutParts{
		footer:         footer,
		footerHeight:   footerHeight,
		bodyWidth:      bodyWidth,
		bodyHeight:     bodyHeight,
		interiorWidth:  ui.Max(1, bodyWidth-2),
		interiorHeight: ui.Max(1, bodyHeight-2),
	}
}
