package user

import (
	"sort"

	"gh-reponark/github"
	"gh-reponark/shared"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// userLoadedMsg carries the authenticated user once the API call completes.
type userLoadedMsg github.User

type Model struct {
	svc   github.Service
	login string
	// items are the accounts to pick from: the user first, then their
	// organizations. cursor is the selected one and offset the first row
	// scrolled into view.
	items  []shared.ListItem
	cursor int
	offset int
	width  int
	height int
	keymap userKeyMap
}

func NewModel(svc github.Service, width, height int) *Model {
	return &Model{svc: svc, width: width, height: height, keymap: newUserKeyMap()}
}

func (m *Model) SetDimensions(width, height int) {
	m.width = width
	m.height = height
	m.scrollToCursor()
}

func (m *Model) Init() tea.Cmd {
	return m.loadUser
}

// loadUser fetches the authenticated user and their organizations.
func (m *Model) loadUser() tea.Msg {
	user, err := m.svc.CurrentUser()
	if err != nil {
		return shared.ErrorMsg{Err: err}
	}
	return userLoadedMsg(user)
}

// SetUser populates the list with the user followed by their organizations.
func (m *Model) SetUser(user github.User) {
	m.login = user.Login
	items := make([]shared.ListItem, len(user.Organizations))
	for i, org := range user.Organizations {
		items[i] = shared.NewListItem(org.Login, org.Url)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].FilterValue() < items[j].FilterValue()
	})

	// Add the user to the top of the list
	// They're not an organization but they also have repositories
	userItem := shared.NewListItem(m.login, user.Url)
	m.items = append([]shared.ListItem{userItem}, items...)
	m.moveCursor(0)
}

// selected returns the highlighted account, if there is one.
func (m *Model) selected() (shared.ListItem, bool) {
	if m.cursor < 0 || m.cursor >= len(m.items) {
		return shared.ListItem{}, false
	}
	return m.items[m.cursor], true
}

// moveCursor selects the account delta rows away, stopping at either end.
func (m *Model) moveCursor(delta int) {
	if len(m.items) == 0 {
		return
	}
	m.cursor = shared.Max(0, shared.Min(m.cursor+delta, len(m.items)-1))
	m.scrollToCursor()
}

func (m *Model) scrollToCursor() {
	m.offset = shared.ScrollOffset(m.offset, m.cursor, m.rows(), len(m.items))
}

// pickerChrome is the heading line above the rows.
const pickerChrome = 1

// rows is how many accounts are shown at once.
func (m *Model) rows() int { return shared.Max(1, m.height-pickerChrome) }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case userLoadedMsg:
		m.SetUser(github.User(msg))

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keymap.Select):
			item, ok := m.selected()
			if !ok {
				return m, nil
			}
			return m, func() tea.Msg {
				isUser := item.Title() == m.login
				orgKey := shared.OrgKey{
					Name:   item.Title(),
					IsUser: isUser,
				}
				return shared.OpenOrgMsg{Key: orgKey}
			}
		case key.Matches(msg, m.keymap.Back):
			return m, func() tea.Msg { return shared.PreviousMsg{} }
		case key.Matches(msg, m.keymap.Up):
			m.moveCursor(-1)
		case key.Matches(msg, m.keymap.Down):
			m.moveCursor(1)
		case key.Matches(msg, m.keymap.PageUp):
			m.moveCursor(-m.rows())
		case key.Matches(msg, m.keymap.PageDown):
			m.moveCursor(m.rows())
		case key.Matches(msg, m.keymap.Top):
			m.moveCursor(-len(m.items))
		case key.Matches(msg, m.keymap.Bottom):
			m.moveCursor(len(m.items))
		}
	}
	return m, nil
}

func (m Model) View() tea.View {
	width, height := shared.Max(1, m.width), shared.Max(1, m.height)
	rows := m.rows()

	nameWidth := 0
	for _, item := range m.items {
		nameWidth = shared.Max(nameWidth, lipgloss.Width(item.Title()))
	}
	nameWidth = shared.Min(nameWidth, shared.Half(width))

	lines := []string{shared.ColumnHeading.Render(shared.Fit("    ACCOUNT", width))}
	// The pane may have been resized since the cursor last moved.
	offset := shared.ScrollOffset(m.offset, m.cursor, rows, len(m.items))
	for i := offset; i < len(m.items) && i < offset+rows; i++ {
		lines = append(lines, m.row(m.items[i], i == m.cursor, nameWidth, width))
	}
	return tea.NewView(shared.Lines(lines, width, height))
}

// row is one account: the accent marker and a full-width band on the
// selected one, the login in bold and its URL in the link colour, dimmed on
// the rows that are not selected.
func (m Model) row(item shared.ListItem, selected bool, nameWidth, width int) string {
	marker, name, url := "  ", shared.TextBodyStyle, shared.DimStyle
	if selected {
		marker = shared.AccentStyle.Render("▌ ")
		name, url = shared.StrongStyle, shared.LinkTextStyle
	}
	row := marker + "  " + name.Render(shared.Fit(item.Title(), nameWidth)) + "  " + url.Render(item.Description())
	if selected {
		return shared.HighlightRow(row, width, true)
	}
	return shared.Fit(row, width)
}

// Status says who is signed in. The screen adds nothing to the breadcrumb:
// it is the root the others open from.
func (m Model) Status() string {
	if m.login == "" {
		return "signing in"
	}
	return "signed in as " + m.login
}

// Help lists the picker's keys, including the paging keys.
func (m Model) Help() shared.Help {
	pages := shared.Combine("←/→", "page", m.keymap.PageUp, m.keymap.PageDown)
	ends := shared.Combine("g/G", "first/last", m.keymap.Top, m.keymap.Bottom)
	return shared.Help{
		Short: []key.Binding{
			shared.Combine("j/k", "org", m.keymap.Down, m.keymap.Up),
			m.keymap.Select,
			m.keymap.Back,
		},
		Full: [][]key.Binding{
			{m.keymap.Up, m.keymap.Down, pages, ends},
			{m.keymap.Select, m.keymap.Back},
		},
	}
}

// userKeyMap holds the bindings for the organization picker. Update matches
// against these and the help footer renders them.
type userKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Top      key.Binding
	Bottom   key.Binding
	Select   key.Binding
	Back     key.Binding
}

func newUserKeyMap() userKeyMap {
	return userKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		// The paging keys are the ones the list component used to answer to.
		PageUp: key.NewBinding(
			key.WithKeys("left", "h", "pgup", "b", "u"),
			key.WithHelp("←/h/pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("right", "l", "pgdown", "f", "d"),
			key.WithHelp("→/l/pgdn", "page down"),
		),
		Top: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "first"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "last"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "open"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "quit"),
		),
	}
}
