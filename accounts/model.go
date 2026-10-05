package accounts

import (
	"sort"

	"gh-reponark/github"
	"gh-reponark/ui"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

// userLoadedMsg carries the authenticated user once the API call completes.
type userLoadedMsg github.User

type Model struct {
	svc   github.Service
	login string
	// items are the accounts to pick from: the user first, then their
	// organizations; cursor is the selected one and the cards in view.
	items   []account
	cursor  ui.Cursor
	width   int
	height  int
	keymap  userKeyMap
	spinner spinner.Model
}

func NewModel(svc github.Service, width, height int) *Model {
	return &Model{svc: svc, width: width, height: height, keymap: newUserKeyMap(), spinner: ui.NewSpinner()}
}

func (m *Model) SetDimensions(width, height int) {
	m.width = width
	m.height = height
	m.cursor.Resize(m.rows(), len(m.items))
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.loadUser, m.spinner.Tick)
}

// loadUser fetches the authenticated user and their organizations.
func (m *Model) loadUser() tea.Msg {
	user, err := m.svc.CurrentUser()
	if err != nil {
		return ui.ErrorMsg{Err: err}
	}
	return userLoadedMsg(user)
}

// SetUser populates the list with the user followed by their organizations.
func (m *Model) SetUser(user github.User) {
	m.login = user.Login
	items := make([]account, len(user.Organizations))
	for i, org := range user.Organizations {
		items[i] = orgAccount(org)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].login < items[j].login
	})

	// Add the user to the top of the list
	// They're not an organization but they also have repositories
	m.items = append([]account{userAccount(user)}, items...)
	m.cursor.Resize(m.rows(), len(m.items))
}

// selected returns the highlighted account, if there is one.
func (m *Model) selected() (account, bool) {
	if m.cursor.Index >= len(m.items) {
		return account{}, false
	}
	return m.items[m.cursor.Index], true
}

// pickerChrome is the heading line above the cards.
const pickerChrome = 1

// rows is how many whole cards are shown at once. Cards are separated by a
// blank line, which the last card on screen does not need.
func (m *Model) rows() int {
	return ui.Max(1, (m.height-pickerChrome+1)/(m.cardLines()+1))
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case userLoadedMsg:
		m.SetUser(github.User(msg))

	case spinner.TickMsg:
		// The spinner only turns until the accounts arrive.
		if m.login != "" {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keymap.Select):
			item, ok := m.selected()
			if !ok {
				return m, nil
			}
			return m, func() tea.Msg {
				orgKey := ui.OrgKey{
					Name:   item.login,
					IsUser: item.isUser,
				}
				return ui.OpenOrgMsg{Key: orgKey}
			}
		case key.Matches(msg, m.keymap.Back):
			return m, func() tea.Msg { return ui.PreviousMsg{} }
		case key.Matches(msg, m.keymap.Up):
			m.cursor.Move(-1)
		case key.Matches(msg, m.keymap.Down):
			m.cursor.Move(1)
		case key.Matches(msg, m.keymap.PageUp):
			m.cursor.Move(-m.rows())
		case key.Matches(msg, m.keymap.PageDown):
			m.cursor.Move(m.rows())
		case key.Matches(msg, m.keymap.Top):
			m.cursor.Set(0)
		case key.Matches(msg, m.keymap.Bottom):
			m.cursor.Set(len(m.items))
		}
	}
	return m, nil
}

func (m Model) View() tea.View {
	width, height := ui.Max(1, m.width), ui.Max(1, m.height)
	if m.login == "" {
		// Nothing to pick from yet: say so in the middle of the pane.
		return tea.NewView(ui.Loading(width, height,
			ui.StrongStyle.Render("reponark"),
			m.spinner.View()+" "+ui.TextBodyStyle.Render("Signing in to GitHub…")))
	}

	lines := []string{m.heading(width)}
	from, to := m.cursor.Visible()
	for i := from; i < to; i++ {
		if i > from {
			lines = append(lines, "")
		}
		lines = append(lines, m.card(m.items[i], i == m.cursor.Index, width)...)
	}
	return tea.NewView(ui.Lines(lines, width, height))
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
func (m Model) Help() ui.Help {
	pages := ui.Combine("←/→", "page", m.keymap.PageUp, m.keymap.PageDown)
	ends := ui.Combine("g/G", "first/last", m.keymap.Top, m.keymap.Bottom)
	return ui.Help{
		Short: []key.Binding{
			ui.Combine("j/k", "org", m.keymap.Down, m.keymap.Up),
			m.keymap.Select,
			m.keymap.Back,
		},
		Full: [][]key.Binding{
			{m.keymap.Up, m.keymap.Down, pages, ends},
			{m.keymap.Select, m.keymap.Back},
		},
	}
}
