package user

import (
	"fmt"
	"sort"

	"gh-reponark/github"
	"gh-reponark/shared"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

// userLoadedMsg carries the authenticated user once the API call completes.
type userLoadedMsg github.User

type Model struct {
	svc     github.Service
	login   string
	orgList list.Model
	width   int
	height  int
	help    help.Model
	keymap  userKeyMap
}

func NewModel(svc github.Service, width, height int) *Model {
	keymap := newUserKeyMap()

	list := list.New([]list.Item{}, shared.DefaultDelegate, width, height)
	list.SetStatusBarItemName("Organization", "Organizations")
	list.Styles.Title = shared.TitleStyle
	list.SetShowTitle(false)
	list.SetShowHelp(false)
	list.SetShowStatusBar(false)
	// The list moves with the same bindings the help footer advertises.
	list.KeyMap.CursorUp = keymap.Up
	list.KeyMap.CursorDown = keymap.Down

	helpModel := shared.NewHelpModel(width)

	return &Model{svc: svc, orgList: list, width: width, height: height, help: helpModel, keymap: keymap}
}

func (m *Model) SetDimensions(width, height int) {
	m.width = width
	m.height = height
	m.help.SetWidth(width)
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
	items := make([]list.Item, len(user.Organizations))
	for i, org := range user.Organizations {
		items[i] = shared.NewListItem(org.Login, org.Url)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].FilterValue() < items[j].FilterValue()
	})

	// Add the user to the top of the list
	// They're not an organization but they also have repositories
	userItem := shared.NewListItem(m.login, user.Url)
	items = append([]list.Item{userItem}, items...)
	m.orgList.SetItems(items)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case userLoadedMsg:
		m.SetUser(github.User(msg))
		return m, nil

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keymap.Select):
			item, ok := m.orgList.SelectedItem().(shared.ListItem)
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
		}
	}

	m.orgList, cmd = m.orgList.Update(msg)

	return m, cmd
}

func (m Model) View() tea.View {
	m.orgList.SetWidth(m.width)
	m.orgList.SetHeight(shared.Max(1, m.height))

	return tea.NewView(fmt.Sprint(shared.AppStyle.Width(m.width).Render(m.orgList.View())))
}

func (m Model) HeaderView() tea.View {
	title := "Organizations"
	if m.login != "" {
		title = fmt.Sprintf("User: %s", m.login)
	}

	return tea.NewView(shared.TitleStyle.Render(title))
}

func (m Model) HelpView() tea.View {
	// Even if width is zero, help will render minimally; SetDimensions sets width on resize.
	return tea.NewView(m.help.View(m.keymap))
}

// userKeyMap holds the bindings for the organization picker. Update matches
// against these and the help footer renders them.
type userKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	Back   key.Binding
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
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "quit"),
		),
	}
}

func (k userKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Select, k.Back}
}

func (k userKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}
