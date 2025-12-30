package user

import (
	"fmt"
	"sort"

	"gh-reponark/shared"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
)

type AuthenticationErrorMsg struct{ Err error }
type ErrMsg struct{ Err error }
type queryCompleteMsg Query

type Model struct {
	login   string
	orgList list.Model
	width   int
	height  int
	help    help.Model
	keymap  userKeyMap
}

func NewModel(width, height int) *Model {
	list := list.New([]list.Item{}, shared.DefaultDelegate, width, height)

	list.SetStatusBarItemName("Organization", "Organizations")
	list.Styles.Title = shared.TitleStyle
	list.SetShowTitle(false)
	list.SetShowHelp(false)
	list.SetShowStatusBar(false)

	helpModel := shared.NewHelpModel(width)
	keymap := userKeyMap{}

	return &Model{orgList: list, width: width, height: height, help: helpModel, keymap: keymap}
}

func (m *Model) SetDimensions(width, height int) {
	m.width = width
	m.height = height
	m.help.SetWidth(width)
}

func (m Model) Init() tea.Cmd {
	return getUser
}

func (m *Model) SetOrgList(query Query) {
	m.login = query.User.Login
	items := make([]list.Item, len(query.User.Organizations.Nodes))
	for i, org := range query.User.Organizations.Nodes {
		items[i] = shared.NewListItem(org.Login, org.Url)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].FilterValue() < items[j].FilterValue()
	})

	// Add the user to the top of the list
	// They're not an organization but they also have repositories
	userItem := shared.NewListItem(m.login, query.User.Url)
	items = append([]list.Item{userItem}, items...)
	m.orgList.SetItems(items)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case queryCompleteMsg:
		m.SetOrgList(Query(msg))

		return m, cmd
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if item, ok := m.orgList.SelectedItem().(shared.ListItem); ok {
				return m, func() tea.Msg {
					isUser := item.Title() == m.login
					orgKey := shared.OrgKey{
						Name:   item.Title(),
						IsUser: isUser,
					}
					return shared.NextMsg{ModelData: orgKey}
				}
			}
			return m, cmd
		default:
			m.orgList, cmd = m.orgList.Update(msg)
			return m, cmd
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

type userKeyMap struct{}

func (k userKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	}
}

func (k userKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

func getUser() tea.Msg {
	login, err := getLogin()
	if err != nil {
		return AuthenticationErrorMsg{Err: err}
	}

	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return ErrMsg{Err: err}
	}

	var userQuery = Query{}

	variables := map[string]interface{}{
		"login": graphql.String(login),
		"first": graphql.Int(100),
	}
	err = client.Query("User", &userQuery, variables)
	if err != nil {
		return ErrMsg{Err: err}
	}

	return queryCompleteMsg(userQuery)
}

func getLogin() (string, error) {
	client, err := api.DefaultRESTClient()
	if err != nil {
		return "", err
	}
	response := User{}

	err = client.Get("user", &response)
	if err != nil {
		return "", err
	}

	return response.Login, nil
}
