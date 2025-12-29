package user

import (
	"fmt"
	"sort"

	"gh-reponark/shared"

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
}

func NewModel(width, height int) *Model {
	list := list.New([]list.Item{}, shared.DefaultDelegate, width, height)

	list.SetStatusBarItemName("Organization", "Organizations")
	list.Styles.Title = shared.TitleStyle
	list.SetShowTitle(true)

	return &Model{orgList: list, width: width, height: height}
}

func (m *Model) SetDimensions(width, height int) {
	m.width = width
	m.height = height
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

	m.orgList.Title = "User: " + m.login
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
	m.orgList.SetHeight(m.height)
	return tea.NewView(fmt.Sprint(shared.AppStyle.Render(m.orgList.View())))
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
