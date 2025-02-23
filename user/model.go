package user

import (
	"sort"

	"gh-reponark/org"
	"gh-reponark/shared"

	"github.com/charmbracelet/bubbles/v2/list"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
)

type AuthenticationErrorMsg struct{ Err error }
type ErrMsg struct{ Err error }
type OrgListMsg struct{ Organisations []org.Organisation }
type queryCompleteMsg Query

type Model struct {
	query          Query
	User           User
	SelectedOrgUrl string
	orgList        list.Model
	width          int
	height         int
}

func NewModel(width, height int) Model {
	list := list.New([]list.Item{}, shared.DefaultDelegate, width, height)

	// list.Title = "User: " + user.Name
	list.SetStatusBarItemName("Organization", "Organizations")
	list.Styles.Title = shared.TitleStyle
	list.SetShowTitle(true)

	return Model{orgList: list, width: width, height: height}
}

func (m *Model) SetDimensions(width, height int) {
	m.width = width
	m.height = height
}

func (m Model) Init() (tea.Model, tea.Cmd) {
	return m, getUser
}

func (m Model) GetOrganizationItems(query Query) []list.Item {
	items := make([]list.Item, len(query.User.Organizations))
	for i, org := range m.query.User.Organizations {
		items[i] = shared.NewListItem(org.Login, org.Url)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].FilterValue() < items[j].FilterValue()
	})

	return items
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case queryCompleteMsg:
		m.query = Query(msg)
		items := m.GetOrganizationItems(m.query)
		cmd = m.orgList.SetItems(items)

		return m, cmd
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if item, ok := m.orgList.SelectedItem().(shared.ListItem); ok {
				return m, func() tea.Msg {
					return shared.NextMessage{ModelData: item.Title()}
				}
			}
			// 	selectedOrg := m.organisations[m.orgList.Index()].Login
			// 	cmd = func() tea.Msg {
			// 		return shared.NextMessage{ModelData: selectedOrg}
			// 	}
			return m, cmd
		default:
			m.orgList, cmd = m.orgList.Update(msg)
			return m, cmd
		}
	}

	m.orgList, cmd = m.orgList.Update(msg)

	return m, cmd
}

func (m Model) View() string {
	m.orgList.SetWidth(m.width)
	m.orgList.SetHeight(m.height)
	return shared.AppStyle.Render(m.orgList.View())
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
