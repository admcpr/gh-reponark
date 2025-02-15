package user

import (
	"sort"

	"gh-reponark/org"
	"gh-reponark/shared"

	"github.com/charmbracelet/bubbles/v2/list"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/cli/go-gh/v2"
	"github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
)

type AuthenticationErrorMsg struct{ Err error }

// type AuthenticatedMsg struct{ login string }

type ErrMsg struct{ Err error }
type OrgListMsg struct{ Organisations []org.Organisation }
type userQueryMsg Query

type Model struct {
	organisations []org.Organisation
	// login          string
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

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case OrgListMsg:
		m.organisations = msg.Organisations
		sort.Slice(m.organisations, func(i, j int) bool {
			return m.organisations[i].Login < m.organisations[j].Login
		})

		items := make([]list.Item, len(m.organisations))
		for i, org := range m.organisations {
			items[i] = shared.NewListItem(org.Login, org.Url)
		}

		cmd = m.orgList.SetItems(items)

		return m, cmd
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			selectedOrg := m.organisations[m.orgList.Index()].Login
			cmd = func() tea.Msg {
				return shared.NextMessage{ModelData: selectedOrg}
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

func (m Model) View() string {
	m.orgList.SetWidth(m.width)
	m.orgList.SetHeight(m.height)
	return shared.AppStyle.Render(m.orgList.View())
}

func (m Model) SelectedOrg() org.Organisation {
	return m.organisations[m.orgList.Index()]
}

func getUser() tea.Msg {
	args := []string{"api", "user", "-q", ".login"}
	stdOut, _, err := gh.Exec(args...)
	if err != nil {
		return AuthenticationErrorMsg{Err: err}
	}
	login := stdOut.String()

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

	return userQueryMsg(userQuery)
}
