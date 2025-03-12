package org

import (
	"fmt"
	"log"
	"sort"

	"gh-reponark/filters"
	"gh-reponark/repo"
	"gh-reponark/shared"

	"github.com/charmbracelet/bubbles/v2/list"
	"github.com/charmbracelet/bubbles/v2/progress"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
)

type ApiErrorMsg struct{ Err error }
type orgQueryMsg OrgQuery
type repoQueryMsg repo.Query

type Model struct {
	Title     string
	repoCount int
	repos     []repo.RepoConfig
	filters   filters.FilterMap

	repoList  list.Model
	repoModel repo.Model

	width  int
	height int

	progress progress.Model
}

func NewModel(modelData interface{}, width, height int) *Model {
	title := modelData.(string)
	return &Model{
		Title:     title,
		width:     width,
		height:    height,
		repoModel: repo.NewModel(width/2, height),
		repoList:  list.New([]list.Item{}, shared.SimpleItemDelegate{}, width/2, height),
		progress:  progress.New(progress.WithoutPercentage()),
	}
}

func (m *Model) SetDimensions(width, height int) {
	m.width = width
	m.height = height
}

func (m *Model) populateRepoList() {
	filteredRepositories := m.filters.FilterRepos(m.repos)
	items := make([]list.Item, len(filteredRepositories))
	for i, repo := range filteredRepositories {
		items[i] = shared.SimpleItem(repo.Name)
	}

	list := list.New(items, shared.SimpleItemDelegate{}, m.width/2, m.height-2)
	list.Title = fmt.Sprintf("Organization: %s ", m.Title)
	list.Styles.Title = shared.TitleStyle
	list.SetStatusBarItemName("Repository", "Repositories")
	list.SetShowHelp(false)
	list.SetShowTitle(true)

	m.repoList = list
	m.repoModel.SelectRepo(m.repos[m.repoList.Index()])
}

func (m Model) Init() (tea.Model, tea.Cmd) {
	return m, getRepoList(m.Title)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case orgQueryMsg:
		repos := msg.Organization.Repositories.Nodes
		cmds := []tea.Cmd{m.progress.SetPercent(0.1)}
		m.repoCount = len(msg.Organization.Repositories.Nodes)
		for _, repo := range repos {
			cmds = append(cmds, getRepoDetails(m.Title, repo.Name))
		}
		return m, tea.Batch(cmds...)

	case repoQueryMsg:
		m.repos = append(m.repos, repo.NewRepoConfig(msg.Repository))

		if m.repoCount == len(m.repos) {
			sort.Slice(m.repos, func(i, j int) bool {
				return m.repos[i].Name < m.repos[j].Name
			})
			m.populateRepoList()
			cmd = m.progress.SetPercent(1.0)
		} else {
			cmd = m.progress.IncrPercent(0.9 / float64(m.repoCount))
		}
		return m, cmd

	case filters.FiltersMsg:
		m.filters = filters.FilterMap(msg)
		m.populateRepoList()
		return m, nil

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	case tea.KeyPressMsg:
		switch msg.String() {
		case "F", "f":
			return m, handleNext
		}
		switch msg.String() {
		case "f":
			return m, func() tea.Msg {
				return shared.NextMsg[filters.FilterMap]{ModelData: m.filters}
			}
		case "esc":
			return m, func() tea.Msg {
				return shared.PreviousMsg{}
			}
		case "tab", "shift+tab":
			repoModel, cmd := m.repoModel.Update(msg)
			m.repoModel = repoModel.(repo.Model)
			return m, cmd
		default:
			m.repoList, cmd = m.repoList.Update(msg)
			return m, cmd
		}
	default:
		m.repoList, cmd = m.repoList.Update(msg)
	}

	return m, cmd
}

func (m Model) View() string {
	if m.progress.Percent() < 1 {
		return m.ProgressView()
	}
	m.repoModel.SelectRepo(m.repos[m.repoList.Index()])

	var repoList = shared.AppStyle.Width(shared.Half(m.width)).Render(m.repoList.View())
	var settings = shared.AppStyle.Width(shared.Half(m.width)).Render(m.repoModel.View())
	var rightPanel = lipgloss.JoinVertical(lipgloss.Center, settings)

	var views = []string{repoList, rightPanel}

	return lipgloss.JoinHorizontal(lipgloss.Top, views...)
}

func (m Model) ProgressView() string {
	m.progress.SetWidth(m.width)
	text := fmt.Sprintf("Getting repositories ... %d of %d\n", len(m.repos), m.repoCount)
	return lipgloss.JoinVertical(lipgloss.Center, text, m.progress.View())
}

func getRepoDetails(owner string, name string) tea.Cmd {
	return func() tea.Msg {
		client, err := api.DefaultGraphQLClient()
		if err != nil {
			log.Fatal(err)
		}
		repoQuery := repo.Query{}

		variables := map[string]interface{}{
			"owner": graphql.String(owner),
			"name":  graphql.String(name),
		}
		err = client.Query("Repository", &repoQuery, variables)
		if err != nil {
			log.Fatal(err)
		}

		return repoQueryMsg(repoQuery)
	}
}

func getRepoList(login string) tea.Cmd {
	return func() tea.Msg {
		client, err := api.DefaultGraphQLClient()
		if err != nil {
			return ApiErrorMsg{Err: err}
		}

		var organizationQuery = OrgQuery{}

		variables := map[string]interface{}{
			"login": graphql.String(login),
			"first": graphql.Int(100),
		}
		err = client.Query("OrganizationRepositories", &organizationQuery, variables)
		if err != nil {
			log.Fatal(err)
		}

		return orgQueryMsg(organizationQuery)
	}
}

func handleNext() tea.Msg {
	return shared.NextMsg[any]{}
}
