package repo

import (
	"gh-reponark/shared"

	"github.com/charmbracelet/bubbles/v2/paginator"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
)

type TabSelectMessage struct{ Index int }

type HeaderModel struct {
	titles    []string
	paginator paginator.Model
	width     int
	height    int
}

func NewHeaderModel(width int, titles []string, index int) HeaderModel {
	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = 1
	p.SetTotalPages(len(titles))
	p.Page = index

	return HeaderModel{
		titles:    titles,
		paginator: p,
		width:     width,
	}
}

func (m *HeaderModel) SetDimensions(width, height int) {
	m.width = width
	m.height = height
}

func (m HeaderModel) Init() (tea.Model, tea.Cmd) {
	return m, nil
}

func (m HeaderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m HeaderModel) View() string {
	heading := shared.TitleStyle.Render(m.titles[m.paginator.Page])
	return lipgloss.JoinVertical(lipgloss.Left, heading+"\n"+m.paginator.View())
}
