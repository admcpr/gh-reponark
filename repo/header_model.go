package repo

import (
	"fmt"
	"gh-reponark/shared"

	"charm.land/bubbles/v2/paginator"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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

func (m HeaderModel) Init() tea.Cmd {
	return nil
}

func (m HeaderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TabSelectMessage:
		if msg.Index >= 0 && msg.Index < len(m.titles) {
			m.paginator.Page = msg.Index
		}
	}
	return m, nil
}

func (m HeaderModel) View() tea.View {
	if len(m.titles) == 0 {
		return tea.NewView("")
	}

	page := m.paginator.Page
	if page >= len(m.titles) {
		page = len(m.titles) - 1
	}

	heading := shared.TitleStyle.Render(m.titles[page])
	return tea.NewView(fmt.Sprint(lipgloss.JoinVertical(lipgloss.Left, heading+"\n"+m.paginator.View())))
}
