package repo

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

var headerTitles = []string{"Overview", "Status", "Metrics"}

func TestNewHeaderModel(t *testing.T) {
	m := NewHeaderModel(80, headerTitles, 1)

	assert.Equal(t, headerTitles, m.titles)
	assert.Equal(t, 80, m.width)
	assert.Equal(t, 1, m.paginator.Page)
	assert.Equal(t, len(headerTitles), m.paginator.TotalPages)
	assert.Equal(t, 1, m.paginator.PerPage)
}

func TestHeaderModel_SetDimensions(t *testing.T) {
	m := NewHeaderModel(80, headerTitles, 0)

	m.SetDimensions(100, 50)

	assert.Equal(t, 100, m.width)
	assert.Equal(t, 50, m.height)
}

func TestHeaderModel_Init(t *testing.T) {
	m := NewHeaderModel(80, headerTitles, 0)
	assert.Nil(t, m.Init())
}

func TestHeaderModel_Update(t *testing.T) {
	tests := []struct {
		name     string
		msg      tea.Msg
		wantPage int
	}{
		{name: "select valid tab", msg: TabSelectMessage{Index: 2}, wantPage: 2},
		{name: "select first tab", msg: TabSelectMessage{Index: 0}, wantPage: 0},
		{name: "negative index is ignored", msg: TabSelectMessage{Index: -1}, wantPage: 1},
		{name: "out of range index is ignored", msg: TabSelectMessage{Index: 3}, wantPage: 1},
		{name: "unrelated message is ignored", msg: tea.WindowSizeMsg{Width: 1, Height: 1}, wantPage: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewHeaderModel(80, headerTitles, 1)

			updated, cmd := m.Update(tt.msg)

			assert.Nil(t, cmd)
			assert.Equal(t, tt.wantPage, updated.(HeaderModel).paginator.Page)
		})
	}
}

func TestHeaderModel_View(t *testing.T) {
	tests := []struct {
		name        string
		titles      []string
		page        int
		wantTitle   string
		wantMissing string
	}{
		{name: "first page", titles: headerTitles, page: 0, wantTitle: "Overview", wantMissing: "Status"},
		{name: "last page", titles: headerTitles, page: 2, wantTitle: "Metrics", wantMissing: "Overview"},
		{name: "page beyond range clamps to last", titles: headerTitles, page: 10, wantTitle: "Metrics", wantMissing: "Overview"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewHeaderModel(80, tt.titles, 0)
			m.paginator.Page = tt.page

			content := plain(m.View())

			assert.Contains(t, content, tt.wantTitle)
			assert.NotContains(t, content, tt.wantMissing)
		})
	}
}

func TestHeaderModel_View_NoTitles(t *testing.T) {
	m := NewHeaderModel(80, []string{}, 0)
	assert.Equal(t, "", plain(m.View()))
}
