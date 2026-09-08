package repo

import (
	"fmt"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

// plain renders a view to a string with all ANSI styling removed so tests can
// assert on the visible text.
func plain(v tea.View) string {
	return ansi.Strip(fmt.Sprint(v.Content))
}

func newTestRepoConfig() RepoConfig {
	return NewRepoConfig(Repository{
		Name:           "test-repo",
		Url:            "https://github.com/owner/test-repo",
		IsArchived:     true,
		StargazerCount: 42,
		CreatedAt:      time.Date(2024, 3, 4, 0, 0, 0, 0, time.UTC),
	})
}

func newSelectedModel() Model {
	m := NewModel(80, 24)
	m.SelectRepo(newTestRepoConfig())
	return m
}

func TestNewModel(t *testing.T) {
	m := NewModel(80, 24)

	assert.Equal(t, 80, m.width)
	assert.Equal(t, 24, m.height)
	assert.Equal(t, 0, m.activeTab)
	assert.NotNil(t, m.repository.Properties)
	assert.NotNil(t, m.repository.PropertyGroups)
	assert.Empty(t, m.repository.GroupKeys)
	assert.Empty(t, m.repoHeader.titles)
}

func TestModel_SetDimensions(t *testing.T) {
	m := NewModel(80, 24)

	m.SetDimensions(120, 40)

	assert.Equal(t, 120, m.width)
	assert.Equal(t, 40, m.height)
	assert.Equal(t, 120, m.repoHeader.width)
	assert.Equal(t, 40, m.repoHeader.height)
	assert.Equal(t, 120, m.help.Width())
}

func TestModel_Init(t *testing.T) {
	m := NewModel(80, 24)
	assert.Nil(t, m.Init())
}

func TestModel_SelectRepo(t *testing.T) {
	m := NewModel(80, 24)
	config := newTestRepoConfig()

	m.SelectRepo(config)

	firstGroup := config.GroupKeys[0]
	assert.Equal(t, config.Name, m.repository.Name)
	assert.Equal(t, config.GroupKeys, m.repoHeader.titles)
	assert.Equal(t, 0, m.repoHeader.paginator.Page)
	assert.Equal(t, firstGroup, m.settingsList.Title)
	assert.Len(t, m.settingsList.Items(), len(config.PropertyGroups[firstGroup]))
}

func TestModel_SelectRepo_KeepsActiveTab(t *testing.T) {
	m := newSelectedModel()
	m.SelectTab(2)

	m.SelectRepo(newTestRepoConfig())

	assert.Equal(t, 2, m.activeTab)
	assert.Equal(t, 2, m.repoHeader.paginator.Page)
	assert.Equal(t, m.repository.GroupKeys[2], m.settingsList.Title)
}

func TestModel_SelectTab(t *testing.T) {
	m := newSelectedModel()

	m.SelectTab(3)

	key := m.repository.GroupKeys[3]
	assert.Equal(t, 3, m.activeTab)
	assert.Equal(t, key, m.settingsList.Title)
	assert.Len(t, m.settingsList.Items(), len(m.repository.PropertyGroups[key]))
}

func TestModel_Update_TabNavigation(t *testing.T) {
	tab := tea.KeyPressMsg{Code: tea.KeyTab}
	shiftTab := tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}

	tests := []struct {
		name       string
		startTab   int
		key        tea.KeyPressMsg
		wantTab    int
		lastOffset int
	}{
		{name: "tab moves forward", startTab: 0, key: tab, wantTab: 1},
		{name: "tab wraps to first", startTab: -1, key: tab, wantTab: 0, lastOffset: -1},
		{name: "shift+tab moves backward", startTab: 2, key: shiftTab, wantTab: 1},
		{name: "shift+tab wraps to last", startTab: 0, key: shiftTab, wantTab: -1, lastOffset: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newSelectedModel()
			groupCount := len(m.repository.PropertyGroups)

			startTab := tt.startTab
			if startTab < 0 {
				startTab = groupCount + startTab
			}
			wantTab := tt.wantTab
			if wantTab < 0 {
				wantTab = groupCount + wantTab
			}
			m.SelectTab(startTab)

			updated, cmd := m.Update(tt.key)
			got := updated.(Model)

			assert.Nil(t, cmd)
			assert.Equal(t, wantTab, got.activeTab)
			assert.Equal(t, wantTab, got.repoHeader.paginator.Page)
			assert.Equal(t, got.repository.GroupKeys[wantTab], got.settingsList.Title)
		})
	}
}

func TestModel_Update_IgnoresOtherKeys(t *testing.T) {
	m := newSelectedModel()

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})

	assert.Nil(t, cmd)
	assert.Equal(t, 0, updated.(Model).activeTab)
}

func TestModel_View(t *testing.T) {
	m := newSelectedModel()

	content := plain(m.View())

	assert.Contains(t, content, m.repository.GroupKeys[0])
	assert.Contains(t, content, "Name")
	assert.Contains(t, content, "test-repo")
}

func TestModel_HeaderView(t *testing.T) {
	m := newSelectedModel()
	m.SelectTab(1)
	repoHeader, _ := m.repoHeader.Update(TabSelectMessage{Index: 1})
	m.repoHeader = repoHeader.(HeaderModel)

	content := plain(m.HeaderView())

	assert.Contains(t, content, m.repository.GroupKeys[1])
	assert.NotContains(t, content, m.repository.GroupKeys[0])
}

func TestModel_HelpView(t *testing.T) {
	m := NewModel(80, 24)

	content := plain(m.HelpView())

	assert.Contains(t, content, "next tab")
	assert.Contains(t, content, "prev tab")
	assert.NotContains(t, content, "quit", "only keys the pane actually handles are advertised")
}

func TestModel_Keys(t *testing.T) {
	m := NewModel(80, 24)

	keys := m.Keys()

	assert.Equal(t, []string{"tab"}, keys.NextTab.Keys())
	assert.Equal(t, []string{"shift+tab"}, keys.PrevTab.Keys())
}

func TestNewSettingsList(t *testing.T) {
	settings := []RepoProperty{
		{Name: "Zulu", Value: true},
		{Name: "Alpha", Value: 7},
		{Name: "Mike", Value: "text"},
	}

	list := NewSettingsList(settings, "Group", 80, 24)

	assert.Equal(t, "Group", list.Title)
	items := list.Items()
	assert.Len(t, items, 3)

	titles := make([]string, len(items))
	descriptions := make([]string, len(items))
	for i, item := range items {
		listItem := item.(interface {
			Title() string
			Description() string
		})
		titles[i] = listItem.Title()
		descriptions[i] = listItem.Description()
	}

	assert.Equal(t, []string{"Alpha", "Mike", "Zulu"}, titles)
	assert.Equal(t, []string{"7", "text", "Yes"}, descriptions)
}

func TestNewSettingsList_Empty(t *testing.T) {
	list := NewSettingsList([]RepoProperty{}, "Empty", 80, 24)

	assert.Equal(t, "Empty", list.Title)
	assert.Empty(t, list.Items())
}
