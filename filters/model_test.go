package filters

import (
	"strings"
	"testing"
	"time"

	"gh-reponark/filter"
	"gh-reponark/repo"
	"gh-reponark/ui"
	"gh-reponark/ui/uitest"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
)

// press sends each key to the model in turn. Named keys are spelled out;
// anything else is typed a character at a time.
func press(m *Model, keys ...string) {
	for _, k := range keys {
		switch k {
		case "enter":
			m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		case "esc":
			m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
		case "tab":
			m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		case "space":
			m.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})
		case "backspace":
			m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
		default:
			for _, r := range k {
				m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
			}
		}
	}
}

func testRepos() []repo.RepoConfig {
	archived := repo.Repository{Name: "old", IsArchived: true, StargazerCount: 3}
	archived.PrimaryLanguage.Name = "Ruby"
	active := repo.Repository{Name: "new", StargazerCount: 40}
	active.PrimaryLanguage.Name = "Go"
	busy := repo.Repository{Name: "busy", StargazerCount: 900}
	busy.PrimaryLanguage.Name = "Go"
	return []repo.RepoConfig{repo.NewRepoConfig(archived), repo.NewRepoConfig(active), repo.NewRepoConfig(busy)}
}

func newTestModel() *Model {
	return NewModel(nil, testRepos(), 100, 30)
}

// selectProperty searches for a property and leaves the list highlighting it.
func selectProperty(m *Model, query string) {
	press(m, "/", query, "enter")
}

func TestNewModel(t *testing.T) {
	m := newTestModel()

	assert.Empty(t, m.filters)
	assert.Len(t, m.properties, len(repo.Schema()), "every property can be filtered on")
	p, ok := m.selected()
	assert.True(t, ok)
	assert.Equal(t, "Id", p.Name, "the first property is highlighted")
	assert.Nil(t, m.Init())
}

func TestNewModel_CopiesCurrentFilters(t *testing.T) {
	current := filter.FilterMap{"Is Archived": filter.NewBoolFilter("Is Archived", true)}

	m := NewModel(current, nil, 100, 30)
	press(m, "X")

	assert.Empty(t, m.Filters())
	assert.Len(t, current, 1, "the caller only sees changes via FiltersMsg")
}

func TestModel_SetDimensions(t *testing.T) {
	m := newTestModel()

	m.SetDimensions(120, 50)

	assert.Equal(t, 120, m.width)
	assert.Equal(t, 50, m.height)
}

func TestModel_Search(t *testing.T) {
	m := newTestModel()

	press(m, "/", "wi")
	assert.True(t, m.searching)
	content := uitest.Plain(m.View())
	assert.Contains(t, content, "Has Wiki Enabled", "search matches anywhere in the name")
	assert.NotContains(t, content, "Is Archived")

	press(m, "enter")
	assert.False(t, m.searching)
	assert.Equal(t, "wi", m.search.Value(), "the search stays applied")

	press(m, "/", "esc")
	assert.Empty(t, m.search.Value(), "esc clears the search")
	assert.Len(t, m.matches, len(m.properties))
}

func TestModel_Search_MatchesGroupTitles(t *testing.T) {
	m := newTestModel()

	selectProperty(m, "merge")

	names := []string{}
	for _, i := range m.matches {
		names = append(names, m.properties[i].Name)
	}
	assert.Contains(t, names, "Default Branch", "every property in the Merge group matches")
}

func TestModel_Search_NoMatches(t *testing.T) {
	m := newTestModel()

	selectProperty(m, "zzz")

	assert.Contains(t, uitest.Plain(m.View()), "No properties match")
	press(m, "enter", "x", "space")
	assert.False(t, m.editing, "there is nothing to edit")
	assert.Empty(t, m.filters)
}

func TestModel_Navigation(t *testing.T) {
	m := newTestModel()

	press(m, "j", "j")
	p, _ := m.selected()
	assert.Equal(t, "Name", p.Name)

	press(m, "G")
	p, _ = m.selected()
	assert.Equal(t, m.properties[len(m.properties)-1].Name, p.Name)

	press(m, "g", "k")
	p, _ = m.selected()
	assert.Equal(t, "Id", p.Name, "the cursor stops at the top")
}

func TestModel_ToggleBoolFromList(t *testing.T) {
	m := newTestModel()
	selectProperty(m, "is archived")

	press(m, "space")
	assert.Equal(t, filter.NewBoolFilter("Is Archived", true), m.filters["Is Archived"])
	assert.Equal(t, "1 of 3 repos match", m.Status())

	press(m, "space")
	assert.Equal(t, filter.NewBoolFilter("Is Archived", false), m.filters["Is Archived"])
	assert.Equal(t, "2 of 3 repos match", m.Status())

	press(m, "space")
	assert.NotContains(t, m.filters, "Is Archived", "the third press goes back to any")
}

func TestModel_EditBool(t *testing.T) {
	m := newTestModel()
	selectProperty(m, "is archived")

	press(m, "enter")
	assert.True(t, m.editing)
	press(m, "n")
	assert.Equal(t, filter.NewBoolFilter("Is Archived", false), m.filters["Is Archived"], "changes apply as they are made")

	press(m, "enter")
	assert.False(t, m.editing)
	assert.Contains(t, m.filters, "Is Archived", "enter keeps the change")
}

func TestModel_EditCancelRestores(t *testing.T) {
	m := NewModel(filter.FilterMap{"Is Archived": filter.NewBoolFilter("Is Archived", true)}, testRepos(), 100, 30)
	selectProperty(m, "is archived")

	press(m, "enter", "a")
	assert.NotContains(t, m.filters, "Is Archived")

	press(m, "esc")
	assert.False(t, m.editing)
	assert.Equal(t, filter.NewBoolFilter("Is Archived", true), m.filters["Is Archived"], "esc puts back the filter as it was")
	assert.Equal(t, 1, m.editor.(*boolEditor).choice, "and the editor shows it again")
}

func TestModel_EditInt(t *testing.T) {
	m := newTestModel()
	selectProperty(m, "stargazer")

	press(m, "enter", "10")
	assert.Equal(t, filter.NewIntFilter("Stargazer Count", 10, filter.NoMax), m.filters["Stargazer Count"])
	assert.Equal(t, "2 of 3 repos match", m.Status())

	press(m, "tab", "x")
	assert.Contains(t, uitest.Plain(m.View()), `"x" is not a whole number`)
	assert.Equal(t, filter.NewIntFilter("Stargazer Count", 10, filter.NoMax), m.filters["Stargazer Count"], "an unreadable bound keeps the last good filter")

	press(m, "backspace", "100", "enter")
	assert.Equal(t, filter.NewIntFilter("Stargazer Count", 10, 100), m.filters["Stargazer Count"])
	assert.Contains(t, uitest.Plain(m.View()), "10 – 100", "the list shows the condition")
}

func TestModel_EditText(t *testing.T) {
	m := newTestModel()
	selectProperty(m, "language")

	press(m, "enter", "go", "enter")

	assert.Equal(t, filter.NewStringFilter("Primary Language", "go"), m.filters["Primary Language"])
	assert.Equal(t, "2 of 3 repos match", m.Status())
}

func TestModel_Clear(t *testing.T) {
	m := NewModel(filter.FilterMap{
		"Is Archived":     filter.NewBoolFilter("Is Archived", true),
		"Stargazer Count": filter.NewIntFilter("Stargazer Count", 1, 5),
	}, nil, 100, 30)
	selectProperty(m, "is archived")

	press(m, "x")
	assert.NotContains(t, m.filters, "Is Archived")
	assert.Contains(t, m.filters, "Stargazer Count")

	press(m, "X")
	assert.Empty(t, m.filters)
}

func TestModel_BackSendsFilters(t *testing.T) {
	m := newTestModel()
	selectProperty(m, "is archived")
	press(m, "space")

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if assert.NotNil(t, cmd) {
		prev, ok := cmd().(ui.PreviousMsg)
		assert.True(t, ok)
		assert.Equal(t, FiltersMsg(m.filters), prev.Message)
	}
}

func TestModel_EscLeavesEditingBeforeScreen(t *testing.T) {
	m := newTestModel()
	press(m, "enter")

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	assert.Nil(t, cmd, "esc while editing only stops editing")
	assert.False(t, m.editing)
}

func TestModel_View(t *testing.T) {
	m := NewModel(filter.FilterMap{"Is Archived": filter.NewBoolFilter("Is Archived", false)}, testRepos(), 100, 30)
	selectProperty(m, "is archived")

	content := uitest.Plain(m.View())

	assert.Contains(t, content, "▐Is Archived: no▌", "active filters are chips")
	assert.Contains(t, content, "STATUS", "properties are grouped under column headings")
	assert.Contains(t, content, "● Is Archived ▐toggle▌", "the editor names the property and its type")
	assert.Contains(t, content, "Indicates if the repository is archived.")
	assert.Contains(t, content, "▐any▌ ▐yes▌ ▐no▌", "the choices are a segmented control")
	assert.Contains(t, content, "● yes 1   ○ no 2", "the chart describes the loaded repos")
	assert.Contains(t, content, "2 of 3", "matching repos are counted")
	assert.Contains(t, content, "● new  ● busy", "and listed")
}

func TestModel_View_FitsDimensions(t *testing.T) {
	m := NewModel(filter.FilterMap{"Is Archived": filter.NewBoolFilter("Is Archived", false)}, testRepos(), 90, 20)

	lines := strings.Split(uitest.Plain(m.View()), "\n")

	assert.Len(t, lines, 20)
	for _, line := range lines {
		assert.Equal(t, 90, lipgloss.Width(line), "%q", line)
	}
}

func TestModel_Breadcrumb(t *testing.T) {
	assert.Equal(t, "Filters", newTestModel().Breadcrumb())
}

func TestModel_Status_WithoutRepos(t *testing.T) {
	m := NewModel(filter.FilterMap{"Is Archived": filter.NewBoolFilter("Is Archived", true)}, nil, 80, 24)
	assert.Equal(t, "1 active", m.Status())
}

func TestModel_Help(t *testing.T) {
	m := newTestModel()
	assert.Equal(t, "j/k property  l/enter edit  / search  x clear  esc done", m.Help().String())
	assert.False(t, m.Typing())

	selectProperty(m, "is archived")
	assert.Equal(t, "j/k property  l/enter edit  space toggle  / search  x clear  esc done", m.Help().String(),
		"space only toggles yes/no properties")

	press(m, "enter")
	assert.Equal(t, "←/→ choose  y yes  n no  a any  enter done  esc cancel", m.Help().String())
	assert.False(t, m.Typing(), "the yes/no editor takes single keys")
	assert.Len(t, m.Help().FullHelp(), 2)

	press(m, "esc", "/")
	assert.Equal(t, "type to search  ↓/enter done  esc clear", m.Help().String())
	assert.True(t, m.Typing())

	press(m, "esc", "/", "stargazer", "enter", "enter")
	assert.Equal(t, "tab/↑/↓ other field  enter done  esc cancel", m.Help().String())
	assert.True(t, m.Typing(), "number fields take text")
}

func TestModel_SearchTypesEveryLetter(t *testing.T) {
	m := newTestModel()

	press(m, "/", "proj")

	assert.True(t, m.searching, "j is typed, not taken as a move")
	assert.Equal(t, "proj", m.search.Value())
	press(m, "?")
	assert.Equal(t, "proj?", m.search.Value())
}

func TestModel_BackspaceDoesNotClear(t *testing.T) {
	m := NewModel(filter.FilterMap{"Id": filter.NewStringFilter("Id", "x")}, nil, 100, 30)

	press(m, "backspace")

	assert.Contains(t, m.filters, "Id", "only x clears a filter")
}

func TestParseDate(t *testing.T) {
	today := time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)
	original := now
	now = func() time.Time { return today }
	t.Cleanup(func() { now = original })

	tests := []struct {
		input string
		want  time.Time
	}{
		{input: "2024-06-30", want: time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)},
		{input: "30d", want: today.AddDate(0, 0, -30)},
		{input: "2w", want: today.AddDate(0, 0, -14)},
		{input: "6m", want: today.AddDate(0, -6, 0)},
		{input: "1Y", want: today.AddDate(-1, 0, 0)},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseDate(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	_, err := parseDate("last week")
	assert.Error(t, err)
}

func TestDateEditor(t *testing.T) {
	e := newDateEditor("Pushed At", nil)
	f, err := e.Filter()
	assert.Nil(t, f, "blank bounds mean no filter")
	assert.NoError(t, err)

	e.Focus()
	e.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	for _, r := range "2024-01-01" {
		e.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	f, err = e.Filter()
	assert.NoError(t, err)
	assert.Equal(t, filter.NewDateFilter("Pushed At", time.Time{}, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)), f)

	e = newDateEditor("Pushed At", filter.NewDateFilter("Pushed At", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)))
	_, err = e.Filter()
	assert.EqualError(t, err, "the start is after the end")
}

func TestIntEditor_SeedsAndValidates(t *testing.T) {
	e := newIntEditor("Stars", filter.NewIntFilter("Stars", 5, filter.NoMax))
	assert.Equal(t, "5", e.inputs[0].Value())
	assert.Equal(t, "", e.inputs[1].Value(), "an open bound is left blank")

	e = newIntEditor("Stars", filter.NewIntFilter("Stars", 9, 1))
	_, err := e.Filter()
	assert.EqualError(t, err, "the minimum is above the maximum")
}

func TestIsSupportedPropertyType(t *testing.T) {
	for _, supported := range []string{"bool", "int", "time.Time", "string"} {
		assert.True(t, isSupportedPropertyType(supported), supported)
	}
	assert.False(t, isSupportedPropertyType("[]string"))
}
