package filters

import (
	"fmt"
	"testing"
	"time"

	"gh-reponark/repo"
	"gh-reponark/shared"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestNewModel(t *testing.T) {
	m := NewModel(nil, 80, 30)

	assert.Equal(t, 80, m.width)
	assert.Equal(t, 30, m.height)
	assert.NotNil(t, m.filters)
	assert.Empty(t, m.filters)
	assert.NotNil(t, m.filterSearch)
}

func TestNewModel_SeedsWithCurrentFilters(t *testing.T) {
	current := FilterMap{"Is Archived": NewBoolFilter("Is Archived", true)}

	m := NewModel(current, 80, 30)

	assert.Equal(t, current, m.filters)
	assert.Contains(t, plain(m.View()), "Is Archived")

	// The screen works on its own copy so the caller only sees changes via FiltersMsg.
	m.Update(AddFilterMsg(NewIntFilter("Stargazer Count", 1, 10)))
	assert.Len(t, m.filters, 2)
	assert.Len(t, current, 1)
}

func TestModel_SetDimensions(t *testing.T) {
	m := NewModel(nil, 80, 30)

	m.SetDimensions(100, 50)

	assert.Equal(t, 100, m.width)
	assert.Equal(t, 50, m.height)
	assert.Equal(t, 100, m.help.Width())
}

func TestModel_Init(t *testing.T) {
	m := NewModel(nil, 80, 30)
	assert.NotNil(t, m.Init())
}

func TestModel_Update_BackSendsFilters(t *testing.T) {
	tests := []struct {
		name string
		key  tea.KeyPressMsg
	}{
		{name: "esc", key: tea.KeyPressMsg{Code: tea.KeyEscape}},
		{name: "ctrl+enter", key: tea.KeyPressMsg{Code: tea.KeyEnter, Mod: tea.ModCtrl}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(nil, 80, 30)
			m.filters["Is Archived"] = NewBoolFilter("Is Archived", true)

			_, cmd := m.Update(tt.key)
			assert.NotNil(t, cmd)

			prev, ok := cmd().(shared.PreviousMsg)
			assert.True(t, ok, "expected a PreviousMsg")

			filters, ok := prev.Message.(FiltersMsg)
			assert.True(t, ok, "expected the PreviousMsg to carry a FiltersMsg")
			assert.Len(t, filters, 1)
			assert.Equal(t, "Is Archived", filters["Is Archived"].Name())
		})
	}
}

func TestModel_Update_AddFilterMsg(t *testing.T) {
	m := NewModel(nil, 80, 30)

	// Type something into the search so we can check it gets reset.
	m.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	assert.Equal(t, "a", m.filterSearch.(FilterSearchModel).textinput.Value())

	_, cmd := m.Update(AddFilterMsg(NewIntFilter("Stargazer Count", 1, 10)))

	assert.NotNil(t, cmd, "adding a filter should re-initialise the search")
	assert.Len(t, m.filters, 1)
	assert.Equal(t, "Stargazer Count between 1 and 10", m.filters["Stargazer Count"].String())
	assert.Equal(t, "", m.filterSearch.(FilterSearchModel).textinput.Value(), "search should be reset")
}

func TestModel_Update_AddFilterMsgReplacesSameName(t *testing.T) {
	m := NewModel(nil, 80, 30)

	m.Update(AddFilterMsg(NewBoolFilter("Is Archived", true)))
	m.Update(AddFilterMsg(NewBoolFilter("Is Archived", false)))

	assert.Len(t, m.filters, 1)
	assert.Equal(t, "Is Archived = No", m.filters["Is Archived"].String())
}

func TestModel_Update_DelegatesToSearch(t *testing.T) {
	m := NewModel(nil, 80, 30)

	m.Update(tea.KeyPressMsg{Code: 'N', Text: "N"})

	search := m.filterSearch.(FilterSearchModel)
	assert.Equal(t, "N", search.textinput.Value())
	prop, ok := search.CurrentPropertySuggestion()
	assert.True(t, ok)
	assert.Equal(t, "Name", prop.Name)
}

func TestNewFilterModel(t *testing.T) {
	tests := []struct {
		name     string
		property repo.PropertySchema
		check    func(t *testing.T, m tea.Model)
	}{
		{
			name:     "bool",
			property: repo.PropertySchema{Name: "Is Archived", Type: "bool"},
			check: func(t *testing.T, m tea.Model) {
				bm, ok := m.(*BoolModel)
				assert.True(t, ok)
				assert.Equal(t, "Is Archived", bm.Name())
				assert.False(t, bm.Value())
			},
		},
		{
			name:     "int",
			property: repo.PropertySchema{Name: "Stargazer Count", Type: "int"},
			check: func(t *testing.T, m tea.Model) {
				im, ok := m.(*IntModel)
				assert.True(t, ok)
				assert.Equal(t, "Stargazer Count", im.Name())
				assert.Equal(t, "0", im.fromInput.Placeholder)
				assert.Equal(t, "100000", im.toInput.Placeholder)
			},
		},
		{
			name:     "time.Time",
			property: repo.PropertySchema{Name: "Created At", Type: "time.Time"},
			check: func(t *testing.T, m tea.Model) {
				dm, ok := m.(*DateModel)
				assert.True(t, ok)
				assert.Equal(t, "Created At", dm.Name())
				assert.Equal(t, time.Time{}.Format("2006-01-02"), dm.fromInput.Placeholder)
				assert.Equal(t, time.Now().Format("2006-01-02"), dm.toInput.Placeholder)
			},
		},
		{
			name:     "string",
			property: repo.PropertySchema{Name: "Primary Language", Type: "string"},
			check: func(t *testing.T, m tea.Model) {
				sm, ok := m.(*StringModel)
				assert.True(t, ok)
				assert.Equal(t, "Primary Language", sm.Name())
				assert.Equal(t, "", sm.Value())
			},
		},
		{
			name:     "unsupported type",
			property: repo.PropertySchema{Name: "Languages", Type: "[]string"},
			check: func(t *testing.T, m tea.Model) {
				assert.Nil(t, m)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t, NewFilterModel(tt.property, 60, 40))
		})
	}
}

func TestIsSupportedPropertyType(t *testing.T) {
	tests := []struct {
		typeName string
		want     bool
	}{
		{"bool", true},
		{"int", true},
		{"time.Time", true},
		{"string", true},
		{"float64", false},
		{"[]string", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.typeName), func(t *testing.T) {
			assert.Equal(t, tt.want, isSupportedPropertyType(tt.typeName))
		})
	}
}

func TestModel_View(t *testing.T) {
	m := NewModel(nil, 100, 40)
	m.filters["Is Archived"] = NewBoolFilter("Is Archived", true)

	content := plain(m.View())

	assert.Contains(t, content, "Selected Filters")
	assert.Contains(t, content, "Is Archived")
}

func TestModel_HeaderView(t *testing.T) {
	m := NewModel(nil, 80, 30)
	assert.Contains(t, plain(m.HeaderView()), "Filters")
}

func TestModel_HelpView(t *testing.T) {
	m := NewModel(nil, 80, 30)
	content := plain(m.HelpView())
	assert.NotEmpty(t, content)
	assert.Contains(t, content, "back")
}

func TestNewFiltersList(t *testing.T) {
	filters := map[string]Filter{
		"Stargazer Count": NewIntFilter("Stargazer Count", 1, 10),
		"Is Archived":     NewBoolFilter("Is Archived", true),
		"Created At":      NewDateFilter("Created At", time.Time{}, time.Now()),
	}

	list := NewFiltersList(filters, 80, 30)

	assert.Equal(t, "Selected Filters", list.Title)
	items := list.Items()
	assert.Len(t, items, 3)
	assert.Equal(t, shared.SimpleItem("Created At"), items[0])
	assert.Equal(t, shared.SimpleItem("Is Archived"), items[1])
	assert.Equal(t, shared.SimpleItem("Stargazer Count"), items[2])
}

func TestNewFiltersList_Empty(t *testing.T) {
	list := NewFiltersList(map[string]Filter{}, 80, 30)
	assert.Empty(t, list.Items())
}

func TestFilterKeyMap(t *testing.T) {
	keymap := newFilterKeyMap()

	assert.Equal(t, []string{"enter"}, keymap.Select.Keys())
	assert.Equal(t, []string{"tab"}, keymap.Complete.Keys())
	assert.Equal(t, []string{"down", "ctrl+n"}, keymap.NextSuggestion.Keys())
	assert.Equal(t, []string{"up", "ctrl+p"}, keymap.PrevSuggestion.Keys())
	assert.Equal(t, []string{"esc", "ctrl+enter"}, keymap.Back.Keys())

	short := keymap.ShortHelp()
	assert.Len(t, short, 5)

	full := keymap.FullHelp()
	assert.Len(t, full, 1)
	assert.Equal(t, short, full[0])
}

func TestEditorKeyMap(t *testing.T) {
	keymap := newEditorKeyMap()

	assert.Equal(t, []string{"enter"}, keymap.Apply.Keys())
	assert.Equal(t, []string{"esc"}, keymap.Back.Keys())
	assert.Equal(t, []string{"tab", "shift+tab"}, keymap.NextField.Keys())
	assert.Equal(t, []string{"left", "right"}, keymap.Toggle.Keys())
	assert.Equal(t, []string{"y", "Y"}, keymap.Yes.Keys())
	assert.Equal(t, []string{"n", "N"}, keymap.No.Keys())
}

func TestModel_HelpView_ListsEveryBinding(t *testing.T) {
	m := NewModel(nil, 120, 30)

	content := plain(m.HelpView())

	for _, want := range []string{"add filter", "complete", "next", "prev", "back"} {
		assert.Contains(t, content, want)
	}
}

// plain renders a view to a string with all ANSI styling removed so tests can
// assert on the visible text.
func plain(v tea.View) string {
	return ansi.Strip(fmt.Sprint(v.Content))
}
