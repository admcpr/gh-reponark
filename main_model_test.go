package main

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"gh-reponark/filters"
	"gh-reponark/org"
	"gh-reponark/shared"
	"gh-reponark/user"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestValidateTransition_AllowsExpectedFlow(t *testing.T) {
	tests := []struct {
		name    string
		current tea.Model
		next    tea.Model
		wantErr bool
	}{
		{"user -> org ok", &user.Model{}, &org.Model{}, false},
		{"org -> filters ok", &org.Model{}, &filters.Model{}, false},
		{"filters -> filter detail ok", &filters.Model{}, &filters.BoolModel{}, false},
		{"detail -> anything ok", &filters.BoolModel{}, &user.Model{}, false},
		{"org -> detail blocked", &org.Model{}, &filters.BoolModel{}, true},
		{"user -> filters blocked", &user.Model{}, &filters.Model{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTransition(tt.current, tt.next)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMainModel_Next_FromFiltersCreatesDetail(t *testing.T) {
	m := MainModel{nav: shared.NewNavigator()}

	base := filters.NewModel(nil, 20, 10)
	assert.NoError(t, m.nav.Push(base))

	prop := filters.Property{Name: "is archived", Type: "bool"}
	cmd := m.Next(shared.NextMsg{ModelData: prop})

	assert.NotNil(t, cmd)
	assert.Equal(t, 2, m.nav.Len())

	top, _ := m.nav.Current()
	_, ok := top.(*filters.BoolModel)
	assert.True(t, ok, "top of stack should be BoolModel")
}

func TestMainModel_Previous_EmptyNavQuits(t *testing.T) {
	m := MainModel{nav: shared.NewNavigator()}

	cmd := m.Previous(shared.PreviousMsg{})
	assert.Equal(t, reflect.ValueOf(tea.Quit).Pointer(), reflect.ValueOf(cmd).Pointer())
}

func TestRenderFooter_OrgModelNotEmpty(t *testing.T) {
	m := MainModel{}
	orgModel := org.NewModel(shared.OrgKey{Name: "demo", IsUser: false}, 80, 24)
	footer := m.renderFooter(orgModel)
	assert.NotEmpty(t, strings.TrimSpace(footer))
	assert.Greater(t, lipgloss.Height(footer), 0)
}

// plainModel is a model that provides neither a header nor help so the
// layout falls back to its defaults.
type plainModel struct{}

func (m *plainModel) Init() tea.Cmd                           { return nil }
func (m *plainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m *plainModel) View() tea.View                          { return tea.NewView("plain content") }

// plain renders a view to a string with all ANSI styling removed so tests can
// assert on the visible text.
func plain(v tea.View) string {
	return ansi.Strip(fmt.Sprint(v.Content))
}

// keyPress builds a printable key press for a single character.
func keyPress(text string) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: []rune(text)[0], Text: text}
}

func TestNewMainModel(t *testing.T) {
	m := NewMainModel()

	assert.Equal(t, 1, m.nav.Len())
	top, err := m.nav.Current()
	assert.NoError(t, err)
	_, ok := top.(*user.Model)
	assert.True(t, ok, "the user model should be the first screen")
}

func TestMainModel_Init(t *testing.T) {
	m := NewMainModel()
	assert.NotNil(t, m.Init())
}

func TestMainModel_SetDimensions(t *testing.T) {
	m := NewMainModel()

	m.SetDimensions(120, 40)

	assert.Equal(t, 120, m.width)
	assert.Equal(t, 40, m.height)
}

func TestMainModel_Update_WindowSize(t *testing.T) {
	m := NewMainModel()

	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})

	assert.Nil(t, cmd)
	got := updated.(MainModel)
	assert.Equal(t, 100, got.width)
	assert.Equal(t, 50, got.height)
}

func TestMainModel_Update_CtrlCQuits(t *testing.T) {
	m := NewMainModel()

	_, cmd := m.Update(tea.KeyPressMsg{Code: []rune("c")[0], Mod: tea.ModCtrl})

	if assert.NotNil(t, cmd) {
		_, ok := cmd().(tea.QuitMsg)
		assert.True(t, ok, "ctrl+c should quit")
	}
}

func TestMainModel_Update_KeyGoesToChild(t *testing.T) {
	m := MainModel{nav: shared.NewNavigator()}
	child := filters.NewBoolModel("Is Archived", false, 20, 10)
	assert.NoError(t, m.nav.Push(child))

	m.Update(keyPress("y"))

	assert.True(t, child.Value(), "the key press should reach the child model")
	assert.Equal(t, 1, m.nav.Len())
}

func TestMainModel_Update_NextMsg(t *testing.T) {
	m := NewMainModel()

	updated, cmd := m.Update(shared.NextMsg{ModelData: shared.OrgKey{Name: "demo"}})

	assert.NotNil(t, cmd, "the new screen's Init command should be returned")
	got := updated.(MainModel)
	assert.Equal(t, 2, got.nav.Len())
	top, _ := got.nav.Current()
	_, ok := top.(*org.Model)
	assert.True(t, ok, "top of stack should be the org model")
}

func TestMainModel_Update_PreviousMsg(t *testing.T) {
	m := NewMainModel()
	m.Next(shared.NextMsg{ModelData: shared.OrgKey{Name: "demo"}})
	assert.Equal(t, 2, m.nav.Len())

	updated, cmd := m.Update(shared.PreviousMsg{})

	assert.Nil(t, cmd)
	got := updated.(MainModel)
	assert.Equal(t, 1, got.nav.Len())
	top, _ := got.nav.Current()
	_, ok := top.(*user.Model)
	assert.True(t, ok)
}

func TestMainModel_Update_OtherMessagesGoToChild(t *testing.T) {
	m := MainModel{nav: shared.NewNavigator()}
	child := filters.NewIntModel("Stars", 0, 10, 20, 10)
	assert.NoError(t, m.nav.Push(child))

	_, cmd := m.Update(tea.WindowSizeMsg{Width: 1, Height: 1})
	assert.Nil(t, cmd)
	assert.Equal(t, 1, m.nav.Len())
}

func TestMainModel_Next_Transitions(t *testing.T) {
	tests := []struct {
		name      string
		start     tea.Model
		modelData any
		wantType  tea.Model
	}{
		{name: "user -> org", start: user.NewModel(0, 0), modelData: shared.OrgKey{Name: "demo"}, wantType: &org.Model{}},
		{name: "org -> filters", start: org.NewModel(shared.OrgKey{Name: "demo"}, 0, 0), modelData: filters.FilterMap{}, wantType: &filters.Model{}},
		{name: "filters -> bool detail", start: filters.NewModel(nil, 20, 10), modelData: filters.Property{Name: "Is Archived", Type: "bool"}, wantType: &filters.BoolModel{}},
		{name: "filters -> int detail", start: filters.NewModel(nil, 20, 10), modelData: filters.Property{Name: "Stargazer Count", Type: "int"}, wantType: &filters.IntModel{}},
		{name: "filters -> date detail", start: filters.NewModel(nil, 20, 10), modelData: filters.Property{Name: "Created At", Type: "time.Time"}, wantType: &filters.DateModel{}},
		{name: "filters -> string detail", start: filters.NewModel(nil, 20, 10), modelData: filters.Property{Name: "Name", Type: "string"}, wantType: &filters.StringModel{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MainModel{nav: shared.NewNavigator(), width: 80, height: 24}
			m.nav.SetValidator(validateTransition)
			assert.NoError(t, m.nav.Push(tt.start))

			cmd := m.Next(shared.NextMsg{ModelData: tt.modelData})

			assert.NotNil(t, cmd)
			assert.Equal(t, 2, m.nav.Len())
			top, _ := m.nav.Current()
			assert.IsType(t, tt.wantType, top)
		})
	}
}

func TestMainModel_Next_FromDetailDoesNothing(t *testing.T) {
	m := MainModel{nav: shared.NewNavigator()}
	assert.NoError(t, m.nav.Push(filters.NewBoolModel("Is Archived", false, 20, 10)))

	cmd := m.Next(shared.NextMsg{ModelData: filters.Property{Name: "Name", Type: "string"}})

	assert.Nil(t, cmd)
	assert.Equal(t, 1, m.nav.Len())
}

func TestMainModel_Next_BlockedByValidator(t *testing.T) {
	m := MainModel{nav: shared.NewNavigator()}
	m.nav.SetValidator(func(current, next tea.Model) error {
		return errors.New("blocked")
	})
	assert.NoError(t, m.nav.Push(user.NewModel(0, 0)))

	cmd := m.Next(shared.NextMsg{ModelData: shared.OrgKey{Name: "demo"}})

	assert.Nil(t, cmd)
	assert.Equal(t, 1, m.nav.Len())
}

func TestMainModel_Previous_ForwardsMessageToNewTop(t *testing.T) {
	m := MainModel{nav: shared.NewNavigator()}
	boolModel := filters.NewBoolModel("Is Archived", false, 20, 10)
	assert.NoError(t, m.nav.Push(boolModel))
	assert.NoError(t, m.nav.Push(filters.NewStringModel("Name", "", 20, 10)))

	m.Previous(shared.PreviousMsg{Message: keyPress("y")})

	assert.Equal(t, 1, m.nav.Len())
	assert.True(t, boolModel.Value(), "the forwarded message should update the new top model")
}

func TestMainModel_Previous_WithoutMessage(t *testing.T) {
	m := MainModel{nav: shared.NewNavigator()}
	assert.NoError(t, m.nav.Push(user.NewModel(0, 0)))
	assert.NoError(t, m.nav.Push(filters.NewBoolModel("Is Archived", false, 20, 10)))

	cmd := m.Previous(shared.PreviousMsg{})

	assert.Nil(t, cmd)
	assert.Equal(t, 1, m.nav.Len())
}

func TestMainModel_View(t *testing.T) {
	m := NewMainModel()
	m.SetDimensions(80, 24)

	view := m.View()

	assert.True(t, view.AltScreen)
	content := plain(view)
	assert.NotEmpty(t, content)
	assert.Contains(t, content, "Organizations", "the user model header should be rendered")
	assert.Contains(t, content, "select", "the user model help should be rendered in the footer")
	assert.LessOrEqual(t, lipgloss.Height(content), 24, "the view must fit the terminal height")
}

func TestMainModel_View_ZeroSize(t *testing.T) {
	m := NewMainModel()
	assert.NotPanics(t, func() { m.View() })
}

func TestViewContent(t *testing.T) {
	assert.Equal(t, "hello", viewContent(tea.NewView("hello")))
	assert.Equal(t, "", viewContent(tea.NewView("")))
}

func TestRenderHeader(t *testing.T) {
	m := MainModel{}

	tests := []struct {
		name  string
		model tea.Model
		want  string
	}{
		{name: "header provider", model: user.NewModel(0, 0), want: "Organizations"},
		{name: "falls back to type name", model: &plainModel{}, want: "plainModel"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, ansi.Strip(m.renderHeader(tt.model)), tt.want)
		})
	}
}

func TestRenderFooter(t *testing.T) {
	m := MainModel{}

	tests := []struct {
		name  string
		model tea.Model
		want  string
	}{
		{name: "help provider", model: user.NewModel(0, 0), want: "select"},
		{name: "default help", model: &plainModel{}, want: "esc: back | ctrl+c: quit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, ansi.Strip(m.renderFooter(tt.model)), tt.want)
		})
	}
}

func TestIsFilterDetail(t *testing.T) {
	tests := []struct {
		name  string
		model tea.Model
		want  bool
	}{
		{name: "bool model", model: &filters.BoolModel{}, want: true},
		{name: "int model", model: &filters.IntModel{}, want: true},
		{name: "date model", model: &filters.DateModel{}, want: true},
		{name: "string model", model: &filters.StringModel{}, want: true},
		{name: "filters model", model: &filters.Model{}, want: false},
		{name: "user model", model: &user.Model{}, want: false},
		{name: "nil", model: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isFilterDetail(tt.model))
		})
	}
}

func TestComputeLayout(t *testing.T) {
	m := MainModel{width: 80, height: 24}

	layout := m.computeLayout(user.NewModel(0, 0))

	assert.Equal(t, 80, layout.bodyWidth)
	assert.Equal(t, 78, layout.interiorWidth)
	assert.Greater(t, layout.headerHeight, 0)
	assert.GreaterOrEqual(t, layout.footerHeight, 1)
	assert.Equal(t, layout.bodyHeight-2, layout.interiorHeight)
	assert.LessOrEqual(t, layout.headerHeight+layout.bodyHeight+layout.footerHeight, 24)
	assert.NotEmpty(t, layout.header)
	assert.NotEmpty(t, layout.footer)
}

func TestComputeLayout_NoHeaderProvider(t *testing.T) {
	m := MainModel{width: 80, height: 24}

	layout := m.computeLayout(&plainModel{})

	assert.Contains(t, ansi.Strip(layout.header), "plainModel")
	assert.Contains(t, ansi.Strip(layout.footer), "esc: back")
}

func TestComputeLayout_ZeroSize(t *testing.T) {
	m := MainModel{}

	layout := m.computeLayout(&plainModel{})

	assert.Equal(t, 4, layout.bodyWidth)
	assert.Equal(t, 2, layout.interiorWidth)
	assert.Equal(t, 1, layout.bodyHeight)
	assert.Equal(t, 1, layout.interiorHeight)
	assert.GreaterOrEqual(t, layout.footerHeight, 1)
}

func TestContentDimensions(t *testing.T) {
	tests := []struct {
		name       string
		width      int
		height     int
		wantWidth  int
		wantHeight int
	}{
		{name: "normal", width: 80, height: 24, wantWidth: 78, wantHeight: 20},
		{name: "zero", width: 0, height: 0, wantWidth: 1, wantHeight: 1},
		{name: "tiny", width: 2, height: 4, wantWidth: 1, wantHeight: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MainModel{width: tt.width, height: tt.height}
			w, h := m.contentDimensions()
			assert.Equal(t, tt.wantWidth, w)
			assert.Equal(t, tt.wantHeight, h)
		})
	}
}
