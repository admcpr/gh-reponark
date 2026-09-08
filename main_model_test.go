package main

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"gh-reponark/filters"
	"gh-reponark/github"
	"gh-reponark/github/githubtest"
	"gh-reponark/org"
	"gh-reponark/repo"
	"gh-reponark/shared"
	"gh-reponark/user"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

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

func newTestMainModel() MainModel {
	return NewMainModel(&githubtest.Fake{User: github.User{Login: "octocat"}})
}

// update runs one message through the model and returns the new model, since
// MainModel.Update has a value receiver.
func update(m MainModel, msg tea.Msg) (MainModel, tea.Cmd) {
	updated, cmd := m.Update(msg)
	return updated.(MainModel), cmd
}

func current(t *testing.T, m MainModel) tea.Model {
	t.Helper()
	top, err := m.nav.Current()
	assert.NoError(t, err)
	return top
}

func TestNewMainModel(t *testing.T) {
	m := newTestMainModel()

	assert.Equal(t, 1, m.nav.Len())
	assert.IsType(t, &user.Model{}, current(t, m), "the user model should be the first screen")
}

func TestMainModel_Init(t *testing.T) {
	m := newTestMainModel()

	cmd := m.Init()

	if assert.NotNil(t, cmd) {
		_, isError := cmd().(shared.ErrorMsg)
		assert.False(t, isError, "loading the user from the fake should succeed")
	}
}

func TestMainModel_SetDimensions(t *testing.T) {
	m := newTestMainModel()

	m.SetDimensions(120, 40)

	assert.Equal(t, 120, m.width)
	assert.Equal(t, 40, m.height)
}

func TestMainModel_Update_WindowSize(t *testing.T) {
	m, cmd := update(newTestMainModel(), tea.WindowSizeMsg{Width: 100, Height: 50})

	assert.Nil(t, cmd)
	assert.Equal(t, 100, m.width)
	assert.Equal(t, 50, m.height)
}

func TestMainModel_Update_CtrlCQuits(t *testing.T) {
	_, cmd := update(newTestMainModel(), tea.KeyPressMsg{Code: []rune("c")[0], Mod: tea.ModCtrl})

	if assert.NotNil(t, cmd) {
		_, ok := cmd().(tea.QuitMsg)
		assert.True(t, ok, "ctrl+c should quit")
	}
}

func TestMainModel_Update_KeyGoesToChild(t *testing.T) {
	m := MainModel{nav: shared.NewNavigator()}
	child := filters.NewBoolModel("Is Archived", false, 20, 10)
	m.nav.Push(child)

	m, _ = update(m, keyPress("y"))

	assert.True(t, child.Value(), "the key press should reach the child model")
	assert.Equal(t, 1, m.nav.Len())
}

func TestMainModel_Update_OtherMessagesGoToChild(t *testing.T) {
	m := MainModel{nav: shared.NewNavigator()}
	m.nav.Push(filters.NewIntModel("Stars", 0, 10, 20, 10))

	m, cmd := update(m, tea.WindowSizeMsg{Width: 1, Height: 1})

	assert.Nil(t, cmd)
	assert.Equal(t, 1, m.nav.Len())
}

func TestMainModel_Update_OpenOrgMsg(t *testing.T) {
	m := newTestMainModel()
	m.SetDimensions(80, 24)

	m, cmd := update(m, shared.OpenOrgMsg{Key: shared.OrgKey{Name: "demo", IsUser: true}})

	assert.NotNil(t, cmd, "the org screen's Init command should be returned")
	assert.Equal(t, 2, m.nav.Len())
	orgModel, ok := current(t, m).(*org.Model)
	if assert.True(t, ok, "top of stack should be the org model") {
		assert.Equal(t, "demo", orgModel.Title)
		assert.Contains(t, plain(orgModel.HeaderView()), "User: demo")
	}
}

func TestMainModel_Update_OpenFiltersMsg(t *testing.T) {
	m := newTestMainModel()
	m.SetDimensions(80, 24)
	applied := filters.FilterMap{"Is Archived": filters.NewBoolFilter("Is Archived", true)}

	m, cmd := update(m, filters.OpenFiltersMsg{Filters: applied})

	assert.NotNil(t, cmd, "the filter screen's Init command should be returned")
	assert.Equal(t, 2, m.nav.Len())
	filtersModel, ok := current(t, m).(*filters.Model)
	if assert.True(t, ok, "top of stack should be the filters model") {
		assert.Contains(t, plain(filtersModel.View()), "Is Archived", "the applied filters should be shown")
	}
}

func TestMainModel_Update_EditFilterMsg(t *testing.T) {
	tests := []struct {
		name     string
		property repo.PropertySchema
		wantType tea.Model
	}{
		{name: "bool", property: repo.PropertySchema{Name: "Is Archived", Type: "bool"}, wantType: &filters.BoolModel{}},
		{name: "int", property: repo.PropertySchema{Name: "Stargazer Count", Type: "int"}, wantType: &filters.IntModel{}},
		{name: "date", property: repo.PropertySchema{Name: "Created At", Type: "time.Time"}, wantType: &filters.DateModel{}},
		{name: "string", property: repo.PropertySchema{Name: "Name", Type: "string"}, wantType: &filters.StringModel{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MainModel{nav: shared.NewNavigator(), width: 80, height: 24}
			m.nav.Push(filters.NewModel(nil, 20, 10))

			m, cmd := update(m, filters.EditFilterMsg{Property: tt.property})

			assert.NotNil(t, cmd)
			assert.Equal(t, 2, m.nav.Len())
			assert.IsType(t, tt.wantType, current(t, m))
		})
	}
}

func TestMainModel_Update_EditFilterMsg_UnsupportedType(t *testing.T) {
	m := MainModel{nav: shared.NewNavigator(), width: 80, height: 24}
	m.nav.Push(filters.NewModel(nil, 20, 10))

	m, cmd := update(m, filters.EditFilterMsg{Property: repo.PropertySchema{Name: "Languages", Type: "[]string"}})

	assert.Nil(t, cmd)
	assert.Equal(t, 1, m.nav.Len(), "no editor exists so nothing should be pushed")
}

func TestMainModel_Update_PreviousMsg(t *testing.T) {
	m, _ := update(newTestMainModel(), shared.OpenOrgMsg{Key: shared.OrgKey{Name: "demo"}})
	assert.Equal(t, 2, m.nav.Len())

	m, cmd := update(m, shared.PreviousMsg{})

	assert.Nil(t, cmd)
	assert.Equal(t, 1, m.nav.Len())
	assert.IsType(t, &user.Model{}, current(t, m))
}

func TestMainModel_Update_FullNavigationFlow(t *testing.T) {
	m := newTestMainModel()
	m.SetDimensions(80, 24)

	m, _ = update(m, shared.OpenOrgMsg{Key: shared.OrgKey{Name: "demo"}})
	assert.IsType(t, &org.Model{}, current(t, m))

	m, _ = update(m, filters.OpenFiltersMsg{})
	assert.IsType(t, &filters.Model{}, current(t, m))

	m, _ = update(m, filters.EditFilterMsg{Property: repo.PropertySchema{Name: "Is Archived", Type: "bool"}})
	assert.IsType(t, &filters.BoolModel{}, current(t, m))
	assert.Equal(t, 4, m.nav.Len())

	// Confirming the editor sends the new filter back to the filter screen.
	_, cmd := current(t, m).Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m, _ = update(m, cmd())
	assert.Equal(t, 3, m.nav.Len())
	filtersModel := current(t, m).(*filters.Model)
	assert.Contains(t, plain(filtersModel.View()), "Is Archived")

	// Leaving the filter screen delivers the filters to the org screen.
	_, cmd = filtersModel.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m, _ = update(m, cmd())
	assert.Equal(t, 2, m.nav.Len())
	assert.IsType(t, &org.Model{}, current(t, m))

	m, _ = update(m, shared.PreviousMsg{})
	assert.Equal(t, 1, m.nav.Len())
	assert.IsType(t, &user.Model{}, current(t, m))

	_, cmd = update(m, shared.PreviousMsg{})
	_, isQuit := cmd().(tea.QuitMsg)
	assert.True(t, isQuit, "going back from the first screen quits")
}

func TestMainModel_Update_ErrorMsgShowsErrorScreen(t *testing.T) {
	m := newTestMainModel()
	m.SetDimensions(80, 24)
	boom := errors.New("fetching user octocat: boom")

	m, cmd := update(m, shared.ErrorMsg{Err: boom})

	assert.Nil(t, cmd)
	assert.Equal(t, 2, m.nav.Len())
	errorModel, ok := current(t, m).(*shared.ErrorModel)
	if assert.True(t, ok, "top of stack should be the error screen") {
		assert.Equal(t, boom, errorModel.Err())
	}
	assert.Contains(t, plain(m.View()), "boom")
}

func TestMainModel_Update_ErrorScreenGoesBack(t *testing.T) {
	m, _ := update(newTestMainModel(), shared.ErrorMsg{Err: errors.New("boom")})
	assert.Equal(t, 2, m.nav.Len())

	// The error screen turns esc into a PreviousMsg, which the main model then pops.
	m, cmd := update(m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if assert.NotNil(t, cmd) {
		m, _ = update(m, cmd())
	}

	assert.Equal(t, 1, m.nav.Len())
	assert.IsType(t, &user.Model{}, current(t, m), "should return to the screen that reported the error")
}

func TestMainModel_Update_ErrorFromInit(t *testing.T) {
	m := NewMainModel(&githubtest.Fake{UserErr: errors.New("not logged in")})

	m, _ = update(m, m.Init()())

	assert.IsType(t, &shared.ErrorModel{}, current(t, m), "an error while loading the user should show the error screen")
}

func TestMainModel_Previous_EmptyNavQuits(t *testing.T) {
	m := MainModel{nav: shared.NewNavigator()}

	cmd := m.Previous(shared.PreviousMsg{})
	assert.Equal(t, reflect.ValueOf(tea.Quit).Pointer(), reflect.ValueOf(cmd).Pointer())
}

func TestMainModel_Previous_ForwardsMessageToNewTop(t *testing.T) {
	m := MainModel{nav: shared.NewNavigator()}
	boolModel := filters.NewBoolModel("Is Archived", false, 20, 10)
	m.nav.Push(boolModel)
	m.nav.Push(filters.NewStringModel("Name", "", 20, 10))

	m.Previous(shared.PreviousMsg{Message: keyPress("y")})

	assert.Equal(t, 1, m.nav.Len())
	assert.True(t, boolModel.Value(), "the forwarded message should update the new top model")
}

func TestMainModel_Previous_WithoutMessage(t *testing.T) {
	m := MainModel{nav: shared.NewNavigator()}
	m.nav.Push(user.NewModel(&githubtest.Fake{}, 0, 0))
	m.nav.Push(filters.NewBoolModel("Is Archived", false, 20, 10))

	cmd := m.Previous(shared.PreviousMsg{})

	assert.Nil(t, cmd)
	assert.Equal(t, 1, m.nav.Len())
}

func TestMainModel_View(t *testing.T) {
	m := newTestMainModel()
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
	m := newTestMainModel()
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
		{name: "header provider", model: user.NewModel(&githubtest.Fake{}, 0, 0), want: "Organizations"},
		{name: "error screen", model: shared.NewErrorModel(errors.New("boom"), 0, 0), want: "Error"},
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
		{name: "help provider", model: user.NewModel(&githubtest.Fake{}, 0, 0), want: "select"},
		{name: "org model", model: org.NewModel(&githubtest.Fake{}, shared.OrgKey{Name: "demo"}, 80, 24), want: "filters"},
		{name: "default help", model: &plainModel{}, want: "esc: back | ctrl+c: quit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			footer := m.renderFooter(tt.model)
			assert.NotEmpty(t, strings.TrimSpace(footer))
			assert.Contains(t, ansi.Strip(footer), tt.want)
		})
	}
}

func TestComputeLayout(t *testing.T) {
	m := MainModel{width: 80, height: 24}

	layout := m.computeLayout(user.NewModel(&githubtest.Fake{}, 0, 0))

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
