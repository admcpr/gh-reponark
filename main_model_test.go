package main

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"gh-reponark/filters"
	"gh-reponark/github"
	"gh-reponark/github/githubtest"
	"gh-reponark/org"
	"gh-reponark/shared"
	"gh-reponark/user"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

// plainModel is a model that provides neither a title nor help so the
// layout falls back to its defaults.
type plainModel struct{}

func (m *plainModel) Init() tea.Cmd                           { return nil }
func (m *plainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m *plainModel) View() tea.View                          { return tea.NewView("plain content") }

// recordingModel remembers the messages it receives.
type recordingModel struct {
	plainModel
	received []tea.Msg
}

func (m *recordingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.received = append(m.received, msg)
	return m, nil
}

// titledModel names itself in the breadcrumb and has a status.
type titledModel struct {
	plainModel
	title, status string
}

func (m *titledModel) Breadcrumb() string { return m.title }
func (m *titledModel) Status() string     { return m.status }

// plain renders a view to a string with all ANSI styling removed so tests can
// assert on the visible text.
// loadMsg runs cmd and returns the message its loading command produced,
// skipping the spinner tick that Init batches alongside it.
func loadMsg(cmd tea.Cmd) tea.Msg {
	switch msg := cmd().(type) {
	case tea.BatchMsg:
		for _, c := range msg {
			if m := loadMsg(c); m != nil {
				return m
			}
		}
		return nil
	case spinner.TickMsg:
		return nil
	default:
		return msg
	}
}

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
	child := &recordingModel{}
	m.nav.Push(child)

	m, _ = update(m, keyPress("y"))

	assert.Equal(t, []tea.Msg{keyPress("y")}, child.received, "the key press should reach the child model")
	assert.Equal(t, 1, m.nav.Len())
}

func TestMainModel_Update_OtherMessagesGoToChild(t *testing.T) {
	m := MainModel{nav: shared.NewNavigator()}
	child := &recordingModel{}
	m.nav.Push(child)

	m, cmd := update(m, shared.OrgKey{Name: "other"})

	assert.Nil(t, cmd)
	assert.Equal(t, 1, m.nav.Len())
	assert.Len(t, child.received, 1)
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
		assert.Equal(t, "demo", orgModel.Breadcrumb())
	}
}

func TestMainModel_Update_OpenFiltersMsg(t *testing.T) {
	m := newTestMainModel()
	m.SetDimensions(80, 24)
	applied := filters.FilterMap{"Is Archived": filters.NewBoolFilter("Is Archived", true)}

	m, cmd := update(m, filters.OpenFiltersMsg{Filters: applied})

	assert.Nil(t, cmd, "the filter screen has nothing to load")
	assert.Equal(t, 2, m.nav.Len())
	filtersModel, ok := current(t, m).(*filters.Model)
	if assert.True(t, ok, "top of stack should be the filters model") {
		assert.Contains(t, plain(filtersModel.View()), "Is Archived", "the applied filters should be shown")
	}
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
	assert.Contains(t, plain(m.View()), "reponark › demo › Filters", "the breadcrumb follows the screens")

	// Search for a property and toggle it from the list.
	for _, k := range []tea.KeyPressMsg{keyPress("/"), keyPress("f"), keyPress("o"), keyPress("r"), keyPress("k"), {Code: tea.KeyEnter}, {Code: tea.KeySpace, Text: " "}} {
		m, _ = update(m, k)
	}
	assert.Equal(t, 3, m.nav.Len(), "filters are edited in place, without opening another screen")
	filtersModel := current(t, m).(*filters.Model)
	assert.Contains(t, filtersModel.Filters(), "Is Fork")

	// Leaving the filter screen delivers the filters to the org screen.
	_, cmd := filtersModel.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
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

	m, _ = update(m, loadMsg(m.Init()))

	assert.IsType(t, &shared.ErrorModel{}, current(t, m), "an error while loading the user should show the error screen")
}

func TestMainModel_Previous_EmptyNavQuits(t *testing.T) {
	m := MainModel{nav: shared.NewNavigator()}

	cmd := m.Previous(shared.PreviousMsg{})
	assert.Equal(t, reflect.ValueOf(tea.Quit).Pointer(), reflect.ValueOf(cmd).Pointer())
}

func TestMainModel_Previous_ForwardsMessageToNewTop(t *testing.T) {
	m := MainModel{nav: shared.NewNavigator()}
	below := &recordingModel{}
	m.nav.Push(below)
	m.nav.Push(&plainModel{})

	m.Previous(shared.PreviousMsg{Message: keyPress("y")})

	assert.Equal(t, 1, m.nav.Len())
	assert.Equal(t, []tea.Msg{keyPress("y")}, below.received, "the forwarded message should update the new top model")
}

func TestMainModel_Previous_WithoutMessage(t *testing.T) {
	m := MainModel{nav: shared.NewNavigator()}
	m.nav.Push(user.NewModel(&githubtest.Fake{}, 0, 0))
	m.nav.Push(&plainModel{})

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
	assert.Contains(t, content, "┌─ reponark ", "the breadcrumb is drawn into the top edge")
	assert.Contains(t, content, "signing in", "the current screen's status is on the top edge")
	assert.Contains(t, content, "enter open", "the user model help should be rendered in the footer")
	assert.Equal(t, 24, lipgloss.Height(content), "the view fills the terminal height")
}

func TestMainModel_View_ZeroSize(t *testing.T) {
	m := newTestMainModel()
	assert.NotPanics(t, func() { m.View() })
}

func TestViewContent(t *testing.T) {
	assert.Equal(t, "hello", viewContent(tea.NewView("hello")))
	assert.Equal(t, "", viewContent(tea.NewView("")))
}

func TestRenderTopEdge(t *testing.T) {
	m := MainModel{nav: shared.NewNavigator()}
	m.nav.Push(user.NewModel(&githubtest.Fake{}, 0, 0))
	m.nav.Push(&titledModel{title: "acme"})
	top := &titledModel{title: "Filters", status: "3 of 9 repos match"}
	m.nav.Push(top)

	edge := ansi.Strip(m.renderTopEdge(top, 60, 20))

	assert.Equal(t, 60, lipgloss.Width(edge))
	assert.True(t, strings.HasPrefix(edge, "┌─ reponark › acme › Filters ──"), edge)
	assert.True(t, strings.HasSuffix(edge, "── 3 of 9 repos match ─┐"), edge)
}

func TestRenderTopEdge_ContinuesTheBodyGradient(t *testing.T) {
	m := MainModel{nav: shared.NewNavigator()}
	m.nav.Push(&plainModel{})

	// A frame drawn in one piece by Lip Gloss, and the same frame as the
	// app draws it: a hand-made top edge over a body without one.
	whole := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).
		BorderForegroundBlend(shared.FrameGradient(true)...).
		Width(30).Height(7).Render(strings.Repeat("\n", 4))
	body := lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, true, true, true).
		BorderForegroundBlend(shared.FrameGradient(true)...).
		Width(30).Height(6).Render(strings.Repeat("\n", 4))
	edge := m.renderTopEdge(&plainModel{}, 30, 5)

	wholeLines := strings.Split(whole, "\n")
	assert.Equal(t, wholeLines[1:], strings.Split(body, "\n"), "the body is the frame without its top edge")
	// The corners are the first and last cells of the top edge; they carry
	// the colours either side of the join with the right and left edges.
	foregrounds := regexp.MustCompile(`\x1b\[38;2;\d+;\d+;\d+m`)
	want, got := foregrounds.FindAllString(wholeLines[0], -1), foregrounds.FindAllString(edge, -1)
	assert.Equal(t, 30, len(want))
	assert.Equal(t, want[0], got[0], "the top edge starts on the colour Lip Gloss would use")
	assert.Equal(t, want[len(want)-10:], got[len(got)-10:], "and its rule runs on to the corner in the same colours")
}

func TestRenderTopEdge_Narrow(t *testing.T) {
	m := MainModel{nav: shared.NewNavigator()}
	m.nav.Push(&titledModel{title: "a-long-organization-name"})
	top := &titledModel{title: "Filters", status: "3 of 9 repos match"}
	m.nav.Push(top)

	edge := ansi.Strip(m.renderTopEdge(top, 36, 20))

	assert.Equal(t, 36, lipgloss.Width(edge))
	assert.NotContains(t, edge, "repos match", "the status goes first")
	assert.Contains(t, edge, "… › Filters", "then the oldest crumbs")
}

func TestRenderTopEdge_UntitledScreen(t *testing.T) {
	m := MainModel{nav: shared.NewNavigator()}
	m.nav.Push(&plainModel{})

	edge := ansi.Strip(m.renderTopEdge(&plainModel{}, 30, 20))

	assert.Equal(t, "┌─ reponark ─────────────────┐", edge)
}

func TestRenderFooter(t *testing.T) {
	m := MainModel{}

	tests := []struct {
		name  string
		model tea.Model
		want  string
	}{
		{name: "help provider", model: user.NewModel(&githubtest.Fake{}, 0, 0), want: "enter open  ? help  esc quit"},
		{name: "org model", model: org.NewModel(&githubtest.Fake{}, shared.OrgKey{Name: "demo"}, 80, 24), want: "filters"},
		{name: "default help", model: &plainModel{}, want: "esc back  ? help  ctrl+c quit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			footer := m.renderFooter(tt.model)
			assert.NotEmpty(t, strings.TrimSpace(footer))
			assert.Contains(t, ansi.Strip(footer), tt.want)
		})
	}
}

func TestMainModel_QuestionMarkTogglesFullHelp(t *testing.T) {
	m := newTestMainModel()
	m.SetDimensions(80, 24)
	short := plain(m.View())
	assert.Contains(t, short, "? help")

	m, cmd := update(m, keyPress("?"))
	assert.Nil(t, cmd)
	assert.True(t, m.fullHelp)
	full := plain(m.View())
	assert.Contains(t, full, "hide help")
	assert.Contains(t, full, "ctrl+c quit", "the full view lists keys that work everywhere")
	assert.Contains(t, full, "first/last", "and the screen's extra keys")
	assert.Equal(t, 24, lipgloss.Height(full), "the body shrinks to make room")

	m, _ = update(m, keyPress("?"))
	assert.False(t, m.fullHelp)
}

func TestMainModel_QuestionMarkIsTypedInTextFields(t *testing.T) {
	m := MainModel{nav: shared.NewNavigator(), width: 80, height: 24}
	screen := filters.NewModel(nil, nil, 78, 20)
	m.nav.Push(screen)

	m, _ = update(m, keyPress("/"))
	m, _ = update(m, keyPress("?"))

	assert.False(t, m.fullHelp, "? goes into the search")
	assert.NotContains(t, ansi.Strip(m.renderFooter(screen)), "? help", "and is not offered while typing")
}

func TestMainModel_OpeningAScreenHidesFullHelp(t *testing.T) {
	m := newTestMainModel()
	m, _ = update(m, keyPress("?"))

	m, _ = update(m, shared.OpenOrgMsg{Key: shared.OrgKey{Name: "demo"}})
	assert.False(t, m.fullHelp)

	m, _ = update(m, keyPress("?"))
	m, _ = update(m, shared.PreviousMsg{})
	assert.False(t, m.fullHelp)
}

func TestRenderFooter_FitsEightyColumns(t *testing.T) {
	m := MainModel{width: 80, nav: shared.NewNavigator()}
	loaded := org.NewModel(&githubtest.Fake{}, shared.OrgKey{Name: "demo"}, 78, 20)
	m.nav.Push(loaded)

	for _, full := range []bool{false, true} {
		m.fullHelp = full
		footer := ansi.Strip(m.renderFooter(loaded))
		for _, line := range strings.Split(footer, "\n") {
			assert.LessOrEqual(t, lipgloss.Width(line), 80, "%q", line)
		}
		assert.NotContains(t, footer, "…", "nothing is cut off")
	}
}

func TestComputeLayout(t *testing.T) {
	m := MainModel{width: 80, height: 24}

	layout := m.computeLayout(user.NewModel(&githubtest.Fake{}, 0, 0))

	assert.Equal(t, 80, layout.bodyWidth)
	assert.Equal(t, 78, layout.interiorWidth)
	assert.GreaterOrEqual(t, layout.footerHeight, 1)
	assert.Equal(t, layout.bodyHeight-2, layout.interiorHeight)
	assert.Equal(t, 24, layout.bodyHeight+layout.footerHeight, "the top edge is part of the body")
	assert.NotEmpty(t, layout.footer)
}

func TestComputeLayout_NoHelpProvider(t *testing.T) {
	m := MainModel{width: 80, height: 24, nav: shared.NewNavigator()}

	layout := m.computeLayout(&plainModel{})

	assert.Contains(t, ansi.Strip(m.renderTopEdge(&plainModel{}, layout.bodyWidth, layout.interiorHeight)), "reponark")
	assert.Contains(t, ansi.Strip(layout.footer), "esc back")
}

func TestComputeLayout_ZeroSize(t *testing.T) {
	m := MainModel{}

	layout := m.computeLayout(&plainModel{})

	assert.Equal(t, 4, layout.bodyWidth)
	assert.Equal(t, 2, layout.interiorWidth)
	assert.Equal(t, 3, layout.bodyHeight)
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
		{name: "normal", width: 80, height: 24, wantWidth: 78, wantHeight: 21},
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
