package user

import (
	"errors"
	"fmt"
	"testing"

	"gh-reponark/github"
	"gh-reponark/github/githubtest"
	"gh-reponark/shared"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

// plain renders a view to a string with all ANSI styling removed so tests can
// assert on the visible text.
func plain(v tea.View) string {
	return ansi.Strip(fmt.Sprint(v.Content))
}

func newTestUser(login string, orgs ...string) github.User {
	user := github.User{Login: login, Url: "https://github.com/" + login}
	for _, org := range orgs {
		user.Organizations = append(user.Organizations, github.Organization{
			Login: org,
			Url:   "https://github.com/" + org,
		})
	}
	return user
}

func newTestModel() *Model {
	return NewModel(&githubtest.Fake{}, 80, 24)
}

func itemTitles(m *Model) []string {
	titles := make([]string, len(m.items))
	for i, item := range m.items {
		titles[i] = item.Title()
	}
	return titles
}

func TestNewModel(t *testing.T) {
	m := newTestModel()

	assert.Equal(t, 80, m.width)
	assert.Equal(t, 24, m.height)
	assert.Equal(t, "", m.login)
	assert.Empty(t, m.items)
}

func TestModel_SetDimensions(t *testing.T) {
	m := newTestModel()

	m.SetDimensions(120, 40)

	assert.Equal(t, 120, m.width)
	assert.Equal(t, 40, m.height)
}

func TestModel_Init_LoadsUser(t *testing.T) {
	fake := &githubtest.Fake{User: newTestUser("octocat", "acme")}
	m := NewModel(fake, 80, 24)

	cmd := m.Init()

	if assert.NotNil(t, cmd) {
		msg, ok := cmd().(userLoadedMsg)
		assert.True(t, ok, "expected a userLoadedMsg")
		assert.Equal(t, fake.User, github.User(msg))
	}
}

func TestModel_Init_ReportsErrors(t *testing.T) {
	fake := &githubtest.Fake{UserErr: errors.New("not logged in")}
	m := NewModel(fake, 80, 24)

	msg, ok := m.Init()().(shared.ErrorMsg)

	assert.True(t, ok, "expected an ErrorMsg")
	assert.EqualError(t, msg.Err, "not logged in")
}

func TestModel_SetUser(t *testing.T) {
	m := newTestModel()

	m.SetUser(newTestUser("octocat", "zulu", "alpha", "mike"))

	assert.Equal(t, "octocat", m.login)
	assert.Equal(t, []string{"octocat", "alpha", "mike", "zulu"}, itemTitles(m),
		"the user should be first followed by organisations sorted by login")

	assert.Equal(t, "https://github.com/octocat", m.items[0].Description())
}

func TestModel_SetUser_NoOrganizations(t *testing.T) {
	m := newTestModel()

	m.SetUser(newTestUser("octocat"))

	assert.Equal(t, []string{"octocat"}, itemTitles(m))
}

func TestModel_Update_UserLoadedMsg(t *testing.T) {
	m := newTestModel()

	updated, cmd := m.Update(userLoadedMsg(newTestUser("octocat", "acme")))

	assert.Same(t, m, updated)
	assert.Nil(t, cmd)
	assert.Equal(t, "octocat", m.login)
	assert.Equal(t, []string{"octocat", "acme"}, itemTitles(m))
}

func TestModel_Update_EnterWithoutItems(t *testing.T) {
	m := newTestModel()

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	assert.Nil(t, cmd)
}

func TestModel_Update_EnterSelectsOrgKey(t *testing.T) {
	tests := []struct {
		name      string
		downCount int
		want      shared.OrgKey
	}{
		{name: "user entry", downCount: 0, want: shared.OrgKey{Name: "octocat", IsUser: true}},
		{name: "organisation entry", downCount: 1, want: shared.OrgKey{Name: "acme", IsUser: false}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModel()
			m.SetUser(newTestUser("octocat", "acme"))

			for i := 0; i < tt.downCount; i++ {
				m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
			}

			_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			if assert.NotNil(t, cmd) {
				open, ok := cmd().(shared.OpenOrgMsg)
				assert.True(t, ok, "expected an OpenOrgMsg")
				assert.Equal(t, tt.want, open.Key)
			}
		})
	}
}

func TestModel_Update_ArrowsMoveTheCursor(t *testing.T) {
	m := newTestModel()
	m.SetUser(newTestUser("octocat", "acme"))
	assert.Equal(t, 0, m.cursor)

	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	assert.Equal(t, 1, m.cursor)

	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	assert.Equal(t, 1, m.cursor, "the cursor stops at the end")

	m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	assert.Equal(t, 0, m.cursor)

	m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	assert.Equal(t, 0, m.cursor, "and at the start")
}

func TestModel_Update_PagingKeys(t *testing.T) {
	m := NewModel(&githubtest.Fake{}, 80, 4) // three rows under the heading
	m.SetUser(newTestUser("octocat", "a", "b", "c", "d", "e", "f"))

	m.Update(tea.KeyPressMsg{Code: tea.KeyPgDown})
	assert.Equal(t, 3, m.cursor, "page down moves a screenful")
	assert.Equal(t, 1, m.offset, "and scrolls the list to keep the cursor in view")

	m.Update(tea.KeyPressMsg{Code: 'G', Text: "G"})
	assert.Equal(t, 6, m.cursor)
	assert.Equal(t, 4, m.offset)
	content := plain(m.View())
	assert.Contains(t, content, "f")
	assert.NotContains(t, content, "octocat", "rows scrolled off the top are hidden")

	m.Update(tea.KeyPressMsg{Code: tea.KeyPgUp})
	assert.Equal(t, 3, m.cursor)

	m.Update(tea.KeyPressMsg{Code: 'g', Text: "g"})
	assert.Equal(t, 0, m.cursor)
	assert.Equal(t, 0, m.offset)
}

func TestModel_SetDimensions_KeepsCursorInView(t *testing.T) {
	m := NewModel(&githubtest.Fake{}, 80, 10)
	m.SetUser(newTestUser("octocat", "a", "b", "c", "d", "e", "f"))
	m.Update(tea.KeyPressMsg{Code: 'G', Text: "G"})
	assert.Equal(t, 0, m.offset, "everything fits at ten rows")

	m.SetDimensions(80, 4)

	assert.Equal(t, 4, m.offset, "a shorter pane scrolls to the cursor")
}

func TestModel_Update_NonKeyMessagesAreIgnored(t *testing.T) {
	m := newTestModel()
	m.SetUser(newTestUser("octocat", "acme"))

	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 10, Height: 10})

	assert.Same(t, m, updated)
	assert.Nil(t, cmd)
	assert.Equal(t, []string{"octocat", "acme"}, itemTitles(m))
}

func TestModel_View(t *testing.T) {
	m := newTestModel()
	m.SetUser(newTestUser("octocat", "acme"))

	content := plain(m.View())

	assert.Contains(t, content, "octocat")
	assert.Contains(t, content, "acme")
}

func TestModel_View_ZeroHeight(t *testing.T) {
	m := NewModel(&githubtest.Fake{}, 80, 0)
	assert.NotPanics(t, func() { m.View() })
}

func TestModel_Status(t *testing.T) {
	m := newTestModel()
	assert.Equal(t, "signing in", m.Status())

	m.SetUser(newTestUser("octocat"))
	assert.Equal(t, "signed in as octocat", m.Status())
}

func TestModel_Help(t *testing.T) {
	m := newTestModel()

	help := m.Help()

	assert.Equal(t, "j/k org  enter open  esc quit", help.String())
	if assert.Len(t, help.FullHelp(), 2, "the full view adds the paging keys") {
		assert.Len(t, help.FullHelp()[0], 4)
		assert.Equal(t, "first/last", help.FullHelp()[0][3].Help().Desc)
	}
}

func TestModel_IgnoresUnboundKeys(t *testing.T) {
	m := newTestModel()
	m.SetUser(newTestUser("octocat", "acme"))

	before := plain(m.View())
	_, cmd := m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	assert.Nil(t, cmd)
	assert.Equal(t, before, plain(m.View()))

	_, cmd = m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	assert.Nil(t, cmd, "q does not quit; esc and ctrl+c do")
}

func TestModel_Update_EscGoesBack(t *testing.T) {
	m := newTestModel()
	m.SetUser(newTestUser("octocat", "acme"))

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if assert.NotNil(t, cmd) {
		prev, ok := cmd().(shared.PreviousMsg)
		assert.True(t, ok, "expected a PreviousMsg")
		assert.Nil(t, prev.Message)
	}
}

func TestModel_VimKeysMoveTheCursor(t *testing.T) {
	m := newTestModel()
	m.SetUser(newTestUser("octocat", "acme", "globex"))

	m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	assert.Equal(t, 1, m.cursor)
	m.Update(tea.KeyPressMsg{Code: 'k', Text: "k"})
	assert.Equal(t, 0, m.cursor)
}

func TestUserKeyMap(t *testing.T) {
	keymap := newUserKeyMap()

	assert.Equal(t, []string{"up", "k"}, keymap.Up.Keys())
	assert.Equal(t, []string{"down", "j"}, keymap.Down.Keys())
	assert.Equal(t, []string{"left", "h", "pgup", "b", "u"}, keymap.PageUp.Keys())
	assert.Equal(t, []string{"right", "l", "pgdown", "f", "d"}, keymap.PageDown.Keys())
	assert.Equal(t, []string{"home", "g"}, keymap.Top.Keys())
	assert.Equal(t, []string{"end", "G"}, keymap.Bottom.Keys())
	assert.Equal(t, []string{"enter"}, keymap.Select.Keys())
	assert.Equal(t, []string{"esc"}, keymap.Back.Keys())
}
