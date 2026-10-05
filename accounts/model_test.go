package accounts

import (
	"errors"
	"testing"

	"gh-reponark/github"
	"gh-reponark/github/githubtest"
	"gh-reponark/ui"
	"gh-reponark/ui/uitest"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

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
		titles[i] = item.login
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
		msg, ok := uitest.LoadMsg(cmd).(userLoadedMsg)
		assert.True(t, ok, "expected a userLoadedMsg")
		assert.Equal(t, fake.User, github.User(msg))
	}
}

func TestModel_Init_ReportsErrors(t *testing.T) {
	fake := &githubtest.Fake{UserErr: errors.New("not logged in")}
	m := NewModel(fake, 80, 24)

	msg, ok := uitest.LoadMsg(m.Init()).(ui.ErrorMsg)

	assert.True(t, ok, "expected an ErrorMsg")
	assert.EqualError(t, msg.Err, "not logged in")
}

func TestModel_SetUser(t *testing.T) {
	m := newTestModel()

	m.SetUser(newTestUser("octocat", "zulu", "alpha", "mike"))

	assert.Equal(t, "octocat", m.login)
	assert.Equal(t, []string{"octocat", "alpha", "mike", "zulu"}, itemTitles(m),
		"the user should be first followed by organisations sorted by login")

	assert.Equal(t, "https://github.com/octocat", m.items[0].url)
	assert.True(t, m.items[0].isUser, "the first card is the signed-in user")
	assert.False(t, m.items[1].isUser)
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
		want      ui.OrgKey
	}{
		{name: "user entry", downCount: 0, want: ui.OrgKey{Name: "octocat", IsUser: true}},
		{name: "organisation entry", downCount: 1, want: ui.OrgKey{Name: "acme", IsUser: false}},
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
				open, ok := cmd().(ui.OpenOrgMsg)
				assert.True(t, ok, "expected an OpenOrgMsg")
				assert.Equal(t, tt.want, open.Key)
			}
		})
	}
}

func TestModel_Update_ArrowsMoveTheCursor(t *testing.T) {
	m := newTestModel()
	m.SetUser(newTestUser("octocat", "acme"))
	assert.Equal(t, 0, m.cursor.Index)

	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	assert.Equal(t, 1, m.cursor.Index)

	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	assert.Equal(t, 1, m.cursor.Index, "the cursor stops at the end")

	m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	assert.Equal(t, 0, m.cursor.Index)

	m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	assert.Equal(t, 0, m.cursor.Index, "and at the start")
}

func TestModel_Update_PagingKeys(t *testing.T) {
	// Ten lines is the short tier: three two-line cards and the blank lines
	// between them fit under the heading.
	m := NewModel(&githubtest.Fake{}, 80, 10)
	m.SetUser(newTestUser("octocat", "a", "b", "c", "d", "e", "f"))
	assert.Equal(t, 3, m.rows())

	m.Update(tea.KeyPressMsg{Code: tea.KeyPgDown})
	assert.Equal(t, 3, m.cursor.Index, "page down moves a screenful")
	assert.Equal(t, 1, m.cursor.Offset, "and scrolls the list to keep the cursor in view")

	m.Update(tea.KeyPressMsg{Code: 'G', Text: "G"})
	assert.Equal(t, 6, m.cursor.Index)
	assert.Equal(t, 4, m.cursor.Offset)
	content := uitest.Plain(m.View())
	assert.Contains(t, content, "f")
	assert.NotContains(t, content, "octocat", "cards scrolled off the top are hidden")

	m.Update(tea.KeyPressMsg{Code: tea.KeyPgUp})
	assert.Equal(t, 3, m.cursor.Index)

	m.Update(tea.KeyPressMsg{Code: 'g', Text: "g"})
	assert.Equal(t, 0, m.cursor.Index)
	assert.Equal(t, 0, m.cursor.Offset)
}

func TestModel_SetDimensions_KeepsCursorInView(t *testing.T) {
	m := NewModel(&githubtest.Fake{}, 80, 40) // eight four-line cards
	m.SetUser(newTestUser("octocat", "a", "b", "c", "d", "e", "f"))
	m.Update(tea.KeyPressMsg{Code: 'G', Text: "G"})
	assert.Equal(t, 0, m.cursor.Offset, "everything fits at forty rows")

	m.SetDimensions(80, 10) // three two-line cards

	assert.Equal(t, 4, m.cursor.Offset, "a shorter pane scrolls to the cursor")
}

func TestModel_Rows_ByTier(t *testing.T) {
	tests := []struct {
		width, height, want int
	}{
		{120, 36, 7}, // four-line cards and a blank line each: 36 lines hold seven
		{120, 16, 3}, // the smallest full-height pane
		{120, 15, 5}, // the short tier: two-line cards
		{40, 36, 9},  // narrow: no URL line, three-line cards
		{40, 8, 2},   // narrow and short
		{120, 1, 1},  // never fewer than one card
	}
	for _, tt := range tests {
		m := NewModel(&githubtest.Fake{}, tt.width, tt.height)
		assert.Equal(t, tt.want, m.rows(), "%dx%d", tt.width, tt.height)
	}
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

	content := uitest.Plain(m.View())

	assert.Contains(t, content, "ACCOUNTS")
	assert.Contains(t, content, "1 organisation")
	assert.Contains(t, content, "octocat")
	assert.Contains(t, content, "acme")
	assert.Contains(t, content, "you")
	assert.Contains(t, content, "org")
	assert.Contains(t, content, "https://github.com/acme")
	assert.Contains(t, content, "No description")
}

func TestModel_View_SigningIn(t *testing.T) {
	m := newTestModel()

	content := uitest.Plain(m.View())

	assert.Contains(t, content, "Signing in to GitHub…")
	assert.NotContains(t, content, "ACCOUNTS")
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

	before := uitest.Plain(m.View())
	_, cmd := m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	assert.Nil(t, cmd)
	assert.Equal(t, before, uitest.Plain(m.View()))

	_, cmd = m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	assert.Nil(t, cmd, "q does not quit; esc and ctrl+c do")
}

func TestModel_Update_EscGoesBack(t *testing.T) {
	m := newTestModel()
	m.SetUser(newTestUser("octocat", "acme"))

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if assert.NotNil(t, cmd) {
		prev, ok := cmd().(ui.PreviousMsg)
		assert.True(t, ok, "expected a PreviousMsg")
		assert.Nil(t, prev.Message)
	}
}

func TestModel_VimKeysMoveTheCursor(t *testing.T) {
	m := newTestModel()
	m.SetUser(newTestUser("octocat", "acme", "globex"))

	m.Update(tea.KeyPressMsg{Code: 'j', Text: "j"})
	assert.Equal(t, 1, m.cursor.Index)
	m.Update(tea.KeyPressMsg{Code: 'k', Text: "k"})
	assert.Equal(t, 0, m.cursor.Index)
}
