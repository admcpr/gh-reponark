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
	items := m.orgList.Items()
	titles := make([]string, len(items))
	for i, item := range items {
		titles[i] = item.(shared.ListItem).Title()
	}
	return titles
}

func TestNewModel(t *testing.T) {
	m := newTestModel()

	assert.Equal(t, 80, m.width)
	assert.Equal(t, 24, m.height)
	assert.Equal(t, "", m.login)
	assert.Empty(t, m.orgList.Items())
	assert.Equal(t, 80, m.help.Width())
}

func TestModel_SetDimensions(t *testing.T) {
	m := newTestModel()

	m.SetDimensions(120, 40)

	assert.Equal(t, 120, m.width)
	assert.Equal(t, 40, m.height)
	assert.Equal(t, 120, m.help.Width())
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

	first := m.orgList.Items()[0].(shared.ListItem)
	assert.Equal(t, "https://github.com/octocat", first.Description())
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

func TestModel_Update_OtherKeysGoToList(t *testing.T) {
	m := newTestModel()
	m.SetUser(newTestUser("octocat", "acme"))
	assert.Equal(t, 0, m.orgList.Index())

	m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	assert.Equal(t, 1, m.orgList.Index())

	m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	assert.Equal(t, 0, m.orgList.Index())
}

func TestModel_Update_NonKeyMessagesGoToList(t *testing.T) {
	m := newTestModel()
	m.SetUser(newTestUser("octocat", "acme"))

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 10, Height: 10})

	assert.Same(t, m, updated)
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

func TestModel_HeaderView(t *testing.T) {
	m := newTestModel()
	assert.Contains(t, plain(m.HeaderView()), "Organizations")

	m.SetUser(newTestUser("octocat"))
	assert.Contains(t, plain(m.HeaderView()), "User: octocat")
}

func TestModel_HelpView(t *testing.T) {
	m := newTestModel()

	content := plain(m.HelpView())

	assert.Contains(t, content, "select")
	assert.Contains(t, content, "back")
}

func TestUserKeyMap(t *testing.T) {
	keymap := userKeyMap{}

	short := keymap.ShortHelp()
	assert.Len(t, short, 4)

	full := keymap.FullHelp()
	assert.Len(t, full, 1)
	assert.Equal(t, short, full[0])
}
