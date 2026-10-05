package user

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"gh-reponark/github"
	"gh-reponark/github/githubtest"
	"gh-reponark/shared"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

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
		msg, ok := loadMsg(cmd).(userLoadedMsg)
		assert.True(t, ok, "expected a userLoadedMsg")
		assert.Equal(t, fake.User, github.User(msg))
	}
}

func TestModel_Init_ReportsErrors(t *testing.T) {
	fake := &githubtest.Fake{UserErr: errors.New("not logged in")}
	m := NewModel(fake, 80, 24)

	msg, ok := loadMsg(m.Init()).(shared.ErrorMsg)

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
	// Ten lines is the short tier: three two-line cards and the blank lines
	// between them fit under the heading.
	m := NewModel(&githubtest.Fake{}, 80, 10)
	m.SetUser(newTestUser("octocat", "a", "b", "c", "d", "e", "f"))
	assert.Equal(t, 3, m.rows())

	m.Update(tea.KeyPressMsg{Code: tea.KeyPgDown})
	assert.Equal(t, 3, m.cursor, "page down moves a screenful")
	assert.Equal(t, 1, m.offset, "and scrolls the list to keep the cursor in view")

	m.Update(tea.KeyPressMsg{Code: 'G', Text: "G"})
	assert.Equal(t, 6, m.cursor)
	assert.Equal(t, 4, m.offset)
	content := plain(m.View())
	assert.Contains(t, content, "f")
	assert.NotContains(t, content, "octocat", "cards scrolled off the top are hidden")

	m.Update(tea.KeyPressMsg{Code: tea.KeyPgUp})
	assert.Equal(t, 3, m.cursor)

	m.Update(tea.KeyPressMsg{Code: 'g', Text: "g"})
	assert.Equal(t, 0, m.cursor)
	assert.Equal(t, 0, m.offset)
}

func TestModel_SetDimensions_KeepsCursorInView(t *testing.T) {
	m := NewModel(&githubtest.Fake{}, 80, 40) // eight four-line cards
	m.SetUser(newTestUser("octocat", "a", "b", "c", "d", "e", "f"))
	m.Update(tea.KeyPressMsg{Code: 'G', Text: "G"})
	assert.Equal(t, 0, m.offset, "everything fits at forty rows")

	m.SetDimensions(80, 10) // three two-line cards

	assert.Equal(t, 4, m.offset, "a shorter pane scrolls to the cursor")
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

	content := plain(m.View())

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

	content := plain(m.View())

	assert.Contains(t, content, "Signing in to GitHub…")
	assert.NotContains(t, content, "ACCOUNTS")
}

// newRichUser is a user with a named, described organization so the cards
// have something to show on every line.
func newRichUser() github.User {
	user := newTestUser("octocat")
	user.Name = "The Octocat"
	user.Description = "Mascot"
	user.Repositories = 8
	user.PublicRepositories = 6
	user.Members = 9000
	user.CreatedAt = time.Date(2011, 1, 25, 0, 0, 0, 0, time.UTC)
	user.Organizations = []github.Organization{{
		Login:               "acme",
		Name:                "Acme Robotics",
		Description:         "Anvils  and\nrockets",
		Url:                 "https://github.com/acme",
		Repositories:        1204,
		PublicRepositories:  1000,
		Members:             18,
		ViewerCanAdminister: true,
		IsVerified:          true,
		CreatedAt:           time.Date(2015, 3, 4, 0, 0, 0, 0, time.UTC),
	}}
	return user
}

func TestModel_View_Cards(t *testing.T) {
	m := NewModel(&githubtest.Fake{}, 120, 36)
	m.SetUser(newRichUser())

	lines := strings.Split(plain(m.View()), "\n")

	assert.Equal(t, "  ACCOUNTS  ▐1 organisation▌", strings.TrimRight(lines[0], " "))
	// The user's card: monogram, name and login, the you pill.
	assert.Equal(t, "  O  The Octocat  octocat  ▐you▌", strings.TrimRight(lines[1], " "))
	assert.Equal(t, "     Mascot", strings.TrimRight(lines[2], " "))
	assert.Equal(t, "     8 repos · 6 public · 2 private · 9k followers · since 2011", strings.TrimRight(lines[3], " "))
	assert.Equal(t, "     https://github.com/octocat", strings.TrimRight(lines[4], " "))
	assert.Equal(t, "", strings.TrimRight(lines[5], " "), "a blank line separates the cards")
	// The organization's card: whitespace in the description is collapsed.
	assert.Equal(t, "  A  Acme Robotics  acme  ▐org▌ ▐admin▌ ▐verified▌", strings.TrimRight(lines[6], " "))
	assert.Equal(t, "     Anvils and rockets", strings.TrimRight(lines[7], " "))
	assert.Equal(t, "     1.2k repos · 1k public · 204 private · 18 members · since 2015", strings.TrimRight(lines[8], " "))
	assert.Equal(t, "     https://github.com/acme", strings.TrimRight(lines[9], " "))
}

func TestModel_View_ShortTier(t *testing.T) {
	m := NewModel(&githubtest.Fake{}, 120, 12)
	m.SetUser(newRichUser())

	lines := strings.Split(plain(m.View()), "\n")

	assert.Equal(t, "  O  The Octocat  octocat  ▐you▌", strings.TrimRight(lines[1], " "))
	assert.Equal(t, "     8 repos · 6 public · 2 private · 9k followers · since 2011", strings.TrimRight(lines[2], " "))
	assert.Equal(t, "", strings.TrimRight(lines[3], " "))
	assert.Equal(t, "  A  Acme Robotics  acme  ▐org▌ ▐admin▌ ▐verified▌", strings.TrimRight(lines[4], " "))
	assert.NotContains(t, plain(m.View()), "https://", "short cards drop the URL and description")
}

func TestModel_View_NarrowTier(t *testing.T) {
	m := NewModel(&githubtest.Fake{}, 44, 36)
	m.SetUser(newRichUser())

	content := plain(m.View())
	lines := strings.Split(content, "\n")

	assert.Equal(t, "  The Octocat  octocat  ▐you▌", strings.TrimRight(lines[1], " "), "no monogram when narrow")
	assert.Equal(t, "  Mascot", strings.TrimRight(lines[2], " "))
	assert.Equal(t, "  8 repos · 6 public · 2 private", strings.TrimRight(lines[3], " "), "facts are shed from the end")
	assert.Equal(t, "", strings.TrimRight(lines[4], " "), "and there is no URL line")
	assert.Equal(t, "  Acme Robotics  ▐org▌ ▐admin▌ ▐verified▌", strings.TrimRight(lines[5], " "),
		"the login is dropped whole before the name is cut")
	assert.NotContains(t, content, "https://")
	for _, line := range lines {
		assert.LessOrEqual(t, lipgloss.Width(line), 44)
	}
}

func TestModel_View_PillsAreNeverCut(t *testing.T) {
	m := NewModel(&githubtest.Fake{}, 36, 36)
	user := newRichUser()
	user.Organizations[0].Name = "An organisation with a very long name indeed"
	m.SetUser(user)

	lines := strings.Split(plain(m.View()), "\n")

	assert.Equal(t, "  An orga…  ▐org▌ ▐admin▌ ▐verified▌", strings.TrimRight(lines[5], " "))
}

func TestMonogram(t *testing.T) {
	assert.Equal(t, "O ", ansi.Strip(monogram("octocat")))
	assert.Equal(t, "Ü ", ansi.Strip(monogram("über")))
	assert.Equal(t, "? ", ansi.Strip(monogram("")))
	assert.Equal(t, monogram("acme"), monogram("acme"), "the colour is stable")
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
