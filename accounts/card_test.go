package accounts

import (
	"strings"
	"testing"
	"time"

	"gh-reponark/github"
	"gh-reponark/github/githubtest"
	"gh-reponark/ui/uitest"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
)

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

	lines := strings.Split(uitest.Plain(m.View()), "\n")

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

	lines := strings.Split(uitest.Plain(m.View()), "\n")

	assert.Equal(t, "  O  The Octocat  octocat  ▐you▌", strings.TrimRight(lines[1], " "))
	assert.Equal(t, "     8 repos · 6 public · 2 private · 9k followers · since 2011", strings.TrimRight(lines[2], " "))
	assert.Equal(t, "", strings.TrimRight(lines[3], " "))
	assert.Equal(t, "  A  Acme Robotics  acme  ▐org▌ ▐admin▌ ▐verified▌", strings.TrimRight(lines[4], " "))
	assert.NotContains(t, uitest.Plain(m.View()), "https://", "short cards drop the URL and description")
}

func TestModel_View_NarrowTier(t *testing.T) {
	m := NewModel(&githubtest.Fake{}, 44, 36)
	m.SetUser(newRichUser())

	content := uitest.Plain(m.View())
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

	lines := strings.Split(uitest.Plain(m.View()), "\n")

	assert.Equal(t, "  An orga…  ▐org▌ ▐admin▌ ▐verified▌", strings.TrimRight(lines[5], " "))
}
