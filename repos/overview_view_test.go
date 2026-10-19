package repos

import (
	"errors"
	"strings"
	"testing"
	"time"

	"gh-reponark/filter"
	"gh-reponark/filters"
	"gh-reponark/github"
	"gh-reponark/github/githubtest"
	"gh-reponark/overview"
	"gh-reponark/repo"
	"gh-reponark/ui"
	"gh-reponark/ui/uitest"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// overviewRepos are three repositories that differ in the ways the cards
// count: one public and protected, one private with alerts, one archived.
func overviewRepos() []repo.Repository {
	alpha := repo.Repository{Name: "alpha", Visibility: "PUBLIC", StargazerCount: 100, HasIssuesEnabled: true, HasVulnerabilityAlertsEnabled: true, SquashMergeAllowed: true, PushedAt: time.Now(), ViewerPermission: "ADMIN"}
	alpha.PrimaryLanguage.Name = "Go"
	alpha.LicenseInfo.Name = "MIT License"
	alpha.DefaultBranchRef.Name = "main"
	alpha.Rulesets.TotalCount = 1
	pr := repo.BranchRule{Type: "PULL_REQUEST"}
	pr.Parameters.PullRequest.RequiredApprovingReviewCount = 1
	pr.Ruleset.Name, pr.Ruleset.Source.Typename = "main", "Repository"
	alpha.DefaultBranchRef.Rules.Nodes = []repo.BranchRule{pr}
	bravo := repo.Repository{Name: "bravo", Visibility: "PRIVATE", IsPrivate: true, StargazerCount: 10, HasIssuesEnabled: true, SquashMergeAllowed: true, PushedAt: time.Now(), ViewerPermission: "ADMIN"}
	bravo.PrimaryLanguage.Name = "Go"
	bravo.DefaultBranchRef.Name = "main"
	bravo.VulnerabilityAlerts.TotalCount = 2
	charlie := repo.Repository{Name: "charlie", Visibility: "PUBLIC", IsArchived: true, SquashMergeAllowed: true, PushedAt: time.Now().AddDate(-2, 0, 0), ViewerPermission: "WRITE"}
	charlie.DefaultBranchRef.Name = "master"
	return []repo.Repository{alpha, bravo, charlie}
}

// newOverviewModel loads the repositories, switches to the overview and
// gives it a profile.
func newOverviewModel(width, height int) *Model {
	m := newLoadedModel(overviewRepos()...)
	m.SetDimensions(width, height)
	m.Update(tea.KeyPressMsg{Code: 'v', Text: "v"})
	m.Update(tea.KeyPressMsg{Code: 'v', Text: "v"})
	m.Update(ownerMsg(github.Owner{
		Login: "demo", Name: "Demo Org", Description: "Anvils and rockets.",
		IsVerified: true, ViewerIsAMember: true, Members: 18, Teams: 3, Location: "Springfield",
		CreatedAt: time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC),
	}))
	return m
}

func press(m *Model, keys ...string) {
	for _, k := range keys {
		switch k {
		case "tab":
			m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		case "shift+tab":
			m.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
		case "enter":
			m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		case "esc":
			m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
		default:
			m.Update(tea.KeyPressMsg{Code: rune(k[0]), Text: k})
		}
	}
}

// focused is the label of the focused slice, or "" for none.
func focused(m *Model) string {
	card, focus := m.focusedCard()
	if focus < 0 {
		return ""
	}
	return card.Slices[focus].Label
}

func TestModel_Init_LoadsTheOwner(t *testing.T) {
	svc := &githubtest.Fake{User: github.User{Login: "demo"}}
	m := NewModel(svc, ui.OrgKey{Name: "acme", IsUser: false}, 80, 24)

	drive(m, m.Init())

	require.Equal(t, []githubtest.OwnerCall{{Login: "acme", IsUser: false}}, svc.OwnerCalls)
	assert.True(t, m.ownerLoaded)
	assert.Equal(t, "acme", m.owner.Login)
}

func TestModel_OwnerErrorIsNotFatal(t *testing.T) {
	m := newLoadedModel(overviewRepos()...)
	m.SetDimensions(120, 40)
	m.svc = &githubtest.Fake{OwnerErr: errors.New("fetching organization acme: 403 Forbidden\nmore detail")}

	msg := m.loadOwner()
	_, cmd := m.Update(msg)

	assert.Nil(t, cmd, "no error screen is opened")
	assert.False(t, m.ownerLoaded)
	m.mode = overviewView
	content := uitest.Plain(m.View())
	assert.Contains(t, content, "Profile unavailable: fetching organization acme: 403 Forbidden")
	assert.NotContains(t, content, "more detail", "only the first line fits")
	assert.Contains(t, content, "REPO MIX", "the repositories are still charted")
}

func TestModel_OverviewView_Header(t *testing.T) {
	m := newOverviewModel(120, 40)

	content := uitest.Plain(m.View())

	assert.Contains(t, content, "Demo Org  demo  ▐org▌ ▐verified▌ ▐member▌")
	assert.Contains(t, content, "Anvils and rockets.")
	assert.Contains(t, content, "18 members · 3 teams · since 2015 · Springfield")
	assert.NotContains(t, content, "repos loaded", "the profile's repository counts are not comparable")
}

func TestModel_OverviewView_UserHeader(t *testing.T) {
	m := newOverviewModel(120, 40)
	m.isUser = true
	m.Update(ownerMsg(github.Owner{Login: "octocat", IsUser: true, Members: 9, Following: 4, Company: "GitHub"}))

	content := uitest.Plain(m.View())

	assert.Contains(t, content, "octocat  ▐you▌")
	assert.Contains(t, content, "9 followers · 4 following · GitHub")
}

func TestModel_OverviewView_LoadingProfile(t *testing.T) {
	m := newLoadedModel(overviewRepos()...)
	m.SetDimensions(120, 40)
	m.mode = overviewView

	content := uitest.Plain(m.View())

	assert.Contains(t, content, "demo", "the login stands in for the name")
	assert.Contains(t, content, "Loading profile…")
}

func TestModel_OverviewView_Cards(t *testing.T) {
	m := newOverviewModel(160, 44)

	content := uitest.Plain(m.View())

	for _, title := range []string{"REPO MIX", "LAST PUSH", "LANGUAGES", "LICENSES", "SECURITY", "ALERTS", "BRANCH RULES", "RULE SOURCES", "MERGE STYLES", "FEATURES", "OTHER SETTINGS"} {
		assert.Contains(t, content, title)
	}
	assert.Contains(t, content, "▸ REPO MIX")
	assert.Regexp(t, `● public 1 ● private 1 ● archived 1`, content)
	assert.Regexp(t, ` fork 0 · template 0 · empty 0`, content)
	assert.Regexp(t, `your role admin 2 · write 1`, content)
	assert.Regexp(t, `branch main 2 · master 1`, content)
	assert.Regexp(t, `Go\s+━+╌*\s+2\s+67%`, content, "shares are rounded, not truncated")
	assert.Regexp(t, `MIT\s+━+╌+\s+1\s+33%`, content)
	assert.Contains(t, content, "2 open alerts")
	assert.Contains(t, content, "1 of 3 with any")
	assert.Regexp(t, `pull request\s+━+╌+\s+1\s+33%`, content)
	assert.Regexp(t, `repo\s+━+╌+\s+1\s+33%`, content, "rule sources")
	assert.Contains(t, content, "1 over a year ago")
	assert.Contains(t, content, "most evenly split first")
	for _, judgement := range []string{"guarded", "idle", "contested", "agree", "hold"} {
		assert.NotContains(t, content, judgement)
	}
}

func TestModel_OverviewView_Footer(t *testing.T) {
	m := newOverviewModel(120, 40)
	press(m, "j", "j", "j") // Repo mix › archived, then fork has none, so the role admin
	press(m, "k", "k")

	content := uitest.Plain(m.View())

	assert.Contains(t, content, "REPO MIX › private")
	assert.Contains(t, content, "j/k kind · enter show repos")
	assert.Contains(t, content, "1 repo · 33% of 3")
	assert.Regexp(t, `● bravo`, content)
	assert.NotContains(t, content, "★", "the footer does not rank by stars")
}

func TestModel_OverviewView_FooterListsByName(t *testing.T) {
	m := newOverviewModel(120, 40)
	press(m, "j", "j", "j") // Repo mix › admin: alpha and bravo

	lines := strings.Split(uitest.Plain(m.View()), "\n")

	require.Contains(t, strings.Join(lines, "\n"), "REPO MIX › admin")
	var names string
	for _, line := range lines {
		if strings.Contains(line, "● alpha") {
			names = line
		}
	}
	assert.Less(t, strings.Index(names, "alpha"), strings.Index(names, "bravo"), "by name, not by stars")
}

func TestModel_OverviewView_KeysMoveCardsAndSlices(t *testing.T) {
	m := newOverviewModel(120, 36) // three columns
	m.View()

	press(m, "tab")
	assert.Equal(t, 1, m.card, "tab moves to the next card")
	press(m, "tab", "shift+tab")
	assert.Equal(t, 1, m.card)
	press(m, "shift+tab", "shift+tab")
	assert.Equal(t, 10, m.card, "and wraps round at either end, as between groups")
	press(m, "tab")
	assert.Equal(t, 0, m.card)

	press(m, "l")
	assert.Equal(t, 1, m.card, "l is the card to the right")
	press(m, "h", "h")
	assert.Equal(t, 0, m.card, "h stops at the first card")
	press(m, "G")
	assert.Equal(t, 10, m.card)
	press(m, "l")
	assert.Equal(t, 10, m.card, "l stops at the last card")
	press(m, "g")
	assert.Equal(t, 0, m.card)

	assert.Equal(t, "public", focused(m))
	press(m, "j")
	assert.Equal(t, "private", focused(m), "j is the next slice")
	assert.Contains(t, uitest.Plain(m.View()), "REPO MIX › private")
	press(m, "k", "k")
	assert.Equal(t, "public", focused(m), "k stops at the first slice")
	press(m, "j", "j", "j", "j", "j", "j", "j", "j", "j", "j")
	assert.Equal(t, "master", focused(m), "j stops at the last")
	press(m, "tab")
	assert.Equal(t, "<1w", focused(m), "a new card starts on its first slice")
	assert.Equal(t, 0, m.list.Index, "the repository selection is untouched")
	assert.False(t, m.inspecting)
}

func TestModel_OverviewView_EnterFiltersToSliceAndEscReturns(t *testing.T) {
	m := newOverviewModel(120, 36)
	m.View()
	press(m, "j") // Repo mix › private

	press(m, "enter")

	assert.Equal(t, listView, m.mode)
	assert.Equal(t, []string{"bravo"}, itemNames(m))
	assert.Len(t, m.filters, 2, "private and not archived")
	assert.Equal(t, filter.NewBoolFilter("Is Private", true), m.filters["Is Private"])
	assert.Equal(t, "1 of 3 repos · 2 filters", m.Status())
	assert.Contains(t, m.Help().String(), "esc overview")

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	assert.Nil(t, cmd, "esc does not leave the screen")
	assert.Equal(t, overviewView, m.mode)
	assert.Empty(t, m.filters, "the slice's filters are taken off again")
	assert.Len(t, m.visible, 3)
	assert.Equal(t, "private", focused(m), "the focus is where it was")

	_, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	require.NotNil(t, cmd)
	assert.IsType(t, ui.PreviousMsg{}, cmd(), "from the overview esc leaves as before")
}

func TestModel_OverviewView_EscBackOnlyUntilFiltersAreEdited(t *testing.T) {
	m := newOverviewModel(120, 36)
	m.View()
	press(m, "enter") // Repo mix › public

	m.Update(filters.FiltersMsg(filter.FilterMap{"Is Fork": filter.NewBoolFilter("Is Fork", false)}))
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	require.NotNil(t, cmd, "filters edited by hand are kept, so esc leaves")
	assert.IsType(t, ui.PreviousMsg{}, cmd())

	m = newOverviewModel(120, 36)
	m.View()
	press(m, "enter", "v")
	assert.False(t, m.fromOverview, "moving on with v keeps the filters")
}

func TestModel_OverviewView_EnterNarrowsAnExistingFilter(t *testing.T) {
	m := newOverviewModel(120, 36)
	m.Update(filters.FiltersMsg(filter.FilterMap{"Vulnerability Alerts": filter.NewIntFilter("Vulnerability Alerts", 1, filter.NoMax)}))
	m.mode = overviewView
	m.card = 5 // Alerts, over bravo alone: the 0 bucket is empty
	m.View()
	assert.Equal(t, "2-3", focused(m))

	press(m, "enter")

	assert.Equal(t, []string{"bravo"}, itemNames(m))
	assert.Equal(t, "≥ 1 and 2 – 3", m.filters["Vulnerability Alerts"].Condition(), "combined, not replaced")
}

func TestModel_OverviewView_EnterFiltersByNameWhenNeeded(t *testing.T) {
	m := newOverviewModel(120, 36)
	m.View()
	press(m, "l", "l", "l", "l") // Security: protected is rules or rulesets

	press(m, "enter")

	assert.Equal(t, listView, m.mode)
	assert.Equal(t, []string{"alpha"}, itemNames(m))
	assert.Equal(t, "Security › protected (1 repo)", m.filters["Name"].Condition(), "the chip says where the selection came from")
}

func TestModel_OverviewView_FilterLine(t *testing.T) {
	m := newOverviewModel(120, 40)
	m.Update(filters.FiltersMsg(filter.FilterMap{
		"Is Archived":      filter.NewBoolFilter("Is Archived", false),
		"Primary Language": filter.NewExactStringFilter("Primary Language", "Go"),
	}))
	m.mode = overviewView

	content := uitest.Plain(m.View())

	assert.Contains(t, content, `▐filtered 2 of 3▌ Is Archived: no · Primary Language: is "Go"`)
	assert.Contains(t, content, "2 repos", "the cards count the visible repositories")
}

func TestModel_EmptyStateNamesTheFilters(t *testing.T) {
	m := newOverviewModel(120, 40)
	m.Update(filters.FiltersMsg(filter.FilterMap{"Is Fork": filter.NewBoolFilter("Is Fork", true)}))

	assert.Equal(t, "No repositories match the 1 filter · f to change them", uitest.Plain(m.View()))
}

func TestModel_OverviewLayout(t *testing.T) {
	m := newOverviewModel(120, 36)
	assert.Equal(t, overviewLayout{description: true, facts: true, columns: 3, cardWidth: 38, rows: 4, footer: 5}, m.overviewLayout())

	m.SetDimensions(120, 33)
	assert.Equal(t, overviewLayout{description: true, facts: true, columns: 3, cardWidth: 38, rows: 3, footer: 9}, m.overviewLayout(),
		"spare height goes to the footer")

	m.SetDimensions(160, 44)
	assert.Equal(t, overviewLayout{description: true, facts: true, columns: 4, cardWidth: 38, rows: 3, footer: 20}, m.overviewLayout())

	m.SetDimensions(100, 30)
	assert.Equal(t, overviewLayout{description: true, facts: true, columns: 2, cardWidth: 48, rows: 3, footer: 6}, m.overviewLayout(),
		"two columns share out the width")

	m.SetDimensions(80, 24)
	assert.Equal(t, overviewLayout{facts: true, columns: 2, cardWidth: 38, rows: 2, footer: 8}, m.overviewLayout(),
		"a card a gap narrower than the pane, so there is a gutter")

	m.SetDimensions(60, 16)
	assert.Equal(t, overviewLayout{columns: 1, cardWidth: 56, rows: 1, footer: 8}, m.overviewLayout(), "the facts give way")
	assert.Equal(t, 16, strings.Count(uitest.Plain(m.View()), "\n")+1, "the view still fills the pane")
}

func TestModel_OverviewView_ScrollsCards(t *testing.T) {
	m := newOverviewModel(120, 33) // nine of the eleven cards fit
	content := uitest.Plain(m.View())
	assert.Contains(t, content, "▾ 2 below")
	assert.NotContains(t, content, "OTHER SETTINGS")

	press(m, "G")
	content = uitest.Plain(m.View())

	assert.Contains(t, content, "▸ OTHER SETTINGS")
	assert.Contains(t, content, "▴ 3 above")
	assert.NotContains(t, content, "REPO MIX", "the first row scrolled out")
	assert.NotContains(t, content, "below")
}

func TestModel_OverviewView_NoRepos(t *testing.T) {
	m := newOverviewModel(120, 40)
	m.visible = nil

	assert.Contains(t, uitest.Plain(m.View()), "No repositories found")
}

func TestModel_OverviewCards_Cached(t *testing.T) {
	m := newOverviewModel(120, 40)
	first := m.overviewCards()
	press(m, "l", "tab")
	assert.Same(t, &first[0], &m.overviewCards()[0], "keys reuse the cards")

	m.Update(filters.FiltersMsg(filter.FilterMap{"Is Archived": filter.NewBoolFilter("Is Archived", false)}))
	assert.Equal(t, "2 repos", m.overviewCards()[0].Headline, "new filters rebuild them")
}

func TestModel_OverviewView_SlicesSkipEmptyOnes(t *testing.T) {
	m := newOverviewModel(120, 36)
	m.View()
	press(m, "tab", "tab", "tab", "tab", "tab") // Alerts: 0 has 2 repos, 2-3 has 1, the rest none

	assert.Equal(t, "0", focused(m))
	press(m, "j")
	assert.Equal(t, "2-3", focused(m), "1 is empty, so j goes to 2-3")
	press(m, "j")
	assert.Equal(t, "2-3", focused(m), "4-7, 8-15 and 16+ are empty, so it stays")
	press(m, "k")
	assert.Equal(t, "0", focused(m))
	press(m, "j")
	assert.Contains(t, uitest.Plain(m.View()), "ALERTS › 2-3")
}

func TestModel_OverviewView_NoFocusableSlice(t *testing.T) {
	m := newOverviewModel(120, 36)
	m.card = 4 // Security
	// Turn every Security setting off so the card has nothing to list.
	for i := range m.visible {
		for _, name := range []string{"Is Security Policy Enabled", "Has Vulnerability Alerts Enabled", "Web Commit Signoff Required", "Forking Allowed"} {
			m.visible[i].Properties[name] = repo.RepoProperty{Value: false, Type: "bool"}
		}
		m.visible[i].Properties["Branch Protection Rules"] = repo.RepoProperty{Value: 0, Type: "int"}
		m.visible[i].Properties["Rulesets"] = repo.RepoProperty{Value: 0, Type: "int"}
	}
	m.cards = nil

	content := uitest.Plain(m.View())
	assert.Equal(t, -1, m.slice)
	assert.Contains(t, content, "No repositories to list on this card")

	press(m, "j", "enter")
	assert.Equal(t, overviewView, m.mode, "enter does nothing without a slice")
	assert.Empty(t, m.filters)
}

func TestModel_OverviewView_EnterUsesExactValues(t *testing.T) {
	java := repo.Repository{Name: "java"}
	java.PrimaryLanguage.Name = "Java"
	js := repo.Repository{Name: "js"}
	js.PrimaryLanguage.Name = "JavaScript"
	m := newLoadedModel(java, js)
	m.SetDimensions(120, 36)
	m.mode = overviewView
	press(m, "l", "l") // Languages: Java, then JavaScript
	require.Equal(t, "Java", focused(m))

	press(m, "enter")

	assert.Equal(t, []string{"java"}, itemNames(m), "JavaScript is not swept in")
	assert.Equal(t, `is "Java"`, m.filters["Primary Language"].Condition())
}

func TestModel_OverviewView_OtherSettingsRow(t *testing.T) {
	var repos []repo.Repository
	for i := 0; i < 10; i++ {
		repos = append(repos, repo.Repository{Name: string(rune('a' + i)), HasSponsorshipsEnabled: i == 0})
	}
	m := newLoadedModel(repos...)
	m.SetDimensions(120, 40)
	m.mode = overviewView
	press(m, "G")

	content := uitest.Plain(m.View())

	assert.Regexp(t, `Sponsorships\s+off 9 · on a`, content, "the shared value and the ones set differently")
	assert.Contains(t, content, "OTHER SETTINGS › Sponsorships: on")
	assert.Contains(t, content, "1 repo · 10% of 10")
}

func TestRenderCard_Shapes(t *testing.T) {
	repos := make([]repo.RepoConfig, 0, 3)
	for _, r := range overviewRepos() {
		repos = append(repos, repo.NewRepoConfig(r))
	}
	strip := func(lines []string) string { return strings.Join(uitest.StripAll(lines), "\n") }
	render := func(card overview.Card, focused bool, focus int) []string {
		lines, _ := renderCard(card, focused, focus, minCardWidth)
		return lines
	}

	bars := strip(render(overview.Languages(repos, overview.MaxRows), true, 0))
	assert.Equal(t, cardLines, strings.Count(bars, "\n")+1)
	assert.Regexp(t, `▸ LANGUAGES ─+ 1 distinct`, bars)
	assert.Regexp(t, `Go\s+━+╌+\s+2\s+67%`, bars)
	assert.Regexp(t, `none\s+━+╌+\s+1\s+33%`, bars)

	histogram := strip(render(overview.Alerts(repos), false, -1))
	assert.Regexp(t, `  ALERTS ─+ 2 open alerts`, histogram)
	assert.Regexp(t, `0\s+1\s+2-3\s+4-7\s+8-15\s+16\+`, histogram)
	assert.Regexp(t, `2\s+0\s+1\s+0\s+0\s+0`, histogram)
	assert.Contains(t, histogram, "repos by number of open alerts")

	rules := strip(render(overview.BranchRules(repos), true, 0))
	assert.Regexp(t, `▸ BRANCH RULES ─+ 1 of 3 with any`, rules)
	assert.Regexp(t, `pull request\s+━+╌+\s+1\s+33%`, rules)
	assert.Regexp(t, `1\+ approvals\s+━+╌+\s+1\s+33%`, rules)

	sources := strip(render(overview.RuleSources(repos), false, -1))
	assert.Regexp(t, `repo\s+━+╌+\s+1\s+33%`, sources)
	assert.Regexp(t, `none\s+━+╌+\s+2\s+67%`, sources)

	stacked := strip(render(overview.RepoMix(repos), true, 0))
	assert.Contains(t, stacked, "━━━━")
	assert.Contains(t, stacked, "● public 1 ● private 1 ● archived 1")

	for _, line := range render(overview.OtherSettings(repos, overview.MaxRows), false, -1) {
		assert.Equal(t, minCardWidth, len([]rune(uitest.StripAll([]string{line})[0])), "every line is the card width")
	}
}

func TestRenderCard_FocusShowsWithoutColour(t *testing.T) {
	repos := []repo.RepoConfig{repo.NewRepoConfig(overviewRepos()[0])}
	lines, _ := renderCard(overview.Languages(repos, overview.MaxRows), true, 0, minCardWidth)

	assert.Contains(t, lines[1], "\x1b[1;7", "the focused label is bold and reversed")
}

func TestRenderCard_StackedLastSegmentOnlyWhenNonEmpty(t *testing.T) {
	card := overview.Card{Title: "Branch rules", Kind: overview.Stacked, Total: 3, Slices: []overview.Slice{
		{Label: "one", Count: 2, InBar: true, Tone: overview.Good},
		{Label: "two", Count: 1, InBar: true, Tone: overview.Warn},
		{Label: "none", Count: 0, InBar: true, Tone: overview.Bad},
	}}

	lines, _ := renderCard(card, false, -1, minCardWidth)

	bar := lines[1]
	assert.NotContains(t, bar, ui.BadStyle.Render("━"), "an empty segment gets no cells")
	assert.Equal(t, minCardWidth-1, strings.Count(uitest.StripAll([]string{bar})[0], "━"), "the bar still fills the width")
}

func TestRenderCard_DrawsEverySlice(t *testing.T) {
	repos := make([]repo.RepoConfig, 0, 3)
	for _, r := range overviewRepos() {
		repos = append(repos, repo.NewRepoConfig(r))
	}
	for _, card := range overview.Cards(repos) {
		_, drawn := renderCard(card, false, -1, minCardWidth)
		assert.Len(t, drawn, len(card.Slices), card.Title)
	}
}

func TestRenderCard_StackedLegendWrapDropsTheBar(t *testing.T) {
	// Three-digit counts wrap the legend; the bar gives way so the three
	// group lines still fit and every slice is drawn.
	card := overview.Card{Title: "Repo mix", Kind: overview.Stacked, Total: 400, Slices: []overview.Slice{
		{Label: "public", Count: 230, InBar: true},
		{Label: "private", Count: 130, InBar: true},
		{Label: "archived", Count: 40, InBar: true},
		{Label: "fork", Count: 6},
		{Label: "admin", Count: 400, Group: "your role"},
		{Label: "main", Count: 330, Group: "branch"},
	}}

	lines, drawn := renderCard(card, false, -1, minCardWidth)

	content := strings.Join(uitest.StripAll(lines), "\n")
	assert.NotContains(t, content, "━━━━", "the bar gave way")
	assert.Contains(t, content, "● public 230 ● private 130")
	assert.Contains(t, content, "● archived 40")
	assert.Regexp(t, `branch main 330`, content)
	assert.Equal(t, []int{0, 1, 2, 3, 4, 5}, drawn)
}

func TestRenderCard_ReportsCutGroupItems(t *testing.T) {
	// More roles than fit on the role line.
	card := overview.Card{Title: "Repo mix", Kind: overview.Stacked, Total: 10, Slices: []overview.Slice{
		{Label: "public", Count: 10, InBar: true},
		{Label: "admin", Count: 1, Group: "your role"},
		{Label: "maintain", Count: 2, Group: "your role"},
		{Label: "write", Count: 3, Group: "your role"},
		{Label: "triage", Count: 4, Group: "your role"},
		{Label: "read", Count: 5, Group: "your role"},
	}}

	_, drawn := renderCard(card, false, -1, minCardWidth)

	assert.Equal(t, []int{0, 1, 2}, drawn, "write, triage and read did not fit, so they are not drawn")
}

func TestNameList(t *testing.T) {
	names := []string{"alpha", "bravo", "charlie"}
	assert.Equal(t, "alpha, bravo, charlie", uitest.Plain(tea.NewView(nameList(names, 40))))
	assert.Equal(t, "alpha, bravo +1 more", uitest.Plain(tea.NewView(nameList(names, 20))))
	assert.Equal(t, "alpha +2 more", uitest.Plain(tea.NewView(nameList(names, 14))))
	assert.Equal(t, "alpha +2", uitest.Plain(tea.NewView(nameList(names, 9))), "the count is shortened before a name is dropped")
	assert.Equal(t, "3 repos", uitest.Plain(tea.NewView(nameList(names, 5))), "when not even one fits")
}

func TestPercent(t *testing.T) {
	assert.Equal(t, 38, percent(15, 40), "rounded, not truncated")
	assert.Equal(t, 48, percent(19, 40))
	assert.Equal(t, 0, percent(1, 0))
}

func TestModel_OverviewHelp_NamesTheFocusedCardsRows(t *testing.T) {
	m := newOverviewModel(120, 40)
	m.View()

	want := []string{"kind", "age", "language", "license", "setting", "count", "rule", "source", "style", "setting", "setting"}
	for i, noun := range want {
		help := m.Help().String()
		assert.Contains(t, help, "tab card  j/k "+noun+"  enter show repos", "card %d", i)
		assert.Contains(t, uitest.Plain(m.View()), "j/k "+noun+" · enter show repos", "the footer hint agrees")
		press(m, "tab")
	}
	var full []string
	for _, column := range m.Help().FullHelp() {
		for _, b := range column {
			full = append(full, b.Help().Desc)
		}
	}
	assert.Contains(t, full, "next kind", "back on Repo mix after a full round")
	assert.Contains(t, full, "show the kind's repos")
}
