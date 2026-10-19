package overview

import (
	"fmt"
	"testing"
	"time"

	"gh-reponark/filter"
	"gh-reponark/repo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var today = time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)

func init() { now = func() time.Time { return today } }

func testRepos() []repo.RepoConfig {
	days := func(n int) time.Time { return today.AddDate(0, 0, -n) }
	var repos []repo.RepoConfig
	for _, r := range []repo.Repository{
		{Name: "alpha", PushedAt: days(1), StargazerCount: 100, HasIssuesEnabled: true, IsSecurityPolicyEnabled: true, HasVulnerabilityAlertsEnabled: true, MergeCommitAllowed: true, SquashMergeAllowed: true, RebaseMergeAllowed: true, ViewerPermission: "ADMIN"},
		{Name: "bravo", IsPrivate: true, PushedAt: days(20), StargazerCount: 10, HasIssuesEnabled: true, HasVulnerabilityAlertsEnabled: true, SquashMergeAllowed: true, ViewerPermission: "ADMIN"},
		{Name: "charlie", IsArchived: true, IsFork: true, PushedAt: days(400), StargazerCount: 1, HasIssuesEnabled: false, SquashMergeAllowed: true, ViewerPermission: "WRITE"},
		{Name: "delta", IsTemplate: true, PushedAt: days(200), HasIssuesEnabled: true, HasVulnerabilityAlertsEnabled: true, SquashMergeAllowed: true, ViewerPermission: "ADMIN"},
	} {
		switch r.Name {
		case "alpha", "bravo":
			r.PrimaryLanguage.Name = "Go"
			r.LicenseInfo.Name = "MIT License"
		case "charlie":
			r.PrimaryLanguage.Name = "Rust"
		}
		r.DefaultBranchRef.Name = "main"
		if r.Name == "charlie" {
			r.DefaultBranchRef.Name = "master"
		}
		if r.Name == "alpha" {
			// A repository ruleset requiring a pull request and blocking
			// force pushes.
			r.Rulesets.TotalCount = 1
			pr := branchRule("PULL_REQUEST", "main", "Repository")
			pr.Parameters.PullRequest.RequiredApprovingReviewCount = 1
			r.DefaultBranchRef.Rules.Nodes = []repo.BranchRule{pr, branchRule("NON_FAST_FORWARD", "main", "Repository")}
			r.VulnerabilityAlerts.TotalCount = 3
		}
		if r.Name == "bravo" {
			// Classic protection with approvals and a status check, and an
			// organization ruleset blocking deletion.
			r.BranchProtectionRuleCount.TotalCount = 1
			r.Rulesets.TotalCount = 2
			two := 2
			r.DefaultBranchRef.Classic = repo.ClassicRule{Pattern: "main", RequiredApprovingReviewCount: &two, RequiredStatusCheckContexts: []string{"build"}}
			r.DefaultBranchRef.Rules.Nodes = []repo.BranchRule{branchRule("DELETION", "baseline", "Organization")}
			r.VulnerabilityAlerts.TotalCount = 1
		}
		repos = append(repos, repo.NewRepoConfig(r))
	}
	return repos
}

// branchRule builds a ruleset rule of the given type from a source.
func branchRule(ruleType, ruleset, source string) repo.BranchRule {
	var r repo.BranchRule
	r.Type = ruleType
	r.Ruleset.Name = ruleset
	r.Ruleset.Source.Typename = source
	return r
}

// counts maps a card's labels to their counts.
func counts(c Card) map[string]int {
	out := map[string]int{}
	for _, s := range c.Slices {
		out[s.Label] = s.Count
	}
	return out
}

// slice finds a card's slice by label.
func slice(t *testing.T, c Card, label string) Slice {
	t.Helper()
	for _, s := range c.Slices {
		if s.Label == label {
			return s
		}
	}
	t.Fatalf("no slice %q in %s", label, c.Title)
	return Slice{}
}

// selects checks that a slice's filters pick exactly its names.
func selects(t *testing.T, repos []repo.RepoConfig, s Slice) {
	t.Helper()
	require.NotEmpty(t, s.Filters, "slice %q has no filters", s.Label)
	fm := filter.FilterMap{}
	for _, f := range s.Filters {
		fm[f.Name()] = f
	}
	var names []string
	for _, c := range fm.FilterRepos(repos) {
		names = append(names, c.Name)
	}
	assert.Equal(t, s.Names, names, "filters for %q", s.Label)
}

func TestCards(t *testing.T) {
	cards := Cards(testRepos())
	require.Len(t, cards, 11)
	assert.Equal(t, "Repo mix", cards[0].Title)
	assert.Equal(t, "Branch rules", cards[6].Title)
	assert.Equal(t, "Rule sources", cards[7].Title)
	assert.Equal(t, "Other settings", cards[10].Title)
}

func TestRepoMix(t *testing.T) {
	repos := testRepos()
	c := RepoMix(repos)

	assert.Equal(t, "4 repos", c.Headline)
	assert.Equal(t, Stacked, c.Kind)
	assert.Equal(t, map[string]int{
		"public": 2, "private": 1, "archived": 1, "fork": 1, "template": 1, "empty": 0,
		"admin": 3, "write": 1, "main": 3, "master": 1,
	}, counts(c))
	assert.True(t, slice(t, c, "public").InBar)
	assert.Equal(t, "your role", slice(t, c, "admin").Group)
	assert.Empty(t, slice(t, c, "fork").Group, "fork, template and empty need no label")
	assert.Equal(t, "branch", slice(t, c, "master").Group)
	for _, label := range []string{"public", "private", "archived", "fork", "admin", "master"} {
		selects(t, repos, slice(t, c, label))
	}
	assert.Equal(t, 3, c.Largest())
}

func TestLastPush(t *testing.T) {
	repos := append(testRepos(), repo.NewRepoConfig(repo.Repository{Name: "echo"}))

	c := LastPush(repos)

	assert.Equal(t, Histogram, c.Kind)
	assert.Equal(t, map[string]int{"<1w": 1, "<1m": 1, "<3m": 0, "<1y": 1, "1y+": 1, "never": 1}, counts(c))
	assert.Equal(t, "1 over a year ago", c.Headline)
	assert.Equal(t, "median push 6mo ago", c.Note, "the upper median of 1, 20, 200 and 400 days")
	for _, label := range []string{"<1w", "<1m", "<1y", "1y+"} {
		selects(t, repos, slice(t, c, label))
	}
	assert.Nil(t, slice(t, c, "never").Filters, "selected by name")
}

func TestLastPush_PushAheadOfTheClockIsRecent(t *testing.T) {
	// GitHub's clock can be a little ahead of this one.
	repos := []repo.RepoConfig{repo.NewRepoConfig(repo.Repository{Name: "fresh", PushedAt: today.Add(time.Minute)})}

	c := LastPush(repos)

	assert.Equal(t, []string{"fresh"}, slice(t, c, "<1w").Names)
	assert.Zero(t, slice(t, c, "never").Count)
}

func TestLanguagesAndLicenses(t *testing.T) {
	repos := testRepos()
	c := Languages(repos, MaxRows)
	assert.Equal(t, "2 distinct", c.Headline)
	assert.Equal(t, map[string]int{"Go": 2, "Rust": 1, "none": 1}, counts(c), "everything fits, so there is no other row")
	assert.Equal(t, []string{"alpha", "bravo"}, c.Slices[0].Names)
	selects(t, repos, c.Slices[0])
	assert.Equal(t, "none", c.Slices[2].Label)
	assert.Equal(t, []string{"delta"}, c.Slices[2].Names)

	l := Licenses(repos, MaxRows)
	assert.Equal(t, "2 of 4 licensed", l.Headline)
	assert.Equal(t, map[string]int{"MIT": 2, "none": 2}, counts(l), "the word License is dropped")
	selects(t, repos, slice(t, l, "MIT"))
}

// languageRepos has one repository per language, plus one with none.
func languageRepos(languages ...string) []repo.RepoConfig {
	var repos []repo.RepoConfig
	for i, language := range append(languages, "") {
		r := repo.Repository{Name: fmt.Sprintf("r%d", i)}
		r.PrimaryLanguage.Name = language
		repos = append(repos, repo.NewRepoConfig(r))
	}
	return repos
}

func TestLanguages_KeepsTheUnsetRowWithinTheCard(t *testing.T) {
	// Six languages and an unset row need seven rows; the card has five.
	repos := languageRepos("C", "C++", "CSS", "Go", "Java", "JavaScript")

	c := Languages(repos, MaxRows)

	require.Len(t, c.Slices, MaxRows)
	assert.Equal(t, "other", c.Slices[3].Label)
	assert.Equal(t, 3, c.Slices[3].Count, "the languages that did not fit")
	assert.Equal(t, "none", c.Slices[4].Label, "the unset row is never squeezed out")

	// Exactly as many values as fit: no other row.
	c = Languages(languageRepos("Go", "Java", "Rust", "Zig"), MaxRows)
	assert.Equal(t, []string{"Go", "Java", "Rust", "Zig", "none"}, labels(c))
}

func TestLanguages_SelectsExactValues(t *testing.T) {
	// Every one of these is a substring of another.
	repos := languageRepos("C", "C++", "Java", "JavaScript")

	c := Languages(repos, MaxRows)

	for _, s := range c.Slices[:4] {
		assert.Equal(t, 1, s.Count)
		selects(t, repos, s)
	}
}

// labels lists a card's slice labels in order.
func labels(c Card) []string {
	var out []string
	for _, s := range c.Slices {
		out = append(out, s.Label)
	}
	return out
}

func TestCards_BarsFitTheCard(t *testing.T) {
	for _, c := range Cards(languageRepos("a", "b", "c", "d", "e", "f", "g", "h")) {
		if c.Kind == Bars {
			assert.LessOrEqual(t, len(c.Slices), MaxRows, c.Title)
		}
	}
}

func TestSecurity(t *testing.T) {
	repos := testRepos()
	c := Security(repos)

	assert.Equal(t, map[string]int{"protected": 2, "policy": 1, "alerts on": 3, "signoff": 0, "forkable": 0}, counts(c))
	assert.Equal(t, []string{"alpha", "bravo"}, slice(t, c, "protected").Names)
	assert.Nil(t, slice(t, c, "protected").Filters, "rules or rulesets cannot be one property filter")
	selects(t, repos, slice(t, c, "policy"))
}

func TestAlerts(t *testing.T) {
	repos := testRepos()
	c := Alerts(repos)

	assert.Equal(t, "4 open alerts", c.Headline)
	assert.Equal(t, map[string]int{"0": 2, "1": 1, "2-3": 1, "4-7": 0, "8-15": 0, "16+": 0}, counts(c))
	assert.Equal(t, "repos by number of open alerts", c.Note, "the columns say what they count")
	selects(t, repos, slice(t, c, "2-3"))
	selects(t, repos, slice(t, c, "0"))

	quiet := Alerts([]repo.RepoConfig{repo.NewRepoConfig(repo.Repository{Name: "x"})})
	assert.Equal(t, "0 open alerts", quiet.Headline)
}

func TestBranchRules(t *testing.T) {
	repos := testRepos()
	c := BranchRules(repos)

	assert.Equal(t, "2 of 4 with any", c.Headline)
	assert.Equal(t, Bars, c.Kind)
	assert.Equal(t, []string{"pull request", "1+ approvals", "no force push", "status checks", "no deletion"}, labels(c),
		"the most common first, in checklist order among equals; rules nobody has are left out")
	assert.Equal(t, []string{"alpha", "bravo"}, slice(t, c, "pull request").Names, "from a ruleset and from classic protection")
	assert.Equal(t, []string{"bravo"}, slice(t, c, "no deletion").Names)
	for _, s := range c.Slices {
		selects(t, repos, s)
	}
}

func TestBranchRules_FoldsTheRestIntoMore(t *testing.T) {
	// Seven rules, each on its own repository: four rows and one for the rest.
	var repos []repo.RepoConfig
	for i, ruleType := range []string{"PULL_REQUEST", "REQUIRED_STATUS_CHECKS", "NON_FAST_FORWARD", "DELETION", "CODE_SCANNING", "SECRET_SCANNING", "MAX_FILE_SIZE"} {
		r := repo.Repository{Name: fmt.Sprint(i)}
		r.DefaultBranchRef.Rules.Nodes = []repo.BranchRule{branchRule(ruleType, "main", "Repository")}
		repos = append(repos, repo.NewRepoConfig(r))
	}

	c := BranchRules(repos)

	require.Len(t, c.Slices, MaxRows)
	more := c.Slices[MaxRows-1]
	assert.Equal(t, "3 more rules", more.Label)
	assert.Equal(t, []string{"4", "5", "6"}, more.Names, "code scanning, secret scanning and the other rule")
	assert.Nil(t, more.Filters, "selected by name")
	assert.Equal(t, "7 of 7 with any", c.Headline)

	none := BranchRules([]repo.RepoConfig{repo.NewRepoConfig(repo.Repository{Name: "x"})})
	assert.Empty(t, none.Slices)
	assert.Equal(t, "no rules on any default branch", none.Note)
}

func TestRuleSources(t *testing.T) {
	repos := testRepos()
	c := RuleSources(repos)

	assert.Equal(t, []string{"classic+org", "repo", "none"}, labels(c), "short labels on the card")
	assert.Equal(t, "1 with both kinds", c.Headline)
	selects(t, repos, c.Slices[0])
	assert.Equal(t, `is "classic + organization"`, c.Slices[0].Filters[0].Condition(), "the filter keeps the full words")
}

func TestMergeHabits(t *testing.T) {
	repos := testRepos()
	c := MergeHabits(repos)

	require.Len(t, c.Slices, 2, "only the combinations, which share out every repo")
	assert.Equal(t, "Merge styles", c.Title)
	assert.Equal(t, "2 combinations", c.Headline)
	assert.Equal(t, "squash", c.Slices[0].Label, "most common first")
	assert.Equal(t, 3, c.Slices[0].Count)
	assert.Equal(t, "all three", c.Slices[1].Label)
	selects(t, repos, c.Slices[0])
	selects(t, repos, c.Slices[1])
}

func TestMergeHabits_FoldsTheRestIntoOther(t *testing.T) {
	var repos []repo.RepoConfig
	for i, allowed := range [][3]bool{
		{true, false, false}, {false, true, false}, {false, false, true},
		{true, true, false}, {true, false, true}, {false, true, true}, {true, true, true},
	} {
		repos = append(repos, repo.NewRepoConfig(repo.Repository{Name: fmt.Sprint(i),
			MergeCommitAllowed: allowed[0], SquashMergeAllowed: allowed[1], RebaseMergeAllowed: allowed[2]}))
	}

	c := MergeHabits(repos)

	require.Len(t, c.Slices, MaxRows)
	assert.Equal(t, "7 combinations", c.Headline)
	assert.Equal(t, "other", c.Slices[MaxRows-1].Label)
	assert.Equal(t, 3, c.Slices[MaxRows-1].Count)
}

func TestFeatures(t *testing.T) {
	c := Features(testRepos())

	require.Len(t, c.Slices, MaxRows)
	assert.Equal(t, "most evenly split first", c.Headline)
	assert.Equal(t, "issues", c.Slices[0].Label, "3 of 4 is the closest to an even split")
	assert.Equal(t, 3, c.Slices[0].Count)
	selects(t, testRepos(), c.Slices[0])
}

func TestOtherSettings(t *testing.T) {
	// Twenty repos agree on everything but one turns sponsorships on, one
	// has a different merge commit title, and one turns issues off.
	var repos []repo.RepoConfig
	for i := 0; i < 20; i++ {
		r := repo.Repository{Name: string(rune('a' + i)), HasIssuesEnabled: true, HasPullRequestsEnabled: true, MergeCommitTitle: "PR_TITLE"}
		r.DefaultBranchRef.Name = "main"
		switch i {
		case 3:
			r.HasSponsorshipsEnabled = true
		case 7:
			r.MergeCommitTitle = "MERGE_MESSAGE"
		case 9:
			r.HasIssuesEnabled = false
			r.DefaultBranchRef.Name = "master"
		}
		repos = append(repos, repo.NewRepoConfig(r))
	}

	c := OtherSettings(repos, MaxRows)

	assert.Equal(t, "2 shared by 85%+", c.Headline)
	assert.Equal(t, []string{"Sponsorships", "Merge Commit Title"}, labels(c),
		"issues and the default branch are on other cards, so are left out")
	sponsorships := slice(t, c, "Sponsorships")
	assert.Equal(t, 19, sponsorships.Count, "how many share the value")
	assert.Equal(t, "off", sponsorships.Majority)
	assert.Equal(t, "on", sponsorships.Minority)
	assert.Equal(t, []string{"d"}, sponsorships.Names, "the ones set differently")
	selects(t, repos, sponsorships)
	title := slice(t, c, "Merge Commit Title")
	assert.Equal(t, "PR_TITLE", title.Majority)
	assert.Equal(t, "other", title.Minority)
	assert.Equal(t, []string{"h"}, title.Names)
	assert.Nil(t, title.Filters, "a text minority is selected by name")
	assert.Empty(t, c.Note)

	alone := OtherSettings(repos[:1], MaxRows)
	assert.Equal(t, "0 shared by 85%+", alone.Headline)
	assert.Equal(t, "no other setting is shared this widely", alone.Note)
}

func TestValueCountsAndSum(t *testing.T) {
	values := ValueCounts(testRepos(), "Primary Language")
	assert.Equal(t, "Go", values[0].Value)
	assert.Equal(t, 2, values[0].Count)

	sum, top, largest := Sum(testRepos(), "Stargazer Count")
	assert.Equal(t, 111, sum)
	assert.Equal(t, "alpha", top)
	assert.Equal(t, 100, largest)
}

func schema(name string) repo.PropertySchema {
	for _, p := range repo.Schema() {
		if p.Name == name {
			return p
		}
	}
	panic("no property " + name)
}

func TestSummarize_Bool(t *testing.T) {
	s := Summarize(testRepos(), schema("Has Issues Enabled"))

	assert.Equal(t, 3, s.On)
	assert.Equal(t, []string{"charlie"}, s.OffNames)
	side, names := s.Minority()
	assert.Equal(t, "off", side)
	assert.Equal(t, []string{"charlie"}, names)

	side, names = Summarize(testRepos(), schema("Is Fork")).Minority()
	assert.Equal(t, "on", side)
	assert.Equal(t, []string{"charlie"}, names)
}

func TestSummarize_Text(t *testing.T) {
	s := Summarize(testRepos(), schema("License Name"))
	top, ok := s.Top()
	require.True(t, ok)
	assert.Equal(t, "MIT License", top.Value)
	assert.Equal(t, 2, top.Count)
	assert.Equal(t, 1, s.Distinct)
	assert.Equal(t, 2, s.Unset)
	assert.False(t, s.AllDistinct())
	assert.True(t, Summarize(testRepos(), schema("Name")).AllDistinct())
	_, ok = Summarize(testRepos(), schema("Homepage URL")).Top()
	assert.False(t, ok, "nothing is set")
}

func TestCards_NameTheirRows(t *testing.T) {
	nouns := map[string]string{}
	for _, c := range Cards(testRepos()) {
		nouns[c.Title] = c.Noun
	}
	assert.Equal(t, map[string]string{
		"Repo mix": "kind", "Last push": "age", "Languages": "language", "Licenses": "license",
		"Security": "setting", "Alerts": "count", "Branch rules": "rule", "Rule sources": "source",
		"Merge styles": "style", "Features": "setting", "Other settings": "setting",
	}, nouns)
}
