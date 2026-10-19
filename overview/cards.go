package overview

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"gh-reponark/filter"
	"gh-reponark/repo"
	"gh-reponark/ui"
)

// RepoMix is what the repositories are: the split between public, private
// and archived as the bar, then how many of them are forks, templates or
// empty (a line with no label, since the words say what they count), the
// signed-in user's role on them and the default branch names.
func RepoMix(repos []repo.RepoConfig) Card {
	live := func(c repo.RepoConfig) bool { return !c.Bool("Is Archived") }
	counts := []count{
		{label: "public", tone: Good, inBar: true,
			match:   func(c repo.RepoConfig) bool { return live(c) && !c.Bool("Is Private") },
			filters: []filter.Filter{off("Is Private"), off("Is Archived")}},
		{label: "private", tone: Warn, inBar: true,
			match:   func(c repo.RepoConfig) bool { return live(c) && c.Bool("Is Private") },
			filters: []filter.Filter{on("Is Private"), off("Is Archived")}},
		{label: "archived", tone: Bad, inBar: true,
			match: func(c repo.RepoConfig) bool { return c.Bool("Is Archived") }, filters: []filter.Filter{on("Is Archived")}},
		{label: "fork", group: "", match: func(c repo.RepoConfig) bool { return c.Bool("Is Fork") }, filters: []filter.Filter{on("Is Fork")}},
		{label: "template", group: "", match: func(c repo.RepoConfig) bool { return c.Bool("Is Template") }, filters: []filter.Filter{on("Is Template")}},
		{label: "empty", group: "", match: func(c repo.RepoConfig) bool { return c.Bool("Is Empty") }, filters: []filter.Filter{on("Is Empty")}},
	}
	for _, v := range ValueCounts(repos, "Viewer Permission") {
		if v.Value == "" {
			continue
		}
		value := v.Value
		counts = append(counts, count{label: strings.ToLower(value), group: "your role",
			match:   func(c repo.RepoConfig) bool { return c.Text("Viewer Permission") == value },
			filters: []filter.Filter{filter.NewExactStringFilter("Viewer Permission", value)}})
	}
	for i, v := range ValueCounts(repos, "Default Branch") {
		if v.Value == "" || i >= 3 {
			break
		}
		value := v.Value
		counts = append(counts, count{label: value, group: "branch",
			match:   func(c repo.RepoConfig) bool { return c.Text("Default Branch") == value },
			filters: []filter.Filter{filter.NewExactStringFilter("Default Branch", value)}})
	}
	return Card{
		Title:    "Repo mix",
		Noun:     "kind",
		Headline: fmt.Sprintf("%d repos", len(repos)),
		Kind:     Stacked,
		Slices:   slices(repos, counts),
		Total:    len(repos),
	}
}

// LastPush buckets the repositories by how long ago they were pushed to.
func LastPush(repos []repo.RepoConfig) Card {
	const day = 24 * time.Hour
	today := now()
	// age is how long ago a repository was pushed to, and false when it
	// never was. A push a little ahead of the local clock counts as now.
	age := func(c repo.RepoConfig) (time.Duration, bool) {
		pushed := c.Time("Pushed At")
		if pushed.IsZero() {
			return 0, false
		}
		return max(0, today.Sub(pushed)), true
	}
	// within is the bucket of repositories pushed between from and to ago,
	// selected by a date range on Pushed At.
	within := func(label string, from, to time.Duration) count {
		var after, before time.Time
		if to > 0 {
			after = today.Add(-to)
		}
		if from > 0 {
			before = today.Add(-from)
		}
		return count{label: label,
			match: func(c repo.RepoConfig) bool {
				a, ok := age(c)
				return ok && a >= from && (to == 0 || a < to)
			},
			filters: []filter.Filter{filter.NewDateFilter("Pushed At", after, before)}}
	}
	counts := []count{
		within("<1w", 0, 7*day),
		within("<1m", 7*day, 30*day),
		within("<3m", 30*day, 90*day),
		within("<1y", 90*day, 365*day),
		within("1y+", 365*day, 0),
		{label: "never", match: func(c repo.RepoConfig) bool { _, ok := age(c); return !ok }},
	}
	card := Card{Title: "Last push", Noun: "age", Kind: Histogram, Slices: slices(repos, counts), Total: len(repos)}

	var ages []int
	for _, c := range repos {
		if a, ok := age(c); ok {
			ages = append(ages, int(a.Hours()/24))
		}
	}
	sort.Ints(ages)
	if len(ages) > 0 {
		card.Note = "median push " + ui.AgoDays(ages[len(ages)/2])
	}
	card.Headline = fmt.Sprintf("%d over a year ago", card.Slices[4].Count)
	return card
}

// Languages is the share of each primary language in at most rows rows: the
// most common in their own rows, the rest as "other", and those with none.
func Languages(repos []repo.RepoConfig, rows int) Card {
	card := topValues("Languages", repos, "Primary Language", rows, "none")
	card.Noun = "language"
	distinct := 0
	for _, v := range ValueCounts(repos, "Primary Language") {
		if v.Value != "" {
			distinct++
		}
	}
	card.Headline = fmt.Sprintf("%d distinct", distinct)
	return card
}

// Licenses is the same for licenses.
func Licenses(repos []repo.RepoConfig, rows int) Card {
	card := topValues("Licenses", repos, "License Name", rows, "none")
	card.Noun = "license"
	for i := range card.Slices {
		// "MIT License" and "Apache License 2.0" read as well without the word.
		card.Slices[i].Label = strings.ReplaceAll(card.Slices[i].Label, " License", "")
	}
	licensed := 0
	for _, c := range repos {
		if c.Text("License Name") != "" {
			licensed++
		}
	}
	card.Headline = fmt.Sprintf("%d of %d licensed", licensed, len(repos))
	return card
}

// topValues counts a text property's values in at most rows rows: the most
// common values in their own rows, then "other" for the rest and a row for
// the repositories with it unset, when there are any of either. The unset
// row is always kept, so it is never squeezed off the card. Named values
// are selected by an exact text filter; the other two are selected by name.
func topValues(title string, repos []repo.RepoConfig, property string, rows int, unset string) Card {
	card := Card{Title: title, Kind: Bars, Total: len(repos)}
	other := Slice{Label: "other"}
	none := Slice{Label: unset}
	var named []ValueCount
	for _, v := range ValueCounts(repos, property) {
		if v.Value == "" {
			none.Count, none.Names = v.Count, v.Names
		} else {
			named = append(named, v)
		}
	}
	room := rows
	if none.Count > 0 {
		room--
	}
	if len(named) > room {
		// One row goes to "other", and it takes everything that did not fit.
		room = max(0, room-1)
		for _, v := range named[room:] {
			other.Count += v.Count
			other.Names = append(other.Names, v.Names...)
		}
		named = named[:room]
	}
	for _, v := range named {
		card.Slices = append(card.Slices, Slice{Label: v.Value, Count: v.Count, Names: v.Names,
			Filters: []filter.Filter{filter.NewExactStringFilter(property, v.Value)}})
	}
	if other.Count > 0 {
		card.Slices = append(card.Slices, other)
	}
	if none.Count > 0 {
		card.Slices = append(card.Slices, none)
	}
	return card
}

// Security is the share of repositories with each security setting on.
// "protected" means at least one branch protection rule or ruleset.
func Security(repos []repo.RepoConfig) Card {
	protected := func(c repo.RepoConfig) bool {
		return c.Int("Branch Protection Rules") > 0 || c.Int("Rulesets") > 0
	}
	return Card{
		Title:    "Security",
		Noun:     "setting",
		Headline: "settings on",
		Kind:     Bars,
		Total:    len(repos),
		Slices: slices(repos, []count{
			{label: "protected", tone: Good, match: protected},
			toggle("policy", "Is Security Policy Enabled"),
			toggle("alerts on", "Has Vulnerability Alerts Enabled"),
			toggle("signoff", "Web Commit Signoff Required"),
			toggle("forkable", "Forking Allowed"),
		}),
	}
}

// toggle counts the repositories with a bool property on.
func toggle(label, property string) count {
	return count{label: label, tone: Good,
		match:   func(c repo.RepoConfig) bool { return c.Bool(property) },
		filters: []filter.Filter{on(property)}}
}

// Alerts buckets the repositories by how many open vulnerability alerts
// they have.
func Alerts(repos []repo.RepoConfig) Card {
	const property = "Vulnerability Alerts"
	between := func(label string, lo, hi int) count {
		return count{label: label,
			match:   func(c repo.RepoConfig) bool { n := c.Int(property); return n >= lo && n <= hi },
			filters: []filter.Filter{filter.NewIntFilter(property, lo, hi)}}
	}
	card := Card{Title: "Alerts", Noun: "count", Kind: Histogram, Total: len(repos), Slices: slices(repos, []count{
		between("0", 0, 0),
		between("1", 1, 1),
		between("2-3", 2, 3),
		between("4-7", 4, 7),
		between("8-15", 8, 15),
		between("16+", 16, filter.NoMax),
	})}

	total, _, _ := Sum(repos, property)
	card.Headline = fmt.Sprintf("%d open %s", total, ui.Plural(total, "alert", "alerts"))
	card.Note = "repos by number of open alerts"
	return card
}

// branchRuleLabels are the short labels of the Branch rules checklist's
// toggles on the card, keyed by property name.
var branchRuleLabels = map[string]string{
	"Pull Request Rule": "pull request", "Status Checks Rule": "status checks",
	"Force Pushes Blocked": "no force push", "Deletion Blocked": "no deletion",
	"Signed Commits Rule": "signed commits", "Linear History Rule": "linear history",
	"Copilot Review Rule": "copilot review", "Code Scanning Rule": "code scanning",
	"Deployments Rule": "deployments", "Code Quality Rule": "code quality",
	"Code Coverage Rule": "code coverage", "License Compliance Rule": "license policy",
	"Secret Scanning Rule": "secret scan",
}

// BranchRules is the share of repositories with each rule on the default
// branch, the most common first, leaving out rules none of them have. When
// more rules are in use than fit, the last row counts the repositories with
// any of the rest.
func BranchRules(repos []repo.RepoConfig) Card {
	var counts []count
	for _, name := range repo.BranchRuleNames() {
		counts = append(counts, toggle(branchRuleLabels[name], name))
		if name == "Pull Request Rule" {
			counts = append(counts, count{label: "1+ approvals", tone: Good,
				match:   func(c repo.RepoConfig) bool { return c.Int("Required Approvals") > 0 },
				filters: []filter.Filter{filter.NewIntFilter("Required Approvals", 1, filter.NoMax)}})
		}
	}
	counts = append(counts, count{label: "other rules", tone: Good,
		match:   func(c repo.RepoConfig) bool { return c.Int("Other Rules") > 0 },
		filters: []filter.Filter{filter.NewIntFilter("Other Rules", 1, filter.NoMax)}})

	var used []Slice
	for _, s := range slices(repos, counts) {
		if s.Count > 0 {
			used = append(used, s)
		}
	}
	sort.SliceStable(used, func(i, j int) bool { return used[i].Count > used[j].Count })
	if len(used) > MaxRows {
		rest := used[MaxRows-1:]
		more := Slice{Label: fmt.Sprintf("%d more rules", len(rest)), Tone: Good}
		in := map[string]bool{}
		for _, s := range rest {
			for _, name := range s.Names {
				in[name] = true
			}
		}
		for _, c := range repos {
			if in[c.Name] {
				more.Names = append(more.Names, c.Name)
			}
		}
		more.Count = len(more.Names)
		used = append(used[:MaxRows-1], more)
	}

	withAny := 0
	for _, c := range repos {
		if c.Text("Rule Sources") != "" {
			withAny++
		}
	}
	card := Card{Title: "Branch rules", Noun: "rule", Kind: Bars, Total: len(repos), Slices: used,
		Headline: fmt.Sprintf("%d of %d with any", withAny, len(repos))}
	if len(used) == 0 {
		card.Note = "no rules on any default branch"
	}
	return card
}

// RuleSources is where the default branches' rules come from: classic
// branch protection, rulesets on the repository or its organization, or a
// mix, with the repositories that have none.
func RuleSources(repos []repo.RepoConfig) Card {
	card := topValues("Rule sources", repos, "Rule Sources", MaxRows, "none")
	card.Noun = "source"
	for i := range card.Slices {
		// The card's labels are short; the property and its filter keep the
		// full words.
		label := strings.NewReplacer("repository", "repo", "organization", "org", " + ", "+").Replace(card.Slices[i].Label)
		card.Slices[i].Label = label
	}
	both := 0
	for _, c := range repos {
		if strings.HasPrefix(c.Text("Rule Sources"), "classic + ") {
			both++
		}
	}
	card.Headline = fmt.Sprintf("%d with both kinds", both)
	return card
}

// MergeHabits is which combinations of merge strategies the repositories
// allow, most common first. Every repository allows exactly one
// combination, so the rows share out the whole set; the rest fold into
// "other".
func MergeHabits(repos []repo.RepoConfig) Card {
	strategies := []struct{ label, property string }{
		{"merge", "Merge Commit Allowed"}, {"squash", "Squash Merge Allowed"}, {"rebase", "Rebase Merge Allowed"},
	}
	combo := func(c repo.RepoConfig) string {
		var parts []string
		for _, s := range strategies {
			if c.Bool(s.property) {
				parts = append(parts, s.label)
			}
		}
		if len(parts) == 3 {
			return "all three"
		}
		if len(parts) == 0 {
			return "none"
		}
		return strings.Join(parts, "+")
	}
	byCombo := map[string]*Slice{}
	var order []string
	for _, c := range repos {
		label := combo(c)
		s, ok := byCombo[label]
		if !ok {
			var filters []filter.Filter
			for _, st := range strategies {
				filters = append(filters, filter.NewBoolFilter(st.property, c.Bool(st.property)))
			}
			s = &Slice{Label: label, Filters: filters}
			byCombo[label] = s
			order = append(order, label)
		}
		s.Count++
		s.Names = append(s.Names, c.Name)
	}
	sort.SliceStable(order, func(i, j int) bool { return byCombo[order[i]].Count > byCombo[order[j]].Count })
	card := Card{Title: "Merge styles", Noun: "style", Kind: Bars, Total: len(repos)}
	card.Headline = fmt.Sprintf("%d %s", len(order), ui.Plural(len(order), "combination", "combinations"))
	shown := order
	if len(order) > MaxRows {
		shown = order[:MaxRows-1]
	}
	for _, label := range shown {
		card.Slices = append(card.Slices, *byCombo[label])
	}
	if len(shown) < len(order) {
		other := Slice{Label: "other"}
		for _, label := range order[len(shown):] {
			other.Count += byCombo[label].Count
			other.Names = append(other.Names, byCombo[label].Names...)
		}
		card.Slices = append(card.Slices, other)
	}
	return card
}

// Features is the share of repositories with each feature or option on,
// the most evenly split first.
func Features(repos []repo.RepoConfig) Card {
	var counts []count
	for _, f := range []struct{ label, property string }{
		{"issues", "Has Issues Enabled"},
		{"projects", "Has Projects Enabled"},
		{"wiki", "Has Wiki Enabled"},
		{"discussions", "Has Discussions Enabled"},
		{"blank issues", "Is Blank Issues Enabled"},
		{"auto-merge", "Auto Merge Allowed"},
		{"update branch", "Allow Update Branch"},
		{"delete branch", "Delete Branch On Merge"},
	} {
		counts = append(counts, toggle(f.label, f.property))
	}
	all := slices(repos, counts)
	sort.SliceStable(all, func(i, j int) bool {
		return math.Abs(float64(all[i].Count)/float64(max(1, len(repos)))-0.5) <
			math.Abs(float64(all[j].Count)/float64(max(1, len(repos)))-0.5)
	})
	return Card{Title: "Features", Noun: "setting", Headline: "most evenly split first", Kind: Bars, Total: len(repos), Slices: all[:min(MaxRows, len(all))]}
}

// sharedShare is how many of the repositories must share a setting's value
// for the rest to be listed as the ones set differently.
const sharedShare = 0.85

// shownElsewhere are the settings other cards already chart, left out of
// Other settings so nothing is shown twice.
var shownElsewhere = map[string]bool{
	"Is Security Policy Enabled": true, "Has Vulnerability Alerts Enabled": true,
	"Web Commit Signoff Required": true, "Forking Allowed": true,
	"Merge Commit Allowed": true, "Squash Merge Allowed": true, "Rebase Merge Allowed": true,
	"Has Issues Enabled": true, "Has Projects Enabled": true, "Has Wiki Enabled": true,
	"Has Discussions Enabled": true, "Is Blank Issues Enabled": true, "Auto Merge Allowed": true,
	"Allow Update Branch": true, "Delete Branch On Merge": true, "Default Branch": true,
}

// settingGroups are the property groups that hold settings: what a
// repository is (Overview, Status, Metrics) and what the viewer may do
// (Permissions) are facts, not settings.
var settingGroups = map[string]bool{"4⟭ Features": true, "5⟭ Merge": true, "6⟭ Security": true}

// OtherSettings lists the settings no other card charts whose value most
// repositories share, naming the value and the repositories set
// differently, the fewest exceptions first. A slice selects the exceptions.
func OtherSettings(repos []repo.RepoConfig, rows int) Card {
	card := Card{Title: "Other settings", Noun: "setting", Kind: Bars, Total: len(repos)}
	var found []Slice
	for _, p := range repo.Schema() {
		if !settingGroups[p.Group] || shownElsewhere[p.Name] || len(repos) < 2 {
			continue
		}
		s := Summarize(repos, p)
		var differ Slice
		switch p.Type {
		case "bool":
			side, names := s.Minority()
			if len(names) == 0 {
				continue
			}
			differ = Slice{Names: names, Minority: side, Majority: map[string]string{"on": "off", "off": "on"}[side],
				Filters: []filter.Filter{filter.NewBoolFilter(p.Name, side == "on")}}
		case "string":
			top, ok := s.Top()
			if !ok || s.AllDistinct() || top.Count == len(repos) {
				continue
			}
			differ = Slice{Majority: top.Value, Minority: "other"}
			for _, c := range repos {
				if c.Text(p.Name) != top.Value {
					differ.Names = append(differ.Names, c.Name)
				}
			}
		default:
			continue
		}
		if float64(len(repos)-len(differ.Names)) < sharedShare*float64(len(repos)) {
			continue
		}
		differ.Label = repo.ShortName(p.Name)
		differ.Count = len(repos) - len(differ.Names)
		found = append(found, differ)
	}
	sort.SliceStable(found, func(i, j int) bool { return found[i].Count > found[j].Count })
	card.Headline = fmt.Sprintf("%d shared by %d%%+", len(found), int(sharedShare*100))
	card.Slices = found[:min(rows, len(found))]
	if len(found) == 0 {
		card.Note = "no other setting is shared this widely"
	}
	return card
}
