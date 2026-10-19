package repo

import (
	"sort"
	"strings"
)

// BranchRulesGroup is the group the default branch's rules are shown in.
const BranchRulesGroup = "7⟭ Branch rules"

// BranchRule is one rule from an active ruleset that applies to a branch.
type BranchRule struct {
	Type       string `graphql:"type"`
	Parameters struct {
		PullRequest struct {
			RequiredApprovingReviewCount int `graphql:"requiredApprovingReviewCount"`
		} `graphql:"... on PullRequestParameters"`
	} `graphql:"parameters"`
	Ruleset struct {
		Name   string `graphql:"name"`
		Source struct {
			// Typename is Repository, Organization or Enterprise. The
			// enterprise's name needs a scope most tokens lack, so only the
			// kind of source is asked for.
			Typename string `graphql:"__typename"`
		} `graphql:"source"`
	} `graphql:"repositoryRuleset"`
}

// ClassicRule is the summary of a branch's classic protection rule that
// people without admin access can read. Pattern is empty when the branch
// has none; RequiredApprovingReviewCount is nil when pull requests are not
// required, and zero when they are but need no approvals.
type ClassicRule struct {
	Pattern                      string   `graphql:"pattern"`
	AllowsDeletions              bool     `graphql:"allowsDeletions"`
	AllowsForcePushes            bool     `graphql:"allowsForcePushes"`
	RequiredApprovingReviewCount *int     `graphql:"requiredApprovingReviewCount"`
	RequiredStatusCheckContexts  []string `graphql:"requiredStatusCheckContexts"`
	RequiresLinearHistory        bool     `graphql:"requiresLinearHistory"`
	RequiresSignatures           bool     `graphql:"requiresSignatures"`
}

// exists says whether the branch has a classic protection rule.
func (c ClassicRule) exists() bool { return c.Pattern != "" }

// branchRule is one entry of the Branch rules checklist: the ruleset rule
// type it reads and, for the ones classic protection can also set, how to
// read that.
type branchRule struct {
	name, ruleType, desc string
	classic              func(ClassicRule) bool
}

// branchRules is the checklist, in the order it is shown. Every entry is
// on when an active ruleset has the rule or, where classic protection has
// an equivalent, when that sets it: GitHub applies both, the stricter
// winning, so either is enough.
var branchRules = []branchRule{
	{"Pull Request Rule", "PULL_REQUEST", "Changes reach the default branch through a pull request.",
		func(c ClassicRule) bool { return c.RequiredApprovingReviewCount != nil }},
	{"Status Checks Rule", "REQUIRED_STATUS_CHECKS", "Named status checks pass before the default branch is updated.",
		func(c ClassicRule) bool { return len(c.RequiredStatusCheckContexts) > 0 }},
	{"Force Pushes Blocked", "NON_FAST_FORWARD", "Force pushes to the default branch are refused.",
		func(c ClassicRule) bool { return !c.AllowsForcePushes }},
	{"Deletion Blocked", "DELETION", "The default branch cannot be deleted.",
		func(c ClassicRule) bool { return !c.AllowsDeletions }},
	{"Signed Commits Rule", "REQUIRED_SIGNATURES", "Commits on the default branch have verified signatures.",
		func(c ClassicRule) bool { return c.RequiresSignatures }},
	{"Linear History Rule", "REQUIRED_LINEAR_HISTORY", "Merge commits cannot be pushed to the default branch.",
		func(c ClassicRule) bool { return c.RequiresLinearHistory }},
	{"Copilot Review Rule", "COPILOT_CODE_REVIEW", "Copilot reviews new pull requests automatically.", nil},
	{"Code Scanning Rule", "CODE_SCANNING", "Named code scanning tools report results before the default branch is updated.", nil},
	{"Deployments Rule", "REQUIRED_DEPLOYMENTS", "Named environments deploy successfully before the default branch is updated.", nil},
	{"Code Quality Rule", "CODE_QUALITY", "Code quality results at chosen severities block merging.", nil},
	{"Code Coverage Rule", "CODE_COVERAGE", "Pull requests meet a minimum line coverage.", nil},
	{"License Compliance Rule", "LICENSE_COMPLIANCE_SCANNING", "New or changed dependencies meet the organization's license policy.", nil},
	{"Secret Scanning Rule", "SECRET_SCANNING", "A ruleset rule on secret scanning applies to pushes. This is the rule, not the repository's own secret scanning setting.", nil},
}

// BranchRuleNames lists the Branch rules checklist's toggles in order.
func BranchRuleNames() []string {
	names := make([]string, len(branchRules))
	for i, r := range branchRules {
		names[i] = r.name
	}
	return names
}

// sourceNames are how each kind of rule source reads in Rule Sources, in
// the order they are listed.
var sourceNames = []struct{ typename, name string }{
	{"", "classic"}, {"Repository", "repository"}, {"Organization", "organization"}, {"Enterprise", "enterprise"},
}

// BranchRuleProperties derives the Branch rules group from the rules on the
// default branch: a toggle per checklist entry, the approvals required, a
// count of the other rule types, where the rules come from, the rulesets
// that contribute, and how many rulesets are in evaluate mode.
func BranchRuleProperties(r Repository) []RepoProperty {
	ref := r.DefaultBranchRef
	classic := ref.Classic
	types := map[string]bool{}
	sources := map[string]bool{}
	rulesets := map[string]bool{}
	approvals := 0
	if classic.exists() {
		sources[""] = true
		if n := classic.RequiredApprovingReviewCount; n != nil {
			approvals = max(approvals, *n)
		}
	}
	for _, rule := range ref.Rules.Nodes {
		types[rule.Type] = true
		sources[rule.Ruleset.Source.Typename] = true
		if rule.Ruleset.Name != "" {
			rulesets[rule.Ruleset.Name] = true
		}
		if rule.Type == "PULL_REQUEST" {
			approvals = max(approvals, rule.Parameters.PullRequest.RequiredApprovingReviewCount)
		}
	}

	property := func(name string, value any, typeStr, desc string) RepoProperty {
		return NewRepoProperty(name, BranchRulesGroup, value, typeStr, desc)
	}
	var properties []RepoProperty
	known := map[string]bool{}
	for _, rule := range branchRules {
		known[rule.ruleType] = true
		on := types[rule.ruleType] || (rule.classic != nil && classic.exists() && rule.classic(classic))
		desc := rule.desc + " From an active ruleset"
		if rule.classic != nil {
			desc += " or classic branch protection"
		}
		properties = append(properties, property(rule.name, on, "bool", desc+"."))
		if rule.ruleType == "PULL_REQUEST" {
			properties = append(properties, property("Required Approvals", approvals, "int",
				"The approving reviews a pull request to the default branch needs, the most any rule asks for."))
		}
	}

	other := 0
	for t := range types {
		if !known[t] {
			other++
		}
	}
	properties = append(properties, property("Other Rules", other, "int",
		"How many other kinds of ruleset rule apply to the default branch, such as commit message patterns or file size limits."))

	var from []string
	for _, s := range sourceNames {
		if sources[s.typename] {
			from = append(from, s.name)
		}
	}
	properties = append(properties, property("Rule Sources", strings.Join(from, " + "), "string",
		"Where the default branch's rules come from: classic branch protection, or rulesets on the repository, organization or enterprise."))

	names := make([]string, 0, len(rulesets))
	for name := range rulesets {
		names = append(names, name)
	}
	sort.Strings(names)
	properties = append(properties, property("Rulesets Applied", strings.Join(names, ", "), "string",
		"The active rulesets with rules on the default branch."))

	evaluating := 0
	for _, rs := range r.Rulesets.Nodes {
		if rs.Enforcement == "EVALUATE" {
			evaluating++
		}
	}
	properties = append(properties, property("Evaluate Rulesets", evaluating, "int",
		"Rulesets in evaluate mode, which report what they would block without enforcing it. They can target any branch."))
	return properties
}
