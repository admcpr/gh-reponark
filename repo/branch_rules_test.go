package repo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// rule builds a ruleset rule of the given type from a source.
func rule(ruleType, ruleset, source string) BranchRule {
	var r BranchRule
	r.Type = ruleType
	r.Ruleset.Name = ruleset
	r.Ruleset.Source.Typename = source
	return r
}

func intPtr(n int) *int { return &n }

// branchValues reads the Branch rules group of r by property name.
func branchValues(r Repository) map[string]any {
	values := map[string]any{}
	for _, p := range BranchRuleProperties(r) {
		values[p.Name] = p.Value
	}
	return values
}

func TestBranchRuleProperties_NoRules(t *testing.T) {
	values := branchValues(Repository{})

	for _, name := range BranchRuleNames() {
		assert.Equal(t, false, values[name], name)
	}
	assert.Equal(t, 0, values["Required Approvals"])
	assert.Equal(t, 0, values["Other Rules"])
	assert.Equal(t, "", values["Rule Sources"])
	assert.Equal(t, "", values["Rulesets Applied"])
	assert.Equal(t, 0, values["Evaluate Rulesets"])
}

func TestBranchRuleProperties_Rulesets(t *testing.T) {
	var r Repository
	pr := rule("PULL_REQUEST", "baseline", "Organization")
	pr.Parameters.PullRequest.RequiredApprovingReviewCount = 2
	r.DefaultBranchRef.Rules.Nodes = []BranchRule{
		pr,
		rule("NON_FAST_FORWARD", "baseline", "Organization"),
		rule("DELETION", "main", "Repository"),
		rule("COPILOT_CODE_REVIEW", "main", "Repository"),
		rule("SECRET_SCANNING", "main", "Repository"),
		rule("COMMIT_MESSAGE_PATTERN", "main", "Repository"),
		rule("MAX_FILE_SIZE", "main", "Repository"),
	}
	r.Rulesets.Nodes = append(r.Rulesets.Nodes,
		struct {
			Enforcement string `graphql:"enforcement"`
		}{"ACTIVE"},
		struct {
			Enforcement string `graphql:"enforcement"`
		}{"EVALUATE"})

	values := branchValues(r)

	assert.Equal(t, true, values["Pull Request Rule"])
	assert.Equal(t, 2, values["Required Approvals"])
	assert.Equal(t, true, values["Force Pushes Blocked"])
	assert.Equal(t, true, values["Deletion Blocked"])
	assert.Equal(t, true, values["Copilot Review Rule"])
	assert.Equal(t, true, values["Secret Scanning Rule"])
	assert.Equal(t, false, values["Status Checks Rule"])
	assert.Equal(t, 2, values["Other Rules"], "commit message pattern and max file size")
	assert.Equal(t, "repository + organization", values["Rule Sources"])
	assert.Equal(t, "baseline, main", values["Rulesets Applied"])
	assert.Equal(t, 1, values["Evaluate Rulesets"])
}

func TestBranchRuleProperties_Classic(t *testing.T) {
	var r Repository
	r.DefaultBranchRef.Classic = ClassicRule{
		Pattern:                      "main",
		AllowsDeletions:              true,
		RequiredApprovingReviewCount: intPtr(1),
		RequiredStatusCheckContexts:  []string{"build"},
		RequiresSignatures:           true,
	}

	values := branchValues(r)

	assert.Equal(t, true, values["Pull Request Rule"])
	assert.Equal(t, 1, values["Required Approvals"])
	assert.Equal(t, true, values["Status Checks Rule"])
	assert.Equal(t, true, values["Force Pushes Blocked"], "classic protection blocks force pushes unless it allows them")
	assert.Equal(t, false, values["Deletion Blocked"], "this rule allows deletion")
	assert.Equal(t, true, values["Signed Commits Rule"])
	assert.Equal(t, false, values["Linear History Rule"])
	assert.Equal(t, false, values["Code Scanning Rule"], "only rulesets have it")
	assert.Equal(t, "classic", values["Rule Sources"])
	assert.Equal(t, "", values["Rulesets Applied"])
}

func TestBranchRuleProperties_ClassicPullRequestWithoutApprovals(t *testing.T) {
	var r Repository
	r.DefaultBranchRef.Classic = ClassicRule{Pattern: "main", RequiredApprovingReviewCount: intPtr(0)}
	values := branchValues(r)
	assert.Equal(t, true, values["Pull Request Rule"], "a pull request is required even with no approvals")
	assert.Equal(t, 0, values["Required Approvals"])

	r.DefaultBranchRef.Classic.RequiredApprovingReviewCount = nil
	assert.Equal(t, false, branchValues(r)["Pull Request Rule"], "nil means pull requests are not required")
}

func TestBranchRuleProperties_BothSourcesTakeTheStricter(t *testing.T) {
	var r Repository
	r.DefaultBranchRef.Classic = ClassicRule{Pattern: "main", AllowsForcePushes: true, RequiredApprovingReviewCount: intPtr(3)}
	pr := rule("PULL_REQUEST", "main", "Repository")
	pr.Parameters.PullRequest.RequiredApprovingReviewCount = 1
	r.DefaultBranchRef.Rules.Nodes = []BranchRule{pr, rule("NON_FAST_FORWARD", "main", "Repository")}

	values := branchValues(r)

	assert.Equal(t, 3, values["Required Approvals"], "the most any rule asks for")
	assert.Equal(t, true, values["Force Pushes Blocked"], "the ruleset blocks what classic protection allows")
	assert.Equal(t, "classic + repository", values["Rule Sources"])
}

func TestBranchRuleProperties_InTheSchema(t *testing.T) {
	var group *Group
	for _, g := range Groups() {
		if g.Key == BranchRulesGroup {
			g := g
			group = &g
		}
	}
	require.NotNil(t, group)
	assert.Equal(t, "Branch rules", group.Title)
	assert.Len(t, group.Properties, len(BranchRuleNames())+5, "the toggles, approvals, other rules, sources, rulesets and evaluate count")
	assert.Equal(t, "Pull Request Rule", group.Properties[0].Name)
	assert.Equal(t, "Required Approvals", group.Properties[1].Name)
	assert.Equal(t, "int", group.Properties[1].Type)

	c := NewRepoConfig(Repository{Name: "x"})
	assert.Contains(t, c.Properties, "Secret Scanning Rule", "a repository config carries them too")
	for _, p := range Schema() {
		assert.NotContains(t, []string{"Pattern", "Allows Deletions", "Type"}, p.Name, "raw rule data is not a property")
		assert.NotEmpty(t, p.Group, p.Name)
	}
}
