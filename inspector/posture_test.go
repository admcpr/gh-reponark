package inspector

import (
	"gh-reponark/repo"
	"gh-reponark/ui/uitest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func tidyRepository() repo.Repository {
	r := repo.Repository{
		Name:                          "tidy",
		Description:                   "Does everything right",
		PushedAt:                      time.Now().Add(-24 * time.Hour),
		IsSecurityPolicyEnabled:       true,
		HasVulnerabilityAlertsEnabled: true,
		DeleteBranchOnMerge:           true,
	}
	r.LicenseInfo.Name = "MIT License"
	r.BranchProtectionRuleCount.TotalCount = 1
	return r
}

func TestPosture(t *testing.T) {
	chips := Posture(repo.NewRepoConfig(tidyRepository()))
	labels := make([]string, len(chips))
	for i, chip := range chips {
		labels[i] = chip.Label
		assert.Equal(t, PosturePass, chip.State, chip.Label)
	}
	assert.Equal(t, []string{"Protected", "Policy", "0 alerts", "Licensed", "Active", "Cleanup"}, labels)

	rulesets := tidyRepository()
	rulesets.BranchProtectionRuleCount.TotalCount = 0
	rulesets.Rulesets.TotalCount = 2
	assert.Equal(t, PostureChip{"Protected", PosturePass}, Posture(repo.NewRepoConfig(rulesets))[0], "rulesets count as protection")

	bad := repo.Repository{IsArchived: true}
	bad.VulnerabilityAlerts.TotalCount = 4
	chips = Posture(repo.NewRepoConfig(bad))
	assert.Equal(t, PostureChip{"Unprotected", PostureFail}, chips[0])
	assert.Equal(t, PostureChip{"No policy", PostureFail}, chips[1])
	assert.Equal(t, PostureChip{"4 alerts", PostureFail}, chips[2])
	assert.Equal(t, PostureChip{"No license", PostureWarn}, chips[3])
	assert.Equal(t, PostureChip{"Archived", PostureFail}, chips[4])
	assert.Equal(t, PostureChip{"No cleanup", PostureWarn}, chips[5])

	quiet := repo.Repository{PushedAt: time.Now().Add(-400 * 24 * time.Hour)}
	chips = Posture(repo.NewRepoConfig(quiet))
	assert.Equal(t, PostureChip{"Alerts off", PostureWarn}, chips[2], "no alerts means little when they are disabled")
	assert.Equal(t, PostureChip{"Stale", PostureWarn}, chips[4])
}

func TestPostureRows(t *testing.T) {
	chips := []PostureChip{{"Protected", PosturePass}, {"No policy", PostureFail}, {"0 alerts", PosturePass}, {"Stale", PostureWarn}}

	assert.Equal(t, []string{"▐Protected▌ ▐No policy▌ ▐0 alerts▌ ▐Stale▌", ""}, uitest.StripAll(PostureRows(chips, 80, 2)), "every row is returned, even empty")
	assert.Equal(t, []string{"▐Protected▌ ▐No policy▌", "▐0 alerts▌ ▐Stale▌"}, uitest.StripAll(PostureRows(chips, 24, 2)))
	assert.Equal(t, []string{"▐Protected▌", "▐No policy▌"}, uitest.StripAll(PostureRows(chips, 12, 2)), "chips that do not fit are dropped whole")
	assert.Equal(t, []string{"▐Protected▌ ▐No policy▌"}, uitest.StripAll(PostureRows(chips, 30, 1)))
	assert.Empty(t, PostureRows(chips, 30, 0))

	pass := PostureRows([]PostureChip{{"Protected", PosturePass}}, 80, 1)
	fail := PostureRows([]PostureChip{{"Protected", PostureFail}}, 80, 1)
	warn := PostureRows([]PostureChip{{"Protected", PostureWarn}}, 80, 1)
	assert.NotEqual(t, pass[0], fail[0], "passing and failing chips are coloured differently")
	assert.NotEqual(t, warn[0], fail[0], "warnings and failures are coloured differently")
}
