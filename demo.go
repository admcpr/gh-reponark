package main

import (
	"fmt"
	"time"

	"gh-reponark/github"
	"gh-reponark/github/githubtest"
	"gh-reponark/repo"
)

// demoService is an in-memory GitHub with a user "demo" and a varied set of
// repositories, so the UI can be run and looked at without a token. It is
// used when the REPONARK_DEMO environment variable is set.
func demoService() *githubtest.Fake {
	return &githubtest.Fake{
		User: github.User{
			Login: "demo",
			Url:   "https://github.com/demo",
			Organizations: []github.Organization{
				{Login: "acme-robotics", Url: "https://github.com/acme-robotics"},
				{Login: "nightshade-labs", Url: "https://github.com/nightshade-labs"},
				{Login: "open-tooling", Url: "https://github.com/open-tooling"},
			},
		},
		Repositories: demoRepositories(),
		PageSize:     15,
	}
}

// demoRepositories builds about forty repositories that differ in every way
// the browser can show: visibility, archive state, forks and templates,
// language, popularity, how recently they were pushed and their settings.
func demoRepositories() []repo.Repository {
	type seed struct {
		name, description, language, license string
		stars, daysAgo, alerts               int
	}
	seeds := []seed{
		{"aurora", "Realtime dashboard for fleet telemetry", "TypeScript", "MIT License", 12040, 0, 0},
		{"beacon", "Service discovery with health probes", "Go", "Apache License 2.0", 3420, 1, 0},
		{"cobalt-cli", "Command line for the Cobalt platform", "Go", "MIT License", 870, 3, 2},
		{"dune", "Static site generator with zero config", "Rust", "MIT License", 5611, 6, 0},
		{"ember-forms", "Declarative forms for Ember apps", "JavaScript", "MIT License", 214, 40, 1},
		{"fjord", "Columnar storage engine", "C++", "Apache License 2.0", 2980, 12, 0},
		{"glacier-sync", "Backs up buckets to cold storage", "Python", "BSD 3-Clause", 56, 200, 4},
		{"harbor-ui", "Web console for container registries", "TypeScript", "Apache License 2.0", 1450, 2, 0},
		{"ion", "Tiny reactive state container", "JavaScript", "MIT License", 7820, 9, 0},
		{"juniper", "GraphQL server toolkit", "Rust", "MIT License", 4310, 5, 0},
		{"kestrel", "Message broker with at-least-once delivery", "Java", "Apache License 2.0", 990, 420, 3},
		{"lumen-docs", "Documentation theme and tooling", "CSS", "MIT License", 133, 15, 0},
		{"mariner", "Kubernetes operator for sea charts", "Go", "Apache License 2.0", 612, 30, 0},
		{"nimbus", "Weather data pipelines", "Python", "MIT License", 1205, 7, 1},
		{"obsidian-theme", "", "CSS", "", 48, 800, 0},
		{"pelican", "Email delivery microservice", "Ruby", "MIT License", 322, 95, 2},
		{"quartz", "Scheduler and cron replacement", "Java", "Apache License 2.0", 2210, 365, 0},
		{"rook", "Chess engine written for fun", "C", "GPL-3.0", 1788, 1100, 0},
		{"saffron", "Design tokens and component library", "TypeScript", "MIT License", 905, 4, 0},
		{"tundra", "Infrastructure modules for cold regions", "HCL", "MPL-2.0", 77, 60, 0},
		{"umbra", "Shadow traffic replay for APIs", "Go", "Apache License 2.0", 445, 18, 0},
		{"vortex", "Stream processing DSL", "Scala", "Apache License 2.0", 1360, 540, 5},
		{"willow", "Lightweight ORM for SQLite", "Python", "MIT License", 2890, 2, 0},
		{"xenon-bench", "Benchmark harness and reports", "Rust", "MIT License", 160, 25, 0},
		{"yarrow", "Plant identification model", "Jupyter Notebook", "", 39, 300, 0},
		{"zephyr-router", "HTTP router with zero allocations", "Go", "MIT License", 6130, 1, 0},
		{"template-go-service", "Starter for new Go services", "Go", "MIT License", 310, 10, 0},
		{"template-web-app", "Starter for Vite and React apps", "TypeScript", "MIT License", 528, 14, 0},
		{"legacy-billing", "Old billing system, archived", "PHP", "", 12, 1460, 7},
		{"legacy-portal", "Customer portal, retired 2023", "Java", "", 8, 1200, 3},
		{"dotfiles", "Shell and editor configuration", "Shell", "", 91, 0, 0},
		{"advent-of-code", "Puzzle solutions by year", "Kotlin", "", 23, 280, 0},
		{"talks", "Slides and demos from conference talks", "", "CC-BY-4.0", 67, 150, 0},
		{"rfcs", "Design proposals and discussion", "", "", 140, 21, 0},
		{"sandbox", "Scratch space, nothing to see", "", "", 0, 700, 0},
		{"homebrew-tap", "Formulae for the CLI tools", "Ruby", "BSD 2-Clause", 19, 11, 0},
		{"infra-live", "Live infrastructure, handle with care", "HCL", "", 3, 1, 0},
		{"ml-experiments", "Notebooks and training runs", "Python", "", 410, 33, 2},
		{"mobile-sdk", "SDK for iOS and Android", "Swift", "Apache License 2.0", 1990, 8, 1},
		{"android-sample", "Fork of the upstream sample app", "Kotlin", "Apache License 2.0", 15, 500, 0},
	}

	now := time.Now()
	repos := make([]repo.Repository, 0, len(seeds))
	for i, s := range seeds {
		pushed := now.AddDate(0, 0, -s.daysAgo)
		created := pushed.AddDate(-1-(i%3), -(i % 11), -(i % 27))
		isPrivate := i%3 == 1 || s.name == "infra-live" || s.name == "sandbox"
		isArchived := s.name == "legacy-billing" || s.name == "legacy-portal" || s.name == "obsidian-theme" || s.name == "rook"
		isFork := s.name == "android-sample" || s.name == "homebrew-tap" || i%9 == 4
		isTemplate := s.name == "template-go-service" || s.name == "template-web-app"

		r := repo.Repository{
			Id:              fmt.Sprintf("R_demo%03d", i+1),
			DatabaseID:      100000 + i*37,
			Name:            s.name,
			NameWithOwner:   "demo/" + s.name,
			Url:             "https://github.com/demo/" + s.name,
			ResourcePath:    "/demo/" + s.name,
			Description:     s.description,
			DescriptionHTML: "<div>" + s.description + "</div>",
			Visibility:      "PUBLIC",
			SshURL:          "git@github.com:demo/" + s.name + ".git",

			IsArchived: isArchived,
			IsFork:     isFork,
			IsPrivate:  isPrivate,
			IsTemplate: isTemplate,
			IsEmpty:    s.name == "sandbox",

			DiskUsage:      1200 + i*4321,
			ForkCount:      s.stars / 7,
			StargazerCount: s.stars,
			CreatedAt:      created,
			UpdatedAt:      pushed,
			PushedAt:       pushed,

			HasIssuesEnabled:              i%5 != 3,
			HasProjectsEnabled:            i%2 == 0,
			HasWikiEnabled:                i%4 == 0,
			HasDiscussionsEnabled:         s.stars > 1000,
			HasVulnerabilityAlertsEnabled: i%6 != 5,
			HasPullRequestsEnabled:        true,
			HasSponsorshipsEnabled:        s.stars > 5000,
			ForkingAllowed:                !isPrivate,
			IsBlankIssuesEnabled:          i%3 != 0,

			MergeCommitAllowed:       i%4 != 1,
			RebaseMergeAllowed:       i%3 == 0,
			SquashMergeAllowed:       true,
			AutoMergeAllowed:         i%2 == 1,
			DeleteBranchOnMerge:      i%5 != 4,
			AllowUpdateBranch:        i%2 == 0,
			WebCommitSignoffRequired: i%7 == 0,
			MergeCommitTitle:         "MERGE_MESSAGE",
			MergeCommitMessage:       "PR_TITLE",
			SquashMergeCommitTitle:   "PR_TITLE",
			SquashMergeCommitMessage: "COMMIT_MESSAGES",

			IsSecurityPolicyEnabled: i%3 == 0,
			ViewerPermission:        "ADMIN",
			ViewerCanAdminister:     true,
			ViewerCanCreateProjects: !isArchived,
			ViewerCanSubscribe:      true,
			ViewerCanUpdateTopics:   !isArchived,
			ViewerHasStarred:        i%4 == 2,
		}
		if isPrivate {
			r.Visibility = "PRIVATE"
		}
		if isArchived {
			r.ArchivedAt = pushed.AddDate(0, 0, 1)
		}
		if s.name == "infra-live" {
			r.WebCommitSignoffRequired = true
		}
		if s.name == "aurora" {
			r.HomepageURL = "https://aurora.example.com"
		}
		r.PrimaryLanguage.Name = s.language
		r.LicenseInfo.Name = s.license
		r.DefaultBranchRef.Name = "main"
		if i%6 == 5 {
			r.DefaultBranchRef.Name = "master"
		}
		if s.name == "dune" {
			r.DefaultBranchRef.Name = "trunk"
		}
		r.Issues.TotalCount = (s.stars / 40) + i%4
		r.OpenPullRequests.TotalCount = (s.stars / 300) + i%3
		r.Releases.TotalCount = i % 12
		r.Watchers.TotalCount = s.stars/9 + 1
		r.DirectForks.TotalCount = s.stars / 11
		if r.HasDiscussionsEnabled {
			r.Discussions.TotalCount = s.stars / 60
		}
		r.Labels.TotalCount = 9 + i%7
		r.Milestones.TotalCount = i % 4
		r.Deployments.TotalCount = (i * 13) % 50
		r.Environments.TotalCount = i % 3
		r.Packages.TotalCount = i % 2
		if i%10 == 7 {
			r.Submodules.TotalCount = 1
		}
		r.VulnerabilityAlerts.TotalCount = s.alerts
		r.BranchProtectionRuleCount.TotalCount = i % 3
		r.Rulesets.TotalCount = (i + 1) % 2
		if s.name == "legacy-billing" {
			r.LockReason = "BILLING"
		}
		repos = append(repos, r)
	}
	return repos
}
