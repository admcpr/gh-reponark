package repo

import "time"

type Repository struct {
	// Overview
	Id              string `graphql:"id" group:"1⟭ Overview" desc:"The Node ID of the Repository."`
	DatabaseID      int    `graphql:"databaseId" group:"1⟭ Overview" desc:"Identifies the primary key from the database."`
	Name            string `graphql:"name" group:"1⟭ Overview" desc:"The name of the repository."`
	NameWithOwner   string `graphql:"nameWithOwner" group:"1⟭ Overview" desc:"The repository's name with owner (e.g., octocat/hello-world)."`
	Url             string `graphql:"url" group:"1⟭ Overview" desc:"The HTTP URL for this repository."`
	ResourcePath    string `graphql:"resourcePath" group:"1⟭ Overview" desc:"The repository's URL path."`
	Description     string `graphql:"description" group:"1⟭ Overview" desc:"The description of the repository."`
	DescriptionHTML string `graphql:"descriptionHTML" group:"1⟭ Overview" desc:"The description of the repository rendered to HTML."`
	HomepageURL     string `graphql:"homepageUrl" group:"1⟭ Overview" desc:"The repository's URL."`
	PrimaryLanguage struct {
		Name string `name:"Primary Language" group:"1⟭ Overview" desc:"The primary programming language of the repository."`
		// Color string
	} `graphql:"primaryLanguage"`
	LicenseInfo struct {
		Name string `name:"License Name" group:"1⟭ Overview" desc:"The license associated with the repository."`
		// Key  string
	} `graphql:"licenseInfo"`

	// Status
	IsArchived       bool `graphql:"isArchived" group:"2⟭ Status" desc:"Indicates if the repository is archived."`
	IsDisabled       bool `graphql:"isDisabled" group:"2⟭ Status" desc:"Indicates if the repository is disabled."`
	IsEmpty          bool `graphql:"isEmpty" group:"2⟭ Status" desc:"Indicates if the repository is empty."`
	IsFork           bool `graphql:"isFork" group:"2⟭ Status" desc:"Identifies if the repository is a fork."`
	IsInOrganization bool `graphql:"isInOrganization" group:"2⟭ Status" desc:"Indicates if the repository is part of an organization."`
	IsLocked         bool `graphql:"isLocked" group:"2⟭ Status" desc:"Indicates if the repository is locked."`
	IsMirror         bool `graphql:"isMirror" group:"2⟭ Status" desc:"Identifies if the repository is a mirror."`
	IsPrivate        bool `graphql:"isPrivate" group:"2⟭ Status" desc:"Identifies if the repository is private."`
	IsTemplate       bool `graphql:"isTemplate" group:"2⟭ Status" desc:"Indicates if the repository is a template repository."`

	// Metrics
	DiskUsage      int `graphql:"diskUsage" group:"3⟭ Metrics" desc:"The number of kilobytes this repository occupies on disk."`
	ForkCount      int `graphql:"forkCount" group:"3⟭ Metrics" desc:"Returns how many forks there are of this repository in the whole network."`
	StargazerCount int `graphql:"stargazerCount" group:"3⟭ Metrics" desc:"Returns a count of how many stargazers there are on this repository."`
	// WatcherCount   int `graphql:"watcherCount" group:"3⟭ Metrics" desc:"The number of watchers this repository has."`
	// OpenIssueCount       int       `graphql:"openIssueCount" group:"3⟭ Metrics" desc:"The number of open issues in this repository."`
	// OpenPullRequestCount int       `graphql:"openPullRequestCount" group:"3⟭ Metrics" desc:"The number of open pull requests in this repository."`
	CreatedAt time.Time `graphql:"createdAt" group:"3⟭ Metrics" desc:"Identifies the date and time when the object was created."`
	UpdatedAt time.Time `graphql:"updatedAt" group:"3⟭ Metrics" desc:"Identifies the date and time when the object was last updated."`
	PushedAt  time.Time `graphql:"pushedAt" group:"3⟭ Metrics" desc:"Identifies when the repository was last pushed to."`
	Issues    struct {
		TotalCount int `name:"Open Issues" group:"3⟭ Metrics" desc:"The number of open issues for this repository."`
	} `graphql:"issues"`
	Releases struct {
		TotalCount int `name:"Releases" group:"3⟭ Metrics" desc:"The number of releases for this repository."`
	} `graphql:"releases"`
	OpenPullRequests struct {
		TotalCount int `name:"Open Pull Requests" group:"3⟭ Metrics" desc:"The number of open pull requests for this repository."`
	} `graphql:"pullRequests(states: OPEN)"`

	// Repository Features
	HasIssuesEnabled              bool `graphql:"hasIssuesEnabled" group:"4⟭ Features" desc:"Indicates if the repository has issues feature enabled."`
	HasProjectsEnabled            bool `graphql:"hasProjectsEnabled" group:"4⟭ Features" desc:"Indicates if the repository has projects feature enabled."`
	HasWikiEnabled                bool `graphql:"hasWikiEnabled" group:"4⟭ Features" desc:"Indicates if the repository has wiki feature enabled."`
	HasDiscussionsEnabled         bool `graphql:"hasDiscussionsEnabled" group:"4⟭ Features" desc:"Indicates if the repository has discussions feature enabled."`
	HasVulnerabilityAlertsEnabled bool `graphql:"hasVulnerabilityAlertsEnabled" group:"4⟭ Features" desc:"Indicates if the repository has vulnerability alerts enabled."`

	// Merge Settings
	MergeCommitAllowed  bool `graphql:"mergeCommitAllowed" group:"5⟭ Merge" desc:"Whether merge commits are allowed on this repository."`
	RebaseMergeAllowed  bool `graphql:"rebaseMergeAllowed" group:"5⟭ Merge" desc:"Whether rebase-merging is allowed on this repository."`
	SquashMergeAllowed  bool `graphql:"squashMergeAllowed" group:"5⟭ Merge" desc:"Whether squash-merging is allowed on this repository."`
	AutoMergeAllowed    bool `graphql:"autoMergeAllowed" group:"5⟭ Merge" desc:"Whether auto-merge is allowed on this repository."`
	DeleteBranchOnMerge bool `graphql:"deleteBranchOnMerge" group:"5⟭ Merge" desc:"Whether to delete head branches when pull requests are merged."`
	DefaultBranchRef    struct {
		Name string `name:"Default Branch" group:"5⟭ Merge" desc:"The name of the default branch for this repository."`
		// BranchProtectionRule BranchProtectionRule `graphql:"branchProtectionRule"`
	} `graphql:"defaultBranchRef"`

	// Branch Protection
	// BranchProtectionRules BranchProtectionRuleConnection `graphql:"branchProtectionRules" group:"branch" desc:"A list of branch protection rules for this repository."`

	// Security Features
	IsSecurityPolicyEnabled bool   `graphql:"isSecurityPolicyEnabled" group:"6⟭ Security" desc:"Whether this repository has a security policy."`
	SecurityPolicyURL       string `graphql:"securityPolicyUrl" group:"6⟭ Security" desc:"The URL for the security policy of this repository."`
	VulnerabilityAlerts     struct {
		TotalCount int `name:"Vulnerability Alerts" group:"6⟭ Security" desc:"The number of vulnerability alerts for this repository."`
	} `graphql:"vulnerabilityAlerts"`

	// Permissions and Access
	ViewerPermission        string `graphql:"viewerPermission" group:"6⟭ Permissions" desc:"The current user's permission level on the repository (READ, WRITE, ADMIN)."`
	ViewerCanAdminister     bool   `graphql:"viewerCanAdminister" group:"6⟭ Permissions" desc:"Indicates if the current viewer can administer this repository."`
	ViewerCanCreateProjects bool   `graphql:"viewerCanCreateProjects" group:"6⟭ Permissions" desc:"Indicates if the current viewer can create projects in this repository."`
	// ViewerCanFork           bool   `graphql:"viewerCanFork" group:"viewer" desc:"Whether the viewer can fork this repository."`
	ViewerCanSubscribe    bool `graphql:"viewerCanSubscribe" group:"6⟭ Permissions" desc:"Indicates if the current viewer can subscribe to this repository."`
	ViewerCanUpdateTopics bool `graphql:"viewerCanUpdateTopics" group:"6⟭ Permissions" desc:"Indicates if the current viewer can update topics in this repository."`
	ViewerHasStarred      bool `graphql:"viewerHasStarred" group:"6⟭ Permissions" desc:"Indicates if the current viewer has starred this starrable."`

	// ClosedPullRequests struct {
	// 	TotalCount int
	// } `graphql:"pullRequests(states: CLOSED)"`

	// Languages struct {
	// 	Nodes []struct {
	// 		Name  string
	// 		Color string
	// 	} `graphql:"nodes"`
	// } `graphql:"languages(first: 100)"`
}

type BranchProtectionRuleConnection struct {
	Nodes      []BranchProtectionRule `graphql:"nodes" group:"branch" desc:"A list of branch protection rules."`
	TotalCount int                    `graphql:"totalCount" group:"branch" desc:"The total number of branch protection rules."`
} // `graphql:"branchProtectionRules" group:"branch" desc:"The connection type for BranchProtectionRule."`

type BranchProtectionRule struct {
	AllowsDeletions       bool   `graphql:"allowsDeletions" group:"branch" desc:"Are deletions allowed."`
	AllowsForcePushes     bool   `graphql:"allowsForcePushes" group:"branch" desc:"Are force pushes allowed."`
	BlocksCreations       bool   `graphql:"blocksCreations" group:"branch" desc:"Blocks branch creation."`
	DismissesStaleReviews bool   `graphql:"dismissesStaleReviews" group:"branch" desc:"Will new commits dismiss pull request review approvals."`
	Id                    string `graphql:"id" group:"branch" desc:"The Node ID of the branch protection rule."`
	IsAdminEnforced       bool   `graphql:"isAdminEnforced" group:"branch" desc:"Can admins override branch protection."`
	// IsLocked                       bool     `graphql:"isLocked" group:"branch" desc:"Is the rule locked."`
	LockBranch                     bool     `graphql:"lockBranch" group:"branch" desc:"Is the branch locked."`
	Pattern                        string   `graphql:"pattern" group:"branch" desc:"The glob-like pattern used to determine matching branches."`
	RequireLastPushApproval        bool     `graphql:"requireLastPushApproval" group:"branch" desc:"Are approvals required on the most recent commit."`
	RequiredApprovingReviewCount   int      `graphql:"requiredApprovingReviewCount" group:"branch" desc:"The number of approving reviews required."`
	RequiredStatusCheckContexts    []string `graphql:"requiredStatusCheckContexts" group:"branch" desc:"Required status check contexts that must pass for commits to be accepted."`
	RequiresApprovingReviews       bool     `graphql:"requiresApprovingReviews" group:"branch" desc:"Are reviews required."`
	RequiresCodeOwnerReviews       bool     `graphql:"requiresCodeOwnerReviews" group:"branch" desc:"Are code owner reviews required."`
	RequiresCommitSignatures       bool     `graphql:"requiresCommitSignatures" group:"branch" desc:"Are commit signatures required."`
	RequiresConversationResolution bool     `graphql:"requiresConversationResolution" group:"branch" desc:"Are conversations required."`
	RequiresDeployments            bool     `graphql:"requiresDeployments" group:"branch" desc:"Are deployments required."`
	RequiresLinearHistory          bool     `graphql:"requiresLinearHistory" group:"branch" desc:"Are merge commits prohibited."`
	RequiresStatusChecks           bool     `graphql:"requiresStatusChecks" group:"branch" desc:"Are status checks required."`
	RequiresStrictStatusChecks     bool     `graphql:"requiresStrictStatusChecks" group:"branch" desc:"Are branches required to be up to date before merging."`
	RestrictsPushes                bool     `graphql:"restrictsPushes" group:"branch" desc:"Is pushing restricted."`
	RestrictsReviewDismissals      bool     `graphql:"restrictsReviewDismissals" group:"branch" desc:"Is dismissing reviews restricted."`
} // `graphql:"" group:"branch" desc:"A branch protection rule."`

// Supporting Types for Security Features
type VulnerabilityAlertConnection struct {
	Nodes      []VulnerabilityAlert `graphql:"nodes" group:"security" desc:"A list of vulnerability alerts."`
	TotalCount int                  `graphql:"totalCount" group:"security" desc:"The total number of vulnerability alerts."`
} // `graphql:"vulnerabilityAlerts" group:"security" desc:"The connection type for VulnerabilityAlert."`

type VulnerabilityAlert struct {
	Id                         string `graphql:"id" group:"security" desc:"The Node ID of the vulnerability alert."`
	CreatedAt                  string `graphql:"createdAt" group:"security" desc:"When the alert was created."`
	DismissedAt                string `graphql:"dismissedAt" group:"security" desc:"When the alert was dismissed."`
	DismissReason              string `graphql:"dismissReason" group:"security" desc:"The reason the alert was dismissed."`
	VulnerableManifestFilename string `graphql:"vulnerableManifestFilename" group:"security" desc:"The name of the manifest file that contains the dependency."`
	VulnerableManifestPath     string `graphql:"vulnerableManifestPath" group:"security" desc:"The full path to the manifest file that contains the dependency."`
	VulnerableRequirements     string `graphql:"vulnerableRequirements" group:"security" desc:"The version requirements that contain the vulnerability."`
} // `graphql:"" group:"security" desc:"An individual vulnerability alert."`

// Supporting Types for Connections
type IssueConnection struct {
	Nodes      []Issue `graphql:"nodes" group:"connections" desc:"A list of issues."`
	TotalCount int     `graphql:"totalCount" group:"connections" desc:"The total number of issues."`
} // `graphql:"issues" group:"connections" desc:"The connection type for Issue."`

type Issue struct {
	Id        string    `graphql:"id" group:"connections" desc:"The Node ID of the issue."`
	Number    int       `graphql:"number" group:"connections" desc:"The issue number."`
	Title     string    `graphql:"title" group:"connections" desc:"The title of the issue."`
	Body      string    `graphql:"body" group:"connections" desc:"The body of the issue."`
	State     string    `graphql:"state" group:"connections" desc:"The state of the issue (OPEN, CLOSED)."`
	CreatedAt time.Time `graphql:"createdAt" group:"connections" desc:"When the issue was created."`
	UpdatedAt time.Time `graphql:"updatedAt" group:"connections" desc:"When the issue was last updated."`
	ClosedAt  time.Time `graphql:"closedAt" group:"connections" desc:"When the issue was closed."`
	// Labels    LabelConnection `graphql:"labels" group:"connections" desc:"Labels associated with this issue."`
	// Author    Actor           `graphql:"author" group:"connections" desc:"The author of the issue."`
} // `graphql:"" group:"connections" desc:"Represents an issue in a repository."`

type PullRequestConnection struct {
	Nodes      []PullRequest `graphql:"nodes" group:"connections" desc:"A list of pull requests."`
	TotalCount int           `graphql:"totalCount" group:"connections" desc:"The total number of pull requests."`
} // `graphql:"pullRequests" group:"connections" desc:"The connection type for PullRequest."`

type PullRequest struct {
	ID          string    `graphql:"id" group:"connections" desc:"The Node ID of the pull request."`
	Number      int       `graphql:"number" group:"connections" desc:"The pull request number."`
	Title       string    `graphql:"title" group:"connections" desc:"The title of the pull request."`
	Body        string    `graphql:"body" group:"connections" desc:"The body of the pull request."`
	State       string    `graphql:"state" group:"connections" desc:"The state of the pull request (OPEN, CLOSED, MERGED)."`
	IsDraft     bool      `graphql:"isDraft" group:"connections" desc:"Whether this PR is a draft."`
	Merged      bool      `graphql:"merged" group:"connections" desc:"Whether this PR has been merged."`
	MergedAt    time.Time `graphql:"mergedAt" group:"connections" desc:"When the PR was merged."`
	BaseRefName string    `graphql:"baseRefName" group:"connections" desc:"The name of the base branch."`
	HeadRefName string    `graphql:"headRefName" group:"connections" desc:"The name of the head branch."`
} // `graphql:"" group:"connections" desc:"Represents a pull request in a repository."`

type LanguageConnection struct {
	Nodes      []Language `graphql:"nodes" group:"connections" desc:"A list of languages."`
	TotalCount int        `graphql:"totalCount" group:"connections" desc:"The total number of languages."`
	TotalSize  int        `graphql:"totalSize" group:"connections" desc:"The total size in bytes of files written in each language."`
} // `graphql:"languages" group:"connections" desc:"The connection type for Language."`

type Language struct {
	Name      string `graphql:"name" group:"connections" desc:"The name of the programming language."`
	Color     string `graphql:"color" group:"connections" desc:"The color defined for the programming language."`
	Size      int    `graphql:"size" group:"connections" desc:"The number of bytes of code written in the language."`
	IsViable  bool   `graphql:"isViable" group:"connections" desc:"Whether the language is a viable option for the repository."`
	IsPopular bool   `graphql:"isPopular" group:"connections" desc:"Whether the language is a popular option for the repository."`
} // `graphql:"" group:"connections" desc:"A programming language used in the repository."`

type ProjectV2Connection struct {
	Nodes      []ProjectV2 `graphql:"nodes" group:"connections" desc:"A list of projects (v2)."`
	TotalCount int         `graphql:"totalCount" group:"connections" desc:"The total number of projects."`
} // `graphql:"projectsV2" group:"connections" desc:"The connection type for ProjectV2."`

type ProjectV2 struct {
	Id           string    `graphql:"id" group:"connections" desc:"The Node ID of the project."`
	Name         string    `graphql:"name" group:"connections" desc:"The name of the project."`
	Body         string    `graphql:"body" group:"connections" desc:"The body of the project."`
	Number       int       `graphql:"number" group:"connections" desc:"The project number."`
	State        string    `graphql:"state" group:"connections" desc:"The state of the project (OPEN, CLOSED)."`
	Url          string    `graphql:"url" group:"connections" desc:"The HTTP URL for this project."`
	ResourcePath string    `graphql:"resourcePath" group:"connections" desc:"The project's URL path."`
	BodyHTML     string    `graphql:"bodyHTML" group:"connections" desc:"The body of the project rendered to HTML."`
	UpdatedAt    time.Time `graphql:"updatedAt" group:"connections" desc:"When the project was last updated."`
} // `graphql:"" group:"connections" desc:"A project in a repository."`

type DiscussionConnection struct {
	Nodes      []Discussion `graphql:"nodes" group:"connections" desc:"A list of discussions."`
	TotalCount int          `graphql:"totalCount" group:"connections" desc:"The total number of discussions."`
} // `graphql:"discussions" group:"connections" desc:"The connection type for Discussion."`

type Discussion struct {
	Id        string    `graphql:"id" group:"connections" desc:"The Node ID of the discussion."`
	Title     string    `graphql:"title" group:"connections" desc:"The title of the discussion."`
	Body      string    `graphql:"body" group:"connections" desc:"The body of the discussion."`
	CreatedAt time.Time `graphql:"createdAt" group:"connections" desc:"When the discussion was created."`
	UpdatedAt time.Time `graphql:"updatedAt" group:"connections" desc:"When the discussion was last updated."`
} // `graphql:"" group:"connections" desc:"A discussion in a repository."`

type ReleaseConnection struct {
	Nodes      []Release `graphql:"nodes" group:"connections" desc:"A list of releases."`
	TotalCount int       `graphql:"totalCount" group:"connections" desc:"The total number of releases."`
} // `graphql:"releases" group:"connections" desc:"The connection type for Release."`

type Release struct {
	Id           string    `graphql:"id" group:"connections" desc:"The Node ID of the release."`
	Name         string    `graphql:"name" group:"connections" desc:"The name of the release."`
	TagName      string    `graphql:"tagName" group:"connections" desc:"The tag name of the release."`
	Description  string    `graphql:"description" group:"connections" desc:"The description of the release."`
	IsDraft      bool      `graphql:"isDraft" group:"connections" desc:"Whether this release is a draft."`
	IsPrerelease bool      `graphql:"isPrerelease" group:"connections" desc:"Whether this release is a prerelease."`
	PublishedAt  time.Time `graphql:"publishedAt" group:"connections" desc:"The date and time when the release was published."`
} // `graphql:"" group:"connections" desc:"A release in a repository."`

type PackageConnection struct {
	Nodes      []Package `graphql:"nodes" group:"connections" desc:"A list of packages."`
	TotalCount int       `graphql:"totalCount" group:"connections" desc:"The total number of packages."`
} // `graphql:"packages" group:"connections" desc:"The connection type for Package."`

type Package struct {
	Id         string `graphql:"id" group:"connections" desc:"The Node ID of the package."`
	Name       string `graphql:"name" group:"connections" desc:"The name of the package."`
	Version    string `graphql:"version" group:"connections" desc:"The version of the package."`
	Manifest   string `graphql:"manifest" group:"connections" desc:"The manifest of the package."`
	Repository struct {
		NameWithOwner string
	} `graphql:"repository" group:"connections" desc:"The repository that the package belongs to."`
} // `graphql:"" group:"connections" desc:"A package in a repository."`

type RepositoryTopicConnection struct {
	Nodes      []RepositoryTopic `graphql:"nodes" group:"connections" desc:"A list of repository topics."`
	TotalCount int               `graphql:"totalCount" group:"connections" desc:"The total number of repository topics."`
} // `graphql:"repositoryTopics" group:"connections" desc:"The connection type for RepositoryTopic."`

type RepositoryTopic struct {
	Topic struct {
		Name string
	} `graphql:"topic" group:"connections" desc:"The topic."`
} // `graphql:"" group:"connections" desc:"A repository topic."`

type DependencyConnection struct {
	Nodes      []Dependency `graphql:"nodes" group:"security" desc:"A list of dependencies."`
	TotalCount int          `graphql:"totalCount" group:"security" desc:"The total number of dependencies."`
} // `graphql:"dependencies" group:"security" desc:"The connection type for Repository dependencies."`

type Dependency struct {
	Id              string `graphql:"id" group:"security" desc:"The Node ID of the dependency."`
	HasDependencies bool   `graphql:"hasDependencies" group:"security" desc:"Whether this dependency has dependencies of its own."`
	HasDependents   bool   `graphql:"hasDependents" group:"security" desc:"Whether this dependency is depended upon by others."`
} // `graphql:"" group:"security" desc:"A dependency of a project."`
