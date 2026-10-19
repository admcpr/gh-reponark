// Package github wraps the parts of the GitHub API that gh-reponark needs
// behind a small interface so the UI can be exercised without the network.
package github

import (
	"time"

	"gh-reponark/repo"
)

// RepositoryBatchSize is the most repositories a caller should ask
// GetRepositories for at once. Every repository's configuration includes
// several connection counts (issues, pull requests, releases, vulnerability
// alerts) that cost GitHub roughly a quarter of a second each, and a GraphQL
// request that takes longer than about ten seconds fails with a 504. Ten per
// request leaves a comfortable margin.
const RepositoryBatchSize = 10

// Service is the subset of the GitHub API used by the application. The UI
// depends only on this interface; Client is the real implementation.
type Service interface {
	// CurrentUser returns the authenticated user together with the
	// organizations they are a member of.
	CurrentUser() (User, error)
	// ListRepositories returns one page of the names of the repositories owned
	// by login. isUser selects between the user and organization GraphQL
	// roots. Pass an empty cursor for the first page and the previous page's
	// EndCursor for the following ones.
	ListRepositories(login string, isUser bool, after string) (RepositoryPage, error)
	// GetRepositories returns the full configuration of the named repositories
	// owned by owner, in the order requested, using a single request. Callers
	// should pass at most RepositoryBatchSize names.
	GetRepositories(owner string, names []string) ([]repo.Repository, error)
	// GetOwner returns the profile of the organization or user whose
	// repositories are being browsed: the parts any member of an
	// organization, or anyone at all for a user, is allowed to see.
	GetOwner(login string, isUser bool) (Owner, error)
}

// User is the authenticated GitHub user.
type User struct {
	Login string
	// Name is the display name, which may be empty.
	Name string
	// Description is the profile bio.
	Description string
	Url         string
	// Repositories counts every repository the user owns and
	// PublicRepositories the public ones among them.
	Repositories       int
	PublicRepositories int
	// Members is the follower count: the nearest thing a user has to an
	// organization's membership.
	Members       int
	CreatedAt     time.Time
	Organizations []Organization
}

// Organization is an organization the authenticated user belongs to.
type Organization struct {
	Login       string
	Name        string
	Description string
	Url         string
	// Repositories counts every repository the organization owns and
	// PublicRepositories the public ones among them.
	Repositories       int
	PublicRepositories int
	// Members counts the organization's members.
	Members int
	// ViewerCanAdminister is true when the signed-in user is an owner.
	ViewerCanAdminister bool
	// IsVerified is true when GitHub has verified the organization's domains.
	IsVerified bool
	CreatedAt  time.Time
}

// RepositoryRef identifies a repository without its configuration.
type RepositoryRef struct {
	Name string
	Url  string
}

// RepositoryPage is one page of a repository listing.
type RepositoryPage struct {
	Repositories []RepositoryRef
	// TotalCount is the number of repositories across every page.
	TotalCount int
	// EndCursor is passed as after to fetch the next page.
	EndCursor   string
	HasNextPage bool
}

// Owner is the profile of a repository owner, an organization or a user, as
// seen by someone who merely belongs to it. Nothing here needs the signed-in
// user to be an owner or admin, so it loads for every account on the picker.
type Owner struct {
	Login string
	// Name is the display name, which may be empty.
	Name string
	// Description is an organization's description or a user's bio.
	Description     string
	Url             string
	WebsiteUrl      string
	Location        string
	Email           string
	TwitterUsername string
	// Company is only set for users: the one on their profile.
	Company string
	// IsUser says whether the owner is a user rather than an organization.
	IsUser bool
	// IsVerified is true when GitHub has verified the organization's domains.
	IsVerified bool
	// ViewerIsAMember says whether the signed-in user belongs to the
	// organization, and ViewerCanAdminister whether they own it.
	ViewerIsAMember     bool
	ViewerCanAdminister bool
	// What the signed-in user may create in the organization.
	ViewerCanCreateRepositories bool
	ViewerCanCreateTeams        bool
	ViewerCanCreateProjects     bool
	// HasSponsorsListing is true when the owner can be sponsored.
	HasSponsorsListing bool
	CreatedAt          time.Time
	// Members counts an organization's members or a user's followers;
	// Following and Teams only apply to one kind of owner each.
	Members   int
	Following int
	Teams     int
	// Pinned names the repositories pinned to the profile, in order.
	Pinned []string
}
