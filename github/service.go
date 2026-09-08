// Package github wraps the parts of the GitHub API that gh-reponark needs
// behind a small interface so the UI can be exercised without the network.
package github

import "gh-reponark/repo"

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
}

// User is the authenticated GitHub user.
type User struct {
	Login         string
	Url           string
	Organizations []Organization
}

// Organization is an organization the authenticated user belongs to.
type Organization struct {
	Login string
	Url   string
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
