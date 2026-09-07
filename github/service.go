// Package github wraps the parts of the GitHub API that gh-reponark needs
// behind a small interface so the UI can be exercised without the network.
package github

import "gh-reponark/repo"

// Service is the subset of the GitHub API used by the application. The UI
// depends only on this interface; Client is the real implementation.
type Service interface {
	// CurrentUser returns the authenticated user together with the
	// organizations they are a member of.
	CurrentUser() (User, error)
	// ListRepositories returns one page of the repositories owned by login,
	// including their full configuration. isUser selects between the user and
	// organization GraphQL roots. Pass an empty cursor for the first page and
	// the previous page's EndCursor for the following ones.
	ListRepositories(login string, isUser bool, after string) (RepositoryPage, error)
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

// RepositoryPage is one page of a repository listing.
type RepositoryPage struct {
	Repositories []repo.Repository
	// TotalCount is the number of repositories across every page.
	TotalCount int
	// EndCursor is passed as after to fetch the next page.
	EndCursor   string
	HasNextPage bool
}
