// Package githubtest provides an in-memory github.Service for tests.
package githubtest

import (
	"strconv"

	"gh-reponark/github"
	"gh-reponark/repo"
)

// ListCall records the arguments of a ListRepositories call.
type ListCall struct {
	Login  string
	IsUser bool
	After  string
}

// GetCall records the arguments of a GetRepositories call.
type GetCall struct {
	Owner string
	Names []string
}

// Fake is a github.Service backed by fixed data. Set the *Err fields to make
// the corresponding method fail. Every call is recorded so tests can assert on
// what the UI asked for. Use it via a pointer so the recorded calls are kept.
type Fake struct {
	User    github.User
	UserErr error

	// Repositories back both ListRepositories, which returns their names, and
	// GetRepositories, which returns them by name. When PageSize is greater
	// than zero the names are listed PageSize at a time using the slice index
	// as the cursor; otherwise a single page holds them all.
	Repositories []repo.Repository
	PageSize     int
	ListErr      error
	GetErr       error

	ListCalls []ListCall
	GetCalls  []GetCall
}

var _ github.Service = (*Fake)(nil)

func (f *Fake) CurrentUser() (github.User, error) {
	if f.UserErr != nil {
		return github.User{}, f.UserErr
	}
	return f.User, nil
}

func (f *Fake) ListRepositories(login string, isUser bool, after string) (github.RepositoryPage, error) {
	f.ListCalls = append(f.ListCalls, ListCall{Login: login, IsUser: isUser, After: after})
	if f.ListErr != nil {
		return github.RepositoryPage{}, f.ListErr
	}

	total := len(f.Repositories)
	start := 0
	if after != "" {
		start, _ = strconv.Atoi(after)
	}
	if start > total {
		start = total
	}

	end := total
	if f.PageSize > 0 && start+f.PageSize < total {
		end = start + f.PageSize
	}

	page := github.RepositoryPage{
		Repositories: make([]github.RepositoryRef, 0, end-start),
		TotalCount:   total,
		EndCursor:    strconv.Itoa(end),
		HasNextPage:  end < total,
	}
	for _, repository := range f.Repositories[start:end] {
		page.Repositories = append(page.Repositories, github.RepositoryRef{Name: repository.Name, Url: repository.Url})
	}
	return page, nil
}

func (f *Fake) GetRepositories(owner string, names []string) ([]repo.Repository, error) {
	f.GetCalls = append(f.GetCalls, GetCall{Owner: owner, Names: append([]string(nil), names...)})
	if f.GetErr != nil {
		return nil, f.GetErr
	}

	repositories := make([]repo.Repository, len(names))
	for i, name := range names {
		repositories[i] = repo.Repository{Name: name}
		for _, repository := range f.Repositories {
			if repository.Name == name {
				repositories[i] = repository
				break
			}
		}
	}
	return repositories, nil
}
