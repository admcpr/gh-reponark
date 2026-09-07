package github

import "gh-reponark/repo"

// currentUserResponse is the subset of the REST /user response we read.
type currentUserResponse struct {
	Login string `json:"login"`
}

type userQuery struct {
	User struct {
		Login         string
		Url           string
		Organizations struct {
			Nodes []struct {
				Login string
				Url   string
			} `graphql:"nodes"`
		} `graphql:"organizations(first: $first)"`
	} `graphql:"user(login: $login)"`
}

// repositoryConnection is a page of repositories with their full
// configuration, as described by repo.Repository's graphql tags.
type repositoryConnection struct {
	TotalCount int `graphql:"totalCount"`
	PageInfo   struct {
		HasNextPage bool   `graphql:"hasNextPage"`
		EndCursor   string `graphql:"endCursor"`
	} `graphql:"pageInfo"`
	Nodes []repo.Repository `graphql:"nodes"`
}

type organizationRepositoriesQuery struct {
	Organization struct {
		Repositories repositoryConnection `graphql:"repositories(first: $first, after: $after, affiliations: OWNER)"`
	} `graphql:"organization(login: $login)"`
}

type userRepositoriesQuery struct {
	User struct {
		Repositories repositoryConnection `graphql:"repositories(first: $first, after: $after, affiliations: OWNER)"`
	} `graphql:"user(login: $login)"`
}
