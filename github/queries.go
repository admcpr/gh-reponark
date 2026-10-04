package github

import "time"

// currentUserResponse is the subset of the REST /user response we read.
type currentUserResponse struct {
	Login string `json:"login"`
}

// countConnection is a connection asked only for its size.
type countConnection struct {
	TotalCount int `graphql:"totalCount"`
}

// userQuery fetches the signed-in user and their organizations with enough
// about each to describe it on the account picker: name, bio or
// description, repository counts, size and age. The public repositories are
// a second, aliased connection on the same query.
type userQuery struct {
	User struct {
		Login              string
		Name               string
		Bio                string
		Url                string
		CreatedAt          time.Time       `graphql:"createdAt"`
		Repositories       countConnection `graphql:"repositories"`
		PublicRepositories countConnection `graphql:"publicRepositories: repositories(privacy: PUBLIC)"`
		Followers          countConnection `graphql:"followers"`
		Organizations      struct {
			Nodes []struct {
				Login               string
				Name                string
				Description         string
				Url                 string
				IsVerified          bool            `graphql:"isVerified"`
				ViewerCanAdminister bool            `graphql:"viewerCanAdminister"`
				CreatedAt           time.Time       `graphql:"createdAt"`
				Repositories        countConnection `graphql:"repositories"`
				PublicRepositories  countConnection `graphql:"publicRepositories: repositories(privacy: PUBLIC)"`
				MembersWithRole     countConnection `graphql:"membersWithRole"`
			} `graphql:"nodes"`
		} `graphql:"organizations(first: $first)"`
	} `graphql:"user(login: $login)"`
}

// repositoryNameConnection is a page of repository names. It deliberately asks
// for nothing expensive so a page of 100 stays fast; configurations are
// fetched separately in small batches.
type repositoryNameConnection struct {
	TotalCount int `graphql:"totalCount"`
	PageInfo   struct {
		HasNextPage bool   `graphql:"hasNextPage"`
		EndCursor   string `graphql:"endCursor"`
	} `graphql:"pageInfo"`
	Nodes []struct {
		Name string
		Url  string
	} `graphql:"nodes"`
}

type organizationRepositoriesQuery struct {
	Organization struct {
		Repositories repositoryNameConnection `graphql:"repositories(first: $first, after: $after, affiliations: OWNER)"`
	} `graphql:"organization(login: $login)"`
}

type userRepositoriesQuery struct {
	User struct {
		Repositories repositoryNameConnection `graphql:"repositories(first: $first, after: $after, affiliations: OWNER)"`
	} `graphql:"user(login: $login)"`
}
