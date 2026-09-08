package github

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
