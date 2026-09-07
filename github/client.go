package github

import (
	"fmt"

	"gh-reponark/repo"

	"github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
)

// pageSize is the number of items requested from GraphQL connections.
const pageSize = 100

// Client implements Service against the real GitHub API using the
// authentication configured for the gh CLI.
type Client struct {
	graphql *api.GraphQLClient
	rest    *api.RESTClient
}

var _ Service = (*Client)(nil)

// NewClient creates a Client using the gh CLI's configured host and token.
// It fails when no authentication is available, so callers can report that
// before starting the UI.
func NewClient() (*Client, error) {
	return NewClientWithOptions(api.ClientOptions{})
}

// NewClientWithOptions creates a Client with explicit go-gh options. It is
// mainly useful for tests, which can supply a Host, AuthToken and Transport.
func NewClientWithOptions(opts api.ClientOptions) (*Client, error) {
	graphqlClient, err := api.NewGraphQLClient(opts)
	if err != nil {
		return nil, fmt.Errorf("creating GitHub GraphQL client: %w", err)
	}

	restClient, err := api.NewRESTClient(opts)
	if err != nil {
		return nil, fmt.Errorf("creating GitHub REST client: %w", err)
	}

	return &Client{graphql: graphqlClient, rest: restClient}, nil
}

func (c *Client) CurrentUser() (User, error) {
	var current currentUserResponse
	if err := c.rest.Get("user", &current); err != nil {
		return User{}, fmt.Errorf("looking up the authenticated user: %w", err)
	}

	var query userQuery
	variables := map[string]interface{}{
		"login": graphql.String(current.Login),
		"first": graphql.Int(pageSize),
	}
	if err := c.graphql.Query("User", &query, variables); err != nil {
		return User{}, fmt.Errorf("fetching user %s: %w", current.Login, err)
	}

	user := User{Login: query.User.Login, Url: query.User.Url}
	for _, org := range query.User.Organizations.Nodes {
		user.Organizations = append(user.Organizations, Organization{Login: org.Login, Url: org.Url})
	}

	return user, nil
}

func (c *Client) ListRepositories(login string, isUser bool, after string) (RepositoryPage, error) {
	// A nil cursor asks for the first page; GraphQL treats it as null.
	var cursor *graphql.String
	if after != "" {
		value := graphql.String(after)
		cursor = &value
	}

	variables := map[string]interface{}{
		"login": graphql.String(login),
		"first": graphql.Int(pageSize),
		"after": cursor,
	}

	var connection repositoryConnection
	if isUser {
		var query userRepositoriesQuery
		if err := c.graphql.Query("UserRepositories", &query, variables); err != nil {
			return RepositoryPage{}, fmt.Errorf("fetching repositories for user %s: %w", login, err)
		}
		connection = query.User.Repositories
	} else {
		var query organizationRepositoriesQuery
		if err := c.graphql.Query("OrganizationRepositories", &query, variables); err != nil {
			return RepositoryPage{}, fmt.Errorf("fetching repositories for organization %s: %w", login, err)
		}
		connection = query.Organization.Repositories
	}

	page := RepositoryPage{
		Repositories: connection.Nodes,
		TotalCount:   connection.TotalCount,
		EndCursor:    connection.PageInfo.EndCursor,
		HasNextPage:  connection.PageInfo.HasNextPage,
	}
	if page.Repositories == nil {
		page.Repositories = []repo.Repository{}
	}

	return page, nil
}
