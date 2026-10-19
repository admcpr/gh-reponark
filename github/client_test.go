package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"gh-reponark/repo"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// graphqlRequest is the JSON body shurcooL-graphql sends.
type graphqlRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

// fakeTransport answers HTTP requests from a handler and records what it saw.
type fakeTransport struct {
	handler  func(req *http.Request, gql *graphqlRequest) (status int, body string)
	requests []*http.Request
	graphql  []graphqlRequest
}

func (f *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var gql *graphqlRequest
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		var parsed graphqlRequest
		if len(body) > 0 && json.Unmarshal(body, &parsed) == nil && parsed.Query != "" {
			gql = &parsed
			f.graphql = append(f.graphql, parsed)
		}
	}
	f.requests = append(f.requests, req)

	status, body := f.handler(req, gql)
	return &http.Response{
		StatusCode: status,
		// Real servers send the code and the text, e.g. "504 Gateway Timeout".
		Status:  fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header:  http.Header{"Content-Type": []string{"application/json"}},
		Body:    io.NopCloser(strings.NewReader(body)),
		Request: req,
	}, nil
}

func newTestClient(t *testing.T, handler func(req *http.Request, gql *graphqlRequest) (int, string)) (*Client, *fakeTransport) {
	t.Helper()

	transport := &fakeTransport{handler: handler}
	client, err := NewClientWithOptions(api.ClientOptions{
		Host:         "github.com",
		AuthToken:    "test-token",
		Transport:    transport,
		LogIgnoreEnv: true,
	})
	require.NoError(t, err)

	return client, transport
}

func isOperation(gql *graphqlRequest, name string) bool {
	return gql != nil && strings.HasPrefix(gql.Query, "query "+name+"(")
}

func graphqlError(message string) string {
	return `{"data":null,"errors":[{"message":"` + message + `"}]}`
}

func boolString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// namesResponse builds a repository names connection response under root.
func namesResponse(root string, totalCount int, hasNextPage bool, endCursor string, names ...string) string {
	nodes := make([]string, len(names))
	for i, name := range names {
		nodes[i] = `{"name":"` + name + `","url":"https://github.com/acme/` + name + `"}`
	}
	total, _ := json.Marshal(totalCount)
	return `{"data":{"` + root + `":{"repositories":{` +
		`"totalCount":` + string(total) + `,` +
		`"pageInfo":{"hasNextPage":` + boolString(hasNextPage) + `,"endCursor":"` + endCursor + `"},` +
		`"nodes":[` + strings.Join(nodes, ",") + `]}}}}`
}

func TestClient_CurrentUser(t *testing.T) {
	client, transport := newTestClient(t, func(req *http.Request, gql *graphqlRequest) (int, string) {
		switch {
		case req.URL.Path == "/user":
			return http.StatusOK, `{"login":"octocat","name":"The Octocat"}`
		case isOperation(gql, "User"):
			return http.StatusOK, `{"data":{"user":{
				"login":"octocat",
				"name":"The Octocat",
				"bio":"Mascot",
				"url":"https://github.com/octocat",
				"createdAt":"2011-01-25T18:44:36Z",
				"repositories":{"totalCount":8},
				"publicRepositories":{"totalCount":6},
				"followers":{"totalCount":9000},
				"organizations":{"nodes":[
					{"login":"acme","name":"Acme","description":"Anvils","url":"https://github.com/acme",
					 "isVerified":true,"viewerCanAdminister":true,"createdAt":"2015-03-04T05:06:07Z",
					 "repositories":{"totalCount":42},"publicRepositories":{"totalCount":30},
					 "membersWithRole":{"totalCount":18}},
					{"login":"globex","url":"https://github.com/globex",
					 "repositories":{"totalCount":1},"publicRepositories":{"totalCount":0},
					 "membersWithRole":{"totalCount":2}}
				]}}}}`
		default:
			return http.StatusNotFound, `{"message":"unexpected request"}`
		}
	})

	user, err := client.CurrentUser()

	require.NoError(t, err)
	assert.Equal(t, User{
		Login:              "octocat",
		Name:               "The Octocat",
		Description:        "Mascot",
		Url:                "https://github.com/octocat",
		Repositories:       8,
		PublicRepositories: 6,
		Members:            9000,
		CreatedAt:          time.Date(2011, 1, 25, 18, 44, 36, 0, time.UTC),
		Organizations: []Organization{
			{
				Login: "acme", Name: "Acme", Description: "Anvils", Url: "https://github.com/acme",
				Repositories: 42, PublicRepositories: 30, Members: 18,
				ViewerCanAdminister: true, IsVerified: true,
				CreatedAt: time.Date(2015, 3, 4, 5, 6, 7, 0, time.UTC),
			},
			{Login: "globex", Url: "https://github.com/globex", Repositories: 1, Members: 2},
		},
	}, user)

	require.Len(t, transport.requests, 2)
	for _, req := range transport.requests {
		assert.Contains(t, req.Header.Get("Authorization"), "test-token")
	}
	require.Len(t, transport.graphql, 1)
	assert.Equal(t, "octocat", transport.graphql[0].Variables["login"])
	assert.EqualValues(t, pageSize, transport.graphql[0].Variables["first"])
	assert.Contains(t, transport.graphql[0].Query, "publicRepositories: repositories(privacy: PUBLIC){totalCount}",
		"the public count is an aliased second repositories connection")
	assert.Contains(t, transport.graphql[0].Query, "membersWithRole{totalCount}")
}

func TestClient_CurrentUser_NoOrganizations(t *testing.T) {
	client, _ := newTestClient(t, func(req *http.Request, gql *graphqlRequest) (int, string) {
		if req.URL.Path == "/user" {
			return http.StatusOK, `{"login":"octocat"}`
		}
		return http.StatusOK, `{"data":{"user":{"login":"octocat","url":"https://github.com/octocat","organizations":{"nodes":[]}}}}`
	})

	user, err := client.CurrentUser()

	require.NoError(t, err)
	assert.Equal(t, "octocat", user.Login)
	assert.Empty(t, user.Organizations)
}

func TestClient_CurrentUser_RESTError(t *testing.T) {
	client, transport := newTestClient(t, func(req *http.Request, gql *graphqlRequest) (int, string) {
		return http.StatusUnauthorized, `{"message":"Bad credentials"}`
	})

	_, err := client.CurrentUser()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "looking up the authenticated user")
	assert.Contains(t, err.Error(), "Bad credentials")
	assert.Len(t, transport.requests, 1, "the GraphQL query should not run when the login lookup fails")
}

func TestClient_CurrentUser_GraphQLError(t *testing.T) {
	client, _ := newTestClient(t, func(req *http.Request, gql *graphqlRequest) (int, string) {
		if req.URL.Path == "/user" {
			return http.StatusOK, `{"login":"octocat"}`
		}
		return http.StatusOK, graphqlError("Something exploded")
	})

	_, err := client.CurrentUser()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "fetching user octocat")
	assert.Contains(t, err.Error(), "Something exploded")
}

func TestClient_ListRepositories(t *testing.T) {
	tests := []struct {
		name          string
		isUser        bool
		wantOperation string
		wantRoot      string
	}{
		{name: "organization", isUser: false, wantOperation: "OrganizationRepositories", wantRoot: "organization"},
		{name: "user", isUser: true, wantOperation: "UserRepositories", wantRoot: "user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, transport := newTestClient(t, func(req *http.Request, gql *graphqlRequest) (int, string) {
				if !isOperation(gql, tt.wantOperation) {
					return http.StatusBadRequest, graphqlError("unexpected operation")
				}
				return http.StatusOK, namesResponse(tt.wantRoot, 2, false, "cursor-2", "widgets", "gadgets")
			})

			page, err := client.ListRepositories("acme", tt.isUser, "")

			require.NoError(t, err)
			assert.Equal(t, RepositoryPage{
				Repositories: []RepositoryRef{
					{Name: "widgets", Url: "https://github.com/acme/widgets"},
					{Name: "gadgets", Url: "https://github.com/acme/gadgets"},
				},
				TotalCount:  2,
				EndCursor:   "cursor-2",
				HasNextPage: false,
			}, page)

			require.Len(t, transport.graphql, 1)
			query := transport.graphql[0].Query
			assert.Contains(t, query, tt.wantRoot+"(login: $login)")
			assert.Contains(t, query, "repositories(first: $first, after: $after, affiliations: OWNER)")
			assert.Contains(t, query, "totalCount")
			assert.Contains(t, query, "pageInfo{hasNextPage,endCursor}")
			assert.Contains(t, query, "nodes{name,url}")
			assert.NotContains(t, query, "isArchived", "the listing must stay cheap; configuration is fetched separately")
			assert.Contains(t, query, "$after:String")
			assert.NotContains(t, query, "$after:String!", "the cursor must be nullable so the first page can pass null")

			variables := transport.graphql[0].Variables
			assert.Equal(t, "acme", variables["login"])
			assert.EqualValues(t, pageSize, variables["first"])
			assert.Contains(t, variables, "after")
			assert.Nil(t, variables["after"], "the first page should send a null cursor")
		})
	}
}

func TestClient_ListRepositories_Pagination(t *testing.T) {
	client, transport := newTestClient(t, func(req *http.Request, gql *graphqlRequest) (int, string) {
		if gql.Variables["after"] == nil {
			return http.StatusOK, namesResponse("organization", 3, true, "cursor-2", "alpha", "bravo")
		}
		return http.StatusOK, namesResponse("organization", 3, false, "cursor-3", "charlie")
	})

	first, err := client.ListRepositories("acme", false, "")
	require.NoError(t, err)
	assert.True(t, first.HasNextPage)
	assert.Equal(t, 3, first.TotalCount)
	assert.Equal(t, "cursor-2", first.EndCursor)
	assert.Len(t, first.Repositories, 2)

	second, err := client.ListRepositories("acme", false, first.EndCursor)
	require.NoError(t, err)
	assert.False(t, second.HasNextPage)
	assert.Len(t, second.Repositories, 1)
	assert.Equal(t, "charlie", second.Repositories[0].Name)

	require.Len(t, transport.graphql, 2)
	assert.Nil(t, transport.graphql[0].Variables["after"])
	assert.Equal(t, "cursor-2", transport.graphql[1].Variables["after"])
}

func TestClient_ListRepositories_Empty(t *testing.T) {
	client, _ := newTestClient(t, func(req *http.Request, gql *graphqlRequest) (int, string) {
		return http.StatusOK, namesResponse("organization", 0, false, "")
	})

	page, err := client.ListRepositories("acme", false, "")

	require.NoError(t, err)
	assert.NotNil(t, page.Repositories)
	assert.Empty(t, page.Repositories)
	assert.Equal(t, 0, page.TotalCount)
	assert.False(t, page.HasNextPage)
}

func TestClient_ListRepositories_Error(t *testing.T) {
	tests := []struct {
		name    string
		isUser  bool
		wantMsg string
	}{
		{name: "organization", isUser: false, wantMsg: "fetching repositories for organization acme"},
		{name: "user", isUser: true, wantMsg: "fetching repositories for user acme"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, _ := newTestClient(t, func(req *http.Request, gql *graphqlRequest) (int, string) {
				return http.StatusOK, graphqlError("Could not resolve to a User")
			})

			_, err := client.ListRepositories("acme", tt.isUser, "")

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantMsg)
			assert.Contains(t, err.Error(), "Could not resolve to a User")
		})
	}
}

func TestClient_GetRepositories(t *testing.T) {
	client, transport := newTestClient(t, func(req *http.Request, gql *graphqlRequest) (int, string) {
		if !isOperation(gql, "Repositories") {
			return http.StatusBadRequest, graphqlError("unexpected operation")
		}
		return http.StatusOK, `{"data":{
			"repo0":{"name":"widgets","nameWithOwner":"acme/widgets","isArchived":true,"stargazerCount":42,
			         "createdAt":"2024-03-04T05:06:07Z","primaryLanguage":{"name":"Go"},
			         "issues":{"totalCount":3},"pullRequests":{"totalCount":2},"defaultBranchRef":{"name":"main"}},
			"repo1":{"name":"gadgets","nameWithOwner":"acme/gadgets","isPrivate":true}
		}}`
	})

	repositories, err := client.GetRepositories("acme", []string{"widgets", "gadgets"})

	require.NoError(t, err)
	require.Len(t, repositories, 2)

	widgets := repositories[0]
	assert.Equal(t, "widgets", widgets.Name)
	assert.Equal(t, "acme/widgets", widgets.NameWithOwner)
	assert.True(t, widgets.IsArchived)
	assert.Equal(t, 42, widgets.StargazerCount)
	assert.Equal(t, time.Date(2024, 3, 4, 5, 6, 7, 0, time.UTC), widgets.CreatedAt)
	assert.Equal(t, "Go", widgets.PrimaryLanguage.Name)
	assert.Equal(t, 3, widgets.Issues.TotalCount)
	assert.Equal(t, 2, widgets.OpenPullRequests.TotalCount)
	assert.Equal(t, "main", widgets.DefaultBranchRef.Name)

	gadgets := repositories[1]
	assert.Equal(t, "gadgets", gadgets.Name)
	assert.True(t, gadgets.IsPrivate)

	require.Len(t, transport.graphql, 1)
	query := transport.graphql[0].Query
	assert.Contains(t, query, "repo0: repository(owner: $owner, name: $name0){")
	assert.Contains(t, query, "repo1: repository(owner: $owner, name: $name1){")
	assert.Contains(t, query, "isArchived", "the full repository configuration should be requested")
	assert.Contains(t, query, "pullRequests(states: OPEN){totalCount}")
	assert.Equal(t, 2, strings.Count(query, "repository(owner: $owner"), "one aliased field per name")

	variables := transport.graphql[0].Variables
	assert.Equal(t, "acme", variables["owner"])
	assert.Equal(t, "widgets", variables["name0"])
	assert.Equal(t, "gadgets", variables["name1"])
}

func TestClient_GetRepositories_PreservesRequestOrder(t *testing.T) {
	client, _ := newTestClient(t, func(req *http.Request, gql *graphqlRequest) (int, string) {
		// Respond with the aliases out of order; decoding is by alias, not position.
		return http.StatusOK, `{"data":{"repo2":{"name":"charlie"},"repo0":{"name":"alpha"},"repo1":{"name":"bravo"}}}`
	})

	repositories, err := client.GetRepositories("acme", []string{"alpha", "bravo", "charlie"})

	require.NoError(t, err)
	names := make([]string, len(repositories))
	for i, r := range repositories {
		names[i] = r.Name
	}
	assert.Equal(t, []string{"alpha", "bravo", "charlie"}, names)
}

func TestClient_GetRepositories_Empty(t *testing.T) {
	client, transport := newTestClient(t, func(req *http.Request, gql *graphqlRequest) (int, string) {
		return http.StatusInternalServerError, "should not be called"
	})

	repositories, err := client.GetRepositories("acme", nil)

	require.NoError(t, err)
	assert.NotNil(t, repositories)
	assert.Empty(t, repositories)
	assert.Empty(t, transport.requests, "no request should be made for an empty batch")
}

func TestClient_GetRepositories_Error(t *testing.T) {
	client, _ := newTestClient(t, func(req *http.Request, gql *graphqlRequest) (int, string) {
		return http.StatusGatewayTimeout, `{"message": "We couldn't respond to your request in time."}`
	})

	_, err := client.GetRepositories("acme", []string{"widgets", "gadgets"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "fetching 2 repositories for acme")
	assert.Contains(t, err.Error(), "504")
}

func TestClient_ImplementsService(t *testing.T) {
	var svc Service = &Client{}
	assert.NotNil(t, svc)
}

func TestClient_GetOwner_Organization(t *testing.T) {
	client, transport := newTestClient(t, func(req *http.Request, gql *graphqlRequest) (int, string) {
		if !isOperation(gql, "OrganizationProfile") {
			return http.StatusNotFound, `{"message":"unexpected request"}`
		}
		return http.StatusOK, `{"data":{"organization":{
			"login":"acme","name":"Acme","url":"https://github.com/acme",
			"websiteUrl":"https://acme.example.com","location":"Springfield",
			"email":"hello@acme.example.com","twitterUsername":"acme",
			"hasSponsorsListing":true,"createdAt":"2015-03-04T05:06:07Z",
			"pinnedItems":{"nodes":[{"name":"anvil"},{"name":"rocket"}]},
			"description":"Anvils","isVerified":true,"viewerIsAMember":true,
			"viewerCanAdminister":false,"viewerCanCreateRepositories":true,
			"viewerCanCreateTeams":false,"viewerCanCreateProjects":true,
			"membersWithRole":{"totalCount":18},"teams":{"totalCount":5}}}}`
	})

	owner, err := client.GetOwner("acme", false)

	require.NoError(t, err)
	assert.Equal(t, Owner{
		Login: "acme", Name: "Acme", Description: "Anvils", Url: "https://github.com/acme",
		WebsiteUrl: "https://acme.example.com", Location: "Springfield",
		Email: "hello@acme.example.com", TwitterUsername: "acme",
		IsVerified: true, ViewerIsAMember: true,
		ViewerCanCreateRepositories: true, ViewerCanCreateProjects: true,
		HasSponsorsListing: true,
		CreatedAt:          time.Date(2015, 3, 4, 5, 6, 7, 0, time.UTC),
		Members:            18, Teams: 5,
		Pinned: []string{"anvil", "rocket"},
	}, owner)

	require.Len(t, transport.graphql, 1)
	query := transport.graphql[0]
	assert.Equal(t, "acme", query.Variables["login"])
	assert.NotContains(t, query.Query, "repositories", "repository counts come from the loaded repositories")
	assert.Contains(t, query.Query, "pinnedItems(first: 6, types: REPOSITORY){nodes{... on Repository{name}}}")
	assert.NotContains(t, query.Query, "requiresTwoFactorAuthentication", "owner-only settings would fail the query for members")
}

func TestClient_GetOwner_User(t *testing.T) {
	client, transport := newTestClient(t, func(req *http.Request, gql *graphqlRequest) (int, string) {
		if !isOperation(gql, "UserProfile") {
			return http.StatusNotFound, `{"message":"unexpected request"}`
		}
		return http.StatusOK, `{"data":{"user":{
			"login":"octocat","name":"The Octocat","url":"https://github.com/octocat",
			"createdAt":"2011-01-25T18:44:36Z",
			"pinnedItems":{"nodes":[]},
			"bio":"Mascot","company":"GitHub","viewerCanCreateProjects":true,
			"followers":{"totalCount":9000},"following":{"totalCount":9}}}}`
	})

	owner, err := client.GetOwner("octocat", true)

	require.NoError(t, err)
	assert.Equal(t, Owner{
		Login: "octocat", Name: "The Octocat", Description: "Mascot", Url: "https://github.com/octocat",
		Company: "GitHub", IsUser: true, ViewerCanCreateProjects: true,
		CreatedAt: time.Date(2011, 1, 25, 18, 44, 36, 0, time.UTC),
		Members:   9000, Following: 9,
	}, owner)
	require.Len(t, transport.graphql, 1)
	assert.Contains(t, transport.graphql[0].Query, "user(login: $login){")
}

func TestClient_GetOwner_Error(t *testing.T) {
	client, _ := newTestClient(t, func(req *http.Request, gql *graphqlRequest) (int, string) {
		return http.StatusOK, graphqlError("Could not resolve to an Organization with the login of 'nope'.")
	})

	_, err := client.GetOwner("nope", false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "fetching organization nope")
	assert.Contains(t, err.Error(), "Could not resolve")
}

func TestClient_GetRepositories_BranchRules(t *testing.T) {
	// The shape GitHub returned for a public repository the viewer can only
	// read: an organization ruleset, a repository ruleset, classic
	// protection, and a ruleset in evaluate mode.
	client, transport := newTestClient(t, func(req *http.Request, gql *graphqlRequest) (int, string) {
		return http.StatusOK, `{"data":{"repo0":{"name":"editor","defaultBranchRef":{"name":"main",
			"rules":{"nodes":[
				{"type":"COPILOT_CODE_REVIEW","parameters":null,"repositoryRuleset":{"name":"Copilot Reviews","source":{"__typename":"Repository"}}},
				{"type":"PULL_REQUEST","parameters":{"requiredApprovingReviewCount":1},"repositoryRuleset":{"name":"production","source":{"__typename":"Organization"}}},
				{"type":"NON_FAST_FORWARD","parameters":null,"repositoryRuleset":{"name":"production","source":{"__typename":"Organization"}}}
			]},
			"refUpdateRule":{"pattern":"main","allowsDeletions":false,"allowsForcePushes":false,"requiredApprovingReviewCount":2,
				"requiredStatusCheckContexts":["build"],"requiresLinearHistory":false,"requiresSignatures":false}},
			"rulesets":{"totalCount":2,"nodes":[{"enforcement":"ACTIVE"},{"enforcement":"EVALUATE"}]}}}}`
	})

	repositories, err := client.GetRepositories("acme", []string{"editor"})

	require.NoError(t, err)
	properties := repo.ToProperties(repositories[0])
	assert.Equal(t, true, properties["Pull Request Rule"].Value)
	assert.Equal(t, 2, properties["Required Approvals"].Value, "the stricter of the ruleset and classic protection")
	assert.Equal(t, true, properties["Status Checks Rule"].Value, "from classic protection")
	assert.Equal(t, true, properties["Copilot Review Rule"].Value)
	assert.Equal(t, true, properties["Deletion Blocked"].Value, "classic protection does not allow deletion")
	assert.Equal(t, "classic + repository + organization", properties["Rule Sources"].Value)
	assert.Equal(t, "Copilot Reviews, production", properties["Rulesets Applied"].Value)
	assert.Equal(t, 1, properties["Evaluate Rulesets"].Value)
	assert.Equal(t, 2, properties["Rulesets"].Value, "the existing count is unchanged")

	query := transport.graphql[0].Query
	assert.Contains(t, query, "rules(first: 100){nodes{type,parameters{... on PullRequestParameters{requiredApprovingReviewCount}}")
	assert.Contains(t, query, "source{__typename}", "only the kind of source, which needs no extra scope")
	assert.Contains(t, query, "refUpdateRule{pattern,")
	assert.Contains(t, query, "rulesets(includeParents: true, first: 100){totalCount,nodes{enforcement}}")
}

func TestClient_GetRepositories_NoBranchRules(t *testing.T) {
	client, _ := newTestClient(t, func(req *http.Request, gql *graphqlRequest) (int, string) {
		return http.StatusOK, `{"data":{"repo0":{"name":"plain","defaultBranchRef":{"name":"main","rules":{"nodes":[]},"refUpdateRule":null},
			"rulesets":{"totalCount":0,"nodes":[]}}}}`
	})

	repositories, err := client.GetRepositories("acme", []string{"plain"})

	require.NoError(t, err)
	properties := repo.ToProperties(repositories[0])
	assert.Equal(t, false, properties["Pull Request Rule"].Value)
	assert.Equal(t, false, properties["Force Pushes Blocked"].Value, "no classic rule means nothing is blocked")
	assert.Equal(t, "", properties["Rule Sources"].Value)
}
