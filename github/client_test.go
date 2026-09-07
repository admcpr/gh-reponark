package github

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

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
		Status:     http.StatusText(status),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
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

// repositoriesResponse builds a repositories connection response under root.
func repositoriesResponse(root string, totalCount int, hasNextPage bool, endCursor string, nodes ...string) string {
	return `{"data":{"` + root + `":{"repositories":{` +
		`"totalCount":` + itoa(totalCount) + `,` +
		`"pageInfo":{"hasNextPage":` + boolString(hasNextPage) + `,"endCursor":"` + endCursor + `"},` +
		`"nodes":[` + strings.Join(nodes, ",") + `]}}}}`
}

func itoa(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}

func boolString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func TestClient_CurrentUser(t *testing.T) {
	client, transport := newTestClient(t, func(req *http.Request, gql *graphqlRequest) (int, string) {
		switch {
		case req.URL.Path == "/user":
			return http.StatusOK, `{"login":"octocat","name":"The Octocat"}`
		case isOperation(gql, "User"):
			return http.StatusOK, `{"data":{"user":{
				"login":"octocat",
				"url":"https://github.com/octocat",
				"organizations":{"nodes":[
					{"login":"acme","url":"https://github.com/acme"},
					{"login":"globex","url":"https://github.com/globex"}
				]}}}}`
		default:
			return http.StatusNotFound, `{"message":"unexpected request"}`
		}
	})

	user, err := client.CurrentUser()

	require.NoError(t, err)
	assert.Equal(t, User{
		Login: "octocat",
		Url:   "https://github.com/octocat",
		Organizations: []Organization{
			{Login: "acme", Url: "https://github.com/acme"},
			{Login: "globex", Url: "https://github.com/globex"},
		},
	}, user)

	require.Len(t, transport.requests, 2)
	for _, req := range transport.requests {
		assert.Contains(t, req.Header.Get("Authorization"), "test-token")
	}
	require.Len(t, transport.graphql, 1)
	assert.Equal(t, "octocat", transport.graphql[0].Variables["login"])
	assert.EqualValues(t, pageSize, transport.graphql[0].Variables["first"])
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
				return http.StatusOK, repositoriesResponse(tt.wantRoot, 2, false, "cursor-2",
					`{"name":"widgets","nameWithOwner":"acme/widgets","isArchived":true,"stargazerCount":42,
					  "createdAt":"2024-03-04T05:06:07Z","primaryLanguage":{"name":"Go"},
					  "issues":{"totalCount":3},"defaultBranchRef":{"name":"main"}}`,
					`{"name":"gadgets","nameWithOwner":"acme/gadgets","isPrivate":true}`,
				)
			})

			page, err := client.ListRepositories("acme", tt.isUser, "")

			require.NoError(t, err)
			assert.Equal(t, 2, page.TotalCount)
			assert.False(t, page.HasNextPage)
			assert.Equal(t, "cursor-2", page.EndCursor)
			require.Len(t, page.Repositories, 2)

			widgets := page.Repositories[0]
			assert.Equal(t, "widgets", widgets.Name)
			assert.Equal(t, "acme/widgets", widgets.NameWithOwner)
			assert.True(t, widgets.IsArchived)
			assert.Equal(t, 42, widgets.StargazerCount)
			assert.Equal(t, time.Date(2024, 3, 4, 5, 6, 7, 0, time.UTC), widgets.CreatedAt)
			assert.Equal(t, "Go", widgets.PrimaryLanguage.Name)
			assert.Equal(t, 3, widgets.Issues.TotalCount)
			assert.Equal(t, "main", widgets.DefaultBranchRef.Name)

			gadgets := page.Repositories[1]
			assert.Equal(t, "gadgets", gadgets.Name)
			assert.True(t, gadgets.IsPrivate)

			require.Len(t, transport.graphql, 1)
			query := transport.graphql[0].Query
			assert.Contains(t, query, tt.wantRoot+"(login: $login)")
			assert.Contains(t, query, "repositories(first: $first, after: $after, affiliations: OWNER)")
			assert.Contains(t, query, "totalCount")
			assert.Contains(t, query, "pageInfo{hasNextPage,endCursor}")
			assert.Contains(t, query, "isArchived", "the full repository configuration should be requested")
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
			return http.StatusOK, repositoriesResponse("organization", 3, true, "cursor-2",
				`{"name":"alpha"}`, `{"name":"bravo"}`)
		}
		return http.StatusOK, repositoriesResponse("organization", 3, false, "cursor-3",
			`{"name":"charlie"}`)
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
		return http.StatusOK, repositoriesResponse("organization", 0, false, "")
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

func TestClient_ImplementsService(t *testing.T) {
	var svc Service = &Client{}
	assert.NotNil(t, svc)
}
