package org

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrgQuery_GetCommonFields(t *testing.T) {
	query := OrgQuery{}
	query.Organization.Id = "org-id"
	query.Organization.Login = "acme"
	query.Organization.Url = "https://github.com/acme"
	query.Organization.Repositories.Nodes = append(query.Organization.Repositories.Nodes,
		struct {
			Name string
			Url  string
		}{Name: "widgets", Url: "https://github.com/acme/widgets"},
	)

	fields := query.GetCommonFields()

	assert.Equal(t, "org-id", fields.Id)
	assert.Equal(t, "acme", fields.Login)
	assert.Equal(t, "https://github.com/acme", fields.Url)
	assert.Len(t, fields.Repositories.Nodes, 1)
	assert.Equal(t, "widgets", fields.Repositories.Nodes[0].Name)
}

func TestUserQuery_GetCommonFields(t *testing.T) {
	query := UserQuery{}
	query.User.Id = "user-id"
	query.User.Login = "octocat"
	query.User.Url = "https://github.com/octocat"

	fields := query.GetCommonFields()

	assert.Equal(t, "user-id", fields.Id)
	assert.Equal(t, "octocat", fields.Login)
	assert.Equal(t, "https://github.com/octocat", fields.Url)
	assert.Empty(t, fields.Repositories.Nodes)
}

func TestQueriesImplementQuery(t *testing.T) {
	var _ Query = OrgQuery{}
	var _ Query = UserQuery{}
}
