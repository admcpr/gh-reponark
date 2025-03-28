package org

type CommonFields struct {
	Id           string
	Login        string
	Url          string
	Repositories struct {
		Nodes []struct {
			Name string
			Url  string
		} `graphql:"nodes"`
	} `graphql:"repositories(first: $first, affiliations:OWNER)"`
}

type Query interface {
	GetCommonFields() CommonFields
}

type OrgQuery struct {
	Organization struct {
		CommonFields
	} `graphql:"organization(login: $login)"`
}

func (q OrgQuery) GetCommonFields() CommonFields {
	return q.Organization.CommonFields
}

type UserQuery struct {
	User struct {
		CommonFields
	} `graphql:"user(login: $login)"`
}

func (q UserQuery) GetCommonFields() CommonFields {
	return q.User.CommonFields
}
