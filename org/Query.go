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
	} `graphql:"repositories(first: $first)"`
}

type OrgQuery struct {
	Organization struct {
		CommonFields
	} `graphql:"organization(login: $login)"`
}

type UserQuery struct {
	User struct {
		CommonFields
	} `graphql:"user(login: $login)"`
}
