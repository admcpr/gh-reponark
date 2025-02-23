package user

type Query struct {
	User struct {
		Id           string
		Login        string
		Url          string
		Repositories []struct {
			Name string
			Url  string
		} `graphql:"repositories(first: $first) { nodes }"`
		Organizations []struct {
			Login string
			Url   string
		} `graphql:"organizations(first: $first) { nodes }"`
	} `graphql:"user(login: $login)"`
}
