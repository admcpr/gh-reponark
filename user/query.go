package user

type Query struct {
	User struct {
		Id           string
		Login        string
		Url          string
		Repositories struct {
			Nodes []struct {
				Name string
				Url  string
			} `graphql:"nodes"`
		} `graphql:"repositories(first: $first)"`
		Organizations struct {
			Nodes []struct {
				Login string
				Url   string
			} `graphql:"nodes"`
		} `graphql:"organizations(first: $first)"`
	} `graphql:"user(login: $login)"`
}
