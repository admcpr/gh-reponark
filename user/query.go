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
	} `graphql:"user(login: $login)"`
}

// query {
// 	user(login:"adam7") {
// 	  login
// 	  url
// 	  repositories{
// 		totalCount
// 	  }
// 	  organizations(first:100){
// 		totalCount
// 		nodes{
// 		  login
// 		  url
// 		}
// 	  }
// 	}
//   }
