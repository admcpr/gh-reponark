package github

import "time"

// currentUserResponse is the subset of the REST /user response we read.
type currentUserResponse struct {
	Login string `json:"login"`
}

// countConnection is a connection asked only for its size.
type countConnection struct {
	TotalCount int `graphql:"totalCount"`
}

// userQuery fetches the signed-in user and their organizations with enough
// about each to describe it on the account picker: name, bio or
// description, repository counts, size and age. The public repositories are
// a second, aliased connection on the same query.
type userQuery struct {
	User struct {
		Login              string
		Name               string
		Bio                string
		Url                string
		CreatedAt          time.Time       `graphql:"createdAt"`
		Repositories       countConnection `graphql:"repositories"`
		PublicRepositories countConnection `graphql:"publicRepositories: repositories(privacy: PUBLIC)"`
		Followers          countConnection `graphql:"followers"`
		Organizations      struct {
			Nodes []struct {
				Login               string
				Name                string
				Description         string
				Url                 string
				IsVerified          bool            `graphql:"isVerified"`
				ViewerCanAdminister bool            `graphql:"viewerCanAdminister"`
				CreatedAt           time.Time       `graphql:"createdAt"`
				Repositories        countConnection `graphql:"repositories"`
				PublicRepositories  countConnection `graphql:"publicRepositories: repositories(privacy: PUBLIC)"`
				MembersWithRole     countConnection `graphql:"membersWithRole"`
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

// ownerProfile is what every owner's profile query asks for, inlined into
// the organization and user roots. Repository counts are deliberately left
// out: the repositories connection's defaults differ between users and
// organizations and from the listing's, so its counts would not agree with
// the repositories loaded; those are counted from the loaded ones instead.
type ownerProfile struct {
	Login              string
	Name               string
	Url                string
	WebsiteUrl         string    `graphql:"websiteUrl"`
	Location           string    `graphql:"location"`
	Email              string    `graphql:"email"`
	TwitterUsername    string    `graphql:"twitterUsername"`
	HasSponsorsListing bool      `graphql:"hasSponsorsListing"`
	CreatedAt          time.Time `graphql:"createdAt"`
	PinnedItems        struct {
		Nodes []struct {
			Repository struct {
				Name string
			} `graphql:"... on Repository"`
		} `graphql:"nodes"`
	} `graphql:"pinnedItems(first: 6, types: REPOSITORY)"`
}

// pinned returns the names of the pinned repositories.
func (p ownerProfile) pinned() []string {
	var names []string
	for _, node := range p.PinnedItems.Nodes {
		if node.Repository.Name != "" {
			names = append(names, node.Repository.Name)
		}
	}
	return names
}

// organizationProfileQuery fetches an organization as any of its members
// sees it. Settings such as the two-factor requirement are left out: they
// need an owner, and asking for them fails the whole query for everyone else.
type organizationProfileQuery struct {
	Organization struct {
		ownerProfile
		Description                 string
		IsVerified                  bool            `graphql:"isVerified"`
		ViewerIsAMember             bool            `graphql:"viewerIsAMember"`
		ViewerCanAdminister         bool            `graphql:"viewerCanAdminister"`
		ViewerCanCreateRepositories bool            `graphql:"viewerCanCreateRepositories"`
		ViewerCanCreateTeams        bool            `graphql:"viewerCanCreateTeams"`
		ViewerCanCreateProjects     bool            `graphql:"viewerCanCreateProjects"`
		MembersWithRole             countConnection `graphql:"membersWithRole"`
		Teams                       countConnection `graphql:"teams"`
	} `graphql:"organization(login: $login)"`
}

// userProfileQuery fetches a user's public profile.
type userProfileQuery struct {
	User struct {
		ownerProfile
		Bio                     string
		Company                 string
		ViewerCanCreateProjects bool            `graphql:"viewerCanCreateProjects"`
		Followers               countConnection `graphql:"followers"`
		Following               countConnection `graphql:"following"`
	} `graphql:"user(login: $login)"`
}
