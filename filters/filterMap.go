package filters

import (
	"gh-reponark/repo"
)

type FilterMap map[string]Filter

func (fm FilterMap) FilterRepos(repoConfigs []repo.RepoConfig) []repo.RepoConfig {
	if fm == nil {
		return repoConfigs
	}

	filteredConfigs := []repo.RepoConfig{}

	for _, repo := range repoConfigs {
		matches := true
		for _, filter := range fm {
			if !filter.Matches(repo.Properties[filter.Name()]) {
				matches = false
				break
			}
		}
		if matches {
			filteredConfigs = append(filteredConfigs, repo)
		}
	}

	return filteredConfigs
}
