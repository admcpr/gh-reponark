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

	for _, repoConfig := range repoConfigs {
		repoPasses := true
		for _, filter := range fm {
			if !filter.Matches(repoConfig.Properties[filter.Name()]) {
				repoPasses = false
				break
			}
		}
		if repoPasses {
			filteredConfigs = append(filteredConfigs, repoConfig)
		}
	}

	// Apply any repo-level filters that need full repo inspection
	for _, filter := range fm {
		if rf, ok := filter.(RepoFilter); ok {
			filteredConfigs = rf.FilterRepos(filteredConfigs)
		}
	}

	return filteredConfigs
}
