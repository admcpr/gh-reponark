# gh-reponark

# So what is this?
It's a GitHub CLI extension that pulls the config for all repos in a GitHub organization and displays them in a TUI. You can also filter for repos based on their config, so you can do exciting things like find repos that don't have CODEOWNERS or don't have branch protection enabled. I know, calm down.

# And why should I care?
You probably shouldn't but if you're a GitHub user with a lot of repos or a GitHub organization admin and you want a way to view details of your repos and find ones which aren't configured the way you want them then you might.

# What's a nark?
It's British slang for a [police spy or informer](https://en.m.wiktionary.org/wiki/nark).

## Installation
Install [GitHub CLI]() and run the following command:
```
gh extension install admcpr/gh-reponark
```

## Usage
``` 
gh login
gh hubbub
```

## Development
### Prerequisites

### Build
```
go build .
```

### Run
```
go run .
```

## Thanks
Built using [Bubble Tea](https://charm.land/bubbletea), [Bubbles](https://charm.land/bubbles), and [Lip Gloss](https://charm.land/lipgloss) from [Charm](https://charm.land).