# gh-reponark

![gh-reponark demo tour](https://github.com/admcpr/gh-reponark/releases/latest/download/demo.gif)

## What is this?

`gh reponark` is a [GitHub CLI extension](https://docs.github.com/en/github-cli/github-cli/using-github-cli-extensions) that let's you browse, review, compare and filter the config settings for every repository in your account or any organisation you belong to without clicking through forty Settings pages. I know, calm down.

Some of the almost limitless fun you can have:

- **Pick an account.** You, or any organisation you belong to.
- **Browse the repositories** with a details pane, stats measured against the rest of the organisation, and seven groups of settings — Overview, Status, Metrics, Features, Merge, Permissions and Security.
- **Switch to the matrix** to see every repository against one group of settings as a grid.
- **Filter** on any property with live match counts and a distribution of the values across your repositories, then go back to the list with only the matches showing.

## Why should I care?

You probably shouldn't but if you look after more than a handful of repositories you have probably wondered things like:

- Which repositories have no branch protection or rulesets on their default branch?
- Which ones have open vulnerability alerts — or alerts switched off entirely?
- Which are still public, or have no license, or no security policy?
- Which haven't been pushed to in a year, and which of those aren't archived yet?
- Which allow merge commits when the team standard is squash, or don't delete branches on merge?

GitHub answers each of these one repository at a time `gh reponark` answers them for the whole organisation.

This is an audit tool, it's read-only and will never update any settings. 

## What's a nark?

British slang for a [police spy or informer](https://en.m.wiktionary.org/wiki/nark). This one informs on your repositories.

## Installation
Install the [GitHub CLI](https://cli.github.com) and run:
```
gh extension install admcpr/gh-reponark
```

## Usage
``` 
gh auth login
gh reponark
```
`gh reponark --version` prints the installed release.

To look around without a GitHub account, run it against the built-in sample data:
```
gh reponark --demo
```

## Development
You need Go (the version in `go.mod`) and the [GitHub CLI](https://cli.github.com) signed in.

```
go run . --demo        # the UI against built-in sample data, no GitHub needed
go run .               # against your account
go test ./...
```

Every screen can be captured as text for a visual check without a terminal:
```
CAPTURE_DIR=captures go test -tags capture -run TestCapture ./...
```
This writes `<width>x<height>-<screen>.ans` (with colour) and `.txt` files you can diff or render.

The demo recording at the top of this file is made from `demo.tape` with [VHS](https://github.com/charmbracelet/vhs) by the release workflow and attached to each release (nothing is committed); run the "Demo" workflow by hand to re-record it for the latest release. Locally: put a `gh` shim that runs the built binary on the PATH and run `vhs demo.tape`.

Merging to `main` publishes a release: the version bumps the patch number unless the merge commit says `#minor` / `feat:` or `#major` / `feat!:`; `[skip release]` skips it.

## Built with

gh-reponark is built on [Charm](https://charm.land)'s terminal libraries:

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) A powerful little TUI framework 🏗
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) Style definitions for nice terminal layouts 👄
- [Bubbles](https://github.com/charmbracelet/bubbles) TUI components for Bubble Tea 🫧
- [VHS](https://github.com/charmbracelet/vhs) Records the demo gif from [`demo.tape`](demo.tape).

GitHub CLI access goes through [go-gh](https://github.com/cli/go-gh).

## License

[MIT](LICENSE)
