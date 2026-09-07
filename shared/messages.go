package shared

import (
	tea "charm.land/bubbletea/v2"
)

// OrgKey identifies a repository owner: either an organization or the
// authenticated user.
type OrgKey struct {
	Name   string
	IsUser bool
}

// OpenOrgMsg asks the application to show the repositories owned by Key.
type OpenOrgMsg struct{ Key OrgKey }

// PreviousMsg asks the application to return to the previous screen. When
// Message is set it is delivered to that screen once it is current again.
type PreviousMsg struct{ Message tea.Msg }

// ErrorMsg reports a failure, typically from the GitHub API, so the UI can
// show it to the user instead of exiting.
type ErrorMsg struct{ Err error }
