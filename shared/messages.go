package shared

import (
	tea "github.com/charmbracelet/bubbletea/v2"
)

type NextMsg struct {
	ModelData any
}

type PreviousMsg struct{ Message tea.Msg }

type OrgKey struct {
	Name   string
	IsUser bool
}
