package shared

import tea "github.com/charmbracelet/bubbletea/v2"

type NextMsg[T any] struct{ ModelData T }
type PreviousMsg struct{ Message tea.Msg }
