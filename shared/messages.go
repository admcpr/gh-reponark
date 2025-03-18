package shared

import (
	"reflect"

	tea "github.com/charmbracelet/bubbletea/v2"
)

type NextMsg struct {
	ModelData any
}

type PreviousMsg struct{ Message tea.Msg }

func IsNextMsg(msg tea.Msg) bool {
	msgType := reflect.TypeOf(msg)
	if msgType.Kind() != reflect.Struct {
		return false
	}
	return msgType.Name() == "NextMsg"
}
