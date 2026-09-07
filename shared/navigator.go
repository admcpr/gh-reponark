package shared

import (
	tea "charm.land/bubbletea/v2"
)

// Navigator wraps ModelStack to provide clearer navigation semantics.
type Navigator struct {
	stack ModelStack
}

func NewNavigator() Navigator {
	return Navigator{stack: ModelStack{}}
}

func (n *Navigator) SetDimensions(width, height int) {
	n.stack.SetDimensions(width, height)
}

// Push makes m the current screen.
func (n *Navigator) Push(m tea.Model) {
	n.stack.Push(m)
}

func (n *Navigator) Pop() (tea.Model, error) {
	return n.stack.Pop()
}

func (n *Navigator) Current() (tea.Model, error) {
	return n.stack.Peek()
}

func (n *Navigator) Len() int {
	return n.stack.Len()
}

// ReplaceCurrent swaps the top of stack with the given model.
func (n *Navigator) ReplaceCurrent(m tea.Model) error {
	return n.stack.ReplaceTop(m)
}
