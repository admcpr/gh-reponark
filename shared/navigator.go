package shared

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

// TransitionValidator validates whether a navigation from current -> next is allowed.
type TransitionValidator func(current, next tea.Model) error

// Navigator wraps ModelStack to provide clearer navigation semantics.
type Navigator struct {
	stack    ModelStack
	validate TransitionValidator
}

func NewNavigator() Navigator {
	return Navigator{stack: ModelStack{}}
}

func (n *Navigator) SetValidator(v TransitionValidator) {
	n.validate = v
}

func (n *Navigator) SetDimensions(width, height int) {
	n.stack.SetDimensions(width, height)
}

func (n *Navigator) Push(m tea.Model) error {
	if n.validate != nil && n.stack.Len() > 0 {
		current, _ := n.stack.Peek()
		if err := n.validate(current, m); err != nil {
			return err
		}
	}
	n.stack.Push(m)
	return nil
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
	if n.stack.Len() == 0 {
		return fmt.Errorf("navigator is empty")
	}
	if n.validate != nil && n.stack.Len() > 1 {
		prev := n.stack.elements[n.stack.Len()-2]
		if err := n.validate(prev, m); err != nil {
			return err
		}
	}
	n.stack.elements[n.stack.Len()-1] = m
	return nil
}
