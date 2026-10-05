package main

import (
	"errors"
	"fmt"
	"reflect"

	tea "charm.land/bubbletea/v2"
)

// Navigator is the stack of open screens: the first pushed is the root and
// the last is the one currently shown. Screens must be pointers so that
// SetDimensions reaches the model the stack holds.
type Navigator struct {
	screens []tea.Model
}

func NewNavigator() Navigator {
	return Navigator{}
}

// SetDimensions resizes every open screen that can be resized.
func (n *Navigator) SetDimensions(width, height int) {
	for _, screen := range n.screens {
		if r, ok := screen.(interface{ SetDimensions(width, height int) }); ok {
			r.SetDimensions(width, height)
		}
	}
}

// Push makes m the current screen.
func (n *Navigator) Push(m tea.Model) {
	if m == nil {
		panic("Navigator cannot push nil model")
	}
	if reflect.ValueOf(m).Kind() != reflect.Ptr {
		panic(fmt.Sprintf("Navigator requires pointer models, got %T", m))
	}
	n.screens = append(n.screens, m)
}

// Pop removes and returns the current screen.
func (n *Navigator) Pop() (tea.Model, error) {
	if len(n.screens) == 0 {
		return nil, errors.New("stack is empty")
	}
	m := n.screens[len(n.screens)-1]
	n.screens = n.screens[:len(n.screens)-1]
	return m, nil
}

// Current returns the screen on top without removing it.
func (n *Navigator) Current() (tea.Model, error) {
	if len(n.screens) == 0 {
		return nil, errors.New("stack is empty")
	}
	return n.screens[len(n.screens)-1], nil
}

// Screens returns every open screen, from the first to the current one.
func (n *Navigator) Screens() []tea.Model {
	return append([]tea.Model(nil), n.screens...)
}

func (n *Navigator) Len() int {
	return len(n.screens)
}

// ReplaceCurrent swaps the top of stack with the given model.
func (n *Navigator) ReplaceCurrent(m tea.Model) error {
	if len(n.screens) == 0 {
		return errors.New("stack is empty")
	}
	if m == nil {
		return errors.New("Navigator cannot replace with nil model")
	}
	if reflect.ValueOf(m).Kind() != reflect.Ptr {
		return fmt.Errorf("Navigator requires pointer models, got %T", m)
	}
	n.screens[len(n.screens)-1] = m
	return nil
}
