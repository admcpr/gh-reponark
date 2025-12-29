package shared

import (
	"errors"
	"fmt"
	"reflect"

	tea "charm.land/bubbletea/v2"
)

// Resizable models support Bubble Tea and react to window size changes.
type Resizable interface {
	tea.Model
	SetDimensions(width, height int)
}

type ModelStack struct {
	elements []tea.Model
}

func (s ModelStack) SetDimensions(width, height int) {
	for idx := range s.elements {
		if r, ok := s.elements[idx].(Resizable); ok {
			r.SetDimensions(width, height)
		}
	}
}

// Push adds an element to the top of the stack
func (s *ModelStack) Push(element tea.Model) {
	if element == nil {
		panic("ModelStack cannot push nil model")
	}

	if reflect.ValueOf(element).Kind() != reflect.Ptr {
		panic(fmt.Sprintf("ModelStack requires pointer models, got %T", element))
	}

	s.elements = append(s.elements, element)
}

// ReplaceTop swaps the element at the top of the stack.
func (s *ModelStack) ReplaceTop(element tea.Model) error {
	if len(s.elements) == 0 {
		return errors.New("stack is empty")
	}
	if element == nil {
		return errors.New("ModelStack cannot replace with nil model")
	}
	if reflect.ValueOf(element).Kind() != reflect.Ptr {
		return fmt.Errorf("ModelStack requires pointer models, got %T", element)
	}

	s.elements[len(s.elements)-1] = element
	return nil
}

// Pop removes and returns the top element of the stack
func (s *ModelStack) Pop() (tea.Model, error) {
	if len(s.elements) == 0 {
		return nil, errors.New("stack is empty")
	}
	element := s.elements[len(s.elements)-1]
	s.elements = s.elements[:len(s.elements)-1]
	return element, nil
}

// Peek returns the top element of the stack without removing it
func (s *ModelStack) Peek() (tea.Model, error) {
	if len(s.elements) == 0 {
		return nil, errors.New("stack is empty")
	}
	element := s.elements[len(s.elements)-1]
	return element, nil
}

// PeekBelowTop returns the element directly below the top of the stack.
func (s *ModelStack) PeekBelowTop() (tea.Model, error) {
	if len(s.elements) < 2 {
		return nil, errors.New("stack has fewer than 2 elements")
	}
	return s.elements[len(s.elements)-2], nil
}

func (s *ModelStack) Len() int {
	return len(s.elements)
}

func (s *ModelStack) TypeOfHead() reflect.Type {
	if len(s.elements) == 0 {
		return nil
	}
	return reflect.TypeOf(s.elements[len(s.elements)-1])
}
