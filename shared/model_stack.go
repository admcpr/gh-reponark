package shared

import (
	"errors"
	"fmt"
	"reflect"

	tea "charm.land/bubbletea/v2"
)

type ModelStack struct {
	elements []tea.Model
}

func CallSetDimensions(model tea.Model, width, height int) (tea.Model, error) {
	methodName := "SetDimensions"
	val := reflect.ValueOf(model)

	// Get to the underlying value and create pointer if needed
	if val.Kind() != reflect.Ptr {
		// Make a new pointer to a new struct value
		ptr := reflect.New(val.Type())
		// Set the value at the pointer to our original value
		ptr.Elem().Set(val)
		// Use the pointer for method calling
		val = ptr
	}

	// Find the method
	method := val.MethodByName(methodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("method %s not found", methodName)
	}

	// Call the method
	method.Call([]reflect.Value{reflect.ValueOf(width), reflect.ValueOf(height)})

	// Return the modified object
	return val.Interface().(tea.Model), nil
}

func (s ModelStack) SetDimensions(width, height int) {
	for idx := range s.elements {
		model, err := CallSetDimensions(s.elements[idx], width, height)
		if err != nil {
			// TODO: Log error
		}
		s.elements[idx] = model
	}
}

// Push adds an element to the top of the stack
func (s *ModelStack) Push(element tea.Model) {
	s.elements = append(s.elements, element)
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

func (s *ModelStack) Len() int {
	return len(s.elements)
}

func (s *ModelStack) TypeOfHead() reflect.Type {
	return reflect.TypeOf(s.elements[len(s.elements)-1])
}
