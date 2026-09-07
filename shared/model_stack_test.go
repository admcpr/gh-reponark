package shared

import (
	"reflect"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

type mockModel struct {
	state string
}

func (m mockModel) Init() tea.Cmd {
	return nil
}

func (m mockModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m mockModel) View() tea.View {
	return tea.NewView("")
}

func (m *mockModel) SetDimensions(width, height int) {
}

func TestStack_Push(t *testing.T) {
	stack := &ModelStack{}
	element := &mockModel{}

	stack.Push(element)
	assert.Equal(t, 1, stack.Len())
}

func TestStack_Pop(t *testing.T) {
	stack := &ModelStack{}
	element := &mockModel{}

	stack.Push(element)
	poppedElement, err := stack.Pop()
	assert.NoError(t, err)
	assert.Equal(t, element, poppedElement)
	assert.Equal(t, 0, stack.Len())

	_, err = stack.Pop()
	assert.Error(t, err)
}

func TestStack_Peek(t *testing.T) {
	stack := &ModelStack{}
	element := &mockModel{state: "initial"}

	stack.Push(element)
	peekedElement, err := stack.Peek()
	assert.NoError(t, err)
	assert.Equal(t, element, peekedElement)
	assert.Equal(t, 1, stack.Len())

	stack.Pop()
	_, err = stack.Peek()
	assert.Error(t, err)
}

type resizableModel struct {
	mockModel
	width, height int
}

func (m *resizableModel) SetDimensions(width, height int) {
	m.width = width
	m.height = height
}

func TestStack_Push_PanicsOnNil(t *testing.T) {
	stack := &ModelStack{}
	assert.Panics(t, func() { stack.Push(nil) })
}

func TestStack_Push_PanicsOnNonPointer(t *testing.T) {
	stack := &ModelStack{}
	assert.Panics(t, func() { stack.Push(mockModel{}) })
}

func TestStack_ReplaceTop(t *testing.T) {
	stack := &ModelStack{}

	err := stack.ReplaceTop(&mockModel{})
	assert.Error(t, err, "replacing on an empty stack should fail")

	first := &mockModel{state: "first"}
	stack.Push(first)

	assert.Error(t, stack.ReplaceTop(nil))
	assert.Error(t, stack.ReplaceTop(mockModel{}))

	top, _ := stack.Peek()
	assert.Equal(t, first, top, "failed replacements should leave the top untouched")

	second := &mockModel{state: "second"}
	assert.NoError(t, stack.ReplaceTop(second))
	assert.Equal(t, 1, stack.Len())

	top, _ = stack.Peek()
	assert.Equal(t, second, top)
}

func TestStack_PeekBelowTop(t *testing.T) {
	stack := &ModelStack{}

	_, err := stack.PeekBelowTop()
	assert.Error(t, err, "empty stack has nothing below top")

	first := &mockModel{state: "first"}
	stack.Push(first)
	_, err = stack.PeekBelowTop()
	assert.Error(t, err, "single element stack has nothing below top")

	second := &mockModel{state: "second"}
	stack.Push(second)
	below, err := stack.PeekBelowTop()
	assert.NoError(t, err)
	assert.Equal(t, first, below)
	assert.Equal(t, 2, stack.Len(), "PeekBelowTop must not modify the stack")
}

func TestStack_TypeOfHead(t *testing.T) {
	stack := &ModelStack{}
	assert.Nil(t, stack.TypeOfHead())

	stack.Push(&mockModel{})
	assert.Equal(t, reflect.TypeOf(&mockModel{}), stack.TypeOfHead())

	stack.Push(&resizableModel{})
	assert.Equal(t, reflect.TypeOf(&resizableModel{}), stack.TypeOfHead())
}

func TestStack_SetDimensions(t *testing.T) {
	stack := &ModelStack{}
	resizable := &resizableModel{}
	stack.Push(&stubModel{id: "not resizable"})
	stack.Push(resizable)

	stack.SetDimensions(120, 40)

	assert.Equal(t, 120, resizable.width)
	assert.Equal(t, 40, resizable.height)
}
