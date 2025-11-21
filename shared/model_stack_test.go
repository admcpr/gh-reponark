package shared

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

type mockModel struct {
	state string
}

func (m mockModel) Init() (tea.Model, tea.Cmd) {
	return m, nil
}

func (m mockModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m mockModel) View() string {
	return ""
}

func (m mockModel) SetDimensions(width, height int) {
}

func TestStack_Push(t *testing.T) {
	stack := &ModelStack{}
	element := mockModel{}

	stack.Push(element)
	assert.Equal(t, 1, stack.Len())
}

func TestStack_Pop(t *testing.T) {
	stack := &ModelStack{}
	element := mockModel{}

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
	element := mockModel{state: "initial"}

	stack.Push(element)
	peekedElement, err := stack.Peek()
	assert.NoError(t, err)
	assert.Equal(t, element, peekedElement)
	assert.Equal(t, 1, stack.Len())

	_, err = stack.Peek()
	assert.NoError(t, err)
}
