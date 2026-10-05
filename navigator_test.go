package main

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

type stubModel struct{ id string }

func (m *stubModel) Init() tea.Cmd                           { return nil }
func (m *stubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m *stubModel) View() tea.View                          { return tea.NewView("") }

// mockModel satisfies tea.Model by value, so it can be pushed as a non-pointer.
type mockModel struct {
	state string
}

func (m mockModel) Init() tea.Cmd                           { return nil }
func (m mockModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m mockModel) View() tea.View                          { return tea.NewView("") }

type resizableModel struct {
	mockModel
	width, height int
}

func (m *resizableModel) SetDimensions(width, height int) {
	m.width = width
	m.height = height
}

func TestNavigator_PushAndCurrent(t *testing.T) {
	nav := NewNavigator()

	a := &stubModel{id: "a"}
	b := &stubModel{id: "b"}

	nav.Push(a)
	nav.Push(b)

	top, err := nav.Current()
	assert.NoError(t, err)
	assert.Equal(t, b, top)
	assert.Equal(t, 2, nav.Len())
}

func TestNavigator_Push_PanicsOnNil(t *testing.T) {
	nav := NewNavigator()
	assert.Panics(t, func() { nav.Push(nil) })
}

func TestNavigator_Push_PanicsOnNonPointer(t *testing.T) {
	nav := NewNavigator()
	assert.Panics(t, func() { nav.Push(mockModel{}) })
}

func TestNavigator_Pop(t *testing.T) {
	nav := NewNavigator()
	a := &stubModel{id: "a"}
	b := &stubModel{id: "b"}

	nav.Push(a)
	nav.Push(b)

	got, err := nav.Pop()
	assert.NoError(t, err)
	assert.Equal(t, b, got)
	assert.Equal(t, 1, nav.Len())

	top, _ := nav.Current()
	assert.Equal(t, a, top)
}

func TestNavigator_EmptyErrors(t *testing.T) {
	nav := NewNavigator()

	_, err := nav.Current()
	assert.Error(t, err)

	_, err = nav.Pop()
	assert.Error(t, err)

	assert.Equal(t, 0, nav.Len())
}

func TestNavigator_Screens(t *testing.T) {
	nav := NewNavigator()
	assert.Empty(t, nav.Screens())

	a := &stubModel{id: "a"}
	b := &stubModel{id: "b"}
	nav.Push(a)
	nav.Push(b)

	screens := nav.Screens()
	assert.Equal(t, []tea.Model{a, b}, screens)

	// The returned slice is a copy.
	screens[0] = &stubModel{id: "z"}
	bottom := nav.Screens()[0]
	assert.Equal(t, a, bottom)
}

func TestNavigator_SetDimensions(t *testing.T) {
	nav := NewNavigator()
	resizable := &resizableModel{}
	nav.Push(&stubModel{id: "plain"})
	nav.Push(resizable)

	nav.SetDimensions(100, 30)

	assert.Equal(t, 100, resizable.width)
	assert.Equal(t, 30, resizable.height)
}

func TestNavigator_ReplaceCurrent(t *testing.T) {
	nav := NewNavigator()
	a := &stubModel{id: "a"}
	b := &stubModel{id: "b"}
	c := &stubModel{id: "c"}

	nav.Push(a)
	nav.Push(b)
	assert.NoError(t, nav.ReplaceCurrent(c))

	top, _ := nav.Current()
	assert.Equal(t, c, top)
	assert.Equal(t, 2, nav.Len())
}

func TestNavigator_ReplaceCurrent_Empty(t *testing.T) {
	nav := NewNavigator()
	err := nav.ReplaceCurrent(&stubModel{id: "x"})
	assert.Error(t, err)
}

func TestNavigator_ReplaceCurrent_RejectsNilAndNonPointer(t *testing.T) {
	nav := NewNavigator()
	first := &stubModel{id: "a"}
	nav.Push(first)

	assert.Error(t, nav.ReplaceCurrent(nil))
	assert.Error(t, nav.ReplaceCurrent(mockModel{}))

	top, _ := nav.Current()
	assert.Equal(t, first, top, "failed replacements should leave the top untouched")
}
