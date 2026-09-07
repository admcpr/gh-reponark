package shared

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

type stubModel struct{ id string }

func (m *stubModel) Init() tea.Cmd                           { return nil }
func (m *stubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m *stubModel) View() tea.View                          { return tea.NewView("") }

func TestNavigator_PushAndCurrent(t *testing.T) {
	nav := NewNavigator()

	a := &stubModel{id: "a"}
	b := &stubModel{id: "b"}

	assert.NoError(t, nav.Push(a))
	assert.NoError(t, nav.Push(b))

	top, err := nav.Current()
	assert.NoError(t, err)
	assert.Equal(t, b, top)
	assert.Equal(t, 2, nav.Len())
}

func TestNavigator_PushBlockedByValidator(t *testing.T) {
	nav := NewNavigator()
	nav.SetValidator(func(current, next tea.Model) error {
		if current.(*stubModel).id == "a" && next.(*stubModel).id == "blocked" {
			return errors.New("blocked transition")
		}
		return nil
	})

	a := &stubModel{id: "a"}
	blocked := &stubModel{id: "blocked"}

	assert.NoError(t, nav.Push(a))
	err := nav.Push(blocked)
	assert.Error(t, err)
	assert.Equal(t, 1, nav.Len())
}

func TestNavigator_ReplaceCurrent_ValidatesAgainstPrevious(t *testing.T) {
	nav := NewNavigator()
	nav.SetValidator(func(current, next tea.Model) error {
		if current.(*stubModel).id == "a" && next.(*stubModel).id == "bad" {
			return errors.New("invalid replace")
		}
		return nil
	})

	a := &stubModel{id: "a"}
	b := &stubModel{id: "b"}
	good := &stubModel{id: "good"}
	bad := &stubModel{id: "bad"}

	assert.NoError(t, nav.Push(a))
	assert.NoError(t, nav.Push(b))

	// Invalid replace should fail and keep old top
	err := nav.ReplaceCurrent(bad)
	assert.Error(t, err)
	top, _ := nav.Current()
	assert.Equal(t, b, top)

	// Valid replace should succeed
	assert.NoError(t, nav.ReplaceCurrent(good))
	top, _ = nav.Current()
	assert.Equal(t, good, top)
}

func TestNavigator_ReplaceCurrent_Empty(t *testing.T) {
	nav := NewNavigator()
	err := nav.ReplaceCurrent(&stubModel{id: "x"})
	assert.Error(t, err)
}

func TestNavigator_Pop(t *testing.T) {
	nav := NewNavigator()
	a := &stubModel{id: "a"}
	b := &stubModel{id: "b"}

	assert.NoError(t, nav.Push(a))
	assert.NoError(t, nav.Push(b))

	got, err := nav.Pop()
	assert.NoError(t, err)
	assert.Equal(t, b, got)
	assert.Equal(t, 1, nav.Len())
}

func TestNavigator_EmptyErrors(t *testing.T) {
	nav := NewNavigator()

	_, err := nav.Current()
	assert.Error(t, err)

	_, err = nav.Pop()
	assert.Error(t, err)

	assert.Equal(t, 0, nav.Len())
}

func TestNavigator_SetDimensions(t *testing.T) {
	nav := NewNavigator()
	resizable := &resizableModel{}
	assert.NoError(t, nav.Push(&stubModel{id: "plain"}))
	assert.NoError(t, nav.Push(resizable))

	nav.SetDimensions(100, 30)

	assert.Equal(t, 100, resizable.width)
	assert.Equal(t, 30, resizable.height)
}

func TestNavigator_ReplaceCurrent_SingleElementSkipsValidation(t *testing.T) {
	nav := NewNavigator()
	nav.SetValidator(func(current, next tea.Model) error {
		return errors.New("never allowed")
	})

	a := &stubModel{id: "a"}
	b := &stubModel{id: "b"}
	assert.NoError(t, nav.Push(a))

	// With only one element there is nothing below the top to validate against.
	assert.NoError(t, nav.ReplaceCurrent(b))
	top, _ := nav.Current()
	assert.Equal(t, b, top)
	assert.Equal(t, 1, nav.Len())
}

func TestNavigator_ReplaceCurrent_WithoutValidator(t *testing.T) {
	nav := NewNavigator()
	a := &stubModel{id: "a"}
	b := &stubModel{id: "b"}
	c := &stubModel{id: "c"}

	assert.NoError(t, nav.Push(a))
	assert.NoError(t, nav.Push(b))
	assert.NoError(t, nav.ReplaceCurrent(c))

	top, _ := nav.Current()
	assert.Equal(t, c, top)
	assert.Equal(t, 2, nav.Len())
}

func TestNavigator_ReplaceCurrent_RejectsNonPointer(t *testing.T) {
	nav := NewNavigator()
	assert.NoError(t, nav.Push(&stubModel{id: "a"}))

	err := nav.ReplaceCurrent(mockModel{})
	assert.Error(t, err)
}
