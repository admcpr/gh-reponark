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

func allowAll(current, next tea.Model) error { return nil }

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
