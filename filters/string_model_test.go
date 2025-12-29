package filters

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestStringModel_InitAndValue(t *testing.T) {
	m := NewStringModel("Title", "hello", 40, 10)

	if m.Name() != "Title" {
		t.Fatalf("name = %s, want Title", m.Name())
	}
	if m.Value() != "hello" {
		t.Fatalf("value = %s, want hello", m.Value())
	}

	m.SetDimensions(50, 20)
	if m.width != 50 || m.height != 20 {
		t.Fatalf("dimensions not set")
	}

	// enter should send add message
	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, cmd := m.Update(msg)
	if cmd == nil {
		t.Fatalf("expected command on enter")
	}
}
