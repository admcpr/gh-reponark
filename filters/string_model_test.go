package filters

import (
	"strings"
	"testing"

	"gh-reponark/shared"

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

func TestStringModel_Init(t *testing.T) {
	m := NewStringModel("Title", "", 40, 10)
	if m.Init() == nil {
		t.Fatal("expected a non-nil command from Init")
	}
}

func TestStringModel_InputIsFocused(t *testing.T) {
	m := NewStringModel("Title", "", 40, 10)
	if !m.input.Focused() {
		t.Fatal("expected the input to be focused")
	}
}

func TestStringModel_Update_TypingAppendsToValue(t *testing.T) {
	m := NewStringModel("Title", "go", 40, 10)

	m.Update(tea.KeyPressMsg{Code: 'l', Text: "l"})
	m.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})

	if m.Value() != "gola" {
		t.Fatalf("value = %q, want %q", m.Value(), "gola")
	}
}

func TestStringModel_Update_EnterSendsFilter(t *testing.T) {
	m := NewStringModel("Primary Language", "rust", 40, 10)

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command on enter")
	}

	prev, ok := cmd().(shared.PreviousMsg)
	if !ok {
		t.Fatalf("expected PreviousMsg, got %T", cmd())
	}

	filter, ok := prev.Message.(StringFilter)
	if !ok {
		t.Fatalf("expected StringFilter, got %T", prev.Message)
	}
	if filter.Name() != "Primary Language" || filter.Value() != "rust" {
		t.Fatalf("unexpected filter %+v", filter)
	}
}

func TestStringModel_Update_EscGoesBack(t *testing.T) {
	m := NewStringModel("Title", "", 40, 10)

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected command on esc")
	}

	prev, ok := cmd().(shared.PreviousMsg)
	if !ok {
		t.Fatalf("expected PreviousMsg, got %T", cmd())
	}
	if prev.Message != nil {
		t.Fatalf("expected an empty PreviousMsg, got %+v", prev.Message)
	}
}

func TestStringModel_View(t *testing.T) {
	m := NewStringModel("Primary Language", "rust", 100, 10)

	content := plain(m.View())
	for _, want := range []string{"Primary Language", "Value:", "rust"} {
		if !strings.Contains(content, want) {
			t.Errorf("view should contain %q", want)
		}
	}
}
