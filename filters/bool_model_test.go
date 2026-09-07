package filters

import (
	"strings"
	"testing"

	"gh-reponark/shared"

	tea "charm.land/bubbletea/v2"
)

func TestNewFilterBoolModel(t *testing.T) {
	tests := []struct {
		name  string
		value bool
	}{
		{"Test 1", true},
		{"Test 2", false},
		{"Test 3", true},
		{"Test 4", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewBoolModel(tt.name, tt.value, 60, 40)

			if m.Name() != tt.name {
				t.Errorf("got %q, want %q", m.Name(), tt.name)
			}

			if m.Value() != tt.value {
				t.Errorf("got %t, want %t", m.Value(), tt.value)
			}
		})
	}
}

func TestFilterBoolModel_Update(t *testing.T) {
	tests := []struct {
		name    string
		initial bool
		_rune   rune
		want    bool
	}{
		{name: "'n' should set value to false", initial: true, _rune: 'n', want: false},
		{name: "'N' should set value to false", initial: true, _rune: 'N', want: false},
		{name: "'y' should set value to true", initial: false, _rune: 'y', want: true},
		{name: "'Y' should set value to true", initial: false, _rune: 'Y', want: true},
		{name: "'x' shouldn't change false value", initial: false, _rune: 'x', want: false},
		{name: "'x' shouldn't change true value", initial: true, _rune: 'x', want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewBoolModel(tt.name, tt.initial, 60, 40)
			keyReleaseMsg := tea.KeyPressMsg{Code: tt._rune}
			m, _ := model.Update(keyReleaseMsg)

			filterBooleanModel := m.(*BoolModel)
			got := filterBooleanModel.Value()

			if got != tt.want {
				t.Errorf("FilterBooleanModel.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBoolModel_SetDimensions(t *testing.T) {
	m := NewBoolModel("Test", false, 60, 40)

	m.SetDimensions(100, 50)

	if m.width != 100 || m.height != 50 {
		t.Errorf("dimensions = %dx%d, want 100x50", m.width, m.height)
	}
}

func TestBoolModel_Init(t *testing.T) {
	m := NewBoolModel("Test", false, 60, 40)
	if m.Init() == nil {
		t.Error("expected a non-nil command from Init")
	}
}

func TestBoolModel_Update_ArrowKeysToggle(t *testing.T) {
	tests := []struct {
		name    string
		initial bool
		code    rune
		want    bool
	}{
		{name: "left toggles true to false", initial: true, code: tea.KeyLeft, want: false},
		{name: "left toggles false to true", initial: false, code: tea.KeyLeft, want: true},
		{name: "right toggles true to false", initial: true, code: tea.KeyRight, want: false},
		{name: "right toggles false to true", initial: false, code: tea.KeyRight, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewBoolModel("Test", tt.initial, 60, 40)
			updated, _ := m.Update(tea.KeyPressMsg{Code: tt.code})

			if got := updated.(*BoolModel).Value(); got != tt.want {
				t.Errorf("Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBoolModel_Update_EnterSendsFilter(t *testing.T) {
	m := NewBoolModel("Is Archived", true, 60, 40)

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a command on enter")
	}

	prev, ok := cmd().(shared.PreviousMsg)
	if !ok {
		t.Fatalf("expected PreviousMsg, got %T", cmd())
	}

	filter, ok := prev.Message.(BoolFilter)
	if !ok {
		t.Fatalf("expected BoolFilter, got %T", prev.Message)
	}
	if filter.Name() != "Is Archived" || !filter.Value {
		t.Errorf("unexpected filter %+v", filter)
	}
}

func TestBoolModel_Update_EscGoesBack(t *testing.T) {
	m := NewBoolModel("Is Archived", true, 60, 40)

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected a command on esc")
	}

	prev, ok := cmd().(shared.PreviousMsg)
	if !ok {
		t.Fatalf("expected PreviousMsg, got %T", cmd())
	}
	if prev.Message != nil {
		t.Errorf("expected an empty PreviousMsg, got %+v", prev.Message)
	}
}

func TestBoolModel_Update_IgnoresNonKeyMessages(t *testing.T) {
	m := NewBoolModel("Test", true, 60, 40)

	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 10, Height: 10})

	if cmd != nil {
		t.Error("expected no command")
	}
	if !updated.(*BoolModel).Value() {
		t.Error("value should be unchanged")
	}
}

func TestBoolModel_View(t *testing.T) {
	m := NewBoolModel("Is Archived", true, 100, 40)

	content := plain(m.View())

	for _, want := range []string{"Is Archived", "Yes", "No"} {
		if !strings.Contains(content, want) {
			t.Errorf("view should contain %q", want)
		}
	}
}
