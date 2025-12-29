package filters

import (
	"testing"

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
