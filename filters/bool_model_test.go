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
	trueModel := NewBoolModel("True", true, 60, 40)
	falseModel := NewBoolModel("False", false, 60, 40)

	tests := []struct {
		model BoolModel
		name  string
		_rune rune
		want  bool
	}{
		{name: "'n' should set value to false", model: trueModel, _rune: 'n', want: false},
		{name: "'N' should set value to false", model: trueModel, _rune: 'N', want: false},
		{name: "'y' should set value to true", model: falseModel, _rune: 'y', want: true},
		{name: "'Y' should set value to true", model: falseModel, _rune: 'Y', want: true},
		{name: "'x' shouldn't change false value", model: falseModel, _rune: 'x', want: false},
		{name: "'x' shouldn't change true value", model: trueModel, _rune: 'x', want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyReleaseMsg := tea.KeyPressMsg{Code: tt._rune}
			m, _ := tt.model.Update(keyReleaseMsg)

			filterBooleanModel, _ := m.(BoolModel)
			got := filterBooleanModel.Value()

			if got != tt.want {
				t.Errorf("FilterBooleanModel.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}
