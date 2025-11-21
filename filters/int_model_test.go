package filters

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFilterIntModel(t *testing.T) {
	tests := []struct {
		name string
		from int
		to   int
	}{
		{"Test 1", 0, 10},
		{"Test 2", -5, 5},
		{"Test 3", 100, 200},
		{"Test 4", -100, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewIntModel(tt.name, tt.from, tt.to, 60, 40)

			assert.Equal(t, m.Name(), tt.name)
			assert.Equal(t, m.fromInput.Placeholder, fmt.Sprint(tt.from))
			assert.Equal(t, m.toInput.Placeholder, fmt.Sprint(tt.to))
		})
	}
}

func TestIntValidator(t *testing.T) {
	tests := []struct {
		input  string
		prompt string
		want   error
	}{
		{"123", "Test 1", nil},
		{"-5", "Test 2", nil},
		{"abc", "Test 3", fmt.Errorf("please enter an integer for the `Test 3` value")},
		{"1.23", "Test 4", fmt.Errorf("please enter an integer for the `Test 4` value")},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s: %s", tt.prompt, tt.input), func(t *testing.T) {
			got := intValidator(tt.input, tt.prompt)

			assert.Equal(t, got, tt.want, fmt.Sprintf("got %q, want %q", got, tt.want))
		})
	}
}
