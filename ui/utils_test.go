package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHalf(t *testing.T) {
	tests := []struct {
		width    int
		expected int
	}{
		{width: 0, expected: 0},
		{width: 1, expected: 0},
		{width: 2, expected: 1},
		{width: 3, expected: 1},
		{width: 4, expected: 2},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run("TestHalf", func(t *testing.T) {
			got := Half(tt.width)
			assert.Equal(t, tt.expected, got)
		})
	}
}
func TestQuarter(t *testing.T) {
	tests := []struct {
		width    int
		expected int
	}{
		{width: 0, expected: 0},
		{width: 1, expected: 0},
		{width: 2, expected: 0},
		{width: 3, expected: 0},
		{width: 4, expected: 1},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run("TestQuarter", func(t *testing.T) {
			got := Quarter(tt.width)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		a        int
		b        int
		expected int
	}{
		{a: 0, b: 0, expected: 0},
		{a: 1, b: 0, expected: 1},
		{a: 2, b: 3, expected: 3},
		{a: -1, b: -5, expected: -1},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run("TestMax", func(t *testing.T) {
			got := Max(tt.a, tt.b)
			assert.Equal(t, tt.expected, got)
		})
	}
}
func TestPlural(t *testing.T) {
	assert.Equal(t, "repo", Plural(1, "repo", "repos"))
	assert.Equal(t, "repos", Plural(0, "repo", "repos"))
	assert.Equal(t, "repos", Plural(3, "repo", "repos"))
	assert.Equal(t, "follower", Plural(1, "follower", "followers"))
}

func TestMin(t *testing.T) {
	tests := []struct {
		a        int
		b        int
		expected int
	}{
		{a: 0, b: 0, expected: 0},
		{a: 1, b: 0, expected: 0},
		{a: 2, b: 3, expected: 2},
		{a: -1, b: -5, expected: -5},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run("TestMin", func(t *testing.T) {
			got := Min(tt.a, tt.b)
			assert.Equal(t, tt.expected, got)
		})
	}
}
