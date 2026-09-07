package filters

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsCaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		haystack string
		needle   string
		want     bool
	}{
		{name: "identical strings", haystack: "hello", needle: "hello", want: true},
		{name: "different case", haystack: "Hello World", needle: "hello world", want: true},
		{name: "substring", haystack: "Hello World", needle: "o W", want: true},
		{name: "substring different case", haystack: "Hello World", needle: "WORLD", want: true},
		{name: "not contained", haystack: "Hello World", needle: "xyz", want: false},
		{name: "needle longer than haystack", haystack: "Hi", needle: "Hello", want: false},
		{name: "empty needle", haystack: "anything", needle: "", want: true},
		{name: "empty haystack", haystack: "", needle: "a", want: false},
		{name: "both empty", haystack: "", needle: "", want: true},
		{name: "unicode", haystack: "Ünïcödé", needle: "ünïcödé", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, containsCaseInsensitive(tt.haystack, tt.needle))
		})
	}
}
