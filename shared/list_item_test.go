package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListItem(t *testing.T) {
	title := "title"
	desc := "desc"
	item := NewListItem(title, desc)

	assert.Equal(t, title, item.Title())
	assert.Equal(t, desc, item.Description())
	assert.Equal(t, title, item.FilterValue())
}
