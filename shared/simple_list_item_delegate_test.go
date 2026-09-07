package shared

import (
	"bytes"
	"testing"

	"charm.land/bubbles/v2/list"
	"github.com/stretchr/testify/assert"
)

func TestSimpleItem_FilterValue(t *testing.T) {
	assert.Equal(t, "", SimpleItem("anything").FilterValue())
}

func TestSimpleItemDelegate_Dimensions(t *testing.T) {
	delegate := SimpleItemDelegate{}

	assert.Equal(t, 1, delegate.Height())
	assert.Equal(t, 1, delegate.Spacing())
	assert.Nil(t, delegate.Update(nil, nil))
}

func TestSimpleItemDelegate_Render(t *testing.T) {
	items := []list.Item{SimpleItem("first"), SimpleItem("second")}
	m := list.New(items, SimpleItemDelegate{}, 40, 10)

	tests := []struct {
		name  string
		index int
		item  list.Item
		want  string
	}{
		{name: "selected item", index: 0, item: items[0], want: "first"},
		{name: "unselected item", index: 1, item: items[1], want: "second"},
		{name: "invalid item type", index: 0, item: NewListItem("title", "desc"), want: "invalid item type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			SimpleItemDelegate{}.Render(&buf, m, tt.index, tt.item)
			assert.Contains(t, buf.String(), tt.want)
		})
	}
}

func TestSimpleItemDelegate_RenderSelectedDiffersFromUnselected(t *testing.T) {
	items := []list.Item{SimpleItem("same"), SimpleItem("same")}
	m := list.New(items, SimpleItemDelegate{}, 40, 10)

	var selected, unselected bytes.Buffer
	SimpleItemDelegate{}.Render(&selected, m, 0, items[0])
	SimpleItemDelegate{}.Render(&unselected, m, 1, items[1])

	assert.NotEqual(t, selected.String(), unselected.String())
}
