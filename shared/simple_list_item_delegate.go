package shared

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

// SimpleItem is a plain string list entry.
type SimpleItem string

func (i SimpleItem) FilterValue() string { return "" }
func (i SimpleItem) String() string      { return string(i) }

// SimpleItemDelegate renders single-line list items. Any list.Item that also
// implements fmt.Stringer can be rendered with it, so callers may attach their
// own data to an item instead of looking it up by index.
type SimpleItemDelegate struct{}

func (d SimpleItemDelegate) Height() int                             { return 1 }
func (d SimpleItemDelegate) Spacing() int                            { return 1 }
func (d SimpleItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d SimpleItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(fmt.Stringer)
	if !ok {
		fmt.Fprintf(w, "invalid item type: %T", listItem)
		return
	}

	str := i.String()

	renderFunction := ItemStyle.Render
	if index == m.Index() {
		renderFunction = SelectedItemStyle.Render
	}

	fmt.Fprint(w, renderFunction(str))
}
