package shared

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

type SimpleItem string

func (i SimpleItem) FilterValue() string { return "" }

type SimpleItemDelegate struct{}

func (d SimpleItemDelegate) Height() int                             { return 1 }
func (d SimpleItemDelegate) Spacing() int                            { return 1 }
func (d SimpleItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d SimpleItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(SimpleItem)
	if !ok {
		fmt.Fprintf(w, "invalid item type: %T", listItem)
		return
	}

	str := string(i)

	renderFunction := ItemStyle.Render
	if index == m.Index() {
		renderFunction = SelectedItemStyle.Render
	}

	fmt.Fprint(w, renderFunction(str))
}
