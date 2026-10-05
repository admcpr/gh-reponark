package ui

// Cursor is a position in a list of Total items with a window of Rows of
// them visible from Offset, kept so the position is always on screen. The
// zero value is an empty list with nothing selected.
type Cursor struct {
	Index  int // the selected item
	Offset int // the first item in view
	Total  int // how many items there are
	Rows   int // how many fit on screen at once
}

// Set selects the item at index, stopping at either end of the list, and
// scrolls the window just far enough to show it. An empty list selects 0.
func (c *Cursor) Set(index int) {
	c.Index = Max(0, Min(index, c.Total-1))
	c.Offset = ScrollOffset(c.Offset, c.Index, c.Rows, c.Total)
}

// Move selects the item delta places away from the current one.
func (c *Cursor) Move(delta int) { c.Set(c.Index + delta) }

// Resize sets the window and list sizes, for when the screen is resized or
// the list changes, keeping the selection in range and in view.
func (c *Cursor) Resize(rows, total int) {
	c.Rows, c.Total = rows, total
	c.Set(c.Index)
}

// Visible returns the half-open range [from, to) of items in view.
func (c Cursor) Visible() (from, to int) {
	from = ScrollOffset(c.Offset, c.Index, c.Rows, c.Total)
	return from, Max(from, Min(c.Total, from+c.Rows))
}
