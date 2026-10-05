package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCursor_SetClamps(t *testing.T) {
	c := Cursor{Rows: 5, Total: 10}

	c.Set(-3)
	assert.Equal(t, 0, c.Index)

	c.Set(42)
	assert.Equal(t, 9, c.Index, "the cursor stops at the last item")
	assert.Equal(t, 5, c.Offset, "and the window scrolls to show it")

	c.Move(-100)
	assert.Equal(t, 0, c.Index)
	assert.Equal(t, 0, c.Offset)
}

func TestCursor_EmptyList(t *testing.T) {
	var c Cursor
	c.Move(1)
	assert.Equal(t, Cursor{}, c, "an empty list selects nothing and never scrolls")

	c = Cursor{Index: 4, Offset: 2, Rows: 3, Total: 8}
	c.Resize(3, 0)
	assert.Equal(t, 0, c.Index)
	assert.Equal(t, 0, c.Offset)
	from, to := c.Visible()
	assert.Equal(t, 0, from)
	assert.Equal(t, 0, to)
}

func TestCursor_MoveScrollsOnlyAsNeeded(t *testing.T) {
	c := Cursor{Rows: 3, Total: 10}

	c.Move(2)
	assert.Equal(t, 0, c.Offset, "moving within the window does not scroll")

	c.Move(1)
	assert.Equal(t, 3, c.Index)
	assert.Equal(t, 1, c.Offset, "moving below the window scrolls one row")

	c.Set(0)
	assert.Equal(t, 0, c.Offset, "moving above the window scrolls back")
}

func TestCursor_PagesByRows(t *testing.T) {
	c := Cursor{Rows: 4, Total: 10}

	c.Move(c.Rows)
	assert.Equal(t, 4, c.Index)
	assert.Equal(t, 1, c.Offset)
	from, to := c.Visible()
	assert.Equal(t, []int{1, 5}, []int{from, to})

	c.Move(c.Rows)
	assert.Equal(t, 8, c.Index)
	assert.Equal(t, 5, c.Offset)

	c.Move(c.Rows)
	assert.Equal(t, 9, c.Index, "the last page stops at the end")
	assert.Equal(t, 6, c.Offset, "and the window shows a full last page")

	c.Move(-c.Rows)
	assert.Equal(t, 5, c.Index)
	assert.Equal(t, 5, c.Offset)
}

func TestCursor_ResizeKeepsIndexInView(t *testing.T) {
	c := Cursor{Rows: 10, Total: 10}
	c.Set(7)
	assert.Equal(t, 0, c.Offset, "everything fits")

	c.Resize(3, 10)
	assert.Equal(t, 7, c.Index)
	assert.Equal(t, 5, c.Offset, "a shorter window scrolls to the cursor")
	from, to := c.Visible()
	assert.Equal(t, []int{5, 8}, []int{from, to})

	c.Resize(3, 6)
	assert.Equal(t, 5, c.Index, "a shorter list pulls the cursor back")
	assert.Equal(t, 3, c.Offset)

	c.Resize(20, 6)
	assert.Equal(t, 0, c.Offset, "a taller window shows everything again")
	from, to = c.Visible()
	assert.Equal(t, []int{0, 6}, []int{from, to})
}

func TestCursor_VisibleWithNoRows(t *testing.T) {
	c := Cursor{Index: 2, Total: 5}
	from, to := c.Visible()
	assert.Equal(t, 0, from)
	assert.Equal(t, 0, to)
}
