package shared

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// Fit truncates or pads a possibly styled string to exactly width cells, so
// columns built from it line up regardless of their content.
func Fit(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) > width {
		s = ansi.Truncate(s, width, "…")
	}
	return s + strings.Repeat(" ", width-lipgloss.Width(s))
}

// FitRight is Fit with the content aligned to the right edge.
func FitRight(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) > width {
		return ansi.Truncate(s, width, "…")
	}
	return strings.Repeat(" ", width-lipgloss.Width(s)) + s
}

// Lines pads or trims a block of text to exactly height lines of width cells.
func Lines(lines []string, width, height int) string {
	out := make([]string, height)
	for i := range out {
		if i < len(lines) {
			out[i] = Fit(lines[i], width)
		} else {
			out[i] = strings.Repeat(" ", width)
		}
	}
	return strings.Join(out, "\n")
}

// ScrollOffset returns the first visible row of a window of rows lines over
// total items, moved only as far as needed to keep cursor on screen.
func ScrollOffset(offset, cursor, rows, total int) int {
	if rows <= 0 || total <= rows {
		return 0
	}
	if cursor < offset {
		offset = cursor
	}
	if cursor >= offset+rows {
		offset = cursor - rows + 1
	}
	return Max(0, Min(offset, total-rows))
}

// resetSequence is how lipgloss ends a styled span. It clears the background
// too, so HighlightRow puts the row's background back after each one.
const resetSequence = "\x1b[m"

// HighlightRow fits a styled row to width and tints its whole background, so
// the selected row reads as one band across the pane. The tint is stronger
// when the row's pane has focus. Spans with a background of their own, such
// as a focused matrix cell, keep it.
func HighlightRow(row string, width int, focused bool) string {
	bg := AppColors.SelectionDim
	if focused {
		bg = AppColors.Selection
	}
	set := backgroundSequence(bg)
	row = strings.ReplaceAll(Fit(row, width), resetSequence, resetSequence+set)
	return set + row + resetSequence
}

// backgroundSequence is the escape sequence that sets a 24-bit background.
func backgroundSequence(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r>>8, g>>8, b>>8)
}
