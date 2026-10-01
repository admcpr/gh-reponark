package shared

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestFit(t *testing.T) {
	assert.Equal(t, "ab   ", Fit("ab", 5))
	assert.Equal(t, "abcd…", Fit("abcdefgh", 5))
	assert.Equal(t, "", Fit("abc", 0))
	assert.Equal(t, "ok   ", ansi.Strip(Fit(GoodStyle.Render("ok"), 5)), "styling does not count towards width")
}

func TestFitRight(t *testing.T) {
	assert.Equal(t, "   42", FitRight("42", 5))
	assert.Equal(t, "abcd…", FitRight("abcdefgh", 5))
}

func TestLines(t *testing.T) {
	assert.Equal(t, "a  \nbc…\n   ", Lines([]string{"a", "bcdef"}, 3, 3), "short blocks are padded")
	assert.Equal(t, "a  \nb  ", Lines([]string{"a", "b", "c"}, 3, 2), "tall blocks are trimmed")
}

func TestScrollOffset(t *testing.T) {
	tests := []struct {
		name                        string
		offset, cursor, rows, total int
		want                        int
	}{
		{name: "fits on screen", offset: 3, cursor: 5, rows: 10, total: 8, want: 0},
		{name: "cursor visible stays put", offset: 2, cursor: 5, rows: 5, total: 20, want: 2},
		{name: "cursor below scrolls down", offset: 0, cursor: 7, rows: 5, total: 20, want: 3},
		{name: "cursor above scrolls up", offset: 10, cursor: 4, rows: 5, total: 20, want: 4},
		{name: "never past the end", offset: 18, cursor: 19, rows: 5, total: 20, want: 15},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ScrollOffset(tt.offset, tt.cursor, tt.rows, tt.total))
		})
	}
}

func TestHighlightRow(t *testing.T) {
	row := GoodStyle.Render("ok") + " plain"

	focused := HighlightRow(row, 12, true)
	unfocused := HighlightRow(row, 12, false)

	assert.Equal(t, "ok plain    ", ansi.Strip(focused), "the row is padded to the full width")
	assert.NotEqual(t, focused, unfocused, "focus changes the tint")

	set := backgroundSequence(AppColors.Selection)
	assert.True(t, strings.HasPrefix(focused, set))
	assert.Equal(t, strings.Count(focused, resetSequence), strings.Count(focused, set),
		"every reset inside the row is followed by the background again, except the last")
}

func TestBackgroundSequence(t *testing.T) {
	assert.Equal(t, "\x1b[48;2;18;42;51m", backgroundSequence(AppColors.Selection))
}
