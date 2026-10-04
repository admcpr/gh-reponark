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

func TestWrap(t *testing.T) {
	lines := Wrap(GoodStyle, "one two three four", 9)

	assert.Equal(t, []string{"one two", "three", "four"}, stripAll(lines))
	assert.NotEqual(t, "one two", lines[0], "every line is rendered in the style")
	assert.Equal(t, []string{""}, stripAll(Wrap(GoodStyle, "", 9)))
}

func TestJoinFit(t *testing.T) {
	items := []string{"aa", "bbb", "c"}

	line, dropped := JoinFit(items, " · ", 20)
	assert.Equal(t, "aa · bbb · c", line)
	assert.Equal(t, 0, dropped)

	line, dropped = JoinFit(items, " · ", 9)
	assert.Equal(t, "aa · bbb", line, "an item that does not fit ends the line")
	assert.Equal(t, 1, dropped)

	line, dropped = JoinFit(items, " · ", 1)
	assert.Equal(t, "", line)
	assert.Equal(t, 3, dropped)

	line, dropped = JoinFit(nil, " · ", 10)
	assert.Equal(t, "", line)
	assert.Equal(t, 0, dropped)
}

func stripAll(lines []string) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = ansi.Strip(l)
	}
	return out
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
	assert.Equal(t, "\x1b[48;2;30;42;74m", backgroundSequence(AppColors.Selection))
}
