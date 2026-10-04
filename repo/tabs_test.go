package repo

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestRenderSegments(t *testing.T) {
	tabs := []string{"Overview", "Status", "Metrics", "Features", "Merge", "Permissions", "Security"}

	wide := ansi.Strip(RenderSegments(tabs, 80, 0, true))
	assert.Equal(t, 80, len([]rune(wide)), "the control fills the width exactly")
	assert.True(t, strings.HasPrefix(wide, "▐Overview▌▐Status▌"), "segments are adjacent pills: %q", wide)
	assert.Contains(t, wide, "▐Security▌", "every group fits at eighty columns")

	scrolled := ansi.Strip(RenderSegments(tabs, 30, 6, true))
	assert.Equal(t, 30, len([]rune(scrolled)))
	assert.True(t, strings.HasPrefix(scrolled, "‹ "), "segments cut off on the left are marked: %q", scrolled)
	assert.Contains(t, scrolled, "▐Security▌")
	assert.NotContains(t, scrolled, "Overview")

	assert.NotEqual(t, RenderSegments(tabs, 80, 0, true), RenderSegments(tabs, 80, 0, false), "the active segment changes colour with focus")
	assert.Equal(t, wide, ansi.Strip(RenderSegments(tabs, 80, 0, false)), "but not its text")

	tiny := ansi.Strip(RenderSegments(tabs, 6, 5, true))
	assert.Equal(t, "Permi…", tiny, "a pane narrower than the active pill gets the bare title rather than a cut pill")
	assert.Equal(t, "Metrics    ", ansi.Strip(RenderSegments(tabs, 11, 2, true)), "the same when the scroll markers would push the pill off the edge")
	assert.Equal(t, "‹ ▐Metrics▌ ›", ansi.Strip(RenderSegments(tabs, 13, 2, true)))
	assert.Equal(t, strings.Repeat(" ", 20), RenderSegments(nil, 20, 0, true), "no groups still fills the line")
}
