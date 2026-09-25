package repo

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestRenderTabs(t *testing.T) {
	tabs := []string{"Overview", "Status", "Metrics"}

	rendered := ansi.Strip(RenderTabs(tabs, 60, 0))

	for _, tab := range tabs {
		assert.Contains(t, rendered, tab)
	}
}

func TestRenderTabs_SingleTab(t *testing.T) {
	rendered := ansi.Strip(RenderTabs([]string{"Only"}, 40, 0))
	assert.Contains(t, rendered, "Only")
}

func TestRenderTabs_ActiveTabIsStyledDifferently(t *testing.T) {
	tabs := []string{"Overview", "Status"}

	first := RenderTabs(tabs, 60, 0)
	second := RenderTabs(tabs, 60, 1)

	assert.NotEqual(t, first, second)
	assert.Equal(t, ansi.Strip(first), ansi.Strip(second), "only styling should differ between active tabs")
}

func TestRenderTabs_ScrollsToActiveTab(t *testing.T) {
	tabs := []string{"Overview", "Status", "Metrics", "Features", "Merge", "Permissions", "Security"}

	rendered := ansi.Strip(RenderTabs(tabs, 30, 6))

	assert.Equal(t, 30, len([]rune(rendered)), "the bar fills the width exactly")
	assert.Contains(t, rendered, "Security")
	assert.NotContains(t, rendered, "Overview")
	assert.True(t, strings.HasPrefix(rendered, "‹ "), "tabs cut off on the left are marked: %q", rendered)
	assert.NotContains(t, rendered, "›", "nothing is cut off on the right of the last tab")
}

func TestRenderTabs_ActiveOutOfRange(t *testing.T) {
	rendered := ansi.Strip(RenderTabs([]string{"One", "Two"}, 20, 9))
	assert.Contains(t, rendered, "One")
}
