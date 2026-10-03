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
	labels := func(bar string) string { return ansi.Strip(strings.Split(bar, "\n")[0]) }
	assert.Equal(t, labels(first), labels(second), "only the underline should differ between active tabs")
}

func TestRenderTabs_ScrollsToActiveTab(t *testing.T) {
	tabs := []string{"Overview", "Status", "Metrics", "Features", "Merge", "Permissions", "Security"}

	bar := ansi.Strip(RenderTabs(tabs, 30, 6))
	rendered := strings.Split(bar, "\n")[0]

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

func TestRenderTabs_UnderlinesActiveTab(t *testing.T) {
	lines := strings.Split(ansi.Strip(RenderTabs([]string{"One", "Two"}, 30, 1)), "\n")

	assert.Len(t, lines, 2)
	assert.Equal(t, 30, len([]rune(lines[1])), "the rule spans the pane")
	assert.Equal(t, len([]rune("  Two  ")), strings.Count(lines[1], "━"))
}
