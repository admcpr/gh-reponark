package repo

import (
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
