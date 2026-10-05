package shared

import (
	"testing"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestNewColors(t *testing.T) {
	c := NewColors()

	assert.Equal(t, "Neon Harbour", c.name)
	assert.Equal(t, lipgloss.Color("#0b1020"), c.Background)
	assert.Equal(t, lipgloss.Color("#e6edf3"), c.Foreground)
	assert.Equal(t, lipgloss.Color("#1e2a4a"), c.Selection, "the focused band is visibly lighter than the background")
	assert.NotEqual(t, c.Selection, c.SelectionDim)

	// Green always means good, amber caution and red bad; the accent is cyan.
	assert.Equal(t, lipgloss.Color("#34d399"), c.Good)
	assert.Equal(t, lipgloss.Color("#fbbf24"), c.Warn)
	assert.Equal(t, lipgloss.Color("#fb7185"), c.Bad)
	assert.Equal(t, lipgloss.Color("#22d3ee"), c.Accent)
	assert.Equal(t, c.Blue, c.Link, "links are the palette's blue")
}

func TestPills(t *testing.T) {
	assert.Equal(t, "▐Go▌", ansi.Strip(Pill("Go", AppColors.Blue)))
	assert.Equal(t, 4, lipgloss.Width(Pill("Go", AppColors.Blue)), "a pill is its text plus two caps")
	assert.Equal(t, "▐Policy▌", ansi.Strip(TintedPill("Policy", AppColors.Good, AppColors.GoodTint)))
	assert.Equal(t, "▐Merge▌", ansi.Strip(GhostPill("Merge")))
	assert.NotEqual(t, Pill("x", AppColors.Good), TintedPill("x", AppColors.Good, AppColors.GoodTint), "solid and tinted pills differ in colour")
}

func TestBar(t *testing.T) {
	assert.Equal(t, "━━╌╌╌╌╌╌", ansi.Strip(Bar(2, 8, AccentStyle)))
	assert.Equal(t, "━━━━", ansi.Strip(Bar(9, 4, AccentStyle)), "never wider than width")
	assert.Equal(t, "╌╌╌╌", ansi.Strip(Bar(-1, 4, AccentStyle)))
	assert.Equal(t, "", Bar(2, 0, AccentStyle))
}

func TestShare(t *testing.T) {
	assert.Equal(t, 2, Share(0.25, 8))
	assert.Equal(t, 1, Share(0.01, 8), "any share at all gets a cell")
	assert.Equal(t, 0, Share(0, 8))
	assert.Equal(t, 8, Share(1.5, 8), "never more than width")
	assert.Equal(t, 0, Share(0.5, 0))
}

func TestFrameGradient(t *testing.T) {
	assert.Len(t, FrameGradient(true), 4)
	assert.NotEqual(t, FrameGradient(true), FrameGradient(false), "the unfocused ramp is darker")
	assert.Len(t, PerimeterGradient(20, 5), 20, "one colour per cell of the top edge")
}

type styleTestKeyMap struct{}

func (k styleTestKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit"))}
}

func (k styleTestKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

func TestNewHelpModel(t *testing.T) {
	tests := []struct {
		name      string
		width     int
		wantWidth int
	}{
		{name: "positive width is applied", width: 80, wantWidth: 80},
		{name: "zero width leaves default", width: 0, wantWidth: 0},
		{name: "negative width leaves default", width: -5, wantWidth: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewHelpModel(tt.width)
			assert.Equal(t, tt.wantWidth, m.Width())
			assert.Contains(t, m.View(styleTestKeyMap{}), "quit")
		})
	}
}

func TestMarquee(t *testing.T) {
	assert.Equal(t, "     ", Marquee("abcdefgh", 5, 0), "starts blank")
	assert.Equal(t, "   ab", Marquee("abcdefgh", 5, 2), "the text slides in from the right")
	assert.Equal(t, "abcde", Marquee("abcdefgh", 5, 5))
	assert.Equal(t, "h    ", Marquee("abcdefgh", 5, 12), "the gap follows the end of the text")
	assert.Equal(t, "   ab", Marquee("abcdefgh", 5, 20), "and it comes round again")
	assert.Equal(t, "   ab", Marquee("abc", 5, 2), "short text slides in too")
	assert.Equal(t, "abc  ", Marquee("abc", 5, 5))
	assert.Equal(t, "abc  ", Marquee("abc", 5, 50), "and then holds still")
}
