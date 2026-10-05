package shared

import (
	"strings"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	"charm.land/lipgloss/v2"
)

// NewSpinner is the spinner every loading state shares: a small dot cycle in
// the accent colour. Screens start it with Tick and stop feeding it TickMsgs
// once they have loaded.
func NewSpinner() spinner.Model {
	return spinner.New(spinner.WithSpinner(spinner.MiniDot), spinner.WithStyle(AccentStyle))
}

// NewProgress is the progress bar every loading state shares: the frame's
// gradient over the house bar glyphs, without the percentage (the caller
// labels it).
func NewProgress() progress.Model {
	p := progress.New(
		progress.WithColors(AppColors.Accent, AppColors.Purple, AppColors.Pink),
		progress.WithFillCharacters('━', '╌'),
		progress.WithoutPercentage(),
	)
	p.EmptyColor = AppColors.BrightBlack
	return p
}

// LoadingWidth is how wide a centred loading block is in a pane of width.
func LoadingWidth(width int) int {
	return Max(1, Min(60, width-8))
}

// Loading lays out a loading screen: a title line, a spinner line and an
// optional bar and footnote, centred in the pane. Any empty line is skipped.
func Loading(width, height int, lines ...string) string {
	var kept []string
	for _, l := range lines {
		if l != "" {
			kept = append(kept, l)
		}
	}
	block := lipgloss.JoinVertical(lipgloss.Left, kept...)
	return lipgloss.Place(Max(1, width), Max(1, height), lipgloss.Center, lipgloss.Center, block)
}

// Marquee is a width-cell window onto text that scrolls in from the right
// edge as offset grows, then loops round with a gap so the scroll is
// seamless. Starting with a window of blanks means the text slides into
// place rather than appearing all at once; text that fits stops once it is in.
func Marquee(text string, width, offset int) string {
	if width <= 0 {
		return ""
	}
	loop := append(append(make([]rune, 0, width+len(text)+5), []rune(strings.Repeat(" ", width))...), []rune(text)...)
	loop = append(loop, []rune("     ")...)
	// Text that fits slides in once and then holds still.
	if len([]rune(text)) <= width {
		start := Min(offset, width)
		return string(loop[start : start+width])
	}
	start := offset % len(loop)
	window := make([]rune, 0, width)
	for i := 0; i < width; i++ {
		window = append(window, loop[(start+i)%len(loop)])
	}
	return string(window)
}
