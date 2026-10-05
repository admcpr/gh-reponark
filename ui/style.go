package ui

import (
	"image/color"
	"math"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/lipgloss/v2"
)

type colors struct {
	name         string
	Blue         color.Color
	Purple       color.Color
	BrightBlack  color.Color
	BrightYellow color.Color
	BrightWhite  color.Color
	Background   color.Color
	Foreground   color.Color
	// Dim is for secondary text: headings, hints and absent values.
	Dim color.Color
	// Surface lifts small elements such as chips off the background;
	// BrightBlack is the empty track of every bar.
	Surface color.Color
	// Selection tints the selected row in the focused pane; SelectionDim
	// marks it more faintly when another pane has focus.
	Selection    color.Color
	SelectionDim color.Color

	// Accent is the one colour that means "this has focus": the active tab,
	// the focused property and the border of the focused card.
	Accent color.Color
	// Semantic colours: Good is green, Warn is amber, Bad is red. They are
	// used the same way everywhere, so a colour alone tells the state.
	Good color.Color
	Warn color.Color
	Bad  color.Color
	// Tints are dark versions of the semantic colours, for the background
	// of a toggle cell or a chip whose text is the full colour on top.
	GoodTint   color.Color
	WarnTint   color.Color
	BadTint    color.Color
	AccentTint color.Color
	// Link colours URLs; it is the same blue as Blue. Pink rounds out the
	// accent set for the stat tiles.
	Link color.Color
	Pink color.Color
}

// NewColors returns the "Neon Harbour" palette: a deep navy background with
// saturated emerald, amber, rose and cyan on top. Every text colour clears
// 4.5:1 against the background it is drawn on, including the tints and the
// selection band.
func NewColors() colors {
	blue := lipgloss.Color("#60a5fa")
	return colors{
		name:         "Neon Harbour",
		Blue:         blue,
		Purple:       lipgloss.Color("#c084fc"),
		BrightBlack:  lipgloss.Color("#3b4760"),
		BrightYellow: lipgloss.Color("#fde047"),
		BrightWhite:  lipgloss.Color("#f8fafc"),
		Background:   lipgloss.Color("#0b1020"),
		Foreground:   lipgloss.Color("#e6edf3"),
		Dim:          lipgloss.Color("#94a0b2"),
		Surface:      lipgloss.Color("#1b2437"),
		Selection:    lipgloss.Color("#1e2a4a"),
		SelectionDim: lipgloss.Color("#141b2e"),

		Accent:     lipgloss.Color("#22d3ee"),
		Good:       lipgloss.Color("#34d399"),
		Warn:       lipgloss.Color("#fbbf24"),
		Bad:        lipgloss.Color("#fb7185"),
		GoodTint:   lipgloss.Color("#0d3b2f"),
		WarnTint:   lipgloss.Color("#3d2e08"),
		BadTint:    lipgloss.Color("#45172a"),
		AccentTint: lipgloss.Color("#0a2630"),
		Link:       blue,
		Pink:       lipgloss.Color("#f472b6"),
	}
}

var (
	AppColors = NewColors()

	// FooterStyle frames the help line under the main frame: a cell of
	// breathing room either side so the keys line up with the frame's text.
	FooterStyle = lipgloss.NewStyle().Foreground(AppColors.Foreground).Padding(0, 1)

	// Text styles. Values are coloured by what they mean, so the eye can scan
	// a column without reading every word.
	DimStyle       = lipgloss.NewStyle().Foreground(AppColors.Dim)
	StrongStyle    = lipgloss.NewStyle().Foreground(AppColors.BrightWhite).Bold(true)
	ValueStyle     = lipgloss.NewStyle().Foreground(AppColors.BrightWhite)
	AccentStyle    = lipgloss.NewStyle().Foreground(AppColors.Accent)
	GoodStyle      = lipgloss.NewStyle().Foreground(AppColors.Good)
	WarnStyle      = lipgloss.NewStyle().Foreground(AppColors.Warn)
	BadStyle       = lipgloss.NewStyle().Foreground(AppColors.Bad)
	ForkStyle      = lipgloss.NewStyle().Foreground(AppColors.Purple)
	StarStyle      = lipgloss.NewStyle().Foreground(AppColors.BrightYellow)
	LanguageStyle  = lipgloss.NewStyle().Foreground(AppColors.Blue)
	LinkTextStyle  = lipgloss.NewStyle().Foreground(AppColors.Link)
	LinkStyle      = LinkTextStyle.Underline(true)
	FocusCellStyle = lipgloss.NewStyle().Foreground(AppColors.Background).Background(AppColors.Accent)
	ColumnHeading  = DimStyle.Bold(true)
	TextBodyStyle  = lipgloss.NewStyle().Foreground(AppColors.Foreground)
	// TrackStyle draws the empty part of a bar: structure, never text.
	TrackStyle = lipgloss.NewStyle().Foreground(AppColors.BrightBlack)

	// Toggle cells: a dot on a tinted background says on, a ring on the
	// surface colour says off.
	ToggleOnStyle  = lipgloss.NewStyle().Foreground(AppColors.Good).Background(AppColors.GoodTint).Bold(true)
	ToggleOffStyle = lipgloss.NewStyle().Foreground(AppColors.Dim).Background(AppColors.Surface)

	// The frame gradient: accent through purple to pink and back, so the
	// join is seamless. Unfocused it sinks towards the background rather
	// than going grey. Both ramps are computed once; callers must not
	// modify them.
	frameGradient      = []color.Color{AppColors.Accent, AppColors.Purple, AppColors.Pink, AppColors.Accent}
	frameGradientMuted = darkenAll(frameGradient, 0.55)
)

// Pill draws text as a solid capsule: a half-block cap either side gives the
// ends a rounded look, and the text sits in dark ink on the colour. It is
// always len(text)+2 cells wide, so a row of pills lines up.
// Marker is the two-cell selection marker at the left of a selected row: a
// cell painted in c followed by a space. Painting the background rather than
// drawing a half-block glyph keeps the bar solid down a multi-line selection,
// since a terminal's line spacing would otherwise show as gaps between glyphs.
func Marker(c color.Color) string {
	return lipgloss.NewStyle().Background(c).Render(" ") + " "
}

func Pill(text string, c color.Color) string {
	cap := lipgloss.NewStyle().Foreground(c)
	ink := lipgloss.NewStyle().Foreground(AppColors.Background).Background(c).Bold(true)
	return cap.Render("▐") + ink.Render(text) + cap.Render("▌")
}

// TintedPill is the quieter capsule: the colour as text on its own tint, for
// states that should be read rather than shouted.
func TintedPill(text string, c, tint color.Color) string {
	cap := lipgloss.NewStyle().Foreground(tint)
	return cap.Render("▐") + lipgloss.NewStyle().Foreground(c).Background(tint).Bold(true).Render(text) + cap.Render("▌")
}

// GhostPill is the quiet counterpart of Pill: dim text on Surface, for the
// unselected segments of a control or a secondary tag.
func GhostPill(text string) string {
	return TintedPill(text, AppColors.Dim, AppColors.Surface)
}

// Bar draws a width-cell bar with the first filled cells in the fill style
// and the rest as the dim track every bar shares.
func Bar(filled, width int, fill lipgloss.Style) string {
	if width <= 0 {
		return ""
	}
	n := Max(0, Min(filled, width))
	return fill.Render(strings.Repeat("━", n)) + TrackStyle.Render(strings.Repeat("╌", width-n))
}

// Share is how many of width cells a fraction fills: rounded, never more
// than width, and at least one cell for any share at all so it can be told
// from none.
func Share(fraction float64, width int) int {
	if width <= 0 || fraction <= 0 {
		return 0
	}
	n := int(math.Round(fraction * float64(width)))
	return Max(1, Min(n, width))
}

// NewHelpModel builds the footer's help: keys in the accent colour so they
// stand out from what they do, and two spaces between entries so the short
// line fits an 80-column terminal.
func NewHelpModel(width int) help.Model {
	m := help.New()
	styles := help.DefaultStyles(true)
	styles.ShortKey = AccentStyle
	styles.ShortDesc = TextBodyStyle
	styles.ShortSeparator = DimStyle
	styles.FullKey = AccentStyle
	styles.FullDesc = TextBodyStyle
	styles.FullSeparator = DimStyle
	styles.Ellipsis = DimStyle
	m.Styles = styles
	m.ShortSeparator = "  "
	m.FullSeparator = "    "
	if width > 0 {
		m.SetWidth(width)
	}
	return m
}

// FrameGradient is the ramp that runs around the frame and the inspector's
// card, bright while focused and sunk towards the background when not. The
// slice is shared: read it, do not modify it.
func FrameGradient(focused bool) []color.Color {
	if focused {
		return frameGradient
	}
	return frameGradientMuted
}

func darkenAll(stops []color.Color, percent float64) []color.Color {
	out := make([]color.Color, len(stops))
	for i, c := range stops {
		out[i] = lipgloss.Darken(c, percent)
	}
	return out
}

// PerimeterGradient returns the colours Lip Gloss gives the top edge of a
// BorderForegroundBlend block that is width cells wide around height lines
// of content, so an edge drawn by hand continues the same gradient as the
// body. Lip Gloss blends (height + width + 2) * 2 steps, giving the top
// and bottom edges width+2 each (two more than their cells) and the sides
// height each, and colours the top edge from the first width of them.
func PerimeterGradient(width, height int) []color.Color {
	width, height = Max(width, 1), Max(height, 1)
	return lipgloss.Blend1D((height+width+2)*2, frameGradient...)[:width]
}

// GradientRun colours each cell of s with the matching entry of colors,
// falling back to the last colour when s is longer.
func GradientRun(s string, colors []color.Color) string {
	if len(colors) == 0 {
		return s
	}
	var b strings.Builder
	for i, r := range []rune(s) {
		c := colors[Min(i, len(colors)-1)]
		b.WriteString(lipgloss.NewStyle().Foreground(c).Render(string(r)))
	}
	return b.String()
}
