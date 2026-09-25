package shared

import (
	"image/color"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
)

type colors struct {
	name         string
	Black        color.Color
	Red          color.Color
	Green        color.Color
	Yellow       color.Color
	Blue         color.Color
	Purple       color.Color
	Cyan         color.Color
	White        color.Color
	BrightBlack  color.Color
	BrightRed    color.Color
	BrightGreen  color.Color
	BrightYellow color.Color
	BrightBlue   color.Color
	BrightPurple color.Color
	BrightCyan   color.Color
	BrightWhite  color.Color
	Background   color.Color
	Foreground   color.Color
	// Dim is for secondary text: headings, hints and absent values.
	Dim color.Color
	// Surface lifts small elements such as chips off the background.
	Surface color.Color
}

func NewColors(darkmode bool) colors {
	colors := colors{
		name:         "PencilDark",
		Black:        lipgloss.Color("#212121"),
		Red:          lipgloss.Color("#c30771"),
		Green:        lipgloss.Color("#10a778"),
		Yellow:       lipgloss.Color("#a89c14"),
		Blue:         lipgloss.Color("#008ec4"),
		Purple:       lipgloss.Color("#523c79"),
		Cyan:         lipgloss.Color("#20a5ba"),
		White:        lipgloss.Color("#d9d9d9"),
		BrightBlack:  lipgloss.Color("#424242"),
		BrightRed:    lipgloss.Color("#fb007a"),
		BrightGreen:  lipgloss.Color("#5fd7af"),
		BrightYellow: lipgloss.Color("#f3e430"),
		BrightBlue:   lipgloss.Color("#20bbfc"),
		BrightPurple: lipgloss.Color("#6855de"),
		BrightCyan:   lipgloss.Color("#4fb8cc"),
		BrightWhite:  lipgloss.Color("#f1f1f1"),
		Background:   lipgloss.Color("#212121"),
		Foreground:   lipgloss.Color("#f1f1f1"),
		Dim:          lipgloss.Color("#6c6868"),
		Surface:      lipgloss.Color("#16323b"),
	}
	if darkmode {
		colors.name = "3024 Night"
		colors.Background = lipgloss.Color("#090300")
		colors.Foreground = lipgloss.Color("#a5a2a2")
	}
	return colors
}

var (
	AppColors = NewColors(true)

	AppStyle = lipgloss.NewStyle().Padding(0, 0).
			Foreground(AppColors.Foreground).
			BorderForeground(AppColors.Blue).
			Border(lipgloss.RoundedBorder(), false)

	TitleStyle = AppStyle.Foreground(AppColors.Blue).
			BorderForeground(AppColors.BrightBlue).
			Border(lipgloss.NormalBorder(), false, false, true, true).
			Padding(0, 1, 0, 1)

	LayoutFooterStyle = AppStyle.
				Foreground(AppColors.Foreground).
				BorderForeground(AppColors.BrightBlack).
				Border(lipgloss.NormalBorder(), false, false, false, false).
				Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().Foreground(AppColors.Red)

	TextStyle = AppStyle.Foreground(AppColors.Foreground).
			PaddingLeft(1)

	DefaultDelegate = BuildDefaultDelegate()

	// Styles for the repository browser. Values are coloured by what they
	// mean, so the eye can scan a column without reading every word.
	DimStyle       = lipgloss.NewStyle().Foreground(AppColors.Dim)
	StrongStyle    = lipgloss.NewStyle().Foreground(AppColors.BrightWhite).Bold(true)
	ValueStyle     = lipgloss.NewStyle().Foreground(AppColors.BrightWhite)
	AccentStyle    = lipgloss.NewStyle().Foreground(AppColors.Cyan)
	GoodStyle      = lipgloss.NewStyle().Foreground(AppColors.Green)
	WarnStyle      = lipgloss.NewStyle().Foreground(AppColors.Yellow)
	BadStyle       = lipgloss.NewStyle().Foreground(AppColors.Red)
	ForkStyle      = lipgloss.NewStyle().Foreground(AppColors.BrightPurple)
	StarStyle      = lipgloss.NewStyle().Foreground(AppColors.BrightYellow)
	ActiveTabLabel = lipgloss.NewStyle().Foreground(AppColors.Background).Background(AppColors.Blue).Bold(true)
	FocusCellStyle = lipgloss.NewStyle().Foreground(AppColors.Background).Background(AppColors.Cyan)
	ColumnHeading  = DimStyle.Bold(true)
	HeadingStyle   = lipgloss.NewStyle().Foreground(AppColors.BrightBlue).Bold(true)
	TextBodyStyle  = lipgloss.NewStyle().Foreground(AppColors.Foreground)
	ChipStyle      = lipgloss.NewStyle().Foreground(AppColors.BrightCyan).Background(AppColors.Surface)
)

// PillStyle is dark text on a solid colour, for labels that should read as
// a single object.
func PillStyle(c color.Color) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(AppColors.Background).Background(c).Bold(true)
}

func NewHelpModel(width int) help.Model {
	m := help.New()
	styles := help.DefaultStyles(true)
	styles.ShortKey = lipgloss.NewStyle().Foreground(AppColors.Foreground)
	styles.ShortDesc = lipgloss.NewStyle().Foreground(AppColors.Foreground)
	styles.ShortSeparator = lipgloss.NewStyle().Foreground(AppColors.BrightBlack)
	styles.FullKey = lipgloss.NewStyle().Foreground(AppColors.Foreground)
	styles.FullDesc = lipgloss.NewStyle().Foreground(AppColors.Foreground)
	styles.FullSeparator = lipgloss.NewStyle().Foreground(AppColors.BrightBlack)
	styles.Ellipsis = lipgloss.NewStyle().Foreground(AppColors.BrightBlack)
	m.Styles = styles
	if width > 0 {
		m.SetWidth(width)
	}
	return m
}

func BuildDefaultDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(AppColors.Cyan).
		BorderForeground(AppColors.Cyan)
	d.Styles.SelectedDesc = d.Styles.SelectedTitle

	return d
}
