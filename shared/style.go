package shared

import (
	"image/color"

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

	TabStyle = AppStyle.Border(lipgloss.NormalBorder(), true, false, false, false).
			Align(lipgloss.Center)

	ActiveTabStyle = TabStyle.BorderForeground(AppColors.BrightBlue).
			Foreground(AppColors.Cyan)

	TitleStyle = AppStyle.Foreground(AppColors.Blue).
			BorderForeground(AppColors.BrightBlue).
			Border(lipgloss.NormalBorder(), false, false, true, true).
			Padding(0, 1, 0, 1)

	ErrorStyle = lipgloss.NewStyle().Foreground(AppColors.Red)

	PromptStyle = AppStyle.Width(7).
			Align(lipgloss.Right).
			PaddingRight(1).
			MarginTop(1)

	TextStyle = AppStyle.Foreground(AppColors.Foreground).
			PaddingLeft(1)

	CursorStyle = AppStyle

	ButtonStyle = AppStyle.
			BorderForeground(AppColors.Purple).
			Padding(0, 3).
			Margin(2)

	ActiveButtonStyle = ButtonStyle.
				Foreground(AppColors.Foreground).
				Background(AppColors.Cyan).
				Underline(true)

	ItemStyle         = lipgloss.NewStyle().PaddingLeft(2)
	SelectedItemStyle = ItemStyle.
				PaddingLeft(1).
				Foreground(AppColors.Cyan).
				BorderForeground(AppColors.Cyan).
				Border(lipgloss.NormalBorder(), false, false, false, true)

	ModalTitleStyle = TitleStyle.
			Align(lipgloss.Center).
			Foreground(AppColors.Blue).
			BorderForeground(AppColors.BrightGreen).
			Border(lipgloss.DoubleBorder(), false, false, true, false).
			Width(60)

	ModalStyle = AppStyle.
			BorderForeground(AppColors.Green).
			Padding(0)

	DefaultDelegate = BuildDefaultDelegate()
)

func BuildDefaultDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(AppColors.Cyan).
		BorderForeground(AppColors.Cyan)
	d.Styles.SelectedDesc = d.Styles.SelectedTitle

	return d
}
