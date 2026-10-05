package inspector

import (
	"gh-reponark/ui"
	"image/color"

	"charm.land/lipgloss/v2"
)

// PropertyType is how each kind of property is marked: a glyph beside it in
// the list, coloured the way the inspector colours that kind of value, and
// the same type pill as the inspector's footer card in the editor.
type PropertyType struct {
	glyph string
	color color.Color
}

func TypeOf(t string) PropertyType {
	switch t {
	case "bool":
		return PropertyType{"●", ui.AppColors.Good}
	case "int":
		return PropertyType{"#", ui.AppColors.Link}
	case "time.Time":
		return PropertyType{"◆", ui.AppColors.Purple}
	default:
		return PropertyType{"¶", ui.AppColors.Dim}
	}
}

func (t PropertyType) Glyph() string {
	return lipgloss.NewStyle().Foreground(t.color).Render(t.glyph)
}

// typeLabel names a property's Go type the way a reader would.
func typeLabel(t string) string {
	switch t {
	case "bool":
		return "toggle"
	case "int":
		return "count"
	case "time.Time":
		return "date"
	default:
		return "text"
	}
}

// TypePill is the quiet capsule every screen tags a property's type with:
// the type label in dim text on the surface tint.
func TypePill(t string) string {
	return ui.GhostPill(typeLabel(t))
}
