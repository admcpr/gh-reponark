package ui

import (
	"fmt"
	"math"
)

func Half(width int) int {
	return int(math.Floor(float64(width) / 2.0))
}

func Quarter(width int) int {
	return int(math.Floor(float64(width) / 4.0))
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Plural picks the singular or plural form of a noun for a count, e.g.
// Plural(1, "repo", "repos") is "repo" and Plural(3, "repo", "repos") is "repos".
func Plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// AgoDays describes a number of days ago compactly, e.g. "today", "3d ago",
// "4mo ago" or "2y ago".
func AgoDays(days int) string {
	switch {
	case days < 1:
		return "today"
	case days < 30:
		return fmt.Sprintf("%dd ago", days)
	case days < 365:
		return fmt.Sprintf("%dmo ago", days/30)
	default:
		return fmt.Sprintf("%dy ago", days/365)
	}
}

// Sparks are the bar glyphs from lowest to full, for sparklines and
// histograms.
var Sparks = []rune("▁▂▃▄▅▆▇█")
