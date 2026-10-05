package ui

import "math"

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
