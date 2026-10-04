package repo

import (
	"strings"

	"gh-reponark/shared"

	"charm.land/lipgloss/v2"
)

// RenderSegments draws the tabs as a one-line segmented control: a row of
// adjacent pills, the active one filled in the accent colour and the rest
// ghosted on the surface tint. It scrolls to keep the active segment in
// view, marking the hidden ones with ‹ and ›. When the pane is
// not focused the active segment steps back to blue so the focused pane's
// accent stands alone.
func RenderSegments(tabs []string, width, activeTab int, focused bool) string {
	if len(tabs) == 0 || width <= 0 {
		return shared.Fit("", width)
	}
	if activeTab < 0 || activeTab >= len(tabs) {
		activeTab = 0
	}

	accent := shared.AppColors.Accent
	if !focused {
		accent = shared.AppColors.Blue
	}
	labels := make([]string, len(tabs))
	for i, t := range tabs {
		if i == activeTab {
			labels[i] = shared.Pill(t, accent)
		} else {
			labels[i] = shared.GhostPill(t)
		}
	}
	lo, hi := tabWindow(labels, width, activeTab)

	var b strings.Builder
	if lo > 0 {
		b.WriteString(shared.DimStyle.Render("‹ "))
	}
	for i := lo; i <= hi; i++ {
		b.WriteString(labels[i])
	}
	if hi < len(tabs)-1 {
		b.WriteString(shared.DimStyle.Render(" ›"))
	}
	row := b.String()
	if lipgloss.Width(row) > width {
		// Even the active pill and its scroll markers do not fit. A pill is
		// never cut through, so a pane this narrow gets the bare title.
		return shared.Fit(lipgloss.NewStyle().Foreground(accent).Bold(true).Render(tabs[activeTab]), width)
	}
	return shared.Fit(row, width)
}

// tabWindow grows a window of adjacent tabs outwards from the active one
// while it fits in width, leaving room for a two-cell scroll marker on each
// side that is cut off. It returns the first and last tab shown.
func tabWindow(labels []string, width, activeTab int) (lo, hi int) {
	markers := func(lo, hi int) int {
		n := 0
		if lo > 0 {
			n += 2
		}
		if hi < len(labels)-1 {
			n += 2
		}
		return n
	}
	lo, hi = activeTab, activeTab
	used := lipgloss.Width(labels[activeTab])
	for grew := true; grew; {
		grew = false
		if hi+1 < len(labels) {
			if w := used + lipgloss.Width(labels[hi+1]); w+markers(lo, hi+1) <= width {
				hi, used, grew = hi+1, w, true
			}
		}
		if lo > 0 {
			if w := used + lipgloss.Width(labels[lo-1]); w+markers(lo-1, hi) <= width {
				lo, used, grew = lo-1, w, true
			}
		}
	}
	return lo, hi
}
