package repo

import (
	"strings"

	"gh-reponark/shared"

	"charm.land/lipgloss/v2"
)

// RenderTabs draws a single-line tab bar with the active tab highlighted.
// When the tabs are wider than width, it shows the tabs around the active one
// with ‹ and › marking the ones scrolled out of view.
func RenderTabs(tabs []string, width, activeTab int) string {
	if len(tabs) == 0 || width <= 0 {
		return shared.Fit("", width)
	}
	if activeTab < 0 || activeTab >= len(tabs) {
		activeTab = 0
	}

	labels := make([]string, len(tabs))
	for i, t := range tabs {
		label := " " + t + " "
		if i == activeTab {
			labels[i] = shared.ActiveTabLabel.Render(label)
		} else {
			labels[i] = shared.DimStyle.Render(label)
		}
	}

	// Grow a window of tabs outwards from the active one while it fits,
	// leaving room for a scroll marker on each side that is cut off.
	markers := func(lo, hi int) int {
		n := 0
		if lo > 0 {
			n += 2
		}
		if hi < len(tabs)-1 {
			n += 2
		}
		return n
	}
	lo, hi := activeTab, activeTab
	used := lipgloss.Width(labels[activeTab])
	for grew := true; grew; {
		grew = false
		if hi+1 < len(tabs) {
			if w := used + 1 + lipgloss.Width(labels[hi+1]); w+markers(lo, hi+1) <= width {
				hi, used, grew = hi+1, w, true
			}
		}
		if lo > 0 {
			if w := used + 1 + lipgloss.Width(labels[lo-1]); w+markers(lo-1, hi) <= width {
				lo, used, grew = lo-1, w, true
			}
		}
	}

	row := strings.Join(labels[lo:hi+1], shared.DimStyle.Render("│"))
	if lo > 0 {
		row = shared.DimStyle.Render("‹ ") + row
	}
	if hi < len(tabs)-1 {
		row += shared.DimStyle.Render(" ›")
	}
	return shared.Fit(row, width)
}
