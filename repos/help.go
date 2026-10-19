package repos

import (
	"gh-reponark/ui"

	"charm.land/bubbles/v2/key"
)

// Help lists the keys for the current view. The short line names one key per
// action; the full view adds the alternatives and paging keys.
func (m Model) Help() ui.Help {
	if m.progress.Percent() < 1 || len(m.visible) == 0 {
		// Until there is something to browse, only filtering and leaving work.
		return ui.Help{Short: []key.Binding{m.keymap.Filters, m.keymap.Back}}
	}

	repoKeys := m.repoModel.Keys()
	tabs := ui.Combine("tab", "group", repoKeys.NextTab, repoKeys.PrevTab)
	filters := m.keymap.Filters
	paging := []key.Binding{m.keymap.PageUp, m.keymap.PageDown, m.keymap.Top, m.keymap.Bottom}
	groups := []key.Binding{repoKeys.NextTab, repoKeys.PrevTab}

	switch {
	case m.mode == overviewView:
		// The focused card names what j/k steps through, e.g. "language".
		noun := m.rowNoun()
		return ui.Help{
			Short: []key.Binding{
				ui.Combine("tab", "card", repoKeys.NextTab, repoKeys.PrevTab),
				ui.Combine("j/k", noun, m.keymap.Down, m.keymap.Up),
				ui.WithHelp(m.keymap.Inspect, "show repos"),
				ui.WithHelp(m.keymap.ToggleView, "list"),
				filters,
				m.keymap.Back,
			},
			Full: [][]key.Binding{
				append([]key.Binding{
					ui.WithHelp(repoKeys.NextTab, "next card"),
					ui.WithHelp(repoKeys.PrevTab, "previous card"),
					ui.WithHelp(m.keymap.Left, "card left"),
					ui.WithHelp(m.keymap.Right, "card right"),
				}, paging...),
				{ui.WithHelp(m.keymap.Down, "next "+noun), ui.WithHelp(m.keymap.Up, "previous "+noun), ui.WithHelp(m.keymap.Inspect, "show the "+noun+"'s repos"), ui.WithHelp(m.keymap.ToggleView, "list view")},
				{filters, m.keymap.Back},
			},
		}

	case m.mode == matrixView:
		return ui.Help{
			Short: []key.Binding{
				// Like every view: j/k names what moves up and down, h/l is
				// listed apart, and enter, which there is no room for at 80
				// columns, is in the full help.
				ui.Combine("j/k", "repo", m.keymap.Down, m.keymap.Up),
				ui.Combine("h/l", "property", m.keymap.Left, m.keymap.Right),
				tabs,
				ui.WithHelp(m.keymap.ToggleView, "overview"),
				filters,
				m.keymap.Back,
			},
			Full: [][]key.Binding{
				append([]key.Binding{
					ui.WithHelp(m.keymap.Up, "repo up"),
					ui.WithHelp(m.keymap.Down, "repo down"),
					ui.WithHelp(m.keymap.Left, "property left"),
					ui.WithHelp(m.keymap.Right, "property right"),
				}, paging...),
				append(groups, ui.WithHelp(m.keymap.Inspect, "inspect repo"), ui.WithHelp(m.keymap.ToggleView, "overview")),
				{filters, m.keymap.Back},
			},
		}

	case m.inspecting:
		return ui.Help{
			Short: []key.Binding{
				ui.Combine("j/k", "property", m.keymap.Down, m.keymap.Up),
				tabs,
				ui.WithHelp(m.keymap.ToggleView, "matrix"),
				filters,
				ui.Combine("h/esc", "repos", m.keymap.Left, m.keymap.Back),
			},
			Full: [][]key.Binding{
				append([]key.Binding{
					ui.WithHelp(m.keymap.Up, "property up"),
					ui.WithHelp(m.keymap.Down, "property down"),
				}, paging...),
				append(groups, ui.Combine("h/esc", "back to repos", m.keymap.Left, m.keymap.Back), ui.WithHelp(m.keymap.ToggleView, "matrix view")),
				{filters},
			},
		}

	default:
		// From a slice of the overview, esc goes back there.
		back := m.keymap.Back
		if m.fromOverview {
			back = ui.WithHelp(m.keymap.Back, "overview")
		}
		return ui.Help{
			Short: []key.Binding{
				ui.Combine("j/k", "repo", m.keymap.Down, m.keymap.Up),
				ui.Combine("enter", "inspect", m.keymap.Inspect, m.keymap.Right),
				tabs,
				ui.WithHelp(m.keymap.ToggleView, "matrix"),
				filters,
				back,
			},
			Full: [][]key.Binding{
				append([]key.Binding{
					ui.WithHelp(m.keymap.Up, "repo up"),
					ui.WithHelp(m.keymap.Down, "repo down"),
				}, paging...),
				append(groups, ui.Combine("l/enter", "inspect repo", m.keymap.Right, m.keymap.Inspect), ui.WithHelp(m.keymap.ToggleView, "matrix view")),
				{filters, back},
			},
		}
	}
}
