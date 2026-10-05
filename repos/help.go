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
	case m.mode == matrixView:
		return ui.Help{
			Short: []key.Binding{
				ui.Combine("hjkl", "move", m.keymap.Left, m.keymap.Down, m.keymap.Up, m.keymap.Right),
				ui.WithHelp(m.keymap.Inspect, "inspect"),
				tabs,
				ui.WithHelp(m.keymap.ToggleView, "list"),
				filters,
				m.keymap.Back,
			},
			Full: [][]key.Binding{
				append([]key.Binding{
					ui.WithHelp(m.keymap.Up, "repo up"),
					ui.WithHelp(m.keymap.Down, "repo down"),
					ui.WithHelp(m.keymap.Left, "column left"),
					ui.WithHelp(m.keymap.Right, "column right"),
				}, paging...),
				append(groups, ui.WithHelp(m.keymap.Inspect, "inspect repo"), ui.WithHelp(m.keymap.ToggleView, "list view")),
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
		return ui.Help{
			Short: []key.Binding{
				ui.Combine("j/k", "repo", m.keymap.Down, m.keymap.Up),
				ui.Combine("enter", "inspect", m.keymap.Inspect, m.keymap.Right),
				tabs,
				ui.WithHelp(m.keymap.ToggleView, "matrix"),
				filters,
				m.keymap.Back,
			},
			Full: [][]key.Binding{
				append([]key.Binding{
					ui.WithHelp(m.keymap.Up, "repo up"),
					ui.WithHelp(m.keymap.Down, "repo down"),
				}, paging...),
				append(groups, ui.Combine("l/enter", "inspect repo", m.keymap.Right, m.keymap.Inspect), ui.WithHelp(m.keymap.ToggleView, "matrix view")),
				{filters, m.keymap.Back},
			},
		}
	}
}
