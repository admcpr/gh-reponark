package shared

import tea "charm.land/bubbletea/v2"

// HeaderProvider models can render a header view for the surrounding layout.
type HeaderProvider interface {
	HeaderView() tea.View
}

// HelpProvider models can render contextual help for the footer.
type HelpProvider interface {
	HelpView() tea.View
}
