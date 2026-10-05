package inspector

// Size tiers. Width decides what the header carries; height decides how the
// footer is drawn and how much of the header survives.
const (
	// minMediumWidth is the full layout: header, tiles, bars and the footer
	// card. Below it the pane is a plain list with a compact header. SIZE
	// joins the tiles from sizeTileWidth.
	minMediumWidth = 50
	// oneRowChipWidth is where all six chips fit a single row even with
	// their longest labels, so the second row is not reserved.
	oneRowChipWidth = 76
	// minListWidth is the narrowest pane with a name column; below it each
	// property is one plain line.
	minListWidth = 12
	// minCardHeight is the shortest pane that draws the footer card; below
	// it the footer is a rule and two plain description lines.
	minCardHeight = 24
	// tallHeight is where the description gets a second line.
	tallHeight = 40
	// minRows is how many property rows the header gives way to keep.
	minRows = 12

	footerCardLines  = 6
	footerPlainLines = 3
	segmentLines     = 1
	gapLines         = 1
)

// layout is what the pane draws at its current size.
type layout struct {
	// plain is the sub-minListWidth pane: the name, the groups and one line
	// per property, nothing else.
	plain       bool
	description int
	facts       bool
	chipRows    int
	// tiles is how many stat tiles the strip shows; zero for no strip.
	tiles int
	// bars draws a bar beside each count; card draws the footer card.
	bars bool
	card bool
}

// headerLines is how many lines the header takes.
func (l layout) headerLines() int {
	return 1 + l.description + boolInt(l.facts) + l.chipRows
}

// footerLines is how many lines sit below the property rows.
func (l layout) footerLines() int {
	switch {
	case l.plain:
		return 0
	case l.card:
		return footerCardLines
	default:
		return footerPlainLines
	}
}

// chrome is how many lines the layout spends around the property rows.
func (l layout) chrome() int {
	if l.plain {
		return 1 + segmentLines
	}
	return l.headerLines() + boolInt(l.tiles > 0)*statsLines + gapLines + segmentLines + l.footerLines()
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// layout decides what fits at the pane's size. Width sets the tier; then
// the header sheds, in order, the description, the second chip row, the
// tiles, the facts and the chips until at least minRows property rows are
// left.
func (m Model) layout() layout {
	w, h := m.width, m.height
	switch {
	case w < minListWidth:
		return layout{plain: true}
	case w < minMediumWidth:
		return layout{description: 1}
	}
	l := layout{description: 1, facts: true, chipRows: 1, tiles: tileCount(w), bars: true, card: h >= minCardHeight}
	if h >= tallHeight {
		l.description = 2
	}
	if w < oneRowChipWidth {
		l.chipRows = 2
	}
	for h-l.chrome() < minRows {
		switch {
		case l.description > 0:
			l.description--
		case l.chipRows > 1:
			l.chipRows = 1
		case l.tiles > 0:
			l.tiles = 0
		case l.facts:
			l.facts = false
		case l.chipRows > 0:
			l.chipRows = 0
		default:
			return l
		}
	}
	return l
}

// chrome is the number of lines around the property rows at this size.
func (m Model) chrome() int { return m.layout().chrome() }
