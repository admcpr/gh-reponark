package repos

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"gh-reponark/filter"
	"gh-reponark/inspector"
	"gh-reponark/overview"
	"gh-reponark/repo"
	"gh-reponark/ui"

	"charm.land/lipgloss/v2"
)

// The overview is the owner's profile over a grid of cards, each one
// question about the repositories answered by a small chart, and a footer
// listing the repositories behind the focused slice of the focused card.
const (
	// Cards are at least minCardWidth wide, so every chart fits, and grow
	// to share out the width up to maxCardWidth; cardGap separates them.
	minCardWidth = 38
	maxCardWidth = 56
	cardGap      = 2
	// cardBody is a card's chart and cardLines its height with the title
	// rule; a blank line separates the rows of the grid.
	cardBody  = overview.MaxRows
	cardLines = cardBody + 1
	// cardLabelWidth is a Bars row's label; the bar takes what the label,
	// count and share leave.
	cardLabelWidth = 14
	// minFooterLines is the footer's rule, figures and one line of names.
	minFooterLines = 3
	// Pane heights from which the header carries the facts line and the
	// owner's description.
	factsHeight       = 20
	descriptionHeight = 28
	// footerNameWidth is one repository's cell in the footer's columns.
	footerNameWidth = 24
)

// overviewCardCount is how many cards overview.Cards builds.
var overviewCardCount = len(overview.Cards(nil))

// overviewLayout is what the overview draws at its current size.
type overviewLayout struct {
	description bool
	facts       bool
	filters     bool
	columns     int
	cardWidth   int
	// rows is how many rows of cards are drawn; footer is how many lines
	// the footer has, which is whatever the grid leaves.
	rows   int
	footer int
}

// headerLines is the title, the description, facts and filters when shown,
// and a gap before the grid.
func (l overviewLayout) headerLines() int {
	n := 2
	for _, shown := range []bool{l.description, l.facts, l.filters} {
		if shown {
			n++
		}
	}
	return n
}

// overviewLayout fits the grid to the pane: as many columns as fit cards of
// at least minCardWidth, sharing out the width with a gutter on the right
// when there is room for one, as many rows as fit above a minimal footer,
// and the footer whatever is left.
func (m *Model) overviewLayout() overviewLayout {
	w, h := m.width, m.height
	l := overviewLayout{
		description: h >= descriptionHeight,
		facts:       h >= factsHeight,
		filters:     len(m.filters) > 0,
		columns:     ui.Max(1, (w+cardGap)/(minCardWidth+cardGap)),
	}
	gaps := (l.columns - 1) * cardGap
	l.cardWidth = (w - gaps) / l.columns
	if l.cardWidth > minCardWidth && l.cardWidth*l.columns+gaps >= w {
		l.cardWidth--
	}
	l.cardWidth = ui.Max(1, ui.Min(maxCardWidth, l.cardWidth))

	totalRows := (overviewCardCount + l.columns - 1) / l.columns
	free := h - l.headerLines() - minFooterLines
	l.rows = ui.Max(1, ui.Min(totalRows, (free+1)/(cardLines+1)))
	l.footer = ui.Max(minFooterLines, h-l.headerLines()-(l.rows*(cardLines+1)-1))
	return l
}

// overviewCards are the cards over the visible repositories, built once
// each time the filters change rather than on every key and redraw.
func (m *Model) overviewCards() []overview.Card {
	if m.cards == nil {
		m.cards = overview.Cards(m.visible)
	}
	return m.cards
}

// focusable lists the slices of card that can take focus: those drawn on
// the card at the current width, and that hold at least one repository,
// since an empty slice has nothing to list or filter to.
func (m *Model) focusable(card overview.Card) []int {
	_, drawn := renderCard(card, false, -1, m.overviewLayout().cardWidth)
	var out []int
	for _, i := range drawn {
		if len(card.Slices[i].Names) > 0 {
			out = append(out, i)
		}
	}
	return out
}

// focusedCard clamps the card cursor to the cards and the slice cursor to
// the focused card's focusable slices, and returns the card and the focused
// slice, or -1 when it has none.
func (m *Model) focusedCard() (overview.Card, int) {
	cards := m.overviewCards()
	m.card = ui.Max(0, ui.Min(m.card, len(cards)-1))
	card := cards[m.card]
	focusable := m.focusable(card)
	switch {
	case len(focusable) == 0:
		m.slice = -1
	case indexOf(focusable, m.slice) < 0:
		m.slice = focusable[0]
	}
	return card, m.slice
}

func indexOf(values []int, v int) int {
	for i, x := range values {
		if x == v {
			return i
		}
	}
	return -1
}

// moveCard moves the focus delta cards, stopping at either end, and starts
// on the new card's first focusable slice.
func (m *Model) moveCard(delta int) {
	next := ui.Max(0, ui.Min(m.card+delta, overviewCardCount-1))
	if next != m.card {
		m.card, m.slice = next, -1
		m.focusedCard()
	}
}

// rowNoun is what the focused card's rows are, for the help, e.g.
// "language"; "item" for a card that does not say.
func (m *Model) rowNoun() string {
	card, _ := m.focusedCard()
	if card.Noun == "" {
		return "item"
	}
	return card.Noun
}

// cycleCard moves the focus delta cards, wrapping around at either end,
// the way tab moves between groups elsewhere.
func (m *Model) cycleCard(delta int) {
	m.card = ((m.card+delta)%overviewCardCount + overviewCardCount) % overviewCardCount
	m.slice = -1
	m.focusedCard()
}

// moveSlice moves the focus delta focusable slices, stopping at either end
// like a list, so it never lands on a slice that is not drawn or holds no
// repositories.
func (m *Model) moveSlice(delta int) {
	card, _ := m.focusedCard()
	focusable := m.focusable(card)
	if len(focusable) == 0 {
		return
	}
	at := ui.Max(0, ui.Min(indexOf(focusable, m.slice)+delta, len(focusable)-1))
	m.slice = focusable[at]
}

// sliceFilters are the filters that select a slice's repositories: its own,
// or its names when no property filter can express it.
func sliceFilters(card overview.Card, slice overview.Slice) []filter.Filter {
	if len(slice.Filters) > 0 {
		return slice.Filters
	}
	return []filter.Filter{filter.NewLabelledNamesFilter(card.Title+" › "+slice.Label, slice.Names)}
}

// filterToSlice narrows the active filters to the focused slice and shows
// the list, so the repositories behind the slice can be walked. A filter
// on a property that already has one is combined with it rather than
// replacing it, so the list holds exactly the slice's repositories. The
// filters from before are kept, so esc in the list can put them back.
// Without a focused slice it does nothing, so it never filters to nothing.
func (m *Model) filterToSlice() {
	card, focus := m.focusedCard()
	if focus < 0 {
		return
	}
	before := make(filter.FilterMap, len(m.filters))
	for name, f := range m.filters {
		before[name] = f
	}
	next := make(filter.FilterMap, len(m.filters))
	for name, f := range m.filters {
		next[name] = f
	}
	for _, f := range sliceFilters(card, card.Slices[focus]) {
		next[f.Name()] = filter.All(next[f.Name()], f)
	}
	m.filters = next
	m.applyFilters()
	m.mode = listView
	m.fromOverview, m.overviewFilters = true, before
}

// overviewView draws the header, the grid and the footer and pads them to
// the pane.
func (m *Model) overviewView() string {
	width := m.width
	l := m.overviewLayout()
	cards := m.overviewCards()
	card, focus := m.focusedCard()

	lines := m.overviewHeader(l, width)

	// The grid scrolls by row of cards, keeping the focused card in view.
	totalRows := (len(cards) + l.columns - 1) / l.columns
	m.cardRow = ui.ScrollOffset(m.cardRow, m.card/l.columns, l.rows, totalRows)
	for row := m.cardRow; row < ui.Min(totalRows, m.cardRow+l.rows); row++ {
		if row > m.cardRow {
			lines = append(lines, "")
		}
		var rendered [][]string
		for col := 0; col < l.columns && row*l.columns+col < len(cards); col++ {
			i := row*l.columns + col
			slice := -1
			if i == m.card {
				slice = focus
			}
			card, _ := renderCard(cards[i], i == m.card, slice, l.cardWidth)
			rendered = append(rendered, card)
		}
		for line := 0; line < cardLines; line++ {
			var parts []string
			for _, r := range rendered {
				parts = append(parts, r[line])
			}
			lines = append(lines, strings.Join(parts, strings.Repeat(" ", cardGap)))
		}
	}
	for len(lines) < m.height-l.footer {
		lines = append(lines, "")
	}
	lines = lines[:ui.Max(0, ui.Min(len(lines), m.height-l.footer))]

	above := m.cardRow * l.columns
	below := len(cards) - ui.Min(len(cards), (m.cardRow+l.rows)*l.columns)
	lines = append(lines, m.overviewFooter(card, focus, l.footer, above, below, width)...)
	return ui.Lines(lines, width, m.height)
}

// overviewHeader is the owner's name with its pills, the description and
// the facts when there is room, the filters applied, and a gap.
func (m *Model) overviewHeader(l overviewLayout, width int) []string {
	lines := []string{m.overviewTitle(width)}
	if l.description {
		lines = append(lines, m.overviewDescription(width))
	}
	if l.facts {
		facts, _ := ui.JoinFit(m.overviewFacts(), ui.DimStyle.Render(" · "), width)
		lines = append(lines, facts)
	}
	if l.filters {
		lines = append(lines, m.overviewFilterLine(width))
	}
	return append(lines, "")
}

// overviewFilterLine names every filter applied, so it is clear what the
// cards are counting, with how many repositories pass.
func (m *Model) overviewFilterLine(width int) string {
	names := make([]string, 0, len(m.filters))
	for name := range m.filters {
		names = append(names, name)
	}
	sort.Strings(names)
	chips := make([]string, len(names))
	for i, name := range names {
		chips[i] = ui.TextBodyStyle.Render(name+": ") + ui.AccentStyle.Render(m.filters[name].Condition())
	}
	lead := ui.TintedPill(fmt.Sprintf("filtered %d of %d", len(m.visible), len(m.repos)), ui.AppColors.Accent, ui.AppColors.AccentTint) + " "
	line, dropped := ui.JoinFit(chips, ui.DimStyle.Render(" · "), width-lipgloss.Width(lead)-5)
	if dropped > 0 {
		line += ui.DimStyle.Render(fmt.Sprintf(" +%d", dropped))
	}
	return ui.Fit(lead+line, width)
}

// overviewTitle is the display name, the login beside it when they differ,
// and pills for what the account is and what the viewer is to it.
func (m *Model) overviewTitle(width int) string {
	o := m.owner
	name := o.Name
	if name == "" {
		name = m.Title
	}
	title := ui.StrongStyle.Render(name)
	if o.Login != "" && o.Login != name {
		title += "  " + ui.DimStyle.Render(o.Login)
	}

	var pills []string
	if m.isUser {
		pills = append(pills, ui.Pill("you", ui.AppColors.Accent))
	} else {
		pills = append(pills, ui.GhostPill("org"))
	}
	if o.IsVerified {
		pills = append(pills, ui.TintedPill("verified", ui.AppColors.Link, lipgloss.Darken(ui.AppColors.Link, 0.7)))
	}
	switch {
	case o.ViewerCanAdminister:
		pills = append(pills, ui.TintedPill("admin", ui.AppColors.Good, ui.AppColors.GoodTint))
	case o.ViewerIsAMember:
		pills = append(pills, ui.GhostPill("member"))
	}
	tags, _ := ui.JoinFit(pills, " ", width-lipgloss.Width(title)-2)
	if tags == "" {
		return ui.Fit(title, width)
	}
	return ui.Fit(title+"  "+tags, width)
}

// overviewDescription is the description or bio on one line, or why there
// is none: still loading, unavailable, or simply not written.
func (m *Model) overviewDescription(width int) string {
	quiet := ui.DimStyle.Italic(true)
	switch {
	case m.ownerErr != nil:
		return ui.Fit(quiet.Render("Profile unavailable: "+firstLine(m.ownerErr.Error())), width)
	case !m.ownerLoaded:
		return quiet.Render("Loading profile…")
	case strings.TrimSpace(m.owner.Description) == "":
		return quiet.Render("No description")
	default:
		return ui.TextBodyStyle.Render(ui.Fit(strings.Join(strings.Fields(m.owner.Description), " "), width))
	}
}

// firstLine is the first line of an error message, which is all the
// description line has room for.
func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}

// overviewFacts are the figures from the profile, most useful first, as
// chips: a number in the foreground colour with what it counts in dim. The
// repository counts are left to the Repo mix card.
func (m *Model) overviewFacts() []string {
	o := m.owner
	if !m.ownerLoaded {
		return nil
	}
	fact := func(n int, word string) string {
		return ui.TextBodyStyle.Render(repo.CompactNumber(n)) + ui.DimStyle.Render(" "+word)
	}
	var facts []string
	if m.isUser {
		facts = append(facts, fact(o.Members, ui.Plural(o.Members, "follower", "followers")), fact(o.Following, "following"))
	} else {
		facts = append(facts, fact(o.Members, ui.Plural(o.Members, "member", "members")))
		if o.Teams > 0 {
			facts = append(facts, fact(o.Teams, ui.Plural(o.Teams, "team", "teams")))
		}
	}
	if !o.CreatedAt.IsZero() {
		facts = append(facts, ui.DimStyle.Render("since ")+ui.TextBodyStyle.Render(fmt.Sprint(o.CreatedAt.Year())))
	}
	for _, text := range []string{o.Company, o.Location} {
		if text != "" {
			facts = append(facts, ui.TextBodyStyle.Render(text))
		}
	}
	if o.WebsiteUrl != "" {
		facts = append(facts, ui.LinkTextStyle.Render(o.WebsiteUrl))
	}
	if len(o.Pinned) > 0 {
		facts = append(facts, ui.DimStyle.Render("pinned ")+ui.TextBodyStyle.Render(strings.Join(o.Pinned, ", ")))
	}
	return facts
}

// cardLine is one line of a card's chart and the slices drawn on it, so
// the slices that fit can be told from the ones cut off.
type cardLine struct {
	text   string
	slices []int
}

// renderCard draws one card in width cells and cardLines lines: a title
// rule with the headline at its right, then the chart. focus is the index
// of the focused slice, or -1 for none. It also returns the slices drawn,
// in order: a slice on a line past the chart's height is left out.
func renderCard(card overview.Card, focused bool, focus, width int) ([]string, []int) {
	lines := []string{cardTitle(card, focused, width)}
	var body []cardLine
	switch card.Kind {
	case overview.Histogram:
		body = histogramBody(card, focus, width)
	case overview.Stacked:
		body = stackedBody(card, focus, width)
	default:
		body = barsBody(card, focus, width)
	}
	var drawn []int
	for i := 0; i < cardBody; i++ {
		if i >= len(body) {
			lines = append(lines, ui.Fit("", width))
			continue
		}
		lines = append(lines, ui.Fit(body[i].text, width))
		drawn = append(drawn, body[i].slices...)
	}
	sort.Ints(drawn)
	return lines, drawn
}

// cardTitle is "▸ TITLE ──── headline": the marker and rule in the accent
// colour while focused, dim otherwise.
func cardTitle(card overview.Card, focused bool, width int) string {
	marker, rule := "  ", ui.DimStyle
	if focused {
		marker, rule = ui.AccentStyle.Render("▸ "), ui.AccentStyle
	}
	title := ui.ColumnHeading.Render(strings.ToUpper(card.Title))
	headline := ui.DimStyle.Render(card.Headline)
	fill := width - lipgloss.Width(marker) - lipgloss.Width(title) - lipgloss.Width(headline) - 2
	if fill < 2 {
		return ui.Fit(marker+title, width)
	}
	return marker + title + " " + rule.Render(strings.Repeat("─", fill)) + " " + headline
}

// percent is n of total as a whole percentage, rounded to nearest.
func percent(n, total int) int {
	if total <= 0 {
		return 0
	}
	return int(math.Round(100 * float64(n) / float64(total)))
}

// barsBody is one row per slice: the label, a bar of its share of the
// repositories, the count and the share. A row for a value most
// repositories share names that value and lists the repositories set
// differently instead of drawing a bar.
func barsBody(card overview.Card, focus, width int) []cardLine {
	// Labels take cardLabelWidth, or more for a longer label while the bar
	// keeps at least six cells.
	labelWidth := cardLabelWidth
	for _, s := range card.Slices {
		if s.Majority == "" {
			labelWidth = ui.Max(labelWidth, lipgloss.Width(s.Label))
		}
	}
	labelWidth = ui.Max(1, ui.Min(labelWidth, width-17))
	barWidth := ui.Max(3, width-labelWidth-11)
	var lines []cardLine
	for i, s := range card.Slices {
		label := styleLabel(ui.Fit(s.Label, labelWidth), i == focus)
		if s.Majority != "" {
			shared := ui.ValueStyle.Render(s.Majority) + ui.DimStyle.Render(fmt.Sprintf(" %d", s.Count))
			differ := ui.DimStyle.Render(" · " + s.Minority + " ")
			rest := width - cardLabelWidth - 2 - lipgloss.Width(shared) - lipgloss.Width(differ)
			lines = append(lines, cardLine{" " + label + " " + shared + differ + nameList(s.Names, rest), []int{i}})
			continue
		}
		count := ui.ValueStyle.Render(ui.FitRight(fmt.Sprint(s.Count), 3))
		bar := ui.Bar(ui.Share(share(s.Count, card.Total), barWidth), barWidth, toneStyle(s.Tone))
		pct := ui.DimStyle.Render(fmt.Sprintf("%3d%%", percent(s.Count, card.Total)))
		lines = append(lines, cardLine{" " + label + " " + bar + " " + count + " " + pct, []int{i}})
	}
	if card.Note != "" && len(lines) < cardBody {
		lines = append(lines, cardLine{text: " " + ui.DimStyle.Render(card.Note)})
	}
	return lines
}

// histogramBody is a row of columns two lines tall, one per slice, with
// the labels and counts under them and the note below.
func histogramBody(card overview.Card, focus, width int) []cardLine {
	n := len(card.Slices)
	if n == 0 {
		return nil
	}
	colWidth := ui.Max(1, (width-1)/n)
	glyphWidth := ui.Max(1, ui.Min(colWidth-2, 6))
	peak := ui.Max(1, card.Largest())
	var top, bottom, labels, counts strings.Builder
	for i, s := range card.Slices {
		// The column's height in sixteenths: eight glyph levels a line.
		level := 0
		if s.Count > 0 {
			level = ui.Max(1, s.Count*16/peak)
		}
		style := toneStyle(s.Tone)
		if i == focus {
			style = ui.FocusSparkStyle
		}
		top.WriteString(" " + ui.Fit(style.Render(column(level-8, glyphWidth)), colWidth-1))
		bottom.WriteString(" " + ui.Fit(style.Render(column(level, glyphWidth)), colWidth-1))
		labels.WriteString(" " + ui.Fit(styleLabel(s.Label, i == focus), colWidth-1))
		counts.WriteString(" " + ui.Fit(ui.ValueStyle.Render(fmt.Sprint(s.Count)), colWidth-1))
	}
	// Every column is drawn on the label line.
	all := make([]int, n)
	for i := range all {
		all[i] = i
	}
	lines := []cardLine{{text: top.String()}, {text: bottom.String()}, {labels.String(), all}, {text: counts.String()}}
	if card.Note != "" {
		lines = append(lines, cardLine{text: " " + ui.DimStyle.Render(card.Note)})
	}
	return lines
}

// column is a run of width glyphs at level, from 1 (lowest) to 8 (full);
// a level of zero or less leaves the run blank.
func column(level, width int) string {
	if level <= 0 {
		return strings.Repeat(" ", width)
	}
	return strings.Repeat(string(ui.Sparks[ui.Min(len(ui.Sparks), level)-1]), width)
}

// stackedBody is one bar split between the slices drawn in it, a legend
// of those, then the other slices in lines grouped by their Group. When the
// legend needs two lines and everything would not fit, the bar gives way:
// the legend carries the same counts, and every slice stays on the card.
func stackedBody(card overview.Card, focus, width int) []cardLine {
	var bar strings.Builder
	var legend []string
	var inBar []int
	used := 0
	for i, s := range card.Slices {
		if s.InBar {
			inBar = append(inBar, i)
		}
	}
	// The last segment with any repositories takes the rounding remainder,
	// so an empty slice never gets a cell.
	last := -1
	for _, i := range inBar {
		if card.Slices[i].Count > 0 {
			last = i
		}
	}
	for _, i := range inBar {
		s := card.Slices[i]
		cells := ui.Share(share(s.Count, card.Total), width-1)
		if i == last {
			cells = ui.Max(0, width-1-used)
		}
		cells = ui.Min(cells, width-1-used)
		style := toneStyle(s.Tone)
		if i == focus {
			style = ui.FocusSparkStyle
		}
		bar.WriteString(style.Render(strings.Repeat("━", cells)))
		used += cells
		legend = append(legend, toneStyle(s.Tone).Render("●")+" "+styleLabel(s.Label, i == focus)+" "+ui.ValueStyle.Render(fmt.Sprint(s.Count)))
	}

	// The legend takes one line when it fits, otherwise two or more.
	var legendLines []cardLine
	rest := inBar
	for len(legend) > 0 {
		line, dropped := ui.JoinFit(legend, " ", width-1)
		if dropped == len(legend) {
			line, dropped = legend[0], len(legend)-1
		}
		kept := len(legend) - dropped
		legendLines = append(legendLines, cardLine{" " + line, rest[:kept]})
		legend, rest = legend[kept:], rest[kept:]
	}

	// The rest, one line per group, "branch    main 33 · master 6", or
	// with no label for a group whose words say what they count.
	var groups []string
	byGroup := map[string][]int{}
	for i, s := range card.Slices {
		if s.InBar {
			continue
		}
		if _, ok := byGroup[s.Group]; !ok {
			groups = append(groups, s.Group)
		}
		byGroup[s.Group] = append(byGroup[s.Group], i)
	}
	var groupLines []cardLine
	for _, g := range groups {
		items := make([]string, len(byGroup[g]))
		for n, i := range byGroup[g] {
			items[n] = styleLabel(card.Slices[i].Label, i == focus) + " " + ui.ValueStyle.Render(fmt.Sprint(card.Slices[i].Count))
		}
		label := ""
		if g != "" {
			label = ui.DimStyle.Render(g) + " "
		}
		line, dropped := ui.JoinFit(items, ui.DimStyle.Render(" · "), width-1-lipgloss.Width(label))
		groupLines = append(groupLines, cardLine{" " + label + line, byGroup[g][:len(items)-dropped]})
	}

	lines := []cardLine{{text: " " + bar.String()}}
	if len(legendLines) > 1 && 1+len(legendLines)+len(groupLines) > cardBody {
		lines = nil
	}
	lines = append(lines, legendLines...)
	return append(lines, groupLines...)
}

// styleLabel draws a slice label: reversed in the accent colour when
// focused, so the focus shows even without colour.
func styleLabel(label string, focused bool) string {
	if focused {
		return ui.FocusLabelStyle.Render(label)
	}
	return ui.TextBodyStyle.Render(label)
}

func share(n, total int) float64 {
	if total <= 0 {
		return 0
	}
	return float64(n) / float64(total)
}

// toneStyle is the colour a tone is drawn in.
func toneStyle(tone overview.Tone) lipgloss.Style {
	switch tone {
	case overview.Good:
		return ui.GoodStyle
	case overview.Warn:
		return ui.WarnStyle
	case overview.Bad:
		return ui.BadStyle
	default:
		return ui.AccentStyle
	}
}

// overviewFooter lists the repositories behind the focused slice, in name
// order: a rule naming the card and slice, how many there are, then the
// names in columns. above and below count the cards scrolled out of view.
func (m *Model) overviewFooter(card overview.Card, focus, height, above, below, width int) []string {
	title := " " + ui.ColumnHeading.Render(strings.ToUpper(card.Title))
	var slice overview.Slice
	if focus >= 0 {
		slice = card.Slices[focus]
		label := slice.Label
		if slice.Minority != "" {
			label += ": " + slice.Minority
		}
		title += ui.DimStyle.Render(" › ") + ui.AccentStyle.Bold(true).Render(label)
	}
	// The hints after the rule are shed from the front when space is short;
	// the counts of cards out of view are kept longest.
	noun := card.Noun
	if noun == "" {
		noun = "item"
	}
	hints := []string{"j/k " + noun, "enter show repos"}
	if above > 0 {
		hints = append(hints, fmt.Sprintf("▴ %d above", above))
	}
	if below > 0 {
		hints = append(hints, fmt.Sprintf("▾ %d below", below))
	}
	hint := ""
	for len(hints) > 0 {
		hint = ui.DimStyle.Render(strings.Join(hints, " · "))
		if lipgloss.Width(title)+lipgloss.Width(hint)+6 <= width {
			break
		}
		hints, hint = hints[1:], ""
	}
	fill := ui.Max(1, width-lipgloss.Width(title)-lipgloss.Width(hint)-4)
	lines := []string{ui.Fit(ui.DimStyle.Render("─")+title+" "+ui.DimStyle.Render(strings.Repeat("─", fill))+" "+hint+" ", width)}

	if focus < 0 {
		return append(lines, " "+ui.DimStyle.Render("No repositories to list on this card"))
	}
	listed := m.named(slice.Names)
	figures := fmt.Sprintf("%d %s", len(listed), ui.Plural(len(listed), "repo", "repos"))
	if card.Total > 0 {
		figures += fmt.Sprintf(" · %d%% of %d", percent(len(listed), card.Total), card.Total)
	}
	lines = append(lines, " "+ui.TextBodyStyle.Render(figures))

	// Names in columns, as many lines as the footer has left.
	perLine := ui.Max(1, (width-1)/footerNameWidth)
	for i := 0; i < len(listed) && len(lines) < height; i += perLine {
		// The last line gives its final cell to a count of what is left.
		end := ui.Min(len(listed), i+perLine)
		more := 0
		if len(lines) == height-1 && end < len(listed) {
			end = ui.Max(i, i+perLine-1)
			more = len(listed) - end
		}
		var cells []string
		for _, c := range listed[i:end] {
			cells = append(cells, inspector.VisibilityMark(c)+" "+ui.ValueStyle.Render(ui.Fit(c.Name, footerNameWidth-3)))
		}
		if more > 0 {
			cells = append(cells, ui.DimStyle.Render(fmt.Sprintf("+%d more", more)))
		}
		lines = append(lines, " "+strings.Join(cells, " "))
	}
	return lines
}

// named is the visible repositories with the given names, in the order they
// are listed: by name.
func (m *Model) named(names []string) []repo.RepoConfig {
	want := make(map[string]bool, len(names))
	for _, name := range names {
		want[name] = true
	}
	var out []repo.RepoConfig
	for _, c := range m.visible {
		if want[c.Name] {
			out = append(out, c)
		}
	}
	return out
}

// nameList joins as many names as fit in width, then counts the rest, so
// a long list ends in "+12 more" rather than a cut name.
func nameList(names []string, width int) string {
	for _, more := range []string{" +%d more", " +%d"} {
		for shown := len(names); shown > 0; shown-- {
			line := ui.ValueStyle.Render(strings.Join(names[:shown], ", "))
			if rest := len(names) - shown; rest > 0 {
				line += ui.DimStyle.Render(fmt.Sprintf(more, rest))
			}
			if lipgloss.Width(line) <= width {
				return line
			}
		}
	}
	return ui.DimStyle.Render(fmt.Sprintf("%d repos", len(names)))
}
