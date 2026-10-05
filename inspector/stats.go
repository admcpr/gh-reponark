package inspector

import (
	"fmt"
	"image/color"
	"math"
	"sort"
	"strings"
	"time"

	"gh-reponark/repo"
	"gh-reponark/ui"

	"charm.land/lipgloss/v2"
)

// Stats summarises one property across every loaded repository, so the
// inspector can show a value in context: a bar relative to the largest, how
// many repositories share a setting, or where a date falls among the rest.
// The organisation-wide parts come from OrgStats; StatsFor adds the parts
// that depend on the inspected repository.
type Stats struct {
	Total int
	// Ints: the largest value and the sorted values, for histograms.
	Max    int
	Values []int
	// Bools: how many repositories have the setting on.
	On int
	// Strings: how many repositories share the inspected value.
	Same int
	// Times: how many repositories have an older value than the inspected one.
	Older int
}

// histogramBuckets is how many columns an int distribution is drawn with.
const histogramBuckets = 10

// sparkline glyphs from empty to full.
var sparks = []rune("▁▂▃▄▅▆▇█")

// Sparkline colours: the histogram's bars, and the bucket holding the
// inspected value. Empty buckets keep a baseline in ui.TrackStyle.
var (
	sparkStyle     = lipgloss.NewStyle().Foreground(ui.AppColors.Blue)
	sparkHighlight = lipgloss.NewStyle().Foreground(ui.AppColors.Pink)
)

// barWidth is how many cells an inline bar beside an integer takes.
const barWidth = 10

// OrgStats computes, in one pass over repos, the parts of every property's
// Stats that do not depend on which repository is inspected: the total, and
// for ints the largest and sorted values, for bools how many are on.
func OrgStats(repos []repo.RepoConfig) map[string]Stats {
	stats := map[string]Stats{}
	for _, c := range repos {
		for name, p := range c.Properties {
			s := stats[name]
			switch value := p.Value.(type) {
			case int:
				s.Values = append(s.Values, value)
				s.Max = ui.Max(s.Max, value)
			case bool:
				if value {
					s.On++
				}
			}
			stats[name] = s
		}
	}
	for name, s := range stats {
		s.Total = len(repos)
		sort.Ints(s.Values)
		stats[name] = s
	}
	return stats
}

// StatsFor completes the Stats for property name relative to current: how
// many of repos share its string value, or have an older date. It is one
// scan of repos; org holds what OrgStats computed for them.
func StatsFor(org map[string]Stats, name string, repos []repo.RepoConfig, current repo.RepoConfig) Stats {
	s := org[name]
	switch value := current.Properties[name].Value.(type) {
	case string:
		for _, c := range repos {
			if c.Text(name) == value {
				s.Same++
			}
		}
	case time.Time:
		for _, c := range repos {
			if t := c.Time(name); !t.IsZero() && t.Before(value) {
				s.Older++
			}
		}
	}
	return s
}

// Histogram draws the distribution of an int property as a sparkline of
// histogramBuckets columns from zero to the largest value, with the column
// holding value picked out.
func (s Stats) Histogram(value int) string {
	if s.Max <= 0 || len(s.Values) == 0 {
		return ""
	}
	counts := make([]int, histogramBuckets)
	bucket := func(v int) int {
		return ui.Min(histogramBuckets-1, v*histogramBuckets/(s.Max+1))
	}
	for _, v := range s.Values {
		counts[bucket(v)]++
	}
	return Sparkline(counts, bucket(value), sparkStyle)
}

// Sparkline draws counts as a row of bars scaled to the largest count, in
// style, with the bar at highlight (-1 for none) picked out in pink. An
// empty bucket keeps a dim baseline so the shape reads.
func Sparkline(counts []int, highlight int, style lipgloss.Style) string {
	peak := 0
	for _, n := range counts {
		peak = ui.Max(peak, n)
	}
	// Consecutive cells in the same style are rendered as one run.
	styles := [...]lipgloss.Style{ui.TrackStyle, style, sparkHighlight}
	const track, filled, highlighted = 0, 1, 2
	var b, run strings.Builder
	current := -1
	flush := func() {
		if run.Len() > 0 {
			b.WriteString(styles[current].Render(run.String()))
			run.Reset()
		}
	}
	for i, n := range counts {
		glyph, kind := sparks[0], track
		if n > 0 {
			glyph, kind = sparks[ui.Min(len(sparks)-1, (n*len(sparks)-1)/peak)], filled
		}
		if i == highlight {
			kind = highlighted
		}
		if kind != current {
			flush()
			current = kind
		}
		run.WriteRune(glyph)
	}
	flush()
	return b.String()
}

// sqrtShare is value's share of max on a square-root scale. Counts in an
// organisation are long-tailed, so a repository with a tenth of the stars of
// the most popular one still shows a third of a bar rather than a sliver.
func sqrtShare(value, max int) float64 {
	if max <= 0 || value <= 0 {
		return 0
	}
	return math.Sqrt(float64(value) / float64(max))
}

// bar draws value as a width-cell bar relative to max, the largest value in
// the organisation, filled in style; any non-zero value gets at least one
// cell so it can be told from zero.
func bar(value, max, width int, style lipgloss.Style) string {
	return ui.Bar(ui.Share(sqrtShare(value, max), width), width, style)
}

// Context describes where the inspected value sits in the organisation, e.g.
// "on in 31 of 40 repos" or "newer than 29 of 40 repos".
func (s Stats) Context(p repo.RepoProperty) string {
	if s.Total == 0 {
		return ""
	}
	switch value := p.Value.(type) {
	case bool:
		return fmt.Sprintf("on in %d of %d repos", s.On, s.Total)
	case int:
		if p.Identifier {
			return ""
		}
		rank := sort.SearchInts(s.Values, value)
		largest := p
		largest.Value = s.Max
		return fmt.Sprintf("more than %d of %d repos · max %s", rank, s.Total, intText(largest, true))
	case string:
		if value == "" {
			return fmt.Sprintf("unset · %d of %d repos share this", s.Same, s.Total)
		}
		return fmt.Sprintf("shared by %d of %d repos", s.Same, s.Total)
	case time.Time:
		if value.IsZero() {
			return ""
		}
		return fmt.Sprintf("newer than %d of %d repos", s.Older, s.Total)
	}
	return ""
}

// FractionBar is a width-cell bar showing n of total, for the share of
// repositories with a setting on.
func FractionBar(n, total, width int) string {
	if total <= 0 {
		return ""
	}
	return ui.Bar(ui.Share(float64(n)/float64(total), width), width, ui.GoodStyle)
}

// tile is one tile of the inspector's stats strip: a labelled number with a
// bar showing where it sits against the rest of the organization.
type tile struct {
	label string
	icon  string
	// property is the int property the tile shows, or "" for the push age.
	property string
	// The bar runs from the tile's colour to a lighter tint of it; the icon
	// and value take the colour itself.
	color, light color.Color
	iconStyle    lipgloss.Style
	valueStyle   lipgloss.Style
}

func newTile(label, icon string, c color.Color, property string) tile {
	return tile{
		label:      label,
		icon:       icon,
		property:   property,
		color:      c,
		light:      lipgloss.Lighten(c, 0.35),
		iconStyle:  lipgloss.NewStyle().Foreground(c),
		valueStyle: lipgloss.NewStyle().Foreground(c).Bold(true),
	}
}

// tiles are the stat tiles in priority order; a narrow pane shows the first
// few. The icons stay inside the arrows and geometric-shapes blocks that
// every monospace font covers.
var tiles = []tile{
	newTile("STARS", "★", ui.AppColors.Warn, "Stargazer Count"),
	newTile("FORKS", "⎇", ui.AppColors.Purple, "Fork Count"),
	newTile("ISSUES", "◎", ui.AppColors.Good, "Open Issues"),
	newTile("PUSHED", "↑", ui.AppColors.Pink, ""),
	newTile("PRS", "⇄", ui.AppColors.Blue, "Open Pull Requests"),
	newTile("SIZE", "▣", ui.AppColors.Accent, "Disk Usage"),
}

// Tile widths: the narrowest a tile can be and still fit a label, a compact
// number and a bar worth reading, and the pane width from which the sixth
// tile (SIZE) joins the strip.
const (
	minTileWidth  = 11
	sizeTileWidth = 70
	statsLines    = 3
)

// tileCount is how many stat tiles fit across width: none when fewer than
// two would, and SIZE only from sizeTileWidth.
func tileCount(width int) int {
	count := ui.Min(len(tiles), width/minTileWidth)
	if width < sizeTileWidth {
		count = ui.Min(count, len(tiles)-1)
	}
	if count < 2 {
		return 0
	}
	return count
}

// value is the tile's number as text, and how full its bar is.
func (s tile) value(c repo.RepoConfig, stats map[string]Stats) (string, float64) {
	if s.property == "" {
		pushed := c.Time("Pushed At")
		if pushed.IsZero() {
			return "—", 0
		}
		// A repo pushed today fills the bar; one untouched for a year empties it.
		age := now().Sub(pushed).Hours() / 24
		return Ago(pushed), math.Max(0, 1-age/365)
	}
	p := c.Properties[s.property]
	value, _ := p.Value.(int)
	max := stats[s.property].Max
	if max <= 0 {
		max = value
	}
	return intText(p, true), sqrtShare(value, max)
}

// bar draws the tile's progress bar: a gradient from the tile colour to a
// lighter tint of it, spanning the whole width so the filled part shows the
// start of it, over the track every bar shares.
func (s tile) bar(fraction float64, width int) string {
	if width <= 0 {
		return ""
	}
	filled := ui.Max(0, ui.Min(width, int(math.Round(fraction*float64(width)))))
	return ui.GradientRun(strings.Repeat("█", filled), lipgloss.Blend1D(width, s.color, s.light)) +
		ui.TrackStyle.Render(strings.Repeat("░", width-filled))
}

// RenderStats lays out count stat tiles across width in three lines: label,
// value and bar. It returns nothing when count is zero.
func RenderStats(c repo.RepoConfig, stats map[string]Stats, width, count int) []string {
	if count <= 0 {
		return nil
	}
	tileWidth := width / count
	var labels, values, bars strings.Builder
	for _, s := range tiles[:count] {
		text, fraction := s.value(c, stats)
		inner := tileWidth - 1 // a gutter on the right of every tile
		labels.WriteString(ui.Fit(s.iconStyle.Render(s.icon)+" "+ui.ColumnHeading.Render(s.label), tileWidth))
		values.WriteString(ui.Fit(s.valueStyle.Render(text), tileWidth))
		bars.WriteString(ui.Fit(s.bar(fraction, inner), tileWidth))
	}
	return []string{
		ui.Fit(labels.String(), width),
		ui.Fit(values.String(), width),
		ui.Fit(bars.String(), width),
	}
}
