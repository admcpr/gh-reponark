package filters

import (
	"fmt"
	"math"
	"sort"
	"time"

	"gh-reponark/inspector"
	"gh-reponark/repo"
	"gh-reponark/ui"
)

// histogram counts values into buckets spread evenly between the smallest
// and largest value.
func histogram(values []float64, buckets int) []int {
	counts := make([]int, buckets)
	if len(values) == 0 || buckets == 0 {
		return counts
	}
	lo, hi := values[0], values[0]
	for _, v := range values {
		lo, hi = math.Min(lo, v), math.Max(hi, v)
	}
	for _, v := range values {
		i := 0
		if hi > lo {
			i = int((v - lo) / (hi - lo) * float64(buckets))
		}
		counts[min(i, buckets-1)]++
	}
	return counts
}

// chartSpan is the widest a chart's bar or sparkline is drawn.
const chartSpan = 40

// chart shows how the loaded repositories' values of p are spread, so a
// filter can be chosen without guessing.
func (m *Model) chart(p repo.PropertySchema, width int) []string {
	switch p.Type {
	case "bool":
		return boolChart(m.repos, p.Name, width)
	case "int":
		values := make([]float64, len(m.repos))
		for i, c := range m.repos {
			values[i] = float64(c.Int(p.Name))
		}
		return rangeChart(values, width, func(v float64) string { return fmt.Sprint(int(v)) })
	case "time.Time":
		var values []float64
		for _, c := range m.repos {
			if t := c.Time(p.Name); !t.IsZero() {
				values = append(values, float64(t.Unix()))
			}
		}
		if len(values) == 0 {
			return []string{ui.TextBodyStyle.Render("No repos have this date set")}
		}
		lines := rangeChart(values, width, func(v float64) string { return time.Unix(int64(v), 0).UTC().Format("2006-01-02") })
		if missing := len(m.repos) - len(values); missing > 0 {
			lines = append(lines, ui.DimStyle.Render(fmt.Sprintf("%d with no date", missing)))
		}
		return lines
	default:
		return textChart(m.repos, p.Name, width)
	}
}

// boolChart is the share of repos with the property on, as the same
// fraction bar the inspector draws under a toggle, with the counts below.
func boolChart(repos []repo.RepoConfig, name string, width int) []string {
	yes := 0
	for _, c := range repos {
		if c.Bool(name) {
			yes++
		}
	}
	return []string{
		inspector.FractionBar(yes, len(repos), ui.Min(width, chartSpan)),
		ui.GoodStyle.Render("● yes ") + ui.ValueStyle.Render(fmt.Sprint(yes)) + "   " +
			ui.DimStyle.Render("○ no ") + ui.ValueStyle.Render(fmt.Sprint(len(repos)-yes)),
	}
}

// rangeChart is a sparkline of how values spread between the smallest and
// largest, labelled at both ends, with the median below. The bars are in
// the accent colour and empty buckets keep a baseline, like the inspector's
// histogram.
func rangeChart(values []float64, width int, format func(float64) string) []string {
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	lo, hi, median := sorted[0], sorted[len(sorted)-1], sorted[len(sorted)/2]

	span := ui.Min(width, chartSpan)
	labels := ui.Fit(ui.ValueStyle.Render(format(lo)), span/2) + ui.FitRight(ui.ValueStyle.Render(format(hi)), span-span/2)
	return []string{
		inspector.Sparkline(histogram(values, span), -1, ui.AccentStyle),
		labels,
		ui.TextBodyStyle.Render("median ") + ui.ValueStyle.Render(format(median)),
	}
}

// textChart bars the most common values of a text property.
func textChart(repos []repo.RepoConfig, name string, width int) []string {
	counts := map[string]int{}
	for _, c := range repos {
		counts[c.Text(name)]++
	}
	if counts[""] == len(repos) {
		return []string{ui.TextBodyStyle.Render("No repos have a value set")}
	}
	if len(counts) == len(repos) && len(repos) > 1 {
		return []string{ui.TextBodyStyle.Render("Every repo has a different value")}
	}

	values := make([]string, 0, len(counts))
	for v := range counts {
		values = append(values, v)
	}
	sort.Slice(values, func(i, j int) bool {
		if counts[values[i]] != counts[values[j]] {
			return counts[values[i]] > counts[values[j]]
		}
		return values[i] < values[j]
	})
	if len(values) > 5 {
		values = values[:5]
	}

	labelWidth := ui.Min(16, width/3)
	barWidth := ui.Max(1, ui.Min(24, width-labelWidth-5))
	lines := make([]string, len(values))
	for i, v := range values {
		label := ui.ValueStyle.Render(ui.Fit(v, labelWidth))
		if v == "" {
			label = ui.DimStyle.Render(ui.Fit("(none)", labelWidth))
		}
		share := ui.Share(float64(counts[v])/float64(counts[values[0]]), barWidth)
		lines[i] = label + " " + ui.Bar(share, barWidth, ui.AccentStyle) + " " + ui.ValueStyle.Render(fmt.Sprint(counts[v]))
	}
	return lines
}
