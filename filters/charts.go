package filters

import (
	"math"
	"strings"
)

var sparkLevels = []rune("▁▂▃▄▅▆▇█")

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

// sparkline draws counts as a row of bars scaled to the largest count. Empty
// buckets stay blank so gaps in the data show.
func sparkline(counts []int) string {
	most := 0
	for _, c := range counts {
		most = max(most, c)
	}
	var b strings.Builder
	for _, c := range counts {
		switch {
		case c == 0:
			b.WriteRune(' ')
		default:
			level := (c*len(sparkLevels) - 1) / most
			b.WriteRune(sparkLevels[min(level, len(sparkLevels)-1)])
		}
	}
	return b.String()
}

// blocks is a bar of width cells, the first fraction of them filled.
func blocks(fraction float64, width int) (filled, empty string) {
	n := int(math.Round(fraction * float64(width)))
	if fraction > 0 && n == 0 {
		n = 1
	}
	n = max(0, min(n, width))
	return strings.Repeat("█", n), strings.Repeat("█", width-n)
}
