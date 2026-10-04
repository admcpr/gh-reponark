package filters

import "math"

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
