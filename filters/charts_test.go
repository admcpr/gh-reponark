package filters

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHistogram(t *testing.T) {
	assert.Equal(t, []int{2, 0, 1, 1}, histogram([]float64{0, 1, 5, 8}, 4))
	assert.Equal(t, []int{3, 0}, histogram([]float64{2, 2, 2}, 2), "equal values share the first bucket")
	assert.Equal(t, []int{0, 0}, histogram(nil, 2))
}

func TestSparkline(t *testing.T) {
	assert.Equal(t, "█ ▄▁", sparkline([]int{8, 0, 4, 1}))
	assert.Equal(t, "  ", sparkline([]int{0, 0}))
}

func TestBlocks(t *testing.T) {
	filled, empty := blocks(0.25, 8)
	assert.Equal(t, "██", filled)
	assert.Equal(t, "██████", empty)

	filled, _ = blocks(0.01, 8)
	assert.Equal(t, "█", filled, "any share at all gets a cell")
}
