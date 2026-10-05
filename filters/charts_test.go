package filters

import (
	"gh-reponark/ui/uitest"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestHistogram(t *testing.T) {
	assert.Equal(t, []int{2, 0, 1, 1}, histogram([]float64{0, 1, 5, 8}, 4))
	assert.Equal(t, []int{3, 0}, histogram([]float64{2, 2, 2}, 2), "equal values share the first bucket")
	assert.Equal(t, []int{0, 0}, histogram(nil, 2))
}

func TestModel_Charts(t *testing.T) {
	m := newTestModel()

	tests := []struct {
		query string
		want  []string
	}{
		{query: "is archived", want: []string{"● yes 1   ○ no 2"}},
		{query: "stargazer", want: []string{"3", "900", "median 40"}},
		{query: "language", want: []string{"Go", "Ruby"}},
		{query: "pushed", want: []string{"No repos have this date set"}},
		{query: "homepage", want: []string{"No repos have a value set"}},
		{query: "name with owner", want: []string{"No repos have a value set"}},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			press(m, "/", "esc")
			selectProperty(m, tt.query)
			p, _ := m.selected()
			chart := ansi.Strip(strings.Join(m.chart(p, 40), "\n"))
			for _, want := range tt.want {
				assert.Contains(t, chart, want)
			}
		})
	}
}

func TestTextChart(t *testing.T) {
	lines := textChart(testRepos(), "Primary Language", 40)

	assert.Len(t, lines, 2)
	assert.True(t, strings.HasPrefix(ansi.Strip(lines[0]), "Go "), "the most common value comes first")
	assert.True(t, strings.HasSuffix(ansi.Strip(lines[0]), " 2"))
	assert.Less(t, strings.Count(lines[1], "━"), strings.Count(lines[0], "━"), "bars are proportional")

	assert.Equal(t, []string{"Every repo has a different value"}, uitest.StripAll(textChart(testRepos(), "Name", 40)))
}
