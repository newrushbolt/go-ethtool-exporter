package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatPrometheusLineNoLabels(t *testing.T) {
	rec := MetricRecord{Name: "metricNoLabels", Labels: map[string]string{}, Value: 42}
	line, err := rec.FormatPrometheusLine()
	assert.NoError(t, err)
	assert.Equal(t, "metricNoLabels{} 42", line)
}

func TestFormatPrometheusLineLabelSorting(t *testing.T) {
	expectedSortedLine := `metricLabelSorting{a="1",b="2"} 1`
	rec := MetricRecord{
		Name:   "metricLabelSorting",
		Labels: map[string]string{"b": "2", "a": "1"},
		Value:  1,
	}
	line, err := rec.FormatPrometheusLine()

	assert.NoError(t, err)
	assert.Equal(t, line, expectedSortedLine)
}
