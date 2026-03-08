package registry

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRegistrySimpleMetric(t *testing.T) {
	expectedMetricResult := `test_metric{key1="value1"} 16.13`
	var metricList Registry
	metricRecordSimple := MetricRecord{
		Name:   "test_metric",
		Labels: map[string]string{"key1": "value1"},
		Value:  16.13,
	}
	metricList = append(metricList, metricRecordSimple)

	metricsResult := metricList.FormatTextfileString(Prometheus_0_0_4)
	assert.Equal(t, expectedMetricResult, string(metricsResult))
}

func TestRegistryTooManyLabels(t *testing.T) {
	tooMuchLabels := map[string]string{
		"key1":  "value1",
		"key2":  "value2",
		"key3":  "value3",
		"key4":  "value4",
		"key5":  "value5",
		"key6":  "value6",
		"key7":  "value7",
		"key8":  "value8",
		"key9":  "value9",
		"key10": "value10",
		"key11": "value11",
		"key12": "value12",
		"key13": "value13",
		"key14": "value14",
		"key15": "value15",
		"key16": "value16",
		"key17": "value17",
	}
	metricRecordTooMuchLabels := MetricRecord{
		Name:   "test_metric",
		Labels: tooMuchLabels,
		Value:  1,
	}
	var metricList Registry
	metricList = append(metricList, metricRecordTooMuchLabels)

	metricsResult := metricList.FormatTextfileString(Prometheus_0_0_4)
	assert.Empty(t, string(metricsResult))
}

func TestTextfileBrokenPath(t *testing.T) {
	textFilePath := fmt.Sprintf("/non-existed-root-dir/.TestRegistryData-%d.prom", time.Now().UnixNano())
	metrics := ""
	assert.Panics(t, func() { MustWriteTextfile(textFilePath, metrics) })
	defer os.Remove(textFilePath)
}

func TestTextfileWriteError(t *testing.T) {
	var metricList Registry
	metricList = append(metricList, MetricRecord{
		Name:   "test_metric",
		Labels: map[string]string{"key": "value"},
		Value:  1,
	})
	dirPath := os.TempDir()
	metrics := metricList.FormatTextfileString(Prometheus_0_0_4)
	assert.Panics(t, func() { MustWriteTextfile(dirPath, metrics) })
}

func TestGetMetricIndex_NoMatch(t *testing.T) {
	metricList := Registry{
		{Name: "foo", Labels: nil, Value: 1},
		{Name: "bar", Labels: nil, Value: 2},
	}
	idx, err := metricList.GetMetricIndex("baz")
	assert.Equal(t, -1, idx)
	assert.NoError(t, err)
}

func TestGetMetricIndex_OneMatch(t *testing.T) {
	metricList := Registry{
		{Name: "foo", Labels: nil, Value: 1},
		{Name: "bar", Labels: nil, Value: 2},
	}
	idx, err := metricList.GetMetricIndex("foo")
	assert.Equal(t, 0, idx)
	assert.NoError(t, err)
}

func TestGetMetricIndex_MultipleMatches(t *testing.T) {
	metricList := Registry{
		{Name: "foo", Labels: nil, Value: 1},
		{Name: "foo", Labels: nil, Value: 2},
	}
	idx, err := metricList.GetMetricIndex("foo")
	assert.Equal(t, -1, idx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "multiple metrics with the same name <foo>")
}

func TestRegistryOpenMetricsAnnotations(t *testing.T) {
	var metricList Registry
	metricList = append(metricList, MetricRecord{Name: "m1", Labels: nil, Value: 1})
	metricList = append(metricList, MetricRecord{Name: "m2", Labels: map[string]string{"k": "v"}, Value: 2})
	result := metricList.FormatTextfileString(OpenMetrics_1_0_0)
	// should contain TYPE lines for both metrics and EOF at the end
	assert.Contains(t, result, "# TYPE m1 UNKNOWN")
	assert.Contains(t, result, "# TYPE m2 UNKNOWN")
	assert.True(t, strings.HasSuffix(result, "# EOF"))
}

func TestNamespacedMetricName_EmptyNamespace(t *testing.T) {
	// Temporarily overriding OutputMetricNamespace for testing
	// Will be set back to "" by defer so it wont affect other tests
	OutputMetricNamespace = ""
	defer func() { OutputMetricNamespace = "" }()

	assert.Equal(t, "my_metric", namespacedMetricName("my_metric"))
}

func TestNamespacedMetricName_WithNamespace(t *testing.T) {
	OutputMetricNamespace = "ethtool"
	defer func() { OutputMetricNamespace = "" }()

	assert.Equal(t, "ethtool_my_metric", namespacedMetricName("my_metric"))
}

func TestNamespacedMetricName_AlreadyPrefixed(t *testing.T) {
	OutputMetricNamespace = "ethtool"
	defer func() { OutputMetricNamespace = "" }()

	assert.Equal(t, "ethtool_my_metric", namespacedMetricName("ethtool_my_metric"))
}
