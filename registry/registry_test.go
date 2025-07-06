package registry_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/newrushbolt/go-ethtool-exporter/registry"
	"github.com/stretchr/testify/assert"
)

func TestRegistrySimpleMetric(t *testing.T) {
	expectedMetricResult := `test_metric{key1="value1"} 16.13
`

	var metricList registry.Registry
	metricRecordSimple := registry.MetricRecord{
		Name:   "test_metric",
		Labels: map[string]string{"key1": "value1"},
		Value:  16.13,
	}
	metricList = append(metricList, metricRecordSimple)

	textFilePath := fmt.Sprintf("../.TestRegistryData-%d.prom", time.Now().UnixNano())
	metricList.MustWriteTextfile(textFilePath)

	metricsResult, err := os.ReadFile(textFilePath)

	assert.NoError(t, err)
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
	metricRecordTooMuchLabels := registry.MetricRecord{
		Name:   "test_metric",
		Labels: tooMuchLabels,
		Value:  1,
	}
	var metricList registry.Registry
	metricList = append(metricList, metricRecordTooMuchLabels)

	textFilePath := fmt.Sprintf("../.TestRegistryData-%d.prom", time.Now().UnixNano())
	metricList.MustWriteTextfile(textFilePath)

	metricsResult, err := os.ReadFile(textFilePath)
	assert.NoError(t, err)
	assert.Empty(t, string(metricsResult))
}

func TestRegistryBrokenPath(t *testing.T) {
	var metricList registry.Registry

	textFilePath := fmt.Sprintf("/non-existed-root-dir/.TestRegistryData-%d.prom", time.Now().UnixNano())
	assert.Panics(t, func() { metricList.MustWriteTextfile(textFilePath) })
}
