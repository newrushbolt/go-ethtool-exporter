package registry_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/newrushbolt/go-ethtool-exporter/registry"
	"github.com/stretchr/testify/assert"
)

func TestRegistryCommon(t *testing.T) {
	expectedMetricResult := `test_metric{key1="value1"} 16.13
`

	var metricList registry.Registry
	metricRecordSimple := registry.MetricRecord{
		Name:   "test_metric",
		Labels: map[string]string{"key1": "value1"},
		Value:  16.13,
	}
	metricList = append(metricList, metricRecordSimple)

	tooMuchLabels := map[string]string{
		"key1":  "value1",
		"key2":  "value1",
		"key3":  "value1",
		"key4":  "value1",
		"key5":  "value1",
		"key6":  "value1",
		"key7":  "value1",
		"key8":  "value1",
		"ke9":   "value1",
		"key10": "value1",
		"key11": "value1",
		"key12": "value1",
		"key13": "value1",
		"key14": "value1",
		"key15": "value1",
		"key16": "value1",
		"key17": "value1",
	}
	metricRecordTooMuchLabels := registry.MetricRecord{
		Name:   "test_metric",
		Labels: tooMuchLabels,
		Value:  1,
	}
	metricList = append(metricList, metricRecordTooMuchLabels)

	textFilePath := fmt.Sprintf("../.TestRegistryData-%d.prom", time.Now().Unix())
	metricList.MustWriteTextfile(textFilePath)

	metricsResult, err := os.ReadFile(textFilePath)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, expectedMetricResult, string(metricsResult))
}

func TestRegistryBrokenPath(t *testing.T) {
	var metricList registry.Registry

	textFilePath := fmt.Sprintf("/non-existed-root-dir/.TestRegistryData-%d.prom", time.Now().Unix())
	assert.Panics(t, func() { metricList.MustWriteTextfile(textFilePath) }, "Should panic without proper path")
}
