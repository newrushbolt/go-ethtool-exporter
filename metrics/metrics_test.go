package metrics

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newrushbolt/go-ethtool-exporter/registry"

	"github.com/newrushbolt/go-ethtool-metrics/pkg/metrics/statistics"
)

func TestDropAllNils(t *testing.T) {
	expectedMetricResult := `prefix_real_float64{} 16.13
missing_metrics_skipped_total{} 2`

	type NilStruct struct {
		Key   string
		Value string
	}

	type TestStruct struct {
		RealFloat64 *float64
		NilFloat64  *float64
		NilString   *string
		NilStruct   *NilStruct
	}

	nilObject := &TestStruct{}
	realFloat64 := 16.13
	nilObject.RealFloat64 = &realFloat64

	metricRegistry := registry.Registry{}
	prefixes := []string{"prefix"}
	labels := map[string]string{}
	MetricListFromStructs(nilObject, &metricRegistry, prefixes, labels, AbsentMetricsConfig{}, "single-label")

	metricRegistryResult := metricRegistry.FormatTextfileString(registry.Prometheus_0_0_4)
	assert.Equal(t, expectedMetricResult, metricRegistryResult)
}

func TestKeepFloat64Nils(t *testing.T) {
	expectedMetricResult := `prefix_real_float64{} 16.13
prefix_nil_float64{} NaN
missing_metrics_skipped_total{} 2`
	type NilStruct struct {
		Key   string
		Value string
	}

	type TestStruct struct {
		RealFloat64 *float64
		NilFloat64  *float64
		NilString   *string
		NilStruct   *NilStruct
	}

	nilObject := &TestStruct{}
	realFloat64 := 16.13
	nilObject.RealFloat64 = &realFloat64

	metricRegistry := registry.Registry{}
	prefixes := []string{"prefix"}
	labels := map[string]string{}
	absentMetrics := AbsentMetricsConfig{
		ExposeNan:          true,
		ExposeTotalCounter: false,
		ExposeDetailedInfo: false,
	}
	MetricListFromStructs(nilObject, &metricRegistry, prefixes, labels, absentMetrics, "single-label")

	metricRegistryResult := metricRegistry.FormatTextfileString(registry.Prometheus_0_0_4)
	assert.Equal(t, expectedMetricResult, metricRegistryResult)
}

func TestMissingMetricsExposeDetailedInfo(t *testing.T) {
	expectedMetricResult := `missing_metric_info{metric_name="prefix_nil_float64"} 1`
	type TestStruct struct {
		NilFloat64 *float64
	}

	nilObject := &TestStruct{}

	metricRegistry := registry.Registry{}
	prefixes := []string{"prefix"}
	labels := map[string]string{}
	absentMetrics := AbsentMetricsConfig{
		ExposeNan:          false,
		ExposeTotalCounter: false,
		ExposeDetailedInfo: true,
	}
	MetricListFromStructs(nilObject, &metricRegistry, prefixes, labels, absentMetrics, "single-label")

	metricRegistryResult := metricRegistry.FormatTextfileString(registry.Prometheus_0_0_4)
	assert.Equal(t, expectedMetricResult, metricRegistryResult)
}

func TestMissingMetricsExposeTotalCounter(t *testing.T) {
	expectedMetricResult := `missing_metrics_total{} 2`
	type TestStruct struct {
		NilFloat64A *float64
		NilFloat64B *float64
	}

	nilObject := &TestStruct{}

	metricRegistry := registry.Registry{}
	prefixes := []string{"prefix"}
	labels := map[string]string{}
	absentMetrics := AbsentMetricsConfig{
		ExposeNan:          false,
		ExposeTotalCounter: true,
		ExposeDetailedInfo: false,
	}
	MetricListFromStructs(nilObject, &metricRegistry, prefixes, labels, absentMetrics, "single-label")

	metricRegistryResult := metricRegistry.FormatTextfileString(registry.Prometheus_0_0_4)
	assert.Equal(t, expectedMetricResult, metricRegistryResult)
}

func TestSkippedNilPointersTotal(t *testing.T) {
	expectedMetricResult := `missing_metrics_skipped_total{} 3`

	type InnerStruct struct {
		Field string
	}
	type TestStruct struct {
		NilString *string
		NilStruct *InnerStruct
		NilBool   *bool
	}

	nilObject := &TestStruct{}

	metricRegistry := registry.Registry{}
	prefixes := []string{"prefix"}
	labels := map[string]string{}
	MetricListFromStructs(nilObject, &metricRegistry, prefixes, labels, AbsentMetricsConfig{}, "single-label")

	metricRegistryResult := metricRegistry.FormatTextfileString(registry.Prometheus_0_0_4)
	assert.Equal(t, expectedMetricResult, metricRegistryResult)
}

func TestAllDataTypes(t *testing.T) {
	expectedMetricResult := `prefprefix_driver_info_info{DriverName="test_driver",FirmwareVersionParts="version_p1,version_p2",device="test_device"} 1
prefprefix_driver_info_supported_feature_whatever{device="test_device"} 1
metric_parse_error_total{} 1
prefprefix_device_data_device_index{device="test_device"} -1613.246008
prefprefix_device_data_device_index32{device="test_device"} 1613
prefprefix_device_data_device_uindex{device="test_device"} 1614
prefprefix_per_qstats_general_tx_bytes{queue="0"} 123
missing_metrics_skipped_total{} 2`
	txBytesValue := 123.0

	type DriverInfo struct {
		DriverName                   string
		FirmwareVersionParts         []string
		SupportedFeatureWhatever     bool
		SupportedFeatureWhateverType complex128
	}
	type DeviceData struct {
		DeviceIndex   float64
		DeviceIndex32 float32
		DeviceUIndex  uint64
	}
	driverInfo := DriverInfo{
		DriverName: "test_driver",
		FirmwareVersionParts: []string{
			"version_p1",
			"version_p2",
		},
		SupportedFeatureWhatever:     true,
		SupportedFeatureWhateverType: complex(10, 11),
	}
	deviceData := DeviceData{
		DeviceIndex:   -1613.246008,
		DeviceIndex32: 1613,
		DeviceUIndex:  1614,
	}

	perQStatsPerQueue := statistics.QueueStatisticsGeneral{
		TxBytes: &txBytesValue,
	}
	perQStats := statistics.PerQueueStatistics{
		statistics.QueueStatistics{
			General: &perQStatsPerQueue,
		},
	}

	type AbstractData struct {
		DriverInfo *DriverInfo
		DeviceData *DeviceData
		PerQStats  *statistics.PerQueueStatistics
	}
	abstractData := AbstractData{
		DriverInfo: &driverInfo,
		DeviceData: &deviceData,
		PerQStats:  &perQStats,
	}

	metricRegistry := registry.Registry{}
	prefixes := []string{"prefprefix"}
	labels := map[string]string{
		"device": "test_device",
	}
	MetricListFromStructs(abstractData, &metricRegistry, prefixes, labels, AbsentMetricsConfig{}, "single-label")

	metricResultString := metricRegistry.FormatTextfileString(registry.Prometheus_0_0_4)
	assert.Equal(t, expectedMetricResult, metricResultString)
}

func TestMetricListFromStructsMetricIndexError(t *testing.T) {
	type dummyStruct struct {
		Info string
	}
	metricList := registry.Registry{
		{Name: "dummy_info", Labels: map[string]string{"a": "1"}, Value: 1},
		{Name: "dummy_info", Labels: map[string]string{"b": "2"}, Value: 2},
	}
	input := dummyStruct{Info: "test"}

	MetricListFromStructs(input, &metricList, []string{"dummy"}, nil, AbsentMetricsConfig{}, "single-label")
}

func TestMetricListFromStructsListMultipleLabels(t *testing.T) {
	expectedResultMultilabel := `info{DriverName="test_driver",DriverNameWithSpace="test driver",FirmwareVersionPartsP0="version_p1",FirmwareVersionPartsP1="version_p2"} 1`
	expectedResultBoth := `info{DriverName="test_driver",DriverNameWithSpace="test driver",FirmwareVersionParts="version_p1,version_p2",FirmwareVersionPartsP0="version_p1",FirmwareVersionPartsP1="version_p2"} 1`

	type DriverInfo struct {
		DriverName           string
		DriverNameWithSpace  string
		FirmwareVersionParts []string
	}

	driverInfo := DriverInfo{
		DriverName:          "test_driver",
		DriverNameWithSpace: "test driver",
		FirmwareVersionParts: []string{
			"version_p1",
			"version_p2",
		},
	}

	metricRegistry := registry.Registry{}
	prefixes := []string{}
	labels := map[string]string{}

	MetricListFromStructs(driverInfo, &metricRegistry, prefixes, labels, AbsentMetricsConfig{}, "multi-label")
	metricResultString := metricRegistry.FormatTextfileString(registry.Prometheus_0_0_4)
	assert.Equal(t, expectedResultMultilabel, metricResultString)

	metricRegistry = registry.Registry{}
	MetricListFromStructs(driverInfo, &metricRegistry, prefixes, labels, AbsentMetricsConfig{}, "both")
	metricResultString = metricRegistry.FormatTextfileString(registry.Prometheus_0_0_4)
	assert.Equal(t, expectedResultBoth, metricResultString)
}

func TestMetricListFromStructsDefaultNamespace(t *testing.T) {
	oldNS := registry.OutputMetricNamespace
	registry.OutputMetricNamespace = "ethtool"
	defer func() { registry.OutputMetricNamespace = oldNS }()

	type Sample struct {
		RxBytes float64
	}

	input := Sample{RxBytes: 42}
	metricRegistry := registry.Registry{}
	MetricListFromStructs(input, &metricRegistry, []string{"statistics"}, map[string]string{"device": "eth0"}, AbsentMetricsConfig{}, "single-label")

	result := metricRegistry.FormatTextfileString(registry.Prometheus_0_0_4)
	assert.True(t, strings.Contains(result, "ethtool_statistics_rx_bytes{"), "metric name must have ethtool_ prefix")
	assert.False(t, strings.Contains(result, "\nnode_"), "metric names must not start with node_")
	assert.False(t, strings.HasPrefix(result, "node_"), "metric names must not start with node_")
}

// Just a snippet for fast testing with real metrics
// func TestRealIntelMetrics(t *testing.T) {
// 	interfaces := map[string]string{
// 		"eth0": "intel/i40e/00_sfp_10g_sr85",
// 	}
