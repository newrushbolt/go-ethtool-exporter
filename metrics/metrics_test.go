package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newrushbolt/go-ethtool-exporter/registry"

	"github.com/newrushbolt/go-ethtool-metrics/pkg/metrics/statistics"
)

func TestDropAllNils(t *testing.T) {
	expectedMetricResult := `prefix_real_float64{} 16.13`

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
	MetricListFromStructs(nilObject, &metricRegistry, prefixes, labels, false, "single-label")

	metricRegistryResult := metricRegistry.FormatTextfileString()
	assert.Equal(t, expectedMetricResult, metricRegistryResult)
}

func TestKeepFloat64Nils(t *testing.T) {
	expectedMetricResult := `prefix_real_float64{} 16.13
prefix_nil_float64{} NaN`
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
	MetricListFromStructs(nilObject, &metricRegistry, prefixes, labels, true, "single-label")

	metricRegistryResult := metricRegistry.FormatTextfileString()
	assert.Equal(t, expectedMetricResult, metricRegistryResult)
}

func TestAllDataTypes(t *testing.T) {
	expectedMetricResult := `prefprefix_driver_info_info{DriverName="test_driver",FirmwareVersionParts="version_p1,version_p2",device="test_device"} 1
prefprefix_driver_info_supported_feature_whatever{device="test_device"} 1
prefprefix_device_data_device_index{device="test_device"} -1613.246008
prefprefix_device_data_device_index32{device="test_device"} 1613
prefprefix_device_data_device_uindex{device="test_device"} 1614
prefprefix_per_qstats_general_tx_bytes{queue="0"} 123`
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
	MetricListFromStructs(abstractData, &metricRegistry, prefixes, labels, false, "single-label")

	metricResultString := metricRegistry.FormatTextfileString()
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

	MetricListFromStructs(input, &metricList, []string{"dummy"}, nil, false, "single-label")
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

	MetricListFromStructs(driverInfo, &metricRegistry, prefixes, labels, true, "multi-label")
	metricResultString := metricRegistry.FormatTextfileString()
	assert.Equal(t, expectedResultMultilabel, metricResultString)

	metricRegistry = registry.Registry{}
	MetricListFromStructs(driverInfo, &metricRegistry, prefixes, labels, true, "both")
	metricResultString = metricRegistry.FormatTextfileString()
	assert.Equal(t, expectedResultBoth, metricResultString)
}

// Just a snippet for fast testing with real metrics
// func TestRealIntelMetrics(t *testing.T) {
// 	interfaces := map[string]string{
// 		"eth0": "intel/i40e/00_sfp_10g_sr85",
// 	}
