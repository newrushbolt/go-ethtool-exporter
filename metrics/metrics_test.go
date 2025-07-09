package metrics

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/newrushbolt/go-ethtool-exporter/registry"
	"github.com/stretchr/testify/assert"

	"github.com/newrushbolt/go-ethtool-metrics/pkg/metrics/statistics"
)

func TestAllDataTypes(t *testing.T) {
	expectedMetricResult := `prefprefix_driver_info_info{DriverName="test_driver",FirmwareVersionParts="version_p1,version_p2",device="test_device"} 1
prefprefix_driver_info_supported_feature_whatever{device="test_device"} 1
prefprefix_device_data_device_index{device="test_device"} -1613.246008
prefprefix_device_data_device_index32{device="test_device"} 1613
prefprefix_device_data_device_uindex{device="test_device"} 1614
prefprefix_per_qstats_tx_bytes{queue="0"} 123
`
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

	perQStats := statistics.PerQueueStatistics{
		statistics.QueueStatistics{
			TxBytes: &txBytesValue,
			RxBytes: nil,
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
	MetricListFromStructs(abstractData, &metricRegistry, prefixes, labels)

	textFilePath := fmt.Sprintf("../.TestAbstractData-%d.prom", time.Now().Unix())
	metricRegistry.MustWriteTextfile(textFilePath)
	defer os.Remove(textFilePath)
	metricResult, err := os.ReadFile(textFilePath)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, expectedMetricResult, string(metricResult))
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

	MetricListFromStructs(input, &metricList, []string{"dummy"}, nil)
}

// func TestRealIntelMetrics(t *testing.T) {

// 	interfaces := map[string]string{
// 		"eth0": "intel/i40e/00_sfp_10g_sr85",
// 	}

// 	for _, interfacePath := range interfaces {
// 		t.Run(interfacePath, func(t *testing.T) {
// 			metricRegistry := registry{}

// 			// statistics
// 			metricType := "statistics"
// 			fixtureSourcePath := fmt.Sprintf("../../go-ethtool-metrics/testdata/%s/src/%s", interfacePath, metricType)
// 			fixtureSourceData, err := os.ReadFile(fixtureSourcePath)
// 			if err != nil {
// 				panicErr := fmt.Errorf("Cannot read fixture file: %v", err)
// 				panic(panicErr)
// 			}
// 			statisticsDataRaw := string(fixtureSourceData)
// 			statisticsConfig := statistics.CollectConfig{}.Default()
// 			statisticsData := statistics.ParseInfo(statisticsDataRaw, statisticsConfig)
// 			MetricListFromStructs(statisticsData, &metricRegistry, []string{"statistics"}, map[string]string{})

// 			// driver_info
// 			driverInfoConfig := driver_info.CollectConfig{
// 				DriverFeatures: true,
// 			}
// 			metricType = "driver_info"
// 			fixtureSourcePath = fmt.Sprintf("../../go-ethtool-metrics/testdata/%s/src/%s", interfacePath, metricType)
// 			fixtureSourceData, err = os.ReadFile(fixtureSourcePath)
// 			if err != nil {
// 				panicErr := fmt.Errorf("Cannot read fixture file: %v", err)
// 				panic(panicErr)
// 			}
// 			driverInfoDataRaw := string(fixtureSourceData)
// 			driverInfoData := driver_info.ParseInfo(driverInfoDataRaw, &driverInfoConfig)
// 			MetricListFromStructs(driverInfoData, &metricRegistry, []string{"driver_info"}, map[string]string{})

// 			fmt.Printf("Total metrics: %+v", metricRegistry)
// 		})
// 	}
// }
