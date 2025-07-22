package collector

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/newrushbolt/go-ethtool-metrics/pkg/metrics/driver_info"
	"github.com/newrushbolt/go-ethtool-metrics/pkg/metrics/generic_info"
	"github.com/newrushbolt/go-ethtool-metrics/pkg/metrics/module_info"
	"github.com/newrushbolt/go-ethtool-metrics/pkg/metrics/statistics"
)

func TestReadEthtoolData(t *testing.T) {
	stubPath := "../testdata/ethtool.sh"
	// Make sure the stub is executable
	os.Chmod(stubPath, 0755)

	var out string
	timeout, err := time.ParseDuration("1s")
	assert.NoError(t, err)

	// No mode
	out = readEthtoolData("eth0", "", stubPath, timeout)
	assert.Equal(t, "ethtool output for eth0\n", out)

	// -i mode
	out = readEthtoolData("eth0", "-i", stubPath, timeout)
	assert.Equal(t, "driver info for eth0\n", out)

	// -m mode
	out = readEthtoolData("eth0", "-m", stubPath, timeout)
	assert.Equal(t, "module info for eth0\n", out)

	// -S mode
	out = readEthtoolData("eth0", "-S", stubPath, timeout)
	assert.Equal(t, "statistics for eth0\n", out)

	// Unknown interface
	out = readEthtoolData("unknown", "", stubPath, timeout)
	assert.Equal(t, "", out)

	// Timeout
	tinyTimeout, err := time.ParseDuration("10ms")
	assert.NoError(t, err)

	out = readEthtoolData("eth0", "-S", stubPath, tinyTimeout)
	assert.Equal(t, "", out)
}

func TestEmptyCollectInterfaceMetrics(t *testing.T) {

	genericinfoConfig := generic_info.CollectConfig{}.Default()
	driverInfoConfig := driver_info.CollectConfig{}.Default()
	moduleInfoConfig := module_info.CollectConfig{}.Default()
	statisticsConfig := statistics.CollectConfig{}.Default()

	ethtoolPath := "../testdata/non_existed_ethtool.sh"
	ethtoolTimeout := 1 * time.Second
	keepAbsentMetrics := false

	collectorConfig := CollectorConfig{
		GenericInfo:       *genericinfoConfig,
		DriverInfo:        *driverInfoConfig,
		ModuleInfo:        *moduleInfoConfig,
		Statistics:        *statisticsConfig,
		EthtoolPath:       ethtoolPath,
		EthtoolTimeout:    ethtoolTimeout,
		KeepAbsentMetrics: keepAbsentMetrics,
	}

	registry := CollectInterfaceMetrics("eth0", collectorConfig)

	assert.Len(t, registry, 0)
}

func TestGenericIntelCollectInterfaceMetrics(t *testing.T) {
	// driver_info_info should not be presented, it's a bug in go-ethtool-metrics
	expectedMetricResult := `generic_info_supported_settings_info{FecModes="Not reported",LinkModes="10000baseSR/Full",PauseFrameUse="Symmetric",device="eth1"} 1
generic_info_advertised_settings_info{FecModes="Not reported",LinkModes="10000baseSR/Full",PauseFrameUse="No",device="eth1"} 1
generic_info_settings_info{Duplex="Full",Port="FIBRE",Speed="10000Mb/s",Transceiver="internal",device="eth1"} 1
generic_info_settings_speed_bytes{device="eth1"} 1e+10
generic_info_settings_auto_negotiation{device="eth1"} 0
generic_info_settings_link_detected{device="eth1"} 1
driver_info_info{BusAddress="",DriverName="",DriverVersion="",FirmwareVersion="",FirmwareVersionParts="",device="eth1"} 1`

	genericinfoConfig := generic_info.CollectConfig{
		CollectAdvertisedSettings: true,
		CollectSupportedSettings:  true,
		CollectSettings:           true,
	}
	driverInfoConfig := driver_info.CollectConfig{
		DriverFeatures: false,
	}
	moduleInfoConfig := module_info.CollectConfig{
		CollectDiagnosticsAlarms:   false,
		CollectDiagnosticsWarnings: false,
		CollectDiagnosticsValues:   false,
		CollectVendor:              false,
	}
	statisticsConfig := statistics.CollectConfig{
		General:  false,
		PerQueue: false,
	}
	ethtoolPath := "../testdata/ethtool.sh"
	ethtoolTimeout := 1 * time.Second
	keepAbsentMetrics := false

	collectorConfig := CollectorConfig{
		GenericInfo:       genericinfoConfig,
		DriverInfo:        driverInfoConfig,
		ModuleInfo:        moduleInfoConfig,
		Statistics:        statisticsConfig,
		EthtoolPath:       ethtoolPath,
		EthtoolTimeout:    ethtoolTimeout,
		KeepAbsentMetrics: keepAbsentMetrics,
	}

	registry := CollectInterfaceMetrics("eth1", collectorConfig)

	assert.Equal(t, expectedMetricResult, registry.FormatTextfileString())
}
