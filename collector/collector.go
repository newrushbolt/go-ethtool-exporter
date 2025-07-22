package collector

import (
	"context"
	"log/slog"
	"os/exec"
	"strings"
	"time"

	"github.com/newrushbolt/go-ethtool-exporter/metrics"
	"github.com/newrushbolt/go-ethtool-exporter/registry"

	"github.com/newrushbolt/go-ethtool-metrics/pkg/metrics/driver_info"
	"github.com/newrushbolt/go-ethtool-metrics/pkg/metrics/generic_info"
	"github.com/newrushbolt/go-ethtool-metrics/pkg/metrics/module_info"
	"github.com/newrushbolt/go-ethtool-metrics/pkg/metrics/statistics"
)

type CollectorConfig struct {
	GenericInfo       generic_info.CollectConfig
	DriverInfo        driver_info.CollectConfig
	ModuleInfo        module_info.CollectConfig
	Statistics        statistics.CollectConfig
	EthtoolPath       string
	EthtoolTimeout    time.Duration
	KeepAbsentMetrics bool
}

func readEthtoolData(interfaceName string, ethtoolMode string, ethtoolPath string, ethtoolTimeout time.Duration) string {
	var ethtoolOutputRaw []byte
	var err error
	var cancel context.CancelFunc
	var ctx context.Context

	ctx, cancel = context.WithTimeout(context.Background(), ethtoolTimeout)
	defer cancel()

	ethtoolArgs := []string{}
	if ethtoolMode != "" {
		ethtoolArgs = append(ethtoolArgs, ethtoolMode)
	}
	ethtoolArgs = append(ethtoolArgs, interfaceName)

	ethtoolOutputRaw, err = exec.CommandContext(ctx, ethtoolPath, ethtoolArgs...).Output()
	if err != nil {
		slog.Info("Cannot run ethtool command", "ethtoolPath", ethtoolPath, "ethtoolMode", ethtoolMode, "error", err)
		return ""
	}
	ethtoolOutput := string(ethtoolOutputRaw)
	return ethtoolOutput
}

func CollectInterfaceMetrics(interfaceName string, config CollectorConfig) registry.Registry {
	var metricRegistry registry.Registry
	interfaceLogger := slog.With("interfaceName", interfaceName)
	deviceLabels := map[string]string{
		"device": interfaceName,
	}

	// generic_info
	interfaceLogger.Debug("generic_info: collecting metrics")
	genericInfoDataRaw := readEthtoolData(interfaceName, "", config.EthtoolPath, config.EthtoolTimeout)
	interfaceLogger.Debug("generic_info: raw lines", "lines", strings.Count(genericInfoDataRaw, "\n"))
	genericInfoData := generic_info.ParseInfo(genericInfoDataRaw, &config.GenericInfo)
	before := len(metricRegistry)
	metrics.MetricListFromStructs(genericInfoData, &metricRegistry, []string{"generic_info"}, deviceLabels, config.KeepAbsentMetrics)
	interfaceLogger.Debug("generic_info: final metrics", "count", len(metricRegistry)-before)

	// driver_info
	interfaceLogger.Debug("driver_info: collecting metrics")
	driverInfoDataRaw := readEthtoolData(interfaceName, "-i", config.EthtoolPath, config.EthtoolTimeout)
	interfaceLogger.Debug("driver_info: raw lines", "lines", strings.Count(driverInfoDataRaw, "\n"))
	driverInfoData := driver_info.ParseInfo(driverInfoDataRaw, &config.DriverInfo)
	before = len(metricRegistry)
	metrics.MetricListFromStructs(driverInfoData, &metricRegistry, []string{"driver_info"}, deviceLabels, config.KeepAbsentMetrics)
	interfaceLogger.Debug("driver_info: final metrics", "count", len(metricRegistry)-before)

	// module_info
	interfaceLogger.Debug("module_info: collecting metrics")
	moduleInfoDataRaw := readEthtoolData(interfaceName, "-m", config.EthtoolPath, config.EthtoolTimeout)
	interfaceLogger.Debug("module_info: raw lines", "lines", strings.Count(moduleInfoDataRaw, "\n"))
	moduleInfoData := module_info.ParseInfo(moduleInfoDataRaw, &config.ModuleInfo)
	before = len(metricRegistry)
	metrics.MetricListFromStructs(moduleInfoData, &metricRegistry, []string{"module_info"}, deviceLabels, config.KeepAbsentMetrics)
	interfaceLogger.Debug("module_info: final metrics", "count", len(metricRegistry)-before)

	// statistics
	interfaceLogger.Debug("statistics: collecting metrics")
	statisticsDataRaw := readEthtoolData(interfaceName, "-S", config.EthtoolPath, config.EthtoolTimeout)
	interfaceLogger.Debug("statistics: raw lines", "lines", strings.Count(statisticsDataRaw, "\n"))
	statisticsData := statistics.ParseInfo(statisticsDataRaw, &config.Statistics)
	before = len(metricRegistry)
	metrics.MetricListFromStructs(statisticsData, &metricRegistry, []string{"statistics"}, deviceLabels, config.KeepAbsentMetrics)
	interfaceLogger.Debug("statistics: final metrics", "count", len(metricRegistry)-before)

	interfaceLogger.Debug("Total metric count", "metricCount", len(metricRegistry))
	return metricRegistry
}
