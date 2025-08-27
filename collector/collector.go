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
	GenericInfo                  generic_info.CollectConfig
	DriverInfo                   driver_info.CollectConfig
	ModuleInfo                   module_info.CollectConfig
	Statistics                   statistics.CollectConfig
	EthtoolPath                  string
	EthtoolTimeout               time.Duration
	KeepAbsentMetricsModuleInfo  bool
	KeepAbsentMetricsGenericInfo bool
	KeepAbsentMetricsDriverInfo  bool
	KeepAbsentMetricsStatistics  bool
	ListLabelFormat              string
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

type metricCollector struct {
	Name              string
	Enabled           bool
	EthtoolMode       string
	KeepAbsentMetrics bool
	ParseFunc         func(string) any
}

func CollectInterfaceMetrics(interfaceName string, config CollectorConfig) registry.Registry {
	collectors := []metricCollector{
		{
			Name:              "generic_info",
			EthtoolMode:       "",
			Enabled:           config.GenericInfo.CollectAdvertisedSettings || config.GenericInfo.CollectSupportedSettings || config.GenericInfo.CollectSettings,
			ParseFunc:         func(raw string) any { return generic_info.ParseInfo(raw, &config.GenericInfo) },
			KeepAbsentMetrics: config.KeepAbsentMetricsGenericInfo,
		},
		{
			Name:              "driver_info",
			EthtoolMode:       "-i",
			Enabled:           config.DriverInfo.CollectCommon || config.DriverInfo.CollectFeatures,
			ParseFunc:         func(raw string) any { return driver_info.ParseInfo(raw, &config.DriverInfo) },
			KeepAbsentMetrics: config.KeepAbsentMetricsDriverInfo,
		},
		{
			Name:              "module_info",
			EthtoolMode:       "-m",
			Enabled:           config.ModuleInfo.CollectDiagnosticsAlarms || config.ModuleInfo.CollectDiagnosticsValues || config.ModuleInfo.CollectDiagnosticsWarnings || config.ModuleInfo.CollectVendor,
			ParseFunc:         func(raw string) any { return module_info.ParseInfo(raw, &config.ModuleInfo) },
			KeepAbsentMetrics: config.KeepAbsentMetricsModuleInfo,
		},
		{
			Name:              "statistics",
			EthtoolMode:       "-S",
			Enabled:           config.Statistics.General || config.Statistics.PerQueueGeneral || config.Statistics.PerQueuePerType,
			ParseFunc:         func(raw string) any { return statistics.ParseInfo(raw, &config.Statistics) },
			KeepAbsentMetrics: config.KeepAbsentMetricsStatistics,
		},
	}

	var metricRegistry registry.Registry
	interfaceLogger := slog.With("interfaceName", interfaceName)
	deviceLabels := map[string]string{
		"device": interfaceName,
	}

	for _, collector := range collectors {
		collectorLogger := interfaceLogger.With("collector", collector.Name)
		if !collector.Enabled {
			collectorLogger.Debug("Metrics are disabled, skipping")
			continue
		}
		collectorLogger.Debug("Collecting metrics")
		dataRaw := readEthtoolData(interfaceName, collector.EthtoolMode, config.EthtoolPath, config.EthtoolTimeout)
		collectorLogger.Debug("Got raw lines", "count", strings.Count(dataRaw, "\n"))
		data := collector.ParseFunc(dataRaw)
		before := len(metricRegistry)
		metrics.MetricListFromStructs(data, &metricRegistry, []string{collector.Name}, deviceLabels, collector.KeepAbsentMetrics, config.ListLabelFormat)
		collectorLogger.Debug("Final metrics", "count", len(metricRegistry)-before)
	}

	interfaceLogger.Debug("Total metric count", "metricCount", len(metricRegistry))
	return metricRegistry
}
