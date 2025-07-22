package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"

	"github.com/newrushbolt/go-ethtool-exporter/interfaces"
	"github.com/newrushbolt/go-ethtool-exporter/metrics"
	"github.com/newrushbolt/go-ethtool-exporter/registry"

	"github.com/newrushbolt/go-ethtool-metrics/pkg/metrics/driver_info"
	"github.com/newrushbolt/go-ethtool-metrics/pkg/metrics/generic_info"
	"github.com/newrushbolt/go-ethtool-metrics/pkg/metrics/module_info"
	"github.com/newrushbolt/go-ethtool-metrics/pkg/metrics/statistics"
)

func initLogger() {
	var level slog.Level
	envLevel := os.Getenv("GO_ETHTOOL_EXPORTER_LOG_LEVEL")
	err := level.UnmarshalText([]byte(envLevel))
	if err != nil {
		level = slog.LevelInfo
	}
	slog.SetLogLoggerLevel(level)
}

func readEthtoolData(interfaceName string, ethtoolMode string, ethtoolPath string, ethtoolTimeout time.Duration) string {
	var ethtoolOutputRaw []byte
	var err error
	var cancel context.CancelFunc
	var ctx context.Context

	ctx, cancel = context.WithTimeout(context.Background(), ethtoolTimeout)
	defer cancel()

	if ethtoolMode == "" {
		ethtoolOutputRaw, err = exec.CommandContext(ctx, ethtoolPath, interfaceName).Output()
	} else {
		ethtoolOutputRaw, err = exec.CommandContext(ctx, ethtoolPath, ethtoolMode, interfaceName).Output()
	}

	if err != nil {
		slog.Info("Cannot run ethtool command", "ethtoolPath", ethtoolPath, "ethtoolMode", ethtoolMode, "error", err)
		return ""
	}
	ethtoolOutput := string(ethtoolOutputRaw)
	return ethtoolOutput
}

func parseAllowedInterfaceTypes(typesStr string) []int {
	types := []int{}
	for _, rawType := range strings.Split(typesStr, ",") {
		strType := strings.TrimSpace(rawType)
		if strType == "" {
			continue
		}
		intType, err := strconv.Atoi(strType)
		if err != nil {
			slog.Error("Invalid interface type in allowed-interface-types, must be comma separated", "allowed-interface-types", typesStr, "interface-type", strType, "error", err)
			continue
		}
		types = append(types, intType)
	}
	return types
}

func getExporterVersion(readBuildInfo func() (*debug.BuildInfo, bool)) string {
	buildInfo, ok := readBuildInfo()
	if !ok {
		return "go-ethtool-exporter version: unknown"
	}

	versionLines := []string{}
	mainVersion := "unknown"
	if buildInfo.Main.Version != "" {
		mainVersion = buildInfo.Main.Version
	}
	versionLines = append(versionLines, fmt.Sprintf("go-ethtool-exporter version: %s", mainVersion))

	var vcsRevision, vcsTime string
	for _, setting := range buildInfo.Settings {
		switch setting.Key {
		case "vcs.revision":
			vcsRevision = setting.Value
			if vcsRevision != "" {
				versionLines = append(versionLines, fmt.Sprintf("vcs.revision: %s", vcsRevision))
			}
		case "vcs.time":
			vcsTime = setting.Value
			if vcsTime != "" {
				versionLines = append(versionLines, fmt.Sprintf("vcs.time: %s", vcsTime))
			}
		}
	}

	return strings.Join(versionLines, "\n")
}

func collectAllMetrics() map[string]registry.Registry {
	allMetricRegistries := map[string]registry.Registry{}

	allowedTypes := parseAllowedInterfaceTypes(*discoverAllowedPortTypes)
	discoverConfig := interfaces.PortDiscoveryOptions{
		DiscoverAllPorts:   *discoverAllPorts,
		DiscoverBondSlaves: *discoverBondSlaves,
	}
	interfaces := interfaces.GetInterfacesList(*linuxNetClassPath, discoverConfig, allowedTypes)

	// TODO: allow parallel gather
	// TODO: split into different functions, probably move to other pkg
	for _, interfaceName := range interfaces {
		var metricRegistry registry.Registry
		interfaceLogger := slog.With("interfaceName", interfaceName)
		deviceLabels := map[string]string{
			"device": interfaceName,
		}

		// generic_info
		interfaceLogger.Debug("generic_info: collecting metrics")
		genericinfoConfig := generic_info.CollectConfig{
			CollectAdvertisedSettings: *collectGenericInfoModes,
			CollectSupportedSettings:  *collectGenericInfoModes,
			CollectSettings:           *collectGenericInfoSettings,
		}
		genericInfoDataRaw := readEthtoolData(interfaceName, "", *ethtoolPath, *ethtoolTimeout)
		interfaceLogger.Debug("generic_info: raw lines", "lines", strings.Count(genericInfoDataRaw, "\n"))
		genericInfoData := generic_info.ParseInfo(genericInfoDataRaw, &genericinfoConfig)
		before := len(metricRegistry)
		metrics.MetricListFromStructs(genericInfoData, &metricRegistry, []string{"generic_info"}, deviceLabels, *keepAbsentMetrics)
		interfaceLogger.Debug("generic_info: final metrics", "count", len(metricRegistry)-before)

		// driver_info
		interfaceLogger.Debug("driver_info: collecting metrics")
		driverInfoConfig := driver_info.CollectConfig{
			DriverFeatures: *collectDriverInfoFeatures,
		}
		driverInfoDataRaw := readEthtoolData(interfaceName, "-i", *ethtoolPath, *ethtoolTimeout)
		interfaceLogger.Debug("driver_info: raw lines", "lines", strings.Count(driverInfoDataRaw, "\n"))
		driverInfoData := driver_info.ParseInfo(driverInfoDataRaw, &driverInfoConfig)
		before = len(metricRegistry)
		metrics.MetricListFromStructs(driverInfoData, &metricRegistry, []string{"driver_info"}, deviceLabels, *keepAbsentMetrics)
		interfaceLogger.Debug("driver_info: final metrics", "count", len(metricRegistry)-before)

		// module_info
		interfaceLogger.Debug("module_info: collecting metrics")
		moduleInfoConfig := module_info.CollectConfig{
			CollectDiagnosticsAlarms:   *collectModuleInfoDiagnosticsAlarms,
			CollectDiagnosticsWarnings: *collectModuleInfoDiagnosticsWarnings,
			CollectDiagnosticsValues:   *collectModuleInfoDiagnosticsValues,
			CollectVendor:              *collectModuleInfoVendor,
		}
		moduleInfoDataRaw := readEthtoolData(interfaceName, "-m", *ethtoolPath, *ethtoolTimeout)
		interfaceLogger.Debug("module_info: raw lines", "lines", strings.Count(moduleInfoDataRaw, "\n"))
		moduleInfoData := module_info.ParseInfo(moduleInfoDataRaw, &moduleInfoConfig)
		before = len(metricRegistry)
		metrics.MetricListFromStructs(moduleInfoData, &metricRegistry, []string{"module_info"}, deviceLabels, *keepAbsentMetrics)
		interfaceLogger.Debug("module_info: final metrics", "count", len(metricRegistry)-before)

		// statistics
		interfaceLogger.Debug("statistics: collecting metrics")
		statisticsConfig := statistics.CollectConfig{}.Default()
		statisticsDataRaw := readEthtoolData(interfaceName, "-S", *ethtoolPath, *ethtoolTimeout)
		interfaceLogger.Debug("statistics: raw lines", "lines", strings.Count(statisticsDataRaw, "\n"))
		statisticsData := statistics.ParseInfo(statisticsDataRaw, statisticsConfig)
		before = len(metricRegistry)
		metrics.MetricListFromStructs(statisticsData, &metricRegistry, []string{"statistics"}, deviceLabels, *keepAbsentMetrics)
		interfaceLogger.Debug("statistics: final metrics", "count", len(metricRegistry)-before)

		interfaceLogger.Debug("Total metric count", "metricCount", len(metricRegistry))
		allMetricRegistries[interfaceName] = metricRegistry
	}

	return allMetricRegistries
}

func writeAllMetricsToTextfiles(metricRegistries map[string]registry.Registry) {
	allMetrics := make([]string, len(metricRegistries))
	textFileName := "ethtool_exporter.prom"
	textFilePath := path.Join(*textfileDirectory, textFileName)
	for _, metricRegistry := range metricRegistries {
		// Writing file in node_exporter textfile format
		metrics := metricRegistry.FormatTextfileString()
		allMetrics = append(allMetrics, metrics)
	}
	allMetricsString := strings.Join(allMetrics, "\n")
	registry.MustWriteTextfile(textFilePath, allMetricsString)
}

func MustDirectoryExist(dirPath *string) {
	panicMessage := fmt.Sprintf("Directory <%s> does not exist", *textfileDirectory)
	info, err := os.Stat(*dirPath)
	if err != nil {
		panic(panicMessage)
	}
	if !info.IsDir() {
		panic(panicMessage)
	}
}

func init() {
	// Moved to separate `init()` in order to work both in exporter and tests
	initLogger()
}

func runDiscoverPortsCommand() {
	// Discover ports mode
	allowedTypes := parseAllowedInterfaceTypes(*discoverAllowedPortTypes)
	discoverConfig := interfaces.PortDiscoveryOptions{
		DiscoverAllPorts:   *discoverAllPorts,
		DiscoverBondSlaves: *discoverBondSlaves,
	}
	interfaces := interfaces.GetInterfacesList(*linuxNetClassPath, discoverConfig, allowedTypes)
	if len(interfaces) == 0 {
		fmt.Println("No ports discovered, re-run with `GO_ETHTOOL_EXPORTER_LOG_LEVEL=DEBUG` to check the discovery logic")
	} else {
		fmt.Println("Discovered following ports:")
		interfacesString := fmt.Sprintf("  - %s", strings.Join(interfaces, "\n  - "))
		fmt.Println(interfacesString)
	}
}

func runSingleTextfileCommand() {
	// Single textfile mode
	MustDirectoryExist(textfileDirectory)
	metricRegistries := collectAllMetrics()
	writeAllMetricsToTextfiles(metricRegistries)
}

func runLoopTextfileCommand() {
	// Loop textfile mode
	MustDirectoryExist(textfileDirectory)
	for {
		metricRegistries := collectAllMetrics()
		writeAllMetricsToTextfiles(metricRegistries)
		time.Sleep(*loopTextfileUpdateInterval)
	}
}

func main() {
	slog.Info("Starting go-ethtool-exporter")

	for _, arg := range os.Args[1:] {
		if arg == "--help" || arg == "-h" {
			fmt.Print(helpText)
			os.Exit(0)
		}
	}

	kingpin.Version(getExporterVersion(debug.ReadBuildInfo))
	exporterCommand := kingpin.Parse()

	switch exporterCommand {
	case discoverPortsCommand.FullCommand():
		runDiscoverPortsCommand()
	case singleTextfileCommand.FullCommand():
		runSingleTextfileCommand()
	case loopTextfileCommand.FullCommand():
		runLoopTextfileCommand()
	default:
		panicMessage := fmt.Sprintf("Unknown command: %s", exporterCommand)
		panic(panicMessage)
	}
}
