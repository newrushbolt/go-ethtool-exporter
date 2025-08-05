package main

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"

	"github.com/newrushbolt/go-ethtool-exporter/collector"
	"github.com/newrushbolt/go-ethtool-exporter/interfaces"
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

// TODO: to be covered by some kind of tests
func collectMetrics() map[string]registry.Registry {
	allMetricRegistries := map[string]registry.Registry{}

	allowedTypes := parseAllowedInterfaceTypes(*discoverAllowedPortTypes)
	discoverConfig := interfaces.PortDiscoveryOptions{
		DiscoverAllPorts:   *discoverAllPorts,
		DiscoverBondSlaves: *discoverBondSlaves,
	}
	interfaces := interfaces.GetInterfacesList(*linuxNetClassPath, discoverConfig, allowedTypes)

	// Format configs
	genericinfoConfig := generic_info.CollectConfig{
		CollectAdvertisedSettings: *collectGenericInfoModes,
		CollectSupportedSettings:  *collectGenericInfoModes,
		CollectSettings:           *collectGenericInfoSettings,
	}
	driverInfoConfig := driver_info.CollectConfig{
		CollectCommon:   *collectDriverInfoCommon,
		CollectFeatures: *collectDriverInfoFeatures,
	}
	moduleInfoConfig := module_info.CollectConfig{
		CollectDiagnosticsAlarms:   *collectModuleInfoDiagnosticsAlarms,
		CollectDiagnosticsWarnings: *collectModuleInfoDiagnosticsWarnings,
		CollectDiagnosticsValues:   *collectModuleInfoDiagnosticsValues,
		CollectVendor:              *collectModuleInfoVendor,
	}
	statisticsConfig := *(statistics.CollectConfig{}.Default())

	collectorConfig := collector.CollectorConfig{
		GenericInfo:       genericinfoConfig,
		DriverInfo:        driverInfoConfig,
		ModuleInfo:        moduleInfoConfig,
		Statistics:        statisticsConfig,
		EthtoolPath:       *ethtoolPath,
		EthtoolTimeout:    *ethtoolTimeout,
		KeepAbsentMetrics: *keepAbsentMetrics,
		ListLabelFormat:   *listLabelFormat,
	}

	for _, interfaceName := range interfaces {
		// TODO: allow parallel gather
		interfaceRegistry := collector.CollectInterfaceMetrics(interfaceName, collectorConfig)
		allMetricRegistries[interfaceName] = interfaceRegistry
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
		PortsRegexp:          *discoverPortsRegexp,
		DiscoverAllPorts:     *discoverAllPorts,
		DiscoverBondSlaves:   *discoverBondSlaves,
		DiscoverBridgeSlaves: *discoverBridgeSlaves,
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
	metricRegistries := collectMetrics()
	writeAllMetricsToTextfiles(metricRegistries)
}

func runLoopTextfileCommand() {
	// Loop textfile mode
	MustDirectoryExist(textfileDirectory)
	for {
		metricRegistries := collectMetrics()
		writeAllMetricsToTextfiles(metricRegistries)
		time.Sleep(*loopTextfileUpdateInterval)
	}
}

func main() {
	// Covering main() is really hard. Moving logic toward separate function(s) is a better solution
	slog.Info("Starting go-ethtool-exporter")

	for _, arg := range os.Args[1:] {
		if arg == "--help" || arg == "-h" {
			fmt.Print(helpText)
			os.Exit(0)
		}
	}

	kingpin.Version(getExporterVersion(debug.ReadBuildInfo))
	exporterCommand := kingpin.Parse()

	if *collectAllMetrics {
		slog.Warn("Flag --collect-all-metrics is set, ignoring all other --collect-* flags")
		// TODO: find better solution because manually adding flags to this block is not fun
		*collectGenericInfoSettings = true
		*collectDriverInfoCommon = true
		*collectModuleInfoDiagnosticsAlarms = true
		*collectModuleInfoDiagnosticsWarnings = true
		*collectDriverInfoFeatures = true
		*collectGenericInfoModes = true
		*collectModuleInfoDiagnosticsValues = true
		*collectModuleInfoVendor = true
	}

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
