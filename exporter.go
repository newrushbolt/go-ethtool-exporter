package main

import (
	"fmt"
	"log/slog"
	"os/exec"
	"path"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/alecthomas/kingpin/v2"

	"github.com/newrushbolt/go-ethtool-exporter/interfaces"
	"github.com/newrushbolt/go-ethtool-exporter/metrics"
	"github.com/newrushbolt/go-ethtool-exporter/registry"

	"github.com/newrushbolt/go-ethtool-metrics/pkg/metrics/driver_info"
	"github.com/newrushbolt/go-ethtool-metrics/pkg/metrics/generic_info"
	"github.com/newrushbolt/go-ethtool-metrics/pkg/metrics/module_info"
	"github.com/newrushbolt/go-ethtool-metrics/pkg/metrics/statistics"
)

var (
	// TODO: add env support
	ethtoolPath       = kingpin.Flag("ethtool-path", "").Default("/usr/sbin/ethtool").ExistingFile()
	linuxNetClassPath = kingpin.Flag("linux-net-class-path", "").Default("/sys/class/net").ExistingDir()
	textfileDirectory = kingpin.Flag("textfile-directory", "Path to node_exporter textfile directory").Default("var/lib/node-exporter/textfiles").ExistingDir()

	// Collectors, enabled by default
	collectGenericInfoSettings           = kingpin.Flag("collect-generic-info-settings", "").Default("true").Bool()
	collectModuleInfoDiagnosticsAlarms   = kingpin.Flag("collect-module-info-diagnostics-alarms", "").Default("true").Bool()
	collectModuleInfoDiagnosticsWarnings = kingpin.Flag("collect-module-info-diagnostics-warnings", "").Default("true").Bool()

	// Collectors, disabled by default
	collectDriverInfoFeatures          = kingpin.Flag("collect-driver-info-features", "").Default("false").Bool()
	collectGenericInfoModes            = kingpin.Flag("collect-generic-info-modes", "").Default("false").Bool()
	collectModuleInfoDiagnosticsValues = kingpin.Flag("collect-module-info-diagnostics-values", "").Default("false").Bool()
	collectModuleInfoVendor            = kingpin.Flag("collect-module-info-vendor", "").Default("false").Bool()

	// Port detection settings
	skipNonBondedPorts = kingpin.Flag("skip-non-bonded-ports", " ").Default("true").Bool()
	// Not yet implemented
	// skipBondMasterPorts = kingpin.Flag("skip-bond-master-ports", "").Bool()
	// skipOvsSlavePorts   = kingpin.Flag("skip-ovs-slave-ports", "").Bool()
	// detectPortsBlackList  = kingpin.Flag("detect-ports-black-list", "").Default(".*").Regexp()
	// Detect aliases and naming types?

	// Absent metrics settings
	// keepAbsentMetrics = kingpin.Flag("keep-absent-metrics", "").Default("false").Bool()

	// All possible types: https://github.com/torvalds/linux/blob/772b78c2abd85586bb90b23adff89f7303c704c7/include/uapi/linux/if_arp.h#L29
	allowedInterfaceTypesStr = kingpin.Flag("allowed-interface-types", "Comma-separated list of allowed interface types (see if_arp.h)").Default("1,").String()
)

func readEthtoolData(interfaceName string, ethtoolMode string, ethtoolPath string) string {
	var ethtoolOutputRaw []byte
	var err error
	if ethtoolMode == "" {
		ethtoolOutputRaw, err = exec.Command(ethtoolPath, interfaceName).Output()
	} else {
		ethtoolOutputRaw, err = exec.Command(ethtoolPath, ethtoolMode, interfaceName).Output()
	}

	if err != nil {
		slog.Debug("Cannot run ethtool command", "ethtoolPath", ethtoolPath, "ethtoolMode", ethtoolMode, "interfaceName", interfaceName, "error", err)
		return ""
	}
	ethtoolOutput := string(ethtoolOutputRaw)
	return ethtoolOutput
}

func parseAllowedInterfaceTypes(typesStr string) []int {
	types := []int{}
	for _, t := range strings.Split(typesStr, ",") {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		v, err := strconv.Atoi(t)
		if err != nil {
			slog.Error("Invalid interface type in allowed-interface-types", "allowed-interface-types", typesStr, "interface-type", t, "error", err)
			continue
		}
		types = append(types, v)
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

	allowedTypes := parseAllowedInterfaceTypes(*allowedInterfaceTypesStr)
	interfaces := interfaces.GetInterfacesList(*linuxNetClassPath, *skipNonBondedPorts, allowedTypes)
	for _, interfaceName := range interfaces {
		var metricRegistry registry.Registry

		// generic_info
		genericinfoConfig := generic_info.CollectConfig{
			CollectAdvertisedSettings: *collectGenericInfoModes,
			CollectSupportedSettings:  *collectGenericInfoModes,
			CollectSettings:           *collectGenericInfoSettings,
		}
		genericInfoDataRaw := readEthtoolData(interfaceName, "", *ethtoolPath)
		genericInfoData := generic_info.ParseInfo(genericInfoDataRaw, &genericinfoConfig)
		metrics.MetricListFromStructs(genericInfoData, &metricRegistry, []string{"generic_info"}, map[string]string{})

		// driver_info
		driverInfoConfig := driver_info.CollectConfig{
			DriverFeatures: *collectDriverInfoFeatures,
		}
		driverInfoDataRaw := readEthtoolData(interfaceName, "-i", *ethtoolPath)
		driverInfoData := driver_info.ParseInfo(driverInfoDataRaw, &driverInfoConfig)
		metrics.MetricListFromStructs(driverInfoData, &metricRegistry, []string{"driver_info"}, map[string]string{})

		// module_info
		moduleInfoConfig := module_info.CollectConfig{
			CollectDiagnosticsAlarms:   *collectModuleInfoDiagnosticsAlarms,
			CollectDiagnosticsWarnings: *collectModuleInfoDiagnosticsWarnings,
			CollectDiagnosticsValues:   *collectModuleInfoDiagnosticsValues,
			CollectVendor:              *collectModuleInfoVendor,
		}
		moduleInfoDataRaw := readEthtoolData(interfaceName, "-m", *ethtoolPath)
		moduleInfoData := module_info.ParseInfo(moduleInfoDataRaw, &moduleInfoConfig)
		metrics.MetricListFromStructs(moduleInfoData, &metricRegistry, []string{"module_info"}, map[string]string{})

		// statistics
		statisticsConfig := statistics.CollectConfig{}.Default()
		statisticsDataRaw := readEthtoolData(interfaceName, "-S", *ethtoolPath)
		statisticsData := statistics.ParseInfo(statisticsDataRaw, statisticsConfig)
		metrics.MetricListFromStructs(statisticsData, &metricRegistry, []string{"statistics"}, map[string]string{})

		allMetricRegistries[interfaceName] = metricRegistry
	}

	return allMetricRegistries
}

func writeAllMetricsToTextfiles() {
	metricRegistries := collectAllMetrics()

	for interfaceName, metricRegistry := range metricRegistries {
		// Writing file in node_exporter textfile format
		textFileName := fmt.Sprintf("%s.prom", interfaceName)
		textFilePath := path.Join(*textfileDirectory, textFileName)
		metricRegistry.MustWriteTextfile(textFilePath)
	}
}

func main() {
	kingpin.Version(getExporterVersion(debug.ReadBuildInfo))
	kingpin.Parse()
	// TODO: create custom kingpin template to display bool defaults

	// Single textfile
	writeAllMetricsToTextfiles()
}
