package main

import (
	"github.com/alecthomas/kingpin/v2"
)

// TODO: add flag (--report?) for creating full metrics report archive,
// containing output of all supporter metrics modes for all interfaces,
// and all meta for interface discovery: bond modes, interface types, etc
// https://github.com/newrushbolt/go-ethtool-metrics/blob/main/testdata/README.md

var (
	discoverPortsCommand = kingpin.Command("discover-ports", "Show discovered ports and exit")

	singleTextfileCommand = kingpin.Command("single-textfile", "Writes all metrics to textfile ONCE. Usefull for testing or crons")

	loopTextfileCommand        = kingpin.Command("loop-textfile", "Writes all metrics to textfile every loop-interval")
	loopTextfileUpdateInterval = loopTextfileCommand.Flag("loop-textfile-update-interval", "Interval between textfile updates").Default("30s").Duration()

	httpServerCommand = kingpin.Command("http-server", "Starts HTTP server of scraping metrics over HTTP(S), like node-exporter does")
	httpListenAddress = httpServerCommand.Flag("web.listen-address", "Address on which to expose metrics").Default(":9417").String()
	// Without caching it seems like 2 requests is enough. And +1 for /health method
	httpMaxRequests = httpServerCommand.Flag("web.max-requests", "Maximum number of concurrent HTTP requests").Default("3").Int()
	// TODO: implement http server params, such as
	// - inmemory metric cache and it's max ttl
	// - TLS-related stuff

	// TODO: add env support???
	// FLAG GROUP START: Ethtool settings
	ethtoolPath    = kingpin.Flag("path.ethtool", "").Default("/usr/sbin/ethtool").ExistingFile()
	ethtoolTimeout = kingpin.Flag("ethtool-timeout", "Timeout for ethtool command execution.").Default("5s").Duration()
	// FLAG GROUP END

	// FLAG GROUP START: Various paths settings
	linuxNetClassPath = kingpin.Flag("path.sysfs.net.class", "").Default("/sys/class/net").ExistingDir()
	textfileDirectory = kingpin.Flag("path.textfile-directory", "Path to the node_exporter textfile directory. Only used in 'single-textfile' and 'loop-textfile' modes").Default("/var/lib/node-exporter/textfiles").String()
	// FLAG GROUP END

	// FLAG GROUP START: Collectors, enabled by default
	collectGenericInfoSettings           = kingpin.Flag("collect-generic-info-settings", "").Default("true").Bool()
	collectDriverInfoCommon              = kingpin.Flag("collect-driver-info-common", "").Default("true").Bool()
	collectModuleInfoDiagnosticsAlarms   = kingpin.Flag("collect-module-info-diagnostics-alarms", "").Default("true").Bool()
	collectModuleInfoDiagnosticsWarnings = kingpin.Flag("collect-module-info-diagnostics-warnings", "").Default("true").Bool()
	collectStatisticsPerQueueGeneral     = kingpin.Flag("collect-statistics-per-queue-general", "").Default("true").Bool()
	// FLAG GROUP END

	// FLAG GROUP START: Collectors, disabled by default
	collectAllMetrics                  = kingpin.Flag("collect-all-metrics", "Ignores all the flags below. Usefull for testing").Default("false").Bool()
	collectDriverInfoFeatures          = kingpin.Flag("collect-driver-info-features", "").Default("false").Bool()
	collectGenericInfoModes            = kingpin.Flag("collect-generic-info-modes", "").Default("false").Bool()
	collectModuleInfoDiagnosticsValues = kingpin.Flag("collect-module-info-diagnostics-values", "").Default("false").Bool()
	collectModuleInfoVendor            = kingpin.Flag("collect-module-info-vendor", "").Default("false").Bool()
	collectStatisticsGeneral           = kingpin.Flag("collect-statistics-general", "").Default("false").Bool()
	collectStatisticsPerQueuePerType   = kingpin.Flag("collect-statistics-per-queue-per-type", "").Default("false").Bool()
	// FLAG GROUP END

	// FLAG GROUP START: Port detection settings
	// All possible types: https://github.com/torvalds/linux/blob/772b78c2abd85586bb90b23adff89f7303c704c7/include/uapi/linux/if_arp.h#L29
	discoverAllowedPortTypes = kingpin.Flag("discover-allowed-port-types", "Comma-separated list of allowed interface types (see if_arp.h). Set to empty ('') to allow all port types").Default("1,").String()
	discoverAllPorts         = kingpin.Flag("discover-all-ports", "Force discover all the ports, ignoring all the other discover flags, EXCEPT for 'discover-allowed-port-types' and 'discover-ports-regexp'").Default("false").Bool()
	discoverBondSlaves       = kingpin.Flag("discover-bond-slaves", "Whether to discover ports that are enslaved by bonds").Default("true").Bool()
	discoverBridgeSlaves     = kingpin.Flag("discover-bridge-slaves", "Whether to discover ports that are enslaved by bridges").Default("false").Bool()
	discoverPortsRegexp      = kingpin.Flag("discover-ports-regexp", "Only discover ports with names matching this regexp").Default(".+").Regexp()
	// Not yet implemented
	// discoverOvsSlaves
	// discoverBondMasters
	// Detect aliases and naming types?
	// FLAG GROUP END

	// FLAG GROUP START: Keep absent metrics, setting 'Nan' value for every metric that was not found
	// Absent metrics (*float64 nil) behavior
	// https://github.com/newrushbolt/go-ethtool-metrics/tree/v0.0.3?tab=readme-ov-file#missing-metrics
	keepAbsentMetricsModuleInfo  = kingpin.Flag("keep-absent-metrics-module-info", "").Default("true").Bool()
	keepAbsentMetricsGenericInfo = kingpin.Flag("keep-absent-metrics-generic-info", "").Default("false").Bool()
	keepAbsentMetricsDriverInfo  = kingpin.Flag("keep-absent-metrics-driver-info", "").Default("false").Bool()
	keepAbsentMetricsStatistics  = kingpin.Flag("keep-absent-metrics-statistics", "").Default("false").Bool()
	// FLAG GROUP END

	// FLAG GROUP START: Metrics processing settings
	// Check the metrics library for more info
	// https://github.com/newrushbolt/go-ethtool-metrics/blob/9c84000a5e0736e721630447958639d09cc532d1/pkg/metrics/statistics/statistics_structs.go#L6
	statisticsGenerateMissingPerQueueMetrics = kingpin.Flag("statistics-generate-missing-per-queue-metrics", "Generate missing metrics per queue if missing (eg in Broadcom bnxt_en driver)").Default("true").Bool()
	listLabelFormat                          = kingpin.Flag("list-label-format", "How to transform lists of strings to prometheus labels").Default("multi-label").Enum("single-label", "multi-label", "both")
	// FLAG GROUP END
)
