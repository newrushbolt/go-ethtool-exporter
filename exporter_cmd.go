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

	singleTextfileCommand = kingpin.Command("single-textfile", "Writes all metrics to textfiles ONCE. Usefull for testing or crons.")

	loopTextfileCommand        = kingpin.Command("loop-textfile", "Writes all metrics to textfiles every loop-interval.")
	loopTextfileUpdateInterval = loopTextfileCommand.Flag("loop-textfile-update-interval", "Interval between textfiles updates.").Default("30s").Duration()

	// TODO: add env support???
	// FLAG GROUP START: Ethtool settings
	ethtoolPath    = kingpin.Flag("ethtool-path", "").Default("/usr/sbin/ethtool").ExistingFile()
	ethtoolTimeout = kingpin.Flag("ethtool-timeout", "Timeout for ethtool command execution.").Default("5s").Duration()
	// FLAG GROUP END

	// FLAG GROUP START: Various paths settings
	linuxNetClassPath = kingpin.Flag("linux-net-class-path", "").Default("/sys/class/net").ExistingDir()
	textfileDirectory = kingpin.Flag("textfile-directory", "Path to node_exporter textfile directory. Only used in 'single-textfile' and 'loop-textfile' modes.").Default("/var/lib/node-exporter/textfiles").String()
	// FLAG GROUP END

	// FLAG GROUP START: Collectors, enabled by default
	collectGenericInfoSettings           = kingpin.Flag("collect-generic-info-settings", "").Default("true").Bool()
	collectModuleInfoDiagnosticsAlarms   = kingpin.Flag("collect-module-info-diagnostics-alarms", "").Default("true").Bool()
	collectModuleInfoDiagnosticsWarnings = kingpin.Flag("collect-module-info-diagnostics-warnings", "").Default("true").Bool()
	// FLAG GROUP END

	// FLAG GROUP START: Collectors, disabled by default
	collectDriverInfoFeatures          = kingpin.Flag("collect-driver-info-features", "").Default("false").Bool()
	collectGenericInfoModes            = kingpin.Flag("collect-generic-info-modes", "").Default("false").Bool()
	collectModuleInfoDiagnosticsValues = kingpin.Flag("collect-module-info-diagnostics-values", "").Default("false").Bool()
	collectModuleInfoVendor            = kingpin.Flag("collect-module-info-vendor", "").Default("false").Bool()
	// FLAG GROUP END

	// FLAG GROUP START: Port detection settings
	// All possible types: https://github.com/torvalds/linux/blob/772b78c2abd85586bb90b23adff89f7303c704c7/include/uapi/linux/if_arp.h#L29
	discoverAllowedPortTypes = kingpin.Flag("discover-allowed-port-types", "Comma-separated list of allowed interface types (see if_arp.h). Set to empty ('') to allow all port types.").Default("1,").String()
	discoverAllPorts         = kingpin.Flag("discover-all-ports", "Force discover all ports, ignoring all the other discover flags, EXCEPT for 'discover-allowed-port-types' and 'discover-ports-regex'.").Default("false").Bool()
	discoverBondSlaves       = kingpin.Flag("discover-bond-slaves", "Whether we discover ports that are enslaved by bonds").Default("true").Bool()
	// TODO discoverPortsRegex
	// discoverPortsRegex       = kingpin.Flag("discover-ports-regex", "Only discover ports with names matching this regex.").Default(".*").Regexp()
	// Not yet implemented
	// discoverOvsSlaves
	// discoverBondMasters
	// Detect aliases and naming types?
	// FLAG GROUP END

	// FLAG GROUP START: Absent metrics settings
	// Absent metrics (*float64 nil) behavior
	// https://github.com/newrushbolt/go-ethtool-metrics/tree/v0.0.3?tab=readme-ov-file#missing-metrics
	// Maybe they should be per-module?
	keepAbsentMetrics = kingpin.Flag("keep-absent-metrics", "Set 'Nan' value for every metric that wasn't found").Default("false").Bool()
	// FLAG GROUP END
)
