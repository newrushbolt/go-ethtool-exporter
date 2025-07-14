# go-ethtool-exporter

Prometheus ethtool exporter on top of [go-ethtool-metrics](https://github.com/newrushbolt/go-ethtool-metrics) library

Usage:

```
usage: go-ethtool-exporter [<flags>] <command> [<args> ...]

Flags:
  --help
  --version

  --ethtool-path=/usr/sbin/ethtool
  --ethtool-timeout
                            Timeout for ethtool command execution

  --linux-net-class-path=/sys/class/net
  --textfile-directory="/var/lib/node-exporter/textfiles"
                            Path to node_exporter textfile directory.
                            Only used in "single-textfile" and "loop-textfile" modes.

  --keep-absent-metrics
                            Set `Nan` value for every metric that wasn't found

  // Enabled collectors
  --no-collect-generic-info-settings
  --no-collect-module-info-diagnostics-alarms
  --no-collect-module-info-diagnostics-warnings

  // Disabled collectors
  --collect-module-info-diagnostics-values
  --collect-driver-info-features
  --collect-generic-info-modes
  --collect-module-info-vendor

  // Port discovery options
  --discover-allowed-port-types="1,"
                            Comma-separated list of allowed interface types (see if_arp.h).
                            Set to "" to allow all port types.
  --[no-]discover-all-ports
                            Discover all ports, ignoring all the other discover flags,
                            EXCEPT for 'discover-allowed-port-types' and 'discover-ports-regex'.
  --[no-]discover-bond-slaves


Commands:

single-textfile
    Writes all metrics to textfiles ONCE. Usefull for testing or crons.


loop-textfile [<flags>]
    Writes all metrics to textfiles every loop-interval.

    --loop-textfile-update-interval=30s
      Interval between textfiles updates.
```