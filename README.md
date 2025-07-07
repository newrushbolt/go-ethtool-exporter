# go-ethtool-exporter

Prometheus ethtool exporter on top of [go-ethtool-metrics](https://github.com/newrushbolt/go-ethtool-metrics) library

Usage:

```
./go-ethtool-exporter --help
usage: go-ethtool-exporter [<flags>] <command> [<args> ...]

Flags:
  --help                   Show context-sensitive help.
  --ethtool-path=/usr/sbin/ethtool

  --linux-net-class-path=/sys/class/net

  --textfile-directory="var/lib/node-exporter/textfiles"
                                Path to node_exporter textfile directory. Only used in "single-textfile" and "loop-textfile" modes.
  --[no-]collect-generic-info-settings

  --[no-]collect-module-info-diagnostics-alarms

  --[no-]collect-module-info-diagnostics-warnings

  --[no-]collect-driver-info-features

  --[no-]collect-generic-info-modes

  --[no-]collect-module-info-diagnostics-values

  --[no-]collect-module-info-vendor

  --[no-]skip-non-bonded-ports
  --allowed-interface-types="1,"
                                Comma-separated list of allowed interface types (see if_arp.h)
  --version                Show application version.

Commands:

single-textfile
    Writes all metrics to textfiles ONCE. Usefull for testing or cron calls.


loop-textfile [<flags>]
    Writes all metrics to textfile every "loop-interval".

    --loop-textfile-update-interval=30s
      Interval between textfile updates.
```