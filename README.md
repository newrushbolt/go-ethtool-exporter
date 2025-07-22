# go-ethtool-exporter

Prometheus ethtool exporter on top of [go-ethtool-metrics](https://github.com/newrushbolt/go-ethtool-metrics) library

## Usage

### Port discovery

Contrary to other exporters (prometheus-ethtool-exporter,node_exporter, etc) we encourage you to use set of port-discovery options instead of crafting eye-bleeding regexes.  
There're a lot of scenarios when ports could be discovered by having specific interface type, belonging to bond or bridge, etc.

ATM there're several options, but if you have another scenario, don't hesitate to introduce it via pull-request.  
Or just drop an Github issue with info about your discovery scenario, if you are not into Golang.

```
Port detection settings:
  --discover-allowed-port-types=1,
    Comma-separated list of allowed interface types (see if_arp.h). Set to empty ('') to allow all port types.
  --discover-all-ports
    Discover all ports, ignoring all the other discover flags, EXCEPT for 'discover-allowed-port-types' and 'discover-ports-regex'.
  --no-discover-bond-slaves
    Discover ports that are enslaved by bonds
```

As the last resort, you can always use `--discover-ports-regex` together with enabling `--discover-all-ports`.  
But this is the least preferable and the least tested option.

To only test discovery logic, you can run exporter with `discover-ports` command:

```bash
go-ethtool-exporter discover-ports

2025/07/22 13:54:12 INFO Starting go-ethtool-exporter
Discovered following ports:
  - erspan0
  - eth0
  - gretap0
```

If you need more info, enable detailed logging via env:

```bash
GO_ETHTOOL_EXPORTER_LOG_LEVEL=DEBUG ./go-ethtool-exporter discover-ports

INFO Starting go-ethtool-exporter
DEBUG Got unfiltered interface list allInterfaces="[- bonding_masters L erspan0 L eth0 L gre0 L gretap0 L ip6_vti0 L ip6gre0 L ip6tnl0 L ip_vti0 L lo L sit0 L tunl0]"

DEBUG Skipping reserved netclass entry deviceName=bonding_masters

DEBUG Port passed all filters, adding to final list deviceName=erspan0
DEBUG Port passed all filters, adding to final list deviceName=eth0

DEBUG Interface type is not allowed interfaceType=778 allowedTypes=[1] devicePath=/sys/class/net/gre0
DEBUG Not a valid device type, skipping deviceName=gre0

DEBUG Interface type is not allowed interfaceType=772 allowedTypes=[1] devicePath=/sys/class/net/lo
DEBUG Not a valid device type, skipping deviceName=lo

DEBUG Discovered following interfaces, collecting metrics interfaces="[erspan0 eth0 gretap0]"

Discovered following ports:
  - erspan0
  - eth0
  - gretap0
```

### All avaliable options

Read [auto-generated usage in repo](exporter_help.go) or run `go-ethtool-exporter --help`.
