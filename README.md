# go-ethtool-exporter

Prometheus ethtool exporter on top of [go-ethtool-metrics](https://github.com/newrushbolt/go-ethtool-metrics) library

## Why&How

https://github.com/newrushbolt/go-ethtool-metrics?tab=readme-ov-file#motivation-for-library-instead-of-exporter

### When to use

-> USECASES.md

### Comparing to other exporters

<!-- Aligning these tables in GFM is a nightmare. <nobr> does not work, non-breaking hyphen too. nbsp works, but makes source MD very hard to read.
I was one step away from doing the whole thing in pure HTML 😭 -->

#### Metrics

|                                                      Exporter                                                           | Structured and vendor abstract | [Generic info](## "ethtool eth0") | [Module info](## "ethtool -m eth0") | [Driver info](## "ethtool -m eth0") |  [Statistics](## "ethtool -S eth0")   |
|-------------------------------------------------------------------------------------------------------------------------|--------------------------------|-----------------------------------|-------------------------------------|-------------------------------------|---------------------------------------|
| [prometheus/node_exporter](https://github.com/prometheus/node_exporter/blob/master/collector/ethtool_linux.go)          |               ✅               |             ✅                     |       ❌                            |     ✅                              | 🧩 General (bytes, packets and errors) |
| [influxdata/telegraf](https://github.com/influxdata/telegraf/blob/master/plugins/inputs/ethtool)                        |               ❌               |             ✅                     |       ❌                            |     ❌                              | ✅                                     |
| [newrushbolt/go-ethtool-exporter](https://github.com/newrushbolt/go-ethtool-exporter)                                   |               ✅               |             ✅                     |       ✅                            |     ✅                              | ✅ General and per-queue               |
| [newrushbolt/prometheus-ethtool-exporter](https://github.com/newrushbolt/prometheus-ethtool-exporter)                   |               ❌               |             ✅                     |       ✅                            |     ✅                              | ✅                                     |
| [adeverteuil/ethtool_exporter](https://github.com/adeverteuil/ethtool_exporter)                                         |               ❌               |             ✅                     |       ❌                            |     ❌                              | 🧩 Except per-queue for bnxt_en        |
| [Showmax/prometheus-ethtool-exporter](https://github.com/Showmax/prometheus-ethtool-exporter)                           |               ❌               |             ✅                     |       ✅                            |     ❌                              | 🧩 Except per-queue for bnxt_en        |
| [slyngshede/prometheus-ethtool-exporter](https://gitlab-replica-b.wikimedia.org/slyngshede/prometheus-ethtool-exporter) |               ❌               |             ✅                     |       ✅                            |     ❌                              | 🧩 Except per-queue for bnxt_en        |

#### Other features

|                                                    Exporter                                                             | Deploy dependencies   |         Port discovery        |         Test data          | Provides alerts | Provides dashboards |   Provides deploy methods   |
|-------------------------------------------------------------------------------------------------------------------------|-----------------------|-------------------------------|----------------------------|-----------------|---------------------|-----------------------------|
| [prometheus/node_exporter](https://github.com/prometheus/node_exporter/blob/master/collector/ethtool_linux.go)          |   ✅Single binary     | ❌Only regexps                | 🧩Only synthetic            |     🧩3rd party  |         ❌          | ✅                          |
| [influxdata/telegraf](https://github.com/influxdata/telegraf/blob/master/plugins/inputs/ethtool)                        |   ✅Single binary     | ❌Only regexps                | 🧩Partly                    |     🧩3rd party  |         ❌          | ✅                          |
| [newrushbolt/go-ethtool-exporter](https://github.com/newrushbolt/go-ethtool-exporter)                                   |   ✅Single binary     | ✅By types, bridges and bonds | ✅                          |     💪Planned    |         💪Planned   | 💪Planned                   |
| [newrushbolt/prometheus-ethtool-exporter](https://github.com/newrushbolt/prometheus-ethtool-exporter)                   |   ❌Python + modules  | ❌Only regexps                | ✅                          |     ❌           |         ❌          | ❌                          |
| [adeverteuil/ethtool_exporter](https://github.com/adeverteuil/ethtool_exporter)                                         |   ❌Python + modules  | ❌Only regexps                | ❌Not really, only one case |     ❌           |         ❌          | ❌                          |
| [Showmax/prometheus-ethtool-exporter](https://github.com/Showmax/prometheus-ethtool-exporter)                           |   ❌Python + modules  | ❌Only regexps                | ❌                          |     ❌           |         ❌          | 🧩Daemonset                 |
| [slyngshede/prometheus-ethtool-exporter](https://gitlab-replica-b.wikimedia.org/slyngshede/prometheus-ethtool-exporter) |   ❌Python + modules  | ❌Only regexps                | ❌                          |     ❌           |         ❌          | 🧩Daemonset, Debian package |

## Usage

### Port discovery

Contrary to other exporters (prometheus-ethtool-exporter,node_exporter, etc) we encourage you to use **set of port-discovery options** instead of crafting eye-bleeding regexes.  
There are a lot of scenarios when ports could be discovered by having a specific interface type, being part of a bond or bridge, etc.

There are several options, but if you have another scenario, don't hesitate to introduce it via a pull-request.  
Or just drop a Github issue with info about your discovery scenario, if you are not into Golang.

Read avaliable port-discovery-options in [auto-generated usage](exporter_help.go).

As the last resort, you can always use `--discover-ports-regex` together with `--discover-all-ports`.  
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

If you need more info, enable verbose logging via env:

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

### All available options

Read [auto-generated usage in repo](exporter_help.go) or run `go-ethtool-exporter --help`.
