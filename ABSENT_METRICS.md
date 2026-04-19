# Absent metrics

Noticing missing metrics in Prometheus-compatible TSDB is not very intuitive, and alerting is even worse.

For the simplest case when you have one instance of `ethtool_module_info_diagnostics_alarms_output_low` per monitoring server, it's pretty straightforward:

```SQL
absent(ethtool_module_info_diagnostics_alarms_output_low)
```

For more complex cases when you may miss **one metric per exporter per host** it starts to look worse:

```SQL
up(job="ethtool_exporter")
UNLESS on(instance)
ethtool_module_info_diagnostics_alarms_output_low
```

And worse with per-device metrics:

```SQL
-- We have to assume that at least some per-device metrics are collected
ethtool_generic_info_settings_info(job="ethtool_exporter")
UNLESS on(device, instance)
ethtool_module_info_diagnostics_alarms_output_low

-- Or take a slippery slope of joining metrics across different exporters,
-- assuming we have `alias` as an exporter-agnostic label for each host.
-- And don't forget to filter for the right device 🫠
node_network_up{device=~"e[a-z0-9]+"}
UNLESS on(device, alias)
ethtool_module_info_diagnostics_alarms_output_low
```

But all of the tricks above become useless when you have **several metrics per exporter**. You cannot detect their absence via regex like `{__name__=~"some_metric_[a-z]+"}`.

With most exporters, your only option is to **explicitly have an alert for each missing metric**. And to update the list on time with new metrics, of course 🫠

What can we do to make this situation better?

## Nan

### Theory

In theory, this is where Nan could be really handy. If the exporter encounters a missing metric, it could just expose it as usual with the `Nan` value:

```SQL
ethtool_module_info_diagnostics_alarms_output_low{device="eth0"} 1
ethtool_module_info_diagnostics_alarms_output_high{device="eth0"} 1
ethtool_module_info_diagnostics_alarms_output_low{device="eth1"} 1
ethtool_module_info_diagnostics_alarms_output_high{device="eth1"} Nan
```

And Go has a very convenient way to work with it internally: just use `*float64`, and convert `*nil` to Nan.

Also, it makes manual debugging the missing metrics on the host a breeze:

```
curl -s http://127.0.0.1:9999/metrics | grep Nan

ethtool_module_info_diagnostics_alarms_output_high{device="eth1"} Nan
```

Talking alerts-wise, you should be able to just:

```yaml
# We can even filter out only the metrics we really care about, thanks to well-structured metric name
- alert: SomeExporterMissingMetric
  expr: "{__name__=~ethtool_module_info_diagnostics_alarms_[a-z]+} == Nan"
  annotations:
    summary: >
      Ethtool-exporter is missing module diagnostics alarm metrics for device <{{ $labels.device }}>
      We might miss an incorrect module state
```

With a little help of `label_replace()` you can even expose the metric name in the alert. This way you know exactly which metric is missing, and can create a well-targeted silent for specific hosts where you expect it to be missing.

At the moment it looks like the only possible downside to this approach is high metric cardinality, since we expose missing metrics as is, with all the labels.

Sounds amazing, right?

### Real world

Sadly, the outside world has its own thoughts on this matter.

Prometheus scrapes, ingests and returns Nan, but does not provide any proper way to query for it.

`== Nan` won't work and no `isNan()` is available. The way is **to just compare the metric with itself**. Since Nan is not equal to anything, this is the only case when a metric can be not equal to itself.

```SQL
ethtool_module_info_diagnostics_alarms_output_high
!=
ethtool_module_info_diagnostics_alarms_output_high
```

This won't work with `{__name__=~"ethtool_module_info_diagnostics_alarms_[a-z]+"}` straight away, but we can make it work by adding `__name__` to the label matcher:

```SQL
{__name__=~"ethtool_module_info_diagnostics_alarms_[a-z]+"}
!= on(__name__, instance, device)
{__name__=~"ethtool_module_info_diagnostics_alarms_[a-z]+"}
```

Not pretty, but it kinda works. But it gets worse.

As you may or may not know, a lot of folks are using prometheus-compatible engines these days, because vanilla Prometheus does not scale up that well.

VictoriaMetrics handles Nan in a different way, [dropping them at ingestion stage](https://medium.com/@romanhavronenko/victoriametrics-promql-compliance-d4318203f51e#nans).

And OpenMetrics 1.0.0, which will likely become the main standard for Prometheus-compatible engines, explicitly [states](https://github.com/prometheus/OpenMetrics/blob/v1.0.0/specification/OpenMetrics.md#nan) that "Nan MUST NOT be used as a marker for missing or otherwise bad data".

Sadly, it seems like providing only `Nan` for missing metrics **is a dead end**.

I decided to provide more future-proof options, but you can still expose `Nan` via flags below:

```
  --absent-metrics-driver-info-expose-nan
  --absent-metrics-generic-info-expose-nan
  --absent-metrics-module-info-expose-nan
  --absent-metrics-statistics-expose-nan
```

## Counter

This approach is much simpler: just count missing events and expose that counter.

Counters are enabled by default. Getting metrics:

```yaml
# Query
ethtool_missing_metrics_total > 0

# Response
ethtool_missing_metrics_total{collector="generic_info",device="eth5"} 2
```

Setting an alert:

```yaml
- alert: SomeExporterMissingMetric
  expr: rate(ethtool_missing_metrics_total[5m]) > 0
  annotations:
    summary: Ethtool-exporter is missing metrics for device <{{ $labels.device }}> and collector <{{ $labels.collector }}>
```

Figuring out which exact metric is missing is trickier though:

```bash
GO_ETHTOOL_EXPORTER_LOG_LEVEL=DEBUG ./go-ethtool-exporter single-textfile | grep 'missing metric'

DEBUG Incrementing total counter for missing metric metric_name=ethtool_generic_info_settings_speed_bytes, labels="collector=generic_info device=eth5"
```

And what's even worse, you can only silence alerts like that **per-collector**, so later you may not notice if some other metrics went missing.

But with relatively-low cardinality, these counters are enabled by default. You can disable them using flags:

```
--no-absent-metrics-driver-info-expose-total-counter
--no-absent-metrics-generic-info-expose-total-counter
--no-absent-metrics-module-info-expose-total-counter
--no-absent-metrics-statistics-expose-total-counter
```

## Detailed info

Provides a single `missing_metric_info` metric for each absent metric:

```yaml
# Query
ethtool_missing_metric_info

# Response
ethtool_missing_metric_info{collector="generic_info",device="eth5", metric_name="ethtool_generic_info_settings_speed_bytes"} 1
```

This metric can be easily alerted on, and allows creating well-targeted silences.

The downside is the same cardinality as with Nan.

By default this is enabled only for module-info, because these metrics usually are the most useful, and we usually expect them to be fully present for every diagnostics-compatible SFP module, so cardinality shouldn't get any worse.

But if your use case differs, feel free to either disable them completely, or enable for other collectors:

```
--no-absent-metrics-module-info-expose-detailed-info

--absent-metrics-driver-info-expose-detailed-info
--absent-metrics-generic-info-expose-detailed-info
--absent-metrics-statistics-expose-detailed-info
```

  <!-- - alert: NetworkTransceiverMissingData
    expr: >-
      (
        {__name__=~"module_info_diagnostics_(alarms|warnings)_.*"}
        != on(__name__, instance, device)
        {__name__=~"module_info_diagnostics_(alarms|warnings)_.*"}
      )
      AND on(instance, device)
      (
        driver_info_features_test{} == 1
        AND on(instance, device)
        generic_info_settings_info{Port="FIBRE"}
      ) -->
