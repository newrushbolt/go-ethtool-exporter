# Typical use cases

Some of them. Sorted by popularity, have some alert examples

## Enabled by default

Just work :shrug:
Make sure your port discovery work too, and add some alerts\dashboards.

### General statistics

Some of the metrics are duplicated and already provided by default kernel network counters and exported by default exporters (node_exporter, telegraf, etc).
But there are also unique events like TxCollisions or RxCrcErrors, that could really indicate some network issues.

Default alerts may look like that

```yaml

```

### Transceiver warnings and alerts

Bla-bla, why we need it

Default alerts may look like that

```yaml

```

### Driver info

Bla-bla, why we need it

Default alerts may look like that

```yaml

```

### General port info

Bla-bla, why we need it

Default alerts may look like that

```yaml

```


## Needs to be configured

## Detailed transceiver metrics

Sometime you may not want to rely on pre-defined threshold for receive\transmit levels, and monitor the exact values.
```
--collect-module-info-diagnostics-values
```




## Per-queue statistics

General, per-type

--collect-statistics-per-queue-general
--collect-statistics-per-queue-per-type



## Only per-queue XDP

Probably only makes sense for VM's

--collect-statistics-per-queue-xdp
