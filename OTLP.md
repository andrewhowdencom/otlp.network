# OTLP Metrics Documentation

This document lists all the metrics exposed by `otlp-network`, specifically designed to adhere to OpenTelemetry semantic conventions where possible.

## Resource Attributes

All metrics exported by this agent are associated with a Resource that identifies the service.

| Attribute | Description | Example |
| :--- | :--- | :--- |
| `service.name` | The name of the service. | `otlp-network` |
| `service.version` | The version of the application. | `0.1.0` |

## Metric Instruments

### Device Collector (`device`)
Collects basic network interface statistics. Sourced from `/proc/net/dev`.

| Metric Name | Type | Unit | Description | Attributes |
| :--- | :--- | :--- | :--- | :--- |
| `device.io` | Sum | Bytes | Total bytes transmitted or received. | `interface`: Interface name (e.g., `eth0`)<br>`direction`: `receive` \| `transmit` |
| `device.packets` | Sum | {packets} | Total packets transmitted or received. | `interface`, `direction` |
| `device.errors` | Sum | {errors} | Total errors encountered. | `interface`, `direction` |
| `device.dropped` | Sum | {packets} | Total dropped packets. | `interface`, `direction` |

### Wifi Collector (`wifi`)
Collects wireless signal quality statistics. Sourced from `/proc/net/wireless`.
*Only enabled if wireless interfaces are present.*

| Metric Name | Type | Unit | Description | Attributes |
| :--- | :--- | :--- | :--- | :--- |
| `wifi.signal` | Gauge | dBm | Current signal level. | `interface`: Interface name (e.g., `wlan0`) |
| `wifi.quality` | Gauge | 1 | Link quality (value depends on driver, often relative). | `interface` |

### TCP Collector (`tcp`)
Collects global TCP connection statistics. Sourced from `/proc/net/snmp`.

| Metric Name | Type | Unit | Description | Attributes |
| :--- | :--- | :--- | :--- | :--- |
| `tcp.connection.current` | Gauge | {connections} | Currently established TCP connections. | *(none)* |
| `tcp.connection.total` | Sum | {connections} | Total TCP connections opened (active + passive). | *(none)* |
| `tcp.retransmit` | Sum | {segments} | Total TCP segments retransmitted. | *(none)* |

### UDP Collector (`udp`)
Collects global UDP statistics. Sourced from `/proc/net/snmp`.

| Metric Name | Type | Unit | Description | Attributes |
| :--- | :--- | :--- | :--- | :--- |
| `udp.packets` | Sum | {datagrams} | Total UDP datagrams delivered/received. | `direction`: `in` \| `out` |
| `udp.drops` | Sum | {datagrams} | Total UDP datagrams dropped. | `reason`: `no_port` \| `rcv_buf_error` \| `snd_buf_error` \| `ignored_multi` \| `mem_error` |

### Conntrack Collector (`conntrack`)
Collects Netfilter Connection Tracking statistics. Sourced from `/proc/sys/net/netfilter/`.

| Metric Name | Type | Unit | Description | Attributes |
| :--- | :--- | :--- | :--- | :--- |
| `conntrack.entries` | Gauge | {entries} | Number of entries currently in the conntrack table. | *(none)* |
| `conntrack.limit` | Gauge | {entries} | Maximum number of entries allowed in the table. | *(none)* |

### Softnet Collector (`softnet`)
Collects Softnet (software interrupt) processing statistics. Sourced from `/proc/net/softnet_stat`.
*Aggregated globally across all CPUs.*

| Metric Name | Type | Unit | Description | Attributes |
| :--- | :--- | :--- | :--- | :--- |
| `softnet.processed` | Sum | {packets} | Total packets processed by softnet. | *(none)* |
| `softnet.dropped` | Sum | {packets} | Total packets dropped (queue overflow). | *(none)* |
| `softnet.squeezed` | Sum | {times} | Times softnet ran out of quota (time squeezed). | *(none)* |

### Sockstat Collector (`sockstat`)
Collects global socket allocation statistics. Sourced from `/proc/net/sockstat`.

| Metric Name | Type | Unit | Description | Attributes |
| :--- | :--- | :--- | :--- | :--- |
| `sockets.used` | Gauge | {sockets} | Total used sockets (all protocols). | *(none)* |
| `sockets.tcp.inuse` | Gauge | {sockets} | TCP sockets currently in use. | *(none)* |
| `sockets.udp.inuse` | Gauge | {sockets} | UDP sockets currently in use. | *(none)* |

### Uptime Collector (`uptime`)

| Metric Name | Type | Unit | Description | Attributes |
| :--- | :--- | :--- | :--- | :--- |
| `uptime` | Sum (Monotonic) | s | Application uptime in seconds. | *(none)* |
