# otlp-network

`otlp-network` is a lightweight CLI application that exports network statistics from the Linux device it is running on. It exposes metrics for Prometheus scraping and can optionally push metrics to an OpenTelemetry (OTLP) endpoint.

## Installation

### DEB Package (Recommended)

Download the latest `.deb` package from the [GitHub Releases](https://github.com/andrewhowdencom/otlp.network/releases) page.

```bash
sudo dpkg -i otlp-network_*.deb
```

This will:
- Install the binary to `/usr/bin/otlp-network`.
- Install a Systemd service `otlp-network.service` (enabled and started automatically).
- Run the service as `otlp-network` user.

### Binary

Download the binary for your architecture from the [GitHub Releases](https://github.com/andrewhowdencom/otlp.network/releases) page and place it in your `$PATH`.

## Configuration

You can configure `otlp-network` using command-line flags or environment variables.

| Flag | Env Var | Default | Description |
|------|---------|---------|-------------|
| `--prometheus.host` | `PROMETHEUS_HOST` | *(empty)* | Host to expose Prometheus metrics (e.g. `127.0.0.1` or `0.0.0.0`). |
| `--prometheus.port` | `PROMETHEUS_PORT` | `9464` | Port to expose Prometheus metrics. |
| `--otel.endpoint` | `OTEL_ENDPOINT` | *(empty)* | OTLP HTTP endpoint to push metrics to (e.g., `localhost:4318`). |
| `--otel.insecure` | `OTEL_INSECURE` | `true` | Use insecure connection for OTLP. |
| `--otel.interval` | `OTEL_INTERVAL` | `60s` | Interval for pushing OTLP metrics. |

### Systemd Configuration

To change configuration for the Systemd service, create a drop-in override:

```bash
sudo systemctl edit otlp-network
```

Add your flags to `ExecStart`:

```ini
[Service]
ExecStart=
ExecStart=/usr/bin/otlp-network --otel.endpoint=my-collector:4318
```

## Usage

Start the exporter:

```bash
otlp-network
```

Verify metrics are available:

```bash
curl http://localhost:9464/metrics
```

### Metrics

See [OTLP.md](OTLP.md) for a comprehensive list of all exposed metrics and their attributes.
