# barrage

[![CI](https://github.com/nurullahgd/barrage/actions/workflows/ci.yml/badge.svg)](https://github.com/nurullahgd/barrage/actions/workflows/ci.yml)

A lightweight, concurrent HTTP load testing tool written in Go.

`barrage` fires a controlled stream of concurrent requests at an HTTP endpoint and reports latency, throughput, and error statistics — built to understand real-world concurrency patterns from the inside out, not just to use them.

![barrage demo](demo.gif)

## Why

Most load testing tools are black boxes. `barrage` is built from scratch to:
- Be simple enough to read and extend for your own use cases
- Support custom payloads, rate limiting, and HTTP/2 out of the box
- Show a live terminal dashboard while the test runs
- Serve as a hands-on exploration of Go concurrency (goroutines, channels, worker pools, generics)

## Installation

```bash
go install github.com/nurullahgd/barrage@latest
```

Or build from source:

```bash
git clone https://github.com/nurullahgd/barrage.git
cd barrage
make build
```

## Quick start

```bash
# Terminal 1 — start the bundled mock server
go run ./cmd/mockserver -port 8090 -latency 5ms -error-rate 0.02

# Terminal 2 — run barrage against it
go run ./cmd/barrage -url http://localhost:8090 -requests 1000 -workers 20 -rate 100
```

### Mock server flags

| Flag          | Description                                              | Default |
|---------------|----------------------------------------------------------|---------|
| `-port`       | Port to listen on                                        | `8090`  |
| `-latency`    | Simulated response latency (e.g. `10ms`)                 | `0`     |
| `-error-rate` | Fraction of requests returning 500 (e.g. `0.05` for 5%) | `0`     |
| `-h2c`        | Enable HTTP/2 cleartext mode                             | `false` |

## Usage

```bash
barrage -url http://localhost:8090 \
        -method GET \
        -requests 5000 \
        -workers 20 \
        -rate 100 \
        -duration 30s \
        -format dashboard
```

### Flags

| Flag        | Description                                                              | Default     |
|-------------|--------------------------------------------------------------------------|-------------|
| `-url`      | Target URL (required)                                                    | —           |
| `-method`   | HTTP method                                                              | `GET`       |
| `-body`     | Request body (raw JSON string)                                           | —           |
| `-requests` | Total number of requests to send                                         | `20`        |
| `-workers`  | Max concurrent in-flight requests                                        | `5`         |
| `-rate`     | Requests per second (`0` = as fast as workers allow)                     | `0`         |
| `-duration` | Test duration (e.g. `30s`, `2m`); `0` = run until requests are exhausted | `0`        |
| `-format`   | Output format: `dashboard`, `text`, `json`                               | `dashboard` |
| `-http2`    | Enable HTTP/2 (h2c for cleartext)                                        | `false`     |

### Example output (text)

```
Total requests:      5000
Successful:          4987 (99.7%)
Failed:                 13 (0.3%)
  Breakdown:
    Timeouts:             9
    Connection errors:    0
    4xx responses:        0
    5xx responses:        4

Min latency:         320µs
Avg latency:         3.1ms
Max latency:         9.98s
p50:                 890µs
p95:                 12.4ms
p99:                 41.2ms

Requests/sec:        1420.55
```

### Example output (json)

```json
{
  "total_requests": 5000,
  "successful": 4987,
  "failed": 13,
  "error_breakdown": {
    "timeouts": 9,
    "connection_errors": 0,
    "client_errors": 0,
    "server_errors": 4
  },
  "latency": {
    "min": 0.000320,
    "avg": 0.003100,
    "max": 9.980000,
    "p50": 0.000890,
    "p95": 0.012400,
    "p99": 0.041200
  },
  "requests_per_sec": 1420.55
}
```

## Development

```bash
make build        # build barrage binary
make build-mock   # build mockserver binary
make test         # run tests with -race
make test-cover   # run tests + open coverage report
make vet          # go vet
make lint         # golangci-lint
make vuln         # govulncheck
make check        # vet + lint + vuln + test (CI equivalent)
make run-mock     # start mock server on :8090
make run-example  # run barrage against mock server
make clean        # remove build artifacts
```

## Testing

```bash
make test
```

Covers:
- Table-driven tests for the generic `Percentile` function and `ComputeStats`
- `httptest`-based integration tests for `DoRequest` (success, 4xx, 5xx, timeout, body forwarding)
- Goroutine-leak test via [goleak](https://github.com/uber-go/goleak) verifying the worker pool cleans up correctly
- Context cancellation test verifying job feeding stops on `cancel()`
- Bubbletea `Model.Update` unit tests (result counting, done signal)
- `Stats.ToJSON()` and `round6` unit tests

## Roadmap

- [x] v0.1 — Fixed worker pool, latency report (min/avg/max), percentile stats (p50/p95/p99)
- [x] v0.1 — Error classification (timeout / connection error / 4xx / 5xx)
- [x] v0.2 — Rate limiting (`-rate`) via a ticker-fed job queue
- [x] v0.3 — Duration-based runs (`-duration`) + graceful shutdown on Ctrl+C
- [x] v0.4 — JSON output format (`-format json`)
- [x] v0.5 — Live terminal dashboard (`-format dashboard`, default)
- [x] v0.6 — Latency distribution histogram in dashboard
- [x] v0.7 — HTTP/2 support (`-http2`, `-h2c`)
- [x] v1.0 — Stable API, comprehensive test suite, CI, Makefile, CHANGELOG

## How it works

`barrage` uses a fixed-size worker pool of goroutines to bound concurrency, an optional `time.Ticker`-based feeder to pace jobs at a target rate, and a fan-in pattern to collect results via a receive-only channel (`<-chan Result`) without holding all results in memory. The live dashboard is powered by [bubbletea](https://github.com/charmbracelet/bubbletea) and receives result events via `p.Send()` as each request completes. See the code — this project is meant to be read, not just run.

## Contributing

Issues, suggestions, and PRs are welcome.

## License

[MIT](LICENSE)