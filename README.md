# barrage

A lightweight, concurrent HTTP load testing tool written in Go.

`barrage` fires a controlled stream of concurrent requests at an HTTP endpoint and reports latency, throughput, and error statistics — built to understand real-world concurrency patterns from the inside out, not just to use them.

> Status: early development (v0.2). Not production-ready yet.

## Why

Most load testing tools are black boxes. `barrage` is built from scratch to:
- Be simple enough to read and extend for your own use cases
- Support custom payloads out of the box
- Serve as a hands-on exploration of Go concurrency (goroutines, channels, worker pools, generics)

## Installation

```bash
go install github.com/nurullahgd/barrage@latest
```

Or build from source:

```bash
git clone https://github.com/nurullahgd/barrage.git
cd barrage
go build -o barrage ./cmd/barrage
```

## Usage

```bash
barrage -url http://localhost:8080/v1/markets \
        -method GET \
        -requests 5000 \
        -workers 20 \
        -rate 100
```

### Flags

| Flag        | Description                                        | Default |
|-------------|-----------------------------------------------------|---------|
| `-url`      | Target URL (required)                                | —       |
| `-method`   | HTTP method                                          | `GET`   |
| `-body`     | Request body (raw JSON string)                       | —       |
| `-requests` | Total number of requests to send                     | `20`    |
| `-workers`  | Max concurrent in-flight requests                    | `5`     |
| `-rate`     | Requests per second (`0` = as fast as workers allow) | `0`     |

### Example output

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

## Roadmap

- [x] v0.1 — Fixed worker pool, latency report (min/avg/max), percentile stats (p50/p95/p99)
- [x] v0.1 — Error classification (timeout / connection error / 4xx / 5xx)
- [x] v0.2 — Rate limiting (`-rate`) via a ticker-fed job queue
- [ ] v0.3 — Duration-based runs (`-duration`, run for a fixed time instead of a fixed count)
- [ ] v0.4 — JSON output format
- [ ] v0.5 — Live terminal dashboard, custom payload templates

## Testing

```bash
go test ./... -race
```

Covers:
- Table-driven tests for the generic `Percentile` function and `ComputeStats`
- `httptest`-based tests for `DoRequest` across success / 4xx / 5xx paths
- A goroutine-leak test (via [goleak](https://github.com/uber-go/goleak)) verifying the worker pool cleans up correctly

## How it works

`barrage` uses a fixed-size worker pool of goroutines to bound concurrency, an optional `time.Ticker`-based feeder to pace jobs at a target rate, and a fan-in pattern to collect results from workers into a single aggregator without data races. See the code for the full breakdown — this project is meant to be read, not just run.

## Contributing

This is primarily a learning project, but issues, suggestions, and PRs are welcome.

## License

[MIT](LICENSE)