# barrage

A lightweight, concurrent HTTP load testing tool written in Go.

`barrage` fires a controlled stream of concurrent requests at an HTTP endpoint and reports latency, throughput, and error statistics — built to understand real-world concurrency patterns from the inside out, not just to use them.

> Status: early development (v0.1). Not production-ready yet.

## Why

Most load testing tools are black boxes. `barrage` is built from scratch to:
- Be simple enough to read and extend for your own use cases
- Support custom payloads, headers, and auth tokens out of the box
- Serve as a hands-on exploration of Go concurrency (goroutines, channels, context, generics)

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
barrage -url https://api.example.com/orders \
        -method POST \
        -body '{"symbol":"BTCUSDT","amount":10}' \
        -header "Authorization: Bearer <token>" \
        -rate 100 \
        -duration 30s \
        -workers 50
```

### Flags

| Flag        | Description                                      | Default |
|-------------|---------------------------------------------------|---------|
| `-url`      | Target URL (required)                              | —       |
| `-method`   | HTTP method                                        | `GET`   |
| `-body`     | Request body (raw string or `@file.json`)          | —       |
| `-header`   | Request header, repeatable (`-header "K: V"`)      | —       |
| `-rate`     | Requests per second                                | `10`    |
| `-duration` | Total test duration (e.g. `30s`, `2m`)             | `10s`   |
| `-workers`  | Max concurrent in-flight requests                  | `10`    |

### Example output

```
Total requests:       3000
Successful (2xx):     2950 (98.3%)
Failed:                 50 (1.7%)
Avg latency:           45ms
p50:                   38ms
p95:                  120ms
p99:                  310ms
Max latency:           890ms
Requests/sec:         99.8
```

## Roadmap

- [x] v0.1 — Fixed worker pool, sequential requests, basic latency report (min/avg/max)
- [ ] v0.2 — Rate limiting (`-rate`) and duration-based runs (`-duration`)
- [ ] v0.3 — Percentile stats (p50/p95/p99), JSON output format
- [ ] v0.4 — Graceful shutdown via context, error classification (timeout / connection refused / 4xx / 5xx)
- [ ] v0.5 — Live terminal dashboard, custom payload templates

## How it works

`barrage` uses a fixed-size worker pool of goroutines to bound concurrency, a rate limiter to control request pacing, and a fan-in pattern to collect results from workers into a single aggregator without data races. See the code for the full breakdown — this project is meant to be read, not just run.

## Contributing

This is primarily a learning project, but issues, suggestions, and PRs are welcome.

## License

[MIT](LICENSE)
