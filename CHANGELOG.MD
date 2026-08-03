# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - TBD

### Added
- GitHub Actions CI pipeline (vet / lint / vuln / test on every push and PR)
- Makefile with `build`, `test`, `lint`, `vuln`, `check`, `run-mock`, `run-example` targets
- `golangci-lint` configuration (errcheck, staticcheck, gocritic, misspell, exhaustive, noctx)
- Comprehensive test suite: timeout, body forwarding, context cancellation, goroutine leak,
  bubbletea Model.Update unit tests, Stats.ToJSON and round6 coverage
- `toolchain` directive in go.mod for reproducible builds

### Fixed
- Renamed `dashboad.go` → `dashboard.go` (typo)

---

## [0.7.0] - 2025-XX-XX

### Added
- `-http2` flag: enables HTTP/2 cleartext (h2c) transport on the client side
- `-h2c` flag on mockserver: enables HTTP/2 cleartext mode for local end-to-end testing
- Verified with `r.Proto` logging: `HTTP/2.0` confirmed when `-http2` is used

---

## [0.6.0] - 2025-XX-XX

### Added
- Latency distribution histogram in the live dashboard
- Logarithmic bucket series: 0-1ms, 1-2ms, 2-5ms, 5-10ms, 10-25ms, 25-50ms, 50-100ms, 100ms+
- Empty buckets are hidden; bars normalized to the widest bucket

---

## [0.5.0] - 2025-XX-XX

### Added
- Live terminal dashboard powered by [bubbletea](https://github.com/charmbracelet/bubbletea)
  and [lipgloss](https://github.com/charmbracelet/lipgloss)
- Real-time progress bar, RPS, success/failure rate, p50/p95/p99, elapsed time
- `-format` flag: `dashboard` (default), `text`, `json`
- `RunLoadTest` now returns `<-chan client.Result` (streaming) instead of `[]client.Result`

### Changed
- Dashboard is now the default output format

---

## [0.4.0] - 2025-XX-XX

### Added
- `-format json` flag: machine-readable JSON output
- `JSONReport`, `ErrorBreakdown`, `LatencyReport` structs with proper JSON tags
- Latency values serialized as float64 seconds (6 decimal places) — compatible with `jq`

---

## [0.3.0] - 2025-XX-XX

### Added
- `-duration` flag: run for a fixed time (e.g. `-duration 30s`) instead of a fixed request count
- Graceful shutdown: `Ctrl+C` stops job feeding and reports partial results
- `signal.NotifyContext` + `context.WithTimeout` — whichever fires first stops the run

---

## [0.2.0] - 2025-XX-XX

### Added
- `-rate` flag: requests per second via a `time.Ticker`-fed job queue
- `rate=0` (default) preserves unlimited behavior

### Fixed
- HTTP transport tuning: `MaxIdleConnsPerHost` set to worker count, `IdleConnTimeout` set
  to 30s — eliminated stale-connection (`EOF`) errors observed at high concurrency

---

## [0.1.0] - 2025-XX-XX

### Added
- Fixed-size worker pool with configurable worker count
- Per-request latency, status code, and error tracking
- Error classification: timeout, connection error, 4xx, 5xx
- Aggregate stats: min/avg/max latency, p50/p95/p99 percentiles, requests/sec
- Generic `Percentile[T cmp.Ordered]` function
- `-url`, `-method`, `-body`, `-requests`, `-workers` CLI flags
- Bundled mock server with configurable latency and error rate
- `internal/` package structure: `client`, `runner`, `stats`
- Table-driven tests for `Percentile` and `ComputeStats`
- `httptest`-based tests for `DoRequest`
- Goroutine-leak test via goleak

[1.0.0]: https://github.com/nurullahgd/barrage/compare/v0.7.0...v1.0.0
[0.7.0]: https://github.com/nurullahgd/barrage/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/nurullahgd/barrage/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/nurullahgd/barrage/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/nurullahgd/barrage/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/nurullahgd/barrage/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/nurullahgd/barrage/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/nurullahgd/barrage/releases/tag/v0.1.0