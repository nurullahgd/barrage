package stats

import (
	"cmp"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"time"

	"github.com/nurullahgd/barrage.git/internal/client"
)

type Stats struct {
	TotalRequests     int
	SuccessCount      int
	ErrorCount        int
	TimeoutCount      int
	ConnectionErrors  int
	ClientErrorCount  int
	ServerErrorCount  int
	MinLatency        time.Duration
	MaxLatency        time.Duration
	AverageLatency    time.Duration
	Percentile50      time.Duration
	Percentile95      time.Duration
	Percentile99      time.Duration
	RequestsPerSecond float64
}

type JSONReport struct {
	TotalRequests  int            `json:"total_requests"`
	Successful     int            `json:"successful"`
	Failed         int            `json:"failed"`
	ErrorBreakdown ErrorBreakdown `json:"error_breakdown"`
	Latency        LatencyReport  `json:"latency"`
	RequestsPerSec float64        `json:"requests_per_sec"`
}

type ErrorBreakdown struct {
	Timeouts         int `json:"timeouts"`
	ConnectionErrors int `json:"connection_errors"`
	ClientErrors     int `json:"client_errors"`
	ServerErrors     int `json:"server_errors"`
}

type LatencyReport struct {
	Min float64 `json:"min"`
	Avg float64 `json:"avg"`
	Max float64 `json:"max"`
	P50 float64 `json:"p50"`
	P95 float64 `json:"p95"`
	P99 float64 `json:"p99"`
}

func Percentile[T cmp.Ordered](sorted []T, p float64) T {
	n := len(sorted)
	if n == 0 {
		var zero T
		return zero
	}
	index := int(float64(n-1) * p / 100)
	return sorted[index]
}

func isSuccess(r client.Result) bool {
	return r.Error == nil && r.StatusCode >= 200 && r.StatusCode < 300
}

func ComputeStats(results []client.Result, elapsed time.Duration) Stats {
	totalRequests := len(results)
	if totalRequests == 0 {
		return Stats{}
	}

	var successCount, errorCount int
	var timeoutCount, connectionErrors, clientErrorCount, serverErrorCount int
	latencies := make([]time.Duration, 0, totalRequests)
	var sumLatency time.Duration

	for _, result := range results {
		if isSuccess(result) {
			successCount++
		} else {
			errorCount++
			switch result.Category {
			case client.CategoryTimeout:
				timeoutCount++
			case client.CategoryConnection:
				connectionErrors++
			case client.CategoryClientError:
				clientErrorCount++
			case client.CategoryServerError:
				serverErrorCount++
			}
		}
		latencies = append(latencies, result.Latency)
		sumLatency += result.Latency
	}

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	return Stats{
		TotalRequests:     totalRequests,
		SuccessCount:      successCount,
		ErrorCount:        errorCount,
		TimeoutCount:      timeoutCount,
		ConnectionErrors:  connectionErrors,
		ClientErrorCount:  clientErrorCount,
		ServerErrorCount:  serverErrorCount,
		MinLatency:        latencies[0],
		MaxLatency:        latencies[totalRequests-1],
		AverageLatency:    sumLatency / time.Duration(totalRequests),
		Percentile50:      Percentile(latencies, 50),
		Percentile95:      Percentile(latencies, 95),
		Percentile99:      Percentile(latencies, 99),
		RequestsPerSecond: float64(totalRequests) / elapsed.Seconds(),
	}
}

func PrintReport(stats Stats) {
	successRate := float64(stats.SuccessCount) / float64(stats.TotalRequests) * 100
	errorRate := float64(stats.ErrorCount) / float64(stats.TotalRequests) * 100

	fmt.Printf("Total requests:      %d\n", stats.TotalRequests)
	fmt.Printf("Successful:          %d (%.1f%%)\n", stats.SuccessCount, successRate)
	fmt.Printf("Failed:              %d (%.1f%%)\n", stats.ErrorCount, errorRate)
	if stats.ErrorCount > 0 {
		fmt.Println("  Breakdown:")
		fmt.Printf("    Timeouts:           %d\n", stats.TimeoutCount)
		fmt.Printf("    Connection errors:  %d\n", stats.ConnectionErrors)
		fmt.Printf("    4xx responses:      %d\n", stats.ClientErrorCount)
		fmt.Printf("    5xx responses:      %d\n", stats.ServerErrorCount)
	}
	fmt.Println()
	fmt.Printf("Min latency:         %s\n", stats.MinLatency.Round(time.Microsecond*10))
	fmt.Printf("Avg latency:         %s\n", stats.AverageLatency.Round(time.Microsecond*10))
	fmt.Printf("Max latency:         %s\n", stats.MaxLatency.Round(time.Microsecond*10))
	fmt.Printf("p50:                 %s\n", stats.Percentile50.Round(time.Microsecond*10))
	fmt.Printf("p95:                 %s\n", stats.Percentile95.Round(time.Microsecond*10))
	fmt.Printf("p99:                 %s\n", stats.Percentile99.Round(time.Microsecond*10))
	fmt.Println()
	fmt.Printf("Requests/sec:        %.2f\n", stats.RequestsPerSecond)
}
func round6(v float64) float64 {
	return math.Round(v*1e6) / 1e6
}

func (s Stats) ToJSON() JSONReport {
	return JSONReport{
		TotalRequests: s.TotalRequests,
		Successful:    s.SuccessCount,
		Failed:        s.ErrorCount,
		ErrorBreakdown: ErrorBreakdown{
			Timeouts:         s.TimeoutCount,
			ConnectionErrors: s.ConnectionErrors,
			ClientErrors:     s.ClientErrorCount,
			ServerErrors:     s.ServerErrorCount,
		},
		Latency: LatencyReport{
			Min: round6(s.MinLatency.Seconds()),
			Avg: round6(s.AverageLatency.Seconds()),
			Max: round6(s.MaxLatency.Seconds()),
			P50: round6(s.Percentile50.Seconds()),
			P95: round6(s.Percentile95.Seconds()),
			P99: round6(s.Percentile99.Seconds()),
		},
		RequestsPerSec: s.RequestsPerSecond,
	}
}

func PrintJSONReport(s Stats) error {
	report := s.ToJSON()
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}
