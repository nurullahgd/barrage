package main

import (
	"bytes"
	"cmp"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"slices"
	"sync"
	"time"
)

func main() {
	var url, method, bodyStr string
	var totalRequests, workerCount int

	flag.StringVar(&url, "url", "", "target URL (required)")
	flag.StringVar(&method, "method", "GET", "http method")
	flag.StringVar(&bodyStr, "body", "", "request body (raw JSON)")
	flag.IntVar(&totalRequests, "requests", 20, "total requests")
	flag.IntVar(&workerCount, "workers", 5, "worker count")
	flag.Parse()

	if url == "" {
		fmt.Println("url is required")
		os.Exit(1)
	}

	var bodyBytes []byte
	if bodyStr != "" {
		bodyBytes = []byte(bodyStr)
	}

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: workerCount, 
		IdleConnTimeout:     30 * time.Second,
	}
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}
	start := time.Now()
	results := runLoadTest(client, url, method, bodyBytes, totalRequests, workerCount)
	stats := ComputeStats(results, time.Since(start))
	shown := 0
	for _, r := range results {
		if r.Category == CategoryConnection && shown < 5 {
			fmt.Println("sample error:", r.Error)
			shown++
		}
	}
	PrintReport(stats)
}

type ErrorCategory int

const (
	CategoryNone ErrorCategory = iota
	CategoryTimeout
	CategoryConnection
	CategoryClientError
	CategoryServerError
)

type Result struct {
	Error      error         `json:"error"`
	Category   ErrorCategory `json:"category"`
	Latency    time.Duration `json:"latency"`
	StatusCode int           `json:"status_code"`
}

func doRequest(client *http.Client, url string, method string, body []byte) Result {
	start := time.Now()
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return Result{Error: err, Latency: time.Since(start), Category: CategoryConnection}
	}
	var netErr net.Error
	resp, err := client.Do(req)
	if err != nil {
		category := CategoryConnection
		if errors.As(err, &netErr) {
			if netErr.Timeout() {
				category = CategoryTimeout
			}
		}
		return Result{Error: err, Latency: time.Since(start), Category: category}
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("error closing response body:", err)
		}
	}()
	_, _ = io.Copy(io.Discard, resp.Body)

	category := CategoryNone
	switch {
	case resp.StatusCode >= 500:
		category = CategoryServerError
	case resp.StatusCode >= 400:
		category = CategoryClientError
	}
	return Result{Latency: time.Since(start), StatusCode: resp.StatusCode, Category: category}
}

func runLoadTest(client *http.Client, url string, method string, body []byte, totalRequests, workerCount int) []Result {
	start := time.Now()
	jobs := make(chan struct{}, totalRequests)
	results := make(chan Result, totalRequests)
	var wg sync.WaitGroup

	for i := 0; i < totalRequests; i++ {
		jobs <- struct{}{}
	}
	close(jobs)
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				results <- doRequest(client, url, method, body)
			}
		}()
	}
	go func() {
		defer close(results)
		wg.Wait()
	}()
	resultSlice := make([]Result, 0, totalRequests)
	for r := range results {
		resultSlice = append(resultSlice, r)
	}
	elapsed := time.Since(start)
	fmt.Printf("Load test completed in %s\n", elapsed)
	return resultSlice
}

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

func Percentile[T cmp.Ordered](sorted []T, p float64) T {
	n := len(sorted)
	if n == 0 {
		var zero T
		return zero
	}
	index := int(float64(n-1) * p / 100)
	return sorted[index]
}

func ComputeStats(results []Result, elapsed time.Duration) Stats {
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
			switch result.Category {
			case CategoryTimeout:
				timeoutCount++
			case CategoryConnection:
				connectionErrors++
			case CategoryClientError:
				clientErrorCount++
			case CategoryServerError:
				serverErrorCount++
			}
			errorCount++
		}
		latencies = append(latencies, result.Latency)
		sumLatency += result.Latency
	}
	slices.Sort(latencies)
	averageLatency := sumLatency / time.Duration(totalRequests)
	requestsPerSecond := float64(totalRequests) / elapsed.Seconds()
	return Stats{
		TimeoutCount:      timeoutCount,
		ConnectionErrors:  connectionErrors,
		ClientErrorCount:  clientErrorCount,
		ServerErrorCount:  serverErrorCount,
		TotalRequests:     totalRequests,
		SuccessCount:      successCount,
		ErrorCount:        errorCount,
		MinLatency:        latencies[0],
		MaxLatency:        latencies[totalRequests-1],
		AverageLatency:    averageLatency,
		Percentile50:      Percentile(latencies, 50),
		Percentile95:      Percentile(latencies, 95),
		Percentile99:      Percentile(latencies, 99),
		RequestsPerSecond: requestsPerSecond,
	}
}

func PrintReport(stats Stats) {
	successRate := float64(stats.SuccessCount) / float64(stats.TotalRequests) * 100
	errorRate := float64(stats.ErrorCount) / float64(stats.TotalRequests) * 100

	fmt.Println("  Breakdown:")
	fmt.Printf("    Timeouts:           %d\n", stats.TimeoutCount)
	fmt.Printf("    Connection errors:  %d\n", stats.ConnectionErrors)
	fmt.Printf("    4xx responses:      %d\n", stats.ClientErrorCount)
	fmt.Printf("    5xx responses:      %d\n", stats.ServerErrorCount)
	fmt.Printf("Total requests:      %d\n", stats.TotalRequests)
	fmt.Printf("Successful:          %d (%.1f%%)\n", stats.SuccessCount, successRate)
	fmt.Printf("Failed:              %d (%.1f%%)\n", stats.ErrorCount, errorRate)
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

func isSuccess(r Result) bool {
	return r.Error == nil && r.StatusCode >= 200 && r.StatusCode < 300
}
