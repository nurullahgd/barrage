package main

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"sync"
	"time"
)

func main() {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	url := "http://localhost:8080/v1/markets"
	method := "GET"
	totalRequests := 20
	workerCount := 5
	start := time.Now()
	results := runLoadTest(client, url, method, nil, totalRequests, workerCount)
	fmt.Println(results)
	stats := ComputeStats(results, time.Since(start))
	PrintReport(stats)
}

type Result struct {
	Error      error         `json:"error"`
	Latency    time.Duration `json:"latency"`
	StatusCode int           `json:"status_code"`
}

func doRequest(client *http.Client, url string, method string, body any) Result {
	start := time.Now()
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return Result{
				Error:      err,
				Latency:    time.Since(start),
				StatusCode: 0,
			}
		}
		bodyReader = bytes.NewReader(bodyBytes)
	} else {
		bodyReader = nil
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return Result{
			Error:      err,
			Latency:    time.Since(start),
			StatusCode: 0,
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return Result{
			Error:      err,
			Latency:    time.Since(start),
			StatusCode: 0,
		}
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("error closing response body:", err)
		}
	}()
	return Result{
		Error:      nil,
		Latency:    time.Since(start),
		StatusCode: resp.StatusCode,
	}
}

func runLoadTest(client *http.Client, url string, method string, body any, totalRequests, workerCount int) []Result {
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
	latencies := make([]time.Duration, 0, totalRequests)
	var sumLatency time.Duration

	for _, result := range results {
		if result.Error != nil {
			errorCount++
		} else {
			successCount++
		}
		latencies = append(latencies, result.Latency)
		sumLatency += result.Latency
	}
	slices.Sort(latencies)
	averageLatency := sumLatency / time.Duration(totalRequests)
	requestsPerSecond := float64(totalRequests) / elapsed.Seconds()
	return Stats{
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
