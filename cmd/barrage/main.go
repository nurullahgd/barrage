package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/nurullahgd/barrage.git/internal/client"
	"github.com/nurullahgd/barrage.git/internal/runner"
	"github.com/nurullahgd/barrage.git/internal/stats"
)

func main() {
	var url, method, bodyStr string
	var totalRequests, workerCount int
	var rate int
	flag.StringVar(&url, "url", "", "target URL (required)")
	flag.StringVar(&method, "method", "GET", "http method")
	flag.StringVar(&bodyStr, "body", "", "request body (raw JSON)")
	flag.IntVar(&totalRequests, "requests", 20, "total requests")
	flag.IntVar(&workerCount, "workers", 5, "worker count")
	flag.IntVar(&rate, "rate", 0, "requests per second")
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
	httpClient := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	start := time.Now()
	results := runner.RunLoadTest(ctx, httpClient, url, method, bodyBytes, totalRequests, workerCount, rate)

	shown := 0
	for _, r := range results {
		if r.Category == client.CategoryConnection && shown < 5 {
			fmt.Println("sample error:", r.Error)
			shown++
		}
	}

	s := stats.ComputeStats(results, time.Since(start))
	stats.PrintReport(s)
}
