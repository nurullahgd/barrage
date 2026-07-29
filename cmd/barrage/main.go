package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nurullahgd/barrage.git/internal/client"
	"github.com/nurullahgd/barrage.git/internal/runner"
	"github.com/nurullahgd/barrage.git/internal/stats"
	"github.com/nurullahgd/barrage.git/internal/ui"
	"golang.org/x/net/http2"
)

func main() {
	var url, method, bodyStr, format string
	var totalRequests, workerCount, rate int
	var duration time.Duration
	var useHTTP2 bool
	flag.StringVar(&url, "url", "", "target URL (required)")
	flag.StringVar(&method, "method", "GET", "http method")
	flag.StringVar(&bodyStr, "body", "", "request body (raw JSON)")
	flag.IntVar(&totalRequests, "requests", 20, "total requests")
	flag.IntVar(&workerCount, "workers", 5, "worker count")
	flag.IntVar(&rate, "rate", 0, "requests per second")
	flag.DurationVar(&duration, "duration", 0, "test duration (e.g. 30s, 2m); 0 = run until requests are exhausted")
	flag.StringVar(&format, "format", "dashboard", "output format: dashboard, text, json")
	flag.BoolVar(&useHTTP2, "http2", false, "enable HTTP/2 (h2c for cleartext)")
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
	if useHTTP2 {
		h2Transport := &http2.Transport{
			AllowHTTP: true, // h2c için şart
			DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, network, addr)
			},
		}
		httpClient = &http.Client{
			Timeout:   10 * time.Second,
			Transport: h2Transport,
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if duration > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, duration)
		defer cancel()
	}

	start := time.Now()
	resultCh := runner.RunLoadTest(ctx, httpClient, url, method, bodyBytes, totalRequests, workerCount, rate)

	var allResults []client.Result

	switch format {
	case "dashboard":
		p := tea.NewProgram(ui.NewModel(url, workerCount, rate, totalRequests))
		go func() {
			for r := range resultCh {
				allResults = append(allResults, r)
				p.Send(ui.ResultMsg(r))
			}
			p.Send(ui.DoneMsg{})
		}()
		if _, err := p.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "dashboard error:", err)
			os.Exit(1)
		}

	default: // text ve json için
		shown := 0
		for r := range resultCh {
			allResults = append(allResults, r)
			if r.Category == client.CategoryConnection && shown < 5 {
				fmt.Println("sample error:", r.Error)
				shown++
			}
		}
	}

	s := stats.ComputeStats(allResults, time.Since(start))
	switch format {
	case "json":
		if err := stats.PrintJSONReport(s); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "dashboard":
		stats.PrintReport(s) // dashboard bittikten sonra özet raporu bas
	default:
		stats.PrintReport(s)
	}
}
