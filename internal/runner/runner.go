package runner

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/nurullahgd/barrage.git/internal/client"
)

func RunLoadTest(ctx context.Context, httpClient *http.Client, url string, method string, body []byte, totalRequests, workerCount, rate int) <-chan client.Result {
	jobs := make(chan struct{}, totalRequests)
	results := make(chan client.Result, workerCount)
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				results <- client.DoRequest(httpClient, url, method, body)
			}
		}()
	}

	go func() {
		defer close(jobs)
		if rate <= 0 {
			for i := 0; i < totalRequests; i++ {
				select {
				case <-ctx.Done():
					return
				case jobs <- struct{}{}:
				}
			}
			return
		}
		interval := time.Second / time.Duration(rate)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for i := 0; i < totalRequests; i++ {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
			select {
			case <-ctx.Done():
				return
			case jobs <- struct{}{}:
			}
		}
	}()

	go func() {
		defer close(results)
		wg.Wait()
	}()

	return results
}
