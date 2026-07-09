package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	results := runLoadTest(client, url, method, nil, totalRequests, workerCount)
	fmt.Println(results)
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
	return resultSlice
}
