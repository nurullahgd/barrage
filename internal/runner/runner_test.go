package runner

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/goleak"
)

func TestRunLoadTest_NoGoroutineLeak(t *testing.T) {
	defer goleak.VerifyNone(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	httpClient := &http.Client{}
	resultCh := RunLoadTest(context.Background(), httpClient, server.URL, "GET", nil, 50, 5, 0)

	for range resultCh {
	}
}

func TestRunLoadTest_ContextCancellation(t *testing.T) {
	defer goleak.VerifyNone(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	httpClient := &http.Client{}

	resultCh := RunLoadTest(ctx, httpClient, server.URL, "GET", nil, 10000, 5, 0)

	count := 0
	for r := range resultCh {
		_ = r
		count++
		if count == 5 {
			cancel()
			break
		}
	}

	for range resultCh {
	}

	if count > 5 {
		t.Errorf("expected to stop at 5, got %d", count)
	}

	t.Logf("stopped after %d requests", count)
}
