package runner

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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
