package client

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDoRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	httpClient := &http.Client{}
	result := DoRequest(httpClient, server.URL, "GET", nil)
	if result.Error != nil {
		t.Errorf("DoRequest returned error: %v", result.Error)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("DoRequest returned status code: %d", result.StatusCode)
	}
	if result.Category != CategoryNone {
		t.Errorf("DoRequest returned category: %d", result.Category)
	}

}

func TestDoRequest_ServerError(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	httpClient := &http.Client{}
	result := DoRequest(httpClient, server.URL, "GET", nil)
	if result.Error != nil {
		t.Errorf("DoRequest returned error: %v", result.Error)
	}
	if result.StatusCode != http.StatusInternalServerError {
		t.Errorf("DoRequest returned status code: %d", result.StatusCode)
	}
	if result.Category != CategoryServerError {
		t.Errorf("DoRequest returned category: %d", result.Category)
	}
}

func TestDoRequest_ClientError(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	httpClient := &http.Client{}
	result := DoRequest(httpClient, server.URL, "GET", nil)
	if result.Error != nil {
		t.Errorf("DoRequest returned error: %v", result.Error)
	}
	if result.StatusCode != http.StatusNotFound {
		t.Errorf("DoRequest returned status code: %d", result.StatusCode)
	}
	if result.Category != CategoryClientError {
		t.Errorf("DoRequest returned category: %d", result.Category)
	}
}

func TestDoRequest_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
	}))
	defer server.Close()

	httpClient := &http.Client{Timeout: 50 * time.Millisecond}
	result := DoRequest(httpClient, server.URL, "GET", nil)

	if result.Error == nil {
		t.Error("expected timeout error, got nil")
	}
	if result.Category != CategoryTimeout {
		t.Errorf("Category = %v, want CategoryTimeout", result.Category)
	}
}

func TestDoRequest_WithBody(t *testing.T) {
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	body := []byte(`{"symbol":"BTC"}`)
	httpClient := &http.Client{}
	result := DoRequest(httpClient, server.URL, "POST", body)

	if result.Error != nil {
		t.Errorf("unexpected error: %v", result.Error)
	}
	if string(receivedBody) != string(body) {
		t.Errorf("body = %s, want %s", receivedBody, body)
	}
}
