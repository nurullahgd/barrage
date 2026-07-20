package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
