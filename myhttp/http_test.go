package myhttp

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestDoGetRequest(t *testing.T) {
	// Test successful request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	data, err := DoGetRequest(ts.URL, "Bearer token")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if string(data) != "ok" {
		t.Errorf("Expected 'ok', got %s", string(data))
	}

	// Test bad URL (triggers http.NewRequest error)
	_, err = DoGetRequest("://invalid-url", "")
	if err == nil {
		t.Error("Expected error for invalid url, got nil")
	}

	// Test bad status code
	tsBadStatus := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer tsBadStatus.Close()
	_, err = DoGetRequest(tsBadStatus.URL, "")
	if err == nil {
		t.Error("Expected error for bad status, got nil")
	}

	// Test read error (Content-Length expects more than received)
	tsReadErr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		w.Write([]byte("short"))
	}))
	defer tsReadErr.Close()
	_, err = DoGetRequest(tsReadErr.URL, "")
	if err == nil {
		t.Error("Expected error for read missing bytes, got nil")
	}

	// Test connection error
	ts.Close()
	_, err = DoGetRequest(ts.URL, "")
	if err == nil {
		t.Error("Expected connection error, got nil")
	}
}

func TestDoPostRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	values := url.Values{}
	values.Add("key", "value")

	data, err := DoPostRequest(ts.URL, values, "Bearer token")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if string(data) != "ok" {
		t.Errorf("Expected 'ok', got %s", string(data))
	}

	// Test bad URL
	_, err = DoPostRequest("://invalid-url", values, "")
	if err == nil {
		t.Error("Expected error for invalid url, got nil")
	}

	// Test bad status
	tsBadStatus := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer tsBadStatus.Close()
	_, err = DoPostRequest(tsBadStatus.URL, values, "")
	if err != nil {
		// DoPostRequest only logs non-OK, it doesn't return error based on status alone
		t.Error("Did not expect post request to error on bad status")
	}

	// Test read error
	tsReadErr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		w.Write([]byte("short"))
	}))
	defer tsReadErr.Close()
	_, err = DoPostRequest(tsReadErr.URL, values, "")
	if err == nil {
		t.Error("Expected error for read missing bytes, got nil")
	}

	// Test connection error
	ts.Close()
	_, err = DoPostRequest(ts.URL, values, "")
	if err == nil {
		t.Error("Expected connection error, got nil")
	}
}

func TestDoJSONRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	data, err := DoJSONRequest(ts.URL, http.MethodPost, map[string]string{"key": "value"}, "Bearer token")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if string(data) != "ok" {
		t.Errorf("Expected 'ok', got %s", string(data))
	}

	// Test JSON marshal error
	_, err = DoJSONRequest(ts.URL, http.MethodPost, make(chan int), "")
	if err == nil {
		t.Error("Expected marshal error, got nil")
	}

	// Test bad URL
	_, err = DoJSONBytesRequest("://invalid-url", http.MethodPost, []byte("{}"), "")
	if err == nil {
		t.Error("Expected error for invalid url, got nil")
	}

	// Test read error
	tsReadErr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		w.Write([]byte("short"))
	}))
	defer tsReadErr.Close()
	_, err = DoJSONBytesRequest(tsReadErr.URL, http.MethodPost, []byte("{}"), "")
	if err == nil {
		t.Error("Expected error for read missing bytes, got nil")
	}

	// Test connection error
	ts.Close()
	_, err = DoJSONBytesRequest(ts.URL, http.MethodPost, []byte("{}"), "")
	if err == nil {
		t.Error("Expected connection error, got nil")
	}
}

func init() {
	// Silence slog during tests
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
}
