package myhttp_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/lauravuo/vegaanibotti/myhttp"
)

const successResp = "ok"

func TestMain(m *testing.M) {
	// Silence slog during tests
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	os.Exit(m.Run())
}

func TestDoGetRequest(t *testing.T) {
	t.Parallel()

	// Test successful request
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(successResp))
	}))
	defer testServer.Close()

	data, err := myhttp.DoGetRequest(testServer.URL, "Bearer token")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if string(data) != successResp {
		t.Errorf("Expected '%s', got %s", successResp, string(data))
	}

	// Test bad URL (triggers http.NewRequest error)
	_, err = myhttp.DoGetRequest("://invalid-url", "")
	if err == nil {
		t.Error("Expected error for invalid url, got nil")
	}

	// Test bad status code
	badStatusServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer badStatusServer.Close()

	_, err = myhttp.DoGetRequest(badStatusServer.URL, "")
	if err == nil {
		t.Error("Expected error for bad status, got nil")
	}

	// Test read error (Content-Length expects more than received)
	readErrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "100")
		_, _ = w.Write([]byte("short"))
	}))
	defer readErrServer.Close()

	_, err = myhttp.DoGetRequest(readErrServer.URL, "")
	if err == nil {
		t.Error("Expected error for read missing bytes, got nil")
	}

	// Test connection error
	testServer.Close()

	_, err = myhttp.DoGetRequest(testServer.URL, "")
	if err == nil {
		t.Error("Expected connection error, got nil")
	}
}

func TestDoPostRequest(t *testing.T) {
	t.Parallel()

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(successResp))
	}))
	defer testServer.Close()

	values := url.Values{}
	values.Add("key", "value")

	data, err := myhttp.DoPostRequest(testServer.URL, values, "Bearer token")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if string(data) != successResp {
		t.Errorf("Expected '%s', got %s", successResp, string(data))
	}

	// Test bad URL
	_, err = myhttp.DoPostRequest("://invalid-url", values, "")
	if err == nil {
		t.Error("Expected error for invalid url, got nil")
	}

	// Test bad status
	badStatusServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer badStatusServer.Close()

	_, err = myhttp.DoPostRequest(badStatusServer.URL, values, "")
	if err != nil {
		// DoPostRequest only logs non-OK, it doesn't return error based on status alone
		t.Error("Did not expect post request to error on bad status")
	}

	// Test read error
	readErrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "100")
		_, _ = w.Write([]byte("short"))
	}))
	defer readErrServer.Close()

	_, err = myhttp.DoPostRequest(readErrServer.URL, values, "")
	if err == nil {
		t.Error("Expected error for read missing bytes, got nil")
	}

	// Test connection error
	testServer.Close()

	_, err = myhttp.DoPostRequest(testServer.URL, values, "")
	if err == nil {
		t.Error("Expected connection error, got nil")
	}
}

func TestDoJSONRequest(t *testing.T) {
	t.Parallel()

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(successResp))
	}))
	defer testServer.Close()

	data, err := myhttp.DoJSONRequest(testServer.URL, http.MethodPost, map[string]string{"key": "value"}, "Bearer token")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if string(data) != successResp {
		t.Errorf("Expected '%s', got %s", successResp, string(data))
	}

	// Test JSON marshal error
	_, err = myhttp.DoJSONRequest(testServer.URL, http.MethodPost, make(chan int), "")
	if err == nil {
		t.Error("Expected marshal error, got nil")
	}

	// Test bad URL
	_, err = myhttp.DoJSONBytesRequest("://invalid-url", http.MethodPost, []byte("{}"), "")
	if err == nil {
		t.Error("Expected error for invalid url, got nil")
	}

	// Test read error
	readErrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "100")
		_, _ = w.Write([]byte("short"))
	}))
	defer readErrServer.Close()

	_, err = myhttp.DoJSONBytesRequest(readErrServer.URL, http.MethodPost, []byte("{}"), "")
	if err == nil {
		t.Error("Expected error for read missing bytes, got nil")
	}

	// Test connection error
	testServer.Close()

	_, err = myhttp.DoJSONBytesRequest(testServer.URL, http.MethodPost, []byte("{}"), "")
	if err == nil {
		t.Error("Expected connection error, got nil")
	}
}
