package myhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/115.0"

func getClient() *http.Client {
	return &http.Client{
		Timeout: time.Second * 30, // Timeout after 30 seconds
	}
}

func DoGetRequest(path, authHeader string) (data []byte, err error) {
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, path, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	req.Header.Set("User-Agent", userAgent)

	res, err := getClient().Do(req)
	if err != nil {
		slog.Error(err.Error())

		return nil, fmt.Errorf("error doing request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		slog.Warn("HTTP request not OK", "status", res.StatusCode)

		return nil, fmt.Errorf("error doing request %d %w", res.StatusCode, err)
	}

	data, err = io.ReadAll(res.Body)
	if err != nil {
		slog.Error(err.Error())
		err = fmt.Errorf("error reading body: %w", err)
	}

	return data, err
}

func DoPostRequest(path string, values url.Values, authHeader string) (data []byte, err error) {
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, path, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)

	res, err := getClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("error doing request: %w", err)
	}
	defer res.Body.Close()

	data, err = io.ReadAll(res.Body)
	if err != nil {
		slog.Error(err.Error())
		err = fmt.Errorf("error reading body: %w", err)
	} else if res.StatusCode != http.StatusOK {
		slog.Info("Post request", "path", path, "status", res.StatusCode, "payload", string(data))
	}

	return data, err
}

func DoJSONRequest(path, method string, values interface{}, authHeader string) ([]byte, error) {
	payload, err := json.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("error marshaling %w", err)
	}

	return DoJSONBytesRequest(path, method, payload, authHeader)
}

func DoJSONBytesRequest(path, method string, values []byte, authHeader string) ([]byte, error) {
	req, err := http.NewRequestWithContext(context.TODO(), method, path, bytes.NewBuffer(values))
	if err != nil {
		return nil, fmt.Errorf("error with new request %w", err)
	}

	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

	res, err := getClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("error doing request %w", err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error(err.Error())

		return nil, fmt.Errorf("error reading body %w", err)
	}

	return data, nil
}
