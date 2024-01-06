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

func getClient() *http.Client {
	return &http.Client{
		Timeout: time.Second * 30, // Timeout after 30 seconds
	}
}

func DoGetRequest(path, authHeader string) (data []byte, err error) {
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, path, http.NoBody)
	if err != nil {
		return
	}

	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	res, err := getClient().Do(req)
	if err != nil {
		slog.Error(err.Error())

		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		slog.Warn("HTTP request not OK", "status", res.StatusCode)

		return nil, fmt.Errorf("error doing request %d %w", res.StatusCode, err)
	}

	data, err = io.ReadAll(res.Body)
	if err != nil {
		slog.Error(err.Error())
	}

	return
}

func DoPostRequest(path string, values url.Values, authHeader string) (data []byte, err error) {
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, path, strings.NewReader(values.Encode()))
	if err != nil {
		return
	}

	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := getClient().Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	data, err = io.ReadAll(res.Body)
	if err != nil {
		slog.Error(err.Error())
	} else {
		if res.StatusCode != http.StatusOK {
			slog.Info("Post request", "path", path, "status", res.StatusCode, "payload", string(data))
		}
	}

	return
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
