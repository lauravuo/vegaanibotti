package my_http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
		log.Println(err)

		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		slog.Warn("HTTP request not OK", "status", res.StatusCode)
		return nil, fmt.Errorf("Encountered error %d", res.StatusCode)
	}

	data, err = io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
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
		log.Println(err)
	}

	return
}

func DoJSONRequest(path, method string, values interface{}, authHeader string) ([]byte, error) {
	payload, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(context.TODO(), method, path, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	res, err := getClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
	}

	log.Println("JSON request response: ", string(data))

	return data, err
}
