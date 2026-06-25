// Package client provides an HTTP client for the Now & Again REST API.
//
// IMPORTANT: This client must stay in sync with the backend API.
// See COUPLING.md in the repository root for the contract between backend and CLI.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client wraps the HTTP client with auth and base URL.
type Client struct {
	BaseURL      string
	Token        string
	HTTPClient   *http.Client
	customHeader map[string]string
}

// New creates a new API client.
func New(baseURL, token string) *Client {
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Get performs an authenticated GET request.
func (c *Client) Get(path string) ([]byte, error) {
	return c.do("GET", path, nil)
}

// Post performs an authenticated POST request.
func (c *Client) Post(path string, body interface{}) ([]byte, error) {
	return c.do("POST", path, body)
}

// Patch performs an authenticated PATCH request.
func (c *Client) Patch(path string, body interface{}) ([]byte, error) {
	return c.do("PATCH", path, body)
}

// Delete performs an authenticated DELETE request.
func (c *Client) Delete(path string) ([]byte, error) {
	return c.do("DELETE", path, nil)
}

func (c *Client) do(method, path string, body interface{}) ([]byte, error) {
	url := c.BaseURL + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	for k, v := range c.customHeader {
		req.Header.Set(k, v)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
