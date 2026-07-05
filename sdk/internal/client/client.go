// Package client provides HTTP API clients for interacting with the Now & Again server.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient wraps the standard http.Client with base URL and auth token.
type HTTPClient struct {
	BaseURL    string
	Token      string
	FamilyID   string // set for family-scoped requests
	AuthMethod string // "jwt" or "apikey", auto-detected
	HTTPClient *http.Client
}

// NewHTTPClient creates a new HTTP client with sensible defaults.
func NewHTTPClient(baseURL, token string) *HTTPClient {
	return &HTTPClient{
		BaseURL:    baseURL,
		Token:      token,
		AuthMethod: detectAuthMethod(token),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken updates the auth token and re-detects the auth method.
func (c *HTTPClient) SetToken(token string) {
	c.Token = token
	c.AuthMethod = detectAuthMethod(token)
}

// SetFamilyID sets the X-Family-Id header value for family-scoped requests.
func (c *HTTPClient) SetFamilyID(familyID string) {
	c.FamilyID = familyID
}

// detectAuthMethod returns "apikey" if token starts with "na_", otherwise "jwt".
func detectAuthMethod(token string) string {
	if len(token) > 3 && token[:3] == "na_" {
		return "apikey"
	}
	return "jwt"
}

// SetBaseURL updates the base URL.
func (c *HTTPClient) SetBaseURL(url string) {
	c.BaseURL = url
}

// Do performs an HTTP request, sending X-Family-Id if set, and unmarshals the JSON response.
// This is the public version for SDK-level callers that need direct API access.
func (c *HTTPClient) Do(method, path string, body, result interface{}) error {
	return c.do(method, path, body, result)
}

// do performs an HTTP request and unmarshals the JSON response.
func (c *HTTPClient) do(method, path string, body, result interface{}) error {
	url := c.BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.Token != "" {
		if c.AuthMethod == "apikey" {
			req.Header.Set("X-API-Key", c.Token)
		} else {
			req.Header.Set("Authorization", "Bearer "+c.Token)
		}
	}
	if c.FamilyID != "" {
		req.Header.Set("X-Family-Id", c.FamilyID)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	// API response envelope: {"success":true,"data":{...}} or {"success":false,"error":"..."}
	var envelope struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
		Error   string          `json:"error,omitempty"`
	}
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return fmt.Errorf("parse response: %w (body: %s)", err, string(respBody))
	}

	if !envelope.Success {
		return fmt.Errorf("api error: %s", envelope.Error)
	}

	if result != nil && envelope.Data != nil {
		if err := json.Unmarshal(envelope.Data, result); err != nil {
			return fmt.Errorf("unmarshal data: %w", err)
		}
	}

	return nil
}

// ─── Aggregate ────────────────────────────────────────────────────

// AllClients bundles all domain clients. Passed to CLI command handlers.
type AllClients struct {
	User   *UserClient
	Family *FamilyClient
	ApiKey *ApiKeyClient
	Task   *TaskClient
}

// NewAllClients creates all clients from a single HTTPClient.
func NewAllClients(httpClient *HTTPClient) *AllClients {
	return &AllClients{
		User:   NewUserClient(httpClient),
		Family: NewFamilyClient(httpClient),
		ApiKey: NewApiKeyClient(httpClient),
		Task:   NewTaskClient(httpClient),
	}
}
