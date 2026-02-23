package proxmox

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const defaultTimeout = 30 * time.Second

// Client is a Proxmox VE API client that authenticates using API tokens.
// All methods require a context.Context as the first parameter.
type Client struct {
	baseURL    string
	httpClient *http.Client
	authHeader string // "PVEAPIToken=USER@REALM!TOKENID=UUID"
}

// NewClient creates a new Proxmox API client.
//
// apiURL must be the full base URL including the JSON API path prefix,
// e.g. "https://pve.example.com:8006/api2/json".
// tokenID is the full token identifier, e.g. "root@pam!mcp".
// tokenSecret is the UUID secret for the token.
// insecure, when true, disables TLS certificate verification — only use this
// when your Proxmox host uses a self-signed certificate and you understand the risks.
func NewClient(apiURL, tokenID, tokenSecret string, insecure bool) (*Client, error) {
	if apiURL == "" {
		return nil, fmt.Errorf("apiURL must not be empty")
	}
	if tokenID == "" {
		return nil, fmt.Errorf("tokenID must not be empty")
	}
	if tokenSecret == "" {
		return nil, fmt.Errorf("tokenSecret must not be empty")
	}

	transport := &http.Transport{}
	if insecure {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true, /* #nosec G402 */ //nolint:gosec // G402: only set when PROXMOX_INSECURE=true; user must explicitly opt in
		}
	}

	return &Client{
		baseURL: apiURL,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   defaultTimeout,
		},
		authHeader: "PVEAPIToken=" + tokenID + "=" + tokenSecret,
	}, nil
}

// Version calls GET /version as a lightweight health-check and returns the
// raw response map.
func (c *Client) Version(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := c.get(ctx, "/version", &result); err != nil {
		return nil, fmt.Errorf("getting version: %w", err)
	}
	return result, nil
}

// get performs an authenticated GET request to path (relative to baseURL),
// unwraps the Proxmox {"data": ...} envelope, and JSON-decodes the result
// into result (which must be a pointer).
func (c *Client) get(ctx context.Context, path string, result any) error {
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("closing get response body", "err", err)
		}
	}()

	return c.decode(resp.Body, result)
}

// post performs an authenticated POST request to path (relative to baseURL).
// The Proxmox {"data": ...} envelope is unwrapped and decoded into result.
// To POST with a request body, use postWithBody.
func (c *Client) post(ctx context.Context, path string, result any) error {
	resp, err := c.do(ctx, http.MethodPost, path, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("closing post response body", "err", err)
		}
	}()

	return c.decode(resp.Body, result)
}

// do constructs and executes an HTTP request, handling auth headers and error
// status codes. The caller is responsible for closing resp.Body.
func (c *Client) do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("creating request %s %s: %w", method, path, err)
	}

	req.Header.Set("Authorization", c.authHeader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req) /* #nosec G704 */ //nolint:gosec // G704: URL is constructed from PROXMOX_API_URL which the user must explicitly supply
	if err != nil {
		return nil, fmt.Errorf("executing request %s %s: %w", method, path, err)
	}

	if resp.StatusCode == http.StatusNotFound {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("closing not-found response body", "err", err)
		}
		return nil, ErrNotFound
	}

	if resp.StatusCode >= http.StatusBadRequest {
		rawBody, _ := io.ReadAll(resp.Body)
		if err := resp.Body.Close(); err != nil {
			slog.Warn("closing error response body", "err", err)
		}
		return nil, &APIError{StatusCode: resp.StatusCode, Body: string(rawBody)}
	}

	return resp, nil
}

// decode reads r, unwraps the Proxmox {"data": ...} envelope, and decodes the
// inner value into result (which must be a pointer, or nil to discard).
func (c *Client) decode(r io.Reader, result any) error {
	var envelope apiResponse
	if err := json.NewDecoder(r).Decode(&envelope); err != nil {
		return fmt.Errorf("decoding API envelope: %w", err)
	}
	if result == nil {
		return nil
	}
	if err := json.Unmarshal(envelope.Data, result); err != nil {
		return fmt.Errorf("decoding API data: %w", err)
	}
	return nil
}
