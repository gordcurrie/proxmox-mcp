package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const initializeBody = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"0"}}}`

func TestNewHTTPHandlerCrossOrigin(t *testing.T) {
	handler := newHTTPHandler(mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil))

	tests := []struct {
		name       string
		headers    map[string]string
		wantStatus int
	}{
		{
			name:       "cross-site browser request rejected",
			headers:    map[string]string{"Sec-Fetch-Site": "cross-site"},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "mismatched origin rejected",
			headers:    map[string]string{"Origin": "https://evil.example"},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "same-origin browser request allowed",
			headers:    map[string]string{"Sec-Fetch-Site": "same-origin"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "non-browser request allowed",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "http://localhost:8080/", strings.NewReader(initializeBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json, text/event-stream")
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestValidateTokenSecret(t *testing.T) {
	const valid = "0f8b6c1e-2a4d-4e7f-9b3c-5d6e7f8a9b0c"

	tests := []struct {
		name    string
		secret  string
		wantErr bool
	}{
		{name: "canonical lowercase UUID", secret: valid},
		{name: "token ID instead of secret", secret: "root@pam!mcp", wantErr: true}, //nolint:gosec // G101: fake token ID placeholder, not a real credential
		{name: "truncated", secret: valid[:len(valid)-1], wantErr: true},
		{name: "trailing newline", secret: valid + "\n", wantErr: true},
		{name: "uppercase", secret: strings.ToUpper(valid), wantErr: true},
		{name: "braces", secret: "{" + valid + "}", wantErr: true},
		{name: "urn prefix", secret: "urn:uuid:" + valid, wantErr: true},
		{name: "no hyphens", secret: strings.ReplaceAll(valid, "-", ""), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTokenSecret(tt.secret)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateTokenSecret() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.secret != "" && strings.Contains(err.Error(), tt.secret) {
				t.Errorf("error message leaks the secret: %v", err)
			}
		})
	}
}
