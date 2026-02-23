package proxmox

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestClient creates a Client pointed at the provided test server URL.
func newTestClient(t *testing.T, serverURL string) *Client {
	t.Helper()
	c, err := NewClient(serverURL, "root@pam!test", "test-token-secret", false)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

// jsonEnvelope wraps v in the standard Proxmox {"data": ...} envelope.
func jsonEnvelope(t *testing.T, v any) []byte {
	t.Helper()
	inner, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal inner: %v", err)
	}
	outer, err := json.Marshal(map[string]json.RawMessage{"data": inner})
	if err != nil {
		t.Fatalf("json.Marshal outer: %v", err)
	}
	return outer
}

func TestNewClient_validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		apiURL      string
		tokenID     string
		tokenSecret string
		wantErr     bool
	}{
		{ //nolint:gosec // G101: test-token-secret is a fake placeholder, not a real credential
			name:        "all valid",
			apiURL:      "https://pve:8006/api2/json",
			tokenID:     "root@pam!mcp",
			tokenSecret: "test-token-secret",
			wantErr:     false,
		},
		{ //nolint:gosec // G101: test-token-secret is a fake placeholder, not a real credential
			name:        "empty apiURL",
			apiURL:      "",
			tokenID:     "root@pam!mcp",
			tokenSecret: "test-token-secret",
			wantErr:     true,
		},
		{ //nolint:gosec // G101: test-token-secret is a fake placeholder, not a real credential
			name:        "empty tokenID",
			apiURL:      "https://pve:8006/api2/json",
			tokenID:     "",
			tokenSecret: "test-token-secret",
			wantErr:     true,
		},
		{ //nolint:gosec // G101: tokenSecret is intentionally empty — this case validates that an empty secret is rejected
			name:        "empty tokenSecret",
			apiURL:      "https://pve:8006/api2/json",
			tokenID:     "root@pam!mcp",
			tokenSecret: "",
			wantErr:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewClient(tc.apiURL, tc.tokenID, tc.tokenSecret, false)
			if (err != nil) != tc.wantErr {
				t.Errorf("NewClient() error = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}

func TestClient_get_success(t *testing.T) {
	t.Parallel()

	want := []Node{{Node: "pve", Status: "online"}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header is present and correct format.
		auth := r.Header.Get("Authorization")
		if auth != "PVEAPIToken=root@pam!test=test-token-secret" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	var got []Node
	if err := c.get(context.Background(), "/nodes", &got); err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got) != 1 || got[0].Node != "pve" {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestClient_get_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.NotFound(w, nil)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	var got map[string]any
	err := c.get(context.Background(), "/nonexistent", &got)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestClient_get_apiError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "permission denied", http.StatusForbidden)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	var got map[string]any
	err := c.get(context.Background(), "/nodes", &got)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", apiErr.StatusCode)
	}
}

func TestClient_post_returnsUPID(t *testing.T) {
	t.Parallel()

	const upid = "UPID:pve:000015E3:00000000:60F4B3A7:qmstart:100:root@pam:"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, upid))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	var got string
	if err := c.post(context.Background(), "/nodes/pve/qemu/100/status/start", &got); err != nil {
		t.Fatalf("post: %v", err)
	}
	if got != upid {
		t.Errorf("got UPID %q, want %q", got, upid)
	}
}

func TestClient_contextCancellation(t *testing.T) {
	t.Parallel()

	// Server that never responds until the request is cancelled.
	started := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	c := newTestClient(t, srv.URL)

	done := make(chan error, 1)
	go func() {
		var got map[string]any
		done <- c.get(ctx, "/version", &got)
	}()

	// Wait for server to receive the request, then cancel.
	<-started
	cancel()

	if err := <-done; err == nil {
		t.Fatal("expected error after context cancellation, got nil")
	}
}
