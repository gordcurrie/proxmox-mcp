package proxmox

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListUsers_success(t *testing.T) {
	t.Parallel()

	want := []map[string]any{
		{"userid": "root@pam", "enable": float64(1)},
		{"userid": "gord@pve", "enable": float64(1)},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/access/users" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListUsers(context.Background())
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d users, want 2", len(got))
	}
	if got[0]["userid"] != "root@pam" {
		t.Errorf("user[0]: got %v, want root@pam", got[0]["userid"])
	}
}

func TestListUsers_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).ListUsers(context.Background())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestListUserTokens_success(t *testing.T) {
	t.Parallel()

	want := []map[string]any{
		{"tokenid": "mcp", "comment": "MCP server token"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/access/users/root@pam/token" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListUserTokens(context.Background(), "root@pam")
	if err != nil {
		t.Fatalf("ListUserTokens: %v", err)
	}
	if len(got) != 1 || got[0]["tokenid"] != "mcp" {
		t.Errorf("got %+v, want 1 token with tokenid mcp", got)
	}
}

func TestListUserTokens_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).ListUserTokens(context.Background(), "nobody@pam")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
