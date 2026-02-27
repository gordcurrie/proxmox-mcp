package proxmox

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListPools_success(t *testing.T) {
	t.Parallel()

	want := []Pool{
		{PoolID: "dev", Comment: "Development pool"},
		{PoolID: "prod", Comment: "Production pool"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pools" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListPools(context.Background())
	if err != nil {
		t.Fatalf("ListPools: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d pools, want %d", len(got), len(want))
	}
	if got[0].PoolID != "dev" || got[1].PoolID != "prod" {
		t.Errorf("poolids: got %q, %q; want dev, prod", got[0].PoolID, got[1].PoolID)
	}
}

func TestListPools_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).ListPools(context.Background())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetPool_success(t *testing.T) {
	t.Parallel()

	want := Pool{
		PoolID:  "dev",
		Comment: "Development pool",
		Members: []PoolMember{
			{ID: "qemu/100", Type: "qemu", VMID: 100, Node: "pve1"},
			{ID: "storage/pve1/local", Type: "storage", Node: "pve1", Storage: "local"},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pools" || r.URL.Query().Get("poolid") != "dev" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).GetPool(context.Background(), "dev")
	if err != nil {
		t.Fatalf("GetPool: %v", err)
	}
	if got.PoolID != "dev" {
		t.Errorf("poolid: got %q, want dev", got.PoolID)
	}
	if len(got.Members) != 2 {
		t.Fatalf("got %d members, want 2", len(got.Members))
	}
	if got.Members[0].VMID != 100 {
		t.Errorf("member[0].vmid: got %d, want 100", got.Members[0].VMID)
	}
}

func TestGetPool_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).GetPool(context.Background(), "no-such-pool")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreatePool_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "want POST", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/pools" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, nil))
	}))
	defer srv.Close()

	req := &CreatePoolRequest{PoolID: "test-pool", Comment: "A test pool"}
	if err := newTestClient(t, srv.URL).CreatePool(context.Background(), req); err != nil {
		t.Fatalf("CreatePool: %v", err)
	}
}

func TestCreatePool_apiError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "pool already exists", http.StatusInternalServerError)
	}))
	defer srv.Close()

	req := &CreatePoolRequest{PoolID: "existing-pool"}
	err := newTestClient(t, srv.URL).CreatePool(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpdatePool_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "want PUT", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/pools/dev" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, nil))
	}))
	defer srv.Close()

	req := &UpdatePoolRequest{Comment: "Updated comment", VMs: "100,101"}
	if err := newTestClient(t, srv.URL).UpdatePool(context.Background(), "dev", req); err != nil {
		t.Fatalf("UpdatePool: %v", err)
	}
}

func TestUpdatePool_apiError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "pool not found", http.StatusInternalServerError)
	}))
	defer srv.Close()

	req := &UpdatePoolRequest{VMs: "999"}
	err := newTestClient(t, srv.URL).UpdatePool(context.Background(), "no-such-pool", req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeletePool_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "want DELETE", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/pools/old-pool" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, nil))
	}))
	defer srv.Close()

	if err := newTestClient(t, srv.URL).DeletePool(context.Background(), "old-pool"); err != nil {
		t.Fatalf("DeletePool: %v", err)
	}
}

func TestDeletePool_apiError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "pool not found", http.StatusInternalServerError)
	}))
	defer srv.Close()

	err := newTestClient(t, srv.URL).DeletePool(context.Background(), "no-such-pool")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
