package proxmox

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListNodes_success(t *testing.T) {
	t.Parallel()

	want := []Node{
		{Node: "pve1", Status: "online", MaxCPU: 8, MaxMem: 16 * 1024 * 1024 * 1024},
		{Node: "pve2", Status: "online", MaxCPU: 4, MaxMem: 8 * 1024 * 1024 * 1024},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListNodes(context.Background())
	if err != nil {
		t.Fatalf("ListNodes: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d nodes, want %d", len(got), len(want))
	}
	for i, n := range got {
		if n.Node != want[i].Node || n.Status != want[i].Status {
			t.Errorf("node[%d]: got %+v, want %+v", i, n, want[i])
		}
	}
}

func TestListNodes_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).ListNodes(context.Background())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetNodeStatus_success(t *testing.T) {
	t.Parallel()

	want := map[string]any{"cpuinfo": map[string]any{"cpus": float64(8)}, "uptime": float64(12345)}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/pve1/status" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).GetNodeStatus(context.Background(), "pve1")
	if err != nil {
		t.Fatalf("GetNodeStatus: %v", err)
	}
	if got["uptime"] != float64(12345) {
		t.Errorf("uptime: got %v, want 12345", got["uptime"])
	}
}

func TestGetNodeStatus_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).GetNodeStatus(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestListNodeStorage_success(t *testing.T) {
	t.Parallel()

	want := []map[string]any{
		{"storage": "local", "type": "dir", "active": float64(1)},
		{"storage": "Storage", "type": "btrfs", "active": float64(1)},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/pve1/storage" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListNodeStorage(context.Background(), "pve1")
	if err != nil {
		t.Fatalf("ListNodeStorage: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d entries, want 2", len(got))
	}
	if got[0]["storage"] != "local" {
		t.Errorf("storage[0]: got %v, want local", got[0]["storage"])
	}
}

func TestListNodeStorage_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).ListNodeStorage(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestListNodeTasks_success(t *testing.T) {
	t.Parallel()

	want := []map[string]any{
		{"upid": "UPID:pve1:001:vzstart:100:root@pam:", "type": "vzstart", "status": "OK"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/pve1/tasks" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListNodeTasks(context.Background(), "pve1", 0)
	if err != nil {
		t.Fatalf("ListNodeTasks: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d tasks, want 1", len(got))
	}
}

func TestListNodeTasks_withLimit(t *testing.T) {
	t.Parallel()

	want := []map[string]any{
		{"upid": "UPID:pve1:001:vzstart:100:root@pam:", "type": "vzstart"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/pve1/tasks" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("limit") != "10" {
			http.Error(w, "missing limit param", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListNodeTasks(context.Background(), "pve1", 10)
	if err != nil {
		t.Fatalf("ListNodeTasks with limit: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d tasks, want 1", len(got))
	}
}

func TestListNodeTasks_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).ListNodeTasks(context.Background(), "missing", 0)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetNodeDisks_success(t *testing.T) {
	t.Parallel()

	want := []map[string]any{
		{"devpath": "/dev/sda", "size": float64(500107862016), "type": "hdd"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/pve1/disks/list" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).GetNodeDisks(context.Background(), "pve1")
	if err != nil {
		t.Fatalf("GetNodeDisks: %v", err)
	}
	if len(got) != 1 || got[0]["devpath"] != "/dev/sda" {
		t.Errorf("got %+v, want 1 disk with devpath /dev/sda", got)
	}
}

func TestGetNodeDisks_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).GetNodeDisks(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
