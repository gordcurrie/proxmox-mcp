package proxmox

import (
	"context"
	"encoding/json"
	"errors"
	"io"
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

func TestListNodeTasks_negativeLimitError(t *testing.T) {
	t.Parallel()

	// The error is returned before any HTTP call; the server is never contacted.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.ListNodeTasks(context.Background(), "pve1", -1)
	if err == nil {
		t.Fatal("expected error for negative limit, got nil")
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

func TestGetDiskSMART_success(t *testing.T) {
	t.Parallel()

	want := map[string]any{"health": "PASSED", "type": "ata"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/pve1/disks/smart" || r.URL.Query().Get("disk") != "/dev/sda" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).GetDiskSMART(context.Background(), "pve1", "/dev/sda")
	if err != nil {
		t.Fatalf("GetDiskSMART: %v", err)
	}
	if got["health"] != "PASSED" {
		t.Errorf("health: got %v, want PASSED", got["health"])
	}
}

func TestGetDiskSMART_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).GetDiskSMART(context.Background(), "pve1", "/dev/missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestListZFSPools_success(t *testing.T) {
	t.Parallel()

	want := []map[string]any{
		{"name": "Storage", "health": "DEGRADED", "size": float64(1000204886016)},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/pve2/disks/zfs" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListZFSPools(context.Background(), "pve2")
	if err != nil {
		t.Fatalf("ListZFSPools: %v", err)
	}
	if len(got) != 1 || got[0]["name"] != "Storage" {
		t.Errorf("got %+v, want 1 pool named Storage", got)
	}
}

func TestListZFSPools_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).ListZFSPools(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetZFSPool_success(t *testing.T) {
	t.Parallel()

	want := map[string]any{"name": "Storage", "state": "DEGRADED"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/pve2/disks/zfs/Storage" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).GetZFSPool(context.Background(), "pve2", "Storage")
	if err != nil {
		t.Fatalf("GetZFSPool: %v", err)
	}
	if got["state"] != "DEGRADED" {
		t.Errorf("state: got %v, want DEGRADED", got["state"])
	}
}

func TestGetZFSPool_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).GetZFSPool(context.Background(), "pve2", "NoSuchPool")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetNodeJournal_success(t *testing.T) {
	t.Parallel()

	want := []string{
		"Aug 07 11:30:38 pve1 sshd-session[26013]: Invalid user testuser from 192.0.2.10 port 55009",
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/pve1/journal" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).GetNodeJournal(context.Background(), "pve1", 0, 0, 0)
	if err != nil {
		t.Fatalf("GetNodeJournal: %v", err)
	}
	if len(got) != 1 || got[0] != want[0] {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestGetNodeJournal_withParams(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/pve1/journal" {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query()
		if q.Get("since") != "100" || q.Get("until") != "200" || q.Get("lastentries") != "50" {
			http.Error(w, "missing query params: "+q.Encode(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, []string{}))
	}))
	defer srv.Close()

	if _, err := newTestClient(t, srv.URL).GetNodeJournal(context.Background(), "pve1", 100, 200, 50); err != nil {
		t.Fatalf("GetNodeJournal with params: %v", err)
	}
}

func TestGetNodeJournal_negativeLastEntriesError(t *testing.T) {
	t.Parallel()

	// The error is returned before any HTTP call; the server is never contacted.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.GetNodeJournal(context.Background(), "pve1", 0, 0, -1)
	if err == nil {
		t.Fatal("expected error for negative lastEntries, got nil")
	}
}

func TestGetNodeJournal_negativeSinceError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.GetNodeJournal(context.Background(), "pve1", -1, 0, 0)
	if err == nil {
		t.Fatal("expected error for negative since, got nil")
	}
}

func TestGetNodeJournal_negativeUntilError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.GetNodeJournal(context.Background(), "pve1", 0, -1, 0)
	if err == nil {
		t.Fatal("expected error for negative until, got nil")
	}
}

func TestGetNodeJournal_sinceAfterUntilError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.GetNodeJournal(context.Background(), "pve1", 200, 100, 0)
	if err == nil {
		t.Fatal("expected error for since > until, got nil")
	}
}

func TestGetNodeJournal_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).GetNodeJournal(context.Background(), "missing", 0, 0, 0)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestNodeCommand_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "want POST", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/nodes/pve1/status" {
			http.NotFound(w, r)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read body", http.StatusInternalServerError)
			return
		}
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		if payload["command"] != "reboot" {
			http.Error(w, "wrong command", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, nil))
	}))
	defer srv.Close()

	if err := newTestClient(t, srv.URL).NodeCommand(context.Background(), "pve1", "reboot"); err != nil {
		t.Fatalf("NodeCommand: %v", err)
	}
}

func TestNodeCommand_apiError(t *testing.T) {
	t.Parallel()
	srv := vmErrorServer(t)
	defer srv.Close()
	err := newTestClient(t, srv.URL).NodeCommand(context.Background(), "pve1", "reboot")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}
