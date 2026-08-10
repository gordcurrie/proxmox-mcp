package proxmox

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListClusterResources_success(t *testing.T) {
	t.Parallel()

	want := []ClusterResource{
		{ID: "node/pve1", Type: "node", Node: "pve1", Status: "online"},
		{ID: "qemu/100", Type: "qemu", Node: "pve1", VMID: 100, Status: "running"},
		{ID: "lxc/200", Type: "lxc", Node: "pve1", VMID: 200, Status: "stopped"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cluster/resources" {
			http.NotFound(w, r)
			return
		}
		if r.URL.RawQuery != "" {
			http.Error(w, "unexpected query param", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListClusterResources(context.Background(), "")
	if err != nil {
		t.Fatalf("ListClusterResources: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d resources, want %d", len(got), len(want))
	}
	for i, r := range got {
		if r.ID != want[i].ID || r.Type != want[i].Type {
			t.Errorf("resource[%d]: got %+v, want %+v", i, r, want[i])
		}
	}
}

func TestListClusterResources_withTypeFilter(t *testing.T) {
	t.Parallel()

	want := []ClusterResource{
		{ID: "qemu/100", Type: "qemu", Node: "pve1", VMID: 100, Status: "running"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cluster/resources" {
			http.NotFound(w, r)
			return
		}
		if !strings.Contains(r.URL.RawQuery, "type=vm") {
			http.Error(w, "missing type filter", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListClusterResources(context.Background(), "vm")
	if err != nil {
		t.Fatalf("ListClusterResources with type filter: %v", err)
	}
	if len(got) != 1 || got[0].Type != "qemu" {
		t.Errorf("got %+v, want 1 qemu resource", got)
	}
}

func TestListClusterResources_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).ListClusterResources(context.Background(), "")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetClusterStatus_success(t *testing.T) {
	t.Parallel()

	want := []map[string]any{
		{"type": "cluster", "name": "proxmox", "quorate": float64(1)},
		{"type": "node", "name": "pve1", "online": float64(1)},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cluster/status" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).GetClusterStatus(context.Background())
	if err != nil {
		t.Fatalf("GetClusterStatus: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d entries, want %d", len(got), len(want))
	}
	if got[0]["type"] != "cluster" {
		t.Errorf("first entry type: got %v, want cluster", got[0]["type"])
	}
}

func TestGetClusterStatus_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).GetClusterStatus(context.Background())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestListHAGroups_success(t *testing.T) {
	t.Parallel()

	want := []map[string]any{
		{"rule": "no-pve3", "type": "node-affinity", "nodes": "pve1,pve2,pve4"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cluster/ha/rules" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListHAGroups(context.Background())
	if err != nil {
		t.Fatalf("ListHAGroups: %v", err)
	}
	if len(got) != 1 || got[0]["rule"] != "no-pve3" {
		t.Errorf("got %+v, want 1 rule named no-pve3", got)
	}
}

func TestListHAGroups_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).ListHAGroups(context.Background())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestListHAResources_success(t *testing.T) {
	t.Parallel()

	want := []map[string]any{
		{"sid": "ct:104", "state": "started"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cluster/ha/resources" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListHAResources(context.Background())
	if err != nil {
		t.Fatalf("ListHAResources: %v", err)
	}
	if len(got) != 1 || got[0]["sid"] != "ct:104" {
		t.Errorf("got %+v, want 1 resource ct:104", got)
	}
}

func TestListHAResources_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).ListHAResources(context.Background())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetHAStatus_success(t *testing.T) {
	t.Parallel()

	want := []map[string]any{
		{"type": "quorum", "status": "OK"},
		{"type": "node", "node": "pve1", "status": "online"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cluster/ha/status/current" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).GetHAStatus(context.Background())
	if err != nil {
		t.Fatalf("GetHAStatus: %v", err)
	}
	if len(got) != 2 || got[0]["type"] != "quorum" {
		t.Errorf("got %+v, want first entry type quorum", got)
	}
}

func TestGetHAStatus_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).GetHAStatus(context.Background())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestListClusterConfigNodes_success(t *testing.T) {
	t.Parallel()

	want := []map[string]any{
		{"name": "pve1", "nodeid": float64(1), "ring0_addr": "192.168.4.101"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cluster/config/nodes" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListClusterConfigNodes(context.Background())
	if err != nil {
		t.Fatalf("ListClusterConfigNodes: %v", err)
	}
	if len(got) != 1 || got[0]["name"] != "pve1" {
		t.Errorf("got %+v, want 1 node named pve1", got)
	}
}

func TestListClusterConfigNodes_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).ListClusterConfigNodes(context.Background())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
