package proxmox

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListNodeNetwork_success(t *testing.T) {
	t.Parallel()

	want := []map[string]any{
		{"iface": "vmbr0", "type": "bridge", "active": float64(1)},
		{"iface": "eth0", "type": "eth", "active": float64(1)},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/pve1/network" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListNodeNetwork(context.Background(), "pve1", "")
	if err != nil {
		t.Fatalf("ListNodeNetwork: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d interfaces, want %d", len(got), len(want))
	}
	if got[0]["iface"] != "vmbr0" {
		t.Errorf("iface[0]: got %v, want vmbr0", got[0]["iface"])
	}
}

func TestListNodeNetwork_withTypeFilter(t *testing.T) {
	t.Parallel()

	want := []map[string]any{
		{"iface": "vmbr0", "type": "bridge", "active": float64(1)},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/pve1/network" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("type") != "bridge" {
			http.Error(w, "want type=bridge query param", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListNodeNetwork(context.Background(), "pve1", "bridge")
	if err != nil {
		t.Fatalf("ListNodeNetwork with type filter: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d interfaces, want 1", len(got))
	}
	if got[0]["iface"] != "vmbr0" {
		t.Errorf("iface: got %v, want vmbr0", got[0]["iface"])
	}
}

func TestListNodeNetwork_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).ListNodeNetwork(context.Background(), "missing", "")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetNodeNetworkInterface_success(t *testing.T) {
	t.Parallel()

	want := map[string]any{
		"iface":        "vmbr0",
		"type":         "bridge",
		"address":      "192.168.1.1",
		"netmask":      "255.255.255.0",
		"bridge_ports": "eth0",
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/pve1/network/vmbr0" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).GetNodeNetworkInterface(context.Background(), "pve1", "vmbr0")
	if err != nil {
		t.Fatalf("GetNodeNetworkInterface: %v", err)
	}
	if got["iface"] != "vmbr0" {
		t.Errorf("iface: got %v, want vmbr0", got["iface"])
	}
	if got["address"] != "192.168.1.1" {
		t.Errorf("address: got %v, want 192.168.1.1", got["address"])
	}
}

func TestGetNodeNetworkInterface_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).GetNodeNetworkInterface(context.Background(), "pve1", "missing0")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestCreateNodeNetworkInterface_success(t *testing.T) {
	t.Parallel()

	want := map[string]any{"iface": "vmbr1", "type": "bridge"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/nodes/pve1/network" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).CreateNodeNetworkInterface(
		context.Background(), "pve1", "vmbr1",
		&NetworkInterfaceConfig{Type: "bridge", BridgePorts: "eth1"},
	)
	if err != nil {
		t.Fatalf("CreateNodeNetworkInterface: %v", err)
	}
	if got["iface"] != "vmbr1" {
		t.Errorf("iface: got %v, want vmbr1", got["iface"])
	}
}

func TestCreateNodeNetworkInterface_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).CreateNodeNetworkInterface(
		context.Background(), "missing", "vmbr1",
		&NetworkInterfaceConfig{Type: "bridge"},
	)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestCreateNodeNetworkInterface_validation(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, "http://localhost:9")
	tests := []struct {
		name  string
		node  string
		iface string
		cfg   *NetworkInterfaceConfig
	}{
		{"empty node", "", "vmbr1", &NetworkInterfaceConfig{Type: "bridge"}},
		{"empty iface", "pve1", "", &NetworkInterfaceConfig{Type: "bridge"}},
		{"nil cfg", "pve1", "vmbr1", nil},
		{"empty type", "pve1", "vmbr1", &NetworkInterfaceConfig{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := c.CreateNodeNetworkInterface(context.Background(), tt.node, tt.iface, tt.cfg)
			if err == nil {
				t.Errorf("expected error for %q, got nil", tt.name)
			}
		})
	}
}

func TestUpdateNodeNetworkInterface_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/nodes/pve1/network/vmbr0" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, nil))
	}))
	defer srv.Close()

	err := newTestClient(t, srv.URL).UpdateNodeNetworkInterface(
		context.Background(), "pve1", "vmbr0",
		&NetworkInterfaceConfig{Type: "bridge", Comments: "updated by mcp"},
	)
	if err != nil {
		t.Fatalf("UpdateNodeNetworkInterface: %v", err)
	}
}

func TestUpdateNodeNetworkInterface_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	err := newTestClient(t, srv.URL).UpdateNodeNetworkInterface(
		context.Background(), "pve1", "missing0",
		&NetworkInterfaceConfig{Type: "bridge"},
	)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdateNodeNetworkInterface_validation(t *testing.T) {
	t.Parallel()

	c := newTestClient(t, "http://localhost:9")
	tests := []struct {
		name  string
		node  string
		iface string
		cfg   *NetworkInterfaceConfig
	}{
		{"empty node", "", "vmbr0", &NetworkInterfaceConfig{Type: "bridge"}},
		{"empty iface", "pve1", "", &NetworkInterfaceConfig{Type: "bridge"}},
		{"nil cfg", "pve1", "vmbr0", nil},
		{"empty type", "pve1", "vmbr0", &NetworkInterfaceConfig{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := c.UpdateNodeNetworkInterface(context.Background(), tt.node, tt.iface, tt.cfg)
			if err == nil {
				t.Errorf("expected error for %q, got nil", tt.name)
			}
		})
	}
}

func TestApplyNodeNetworkChanges_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/nodes/pve1/network" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, nil))
	}))
	defer srv.Close()

	if err := newTestClient(t, srv.URL).ApplyNodeNetworkChanges(context.Background(), "pve1"); err != nil {
		t.Fatalf("ApplyNodeNetworkChanges: %v", err)
	}
}

func TestApplyNodeNetworkChanges_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	err := newTestClient(t, srv.URL).ApplyNodeNetworkChanges(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteNodeNetworkInterface_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/nodes/pve1/network/vmbr1" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, nil))
	}))
	defer srv.Close()

	if err := newTestClient(t, srv.URL).DeleteNodeNetworkInterface(context.Background(), "pve1", "vmbr1"); err != nil {
		t.Fatalf("DeleteNodeNetworkInterface: %v", err)
	}
}

func TestDeleteNodeNetworkInterface_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	err := newTestClient(t, srv.URL).DeleteNodeNetworkInterface(context.Background(), "pve1", "missing0")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
