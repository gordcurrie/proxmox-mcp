package proxmox

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListClusterFirewallRules_success(t *testing.T) {
	t.Parallel()

	want := []map[string]any{
		{"pos": float64(0), "type": "in", "action": "ACCEPT", "enable": float64(1), "source": "192.168.1.0/24"},
		{"pos": float64(1), "type": "out", "action": "DROP", "enable": float64(1)},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cluster/firewall/rules" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListClusterFirewallRules(context.Background())
	if err != nil {
		t.Fatalf("ListClusterFirewallRules: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d rules, want %d", len(got), len(want))
	}
	if got[0]["action"] != "ACCEPT" {
		t.Errorf("action[0]: got %v, want ACCEPT", got[0]["action"])
	}
}

func TestListClusterFirewallRules_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).ListClusterFirewallRules(context.Background())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetClusterFirewallOptions_success(t *testing.T) {
	t.Parallel()

	want := map[string]any{
		"enable":        float64(1),
		"policy_in":     "DROP",
		"policy_out":    "ACCEPT",
		"log_ratelimit": "1/second:5",
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cluster/firewall/options" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).GetClusterFirewallOptions(context.Background())
	if err != nil {
		t.Fatalf("GetClusterFirewallOptions: %v", err)
	}
	if got["policy_in"] != "DROP" {
		t.Errorf("policy_in: got %v, want DROP", got["policy_in"])
	}
}

func TestGetClusterFirewallOptions_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).GetClusterFirewallOptions(context.Background())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestListVMFirewallRules_success(t *testing.T) {
	t.Parallel()

	want := []map[string]any{
		{"pos": float64(0), "type": "in", "action": "ACCEPT", "dport": "22", "proto": "tcp"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/pve1/qemu/100/firewall/rules" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListVMFirewallRules(context.Background(), "pve1", 100)
	if err != nil {
		t.Fatalf("ListVMFirewallRules: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d rules, want 1", len(got))
	}
	if got[0]["dport"] != "22" {
		t.Errorf("dport: got %v, want 22", got[0]["dport"])
	}
}

func TestListVMFirewallRules_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).ListVMFirewallRules(context.Background(), "pve1", 9999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetVMFirewallOptions_success(t *testing.T) {
	t.Parallel()

	want := map[string]any{
		"enable":     float64(1),
		"policy_in":  "DROP",
		"policy_out": "ACCEPT",
		"dhcp":       float64(1),
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/pve1/qemu/100/firewall/options" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).GetVMFirewallOptions(context.Background(), "pve1", 100)
	if err != nil {
		t.Fatalf("GetVMFirewallOptions: %v", err)
	}
	if got["policy_in"] != "DROP" {
		t.Errorf("policy_in: got %v, want DROP", got["policy_in"])
	}
}

func TestGetVMFirewallOptions_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).GetVMFirewallOptions(context.Background(), "pve1", 9999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestListContainerFirewallRules_success(t *testing.T) {
	t.Parallel()

	want := []map[string]any{
		{"pos": float64(0), "type": "in", "action": "ACCEPT", "dport": "80", "proto": "tcp"},
		{"pos": float64(1), "type": "in", "action": "ACCEPT", "dport": "443", "proto": "tcp"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/pve1/lxc/200/firewall/rules" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).ListContainerFirewallRules(context.Background(), "pve1", 200)
	if err != nil {
		t.Fatalf("ListContainerFirewallRules: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d rules, want 2", len(got))
	}
	if got[1]["dport"] != "443" {
		t.Errorf("dport[1]: got %v, want 443", got[1]["dport"])
	}
}

func TestListContainerFirewallRules_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).ListContainerFirewallRules(context.Background(), "pve1", 9999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetContainerFirewallOptions_success(t *testing.T) {
	t.Parallel()

	want := map[string]any{
		"enable":     float64(1),
		"policy_in":  "ACCEPT",
		"policy_out": "ACCEPT",
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/nodes/pve1/lxc/200/firewall/options" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, want))
	}))
	defer srv.Close()

	got, err := newTestClient(t, srv.URL).GetContainerFirewallOptions(context.Background(), "pve1", 200)
	if err != nil {
		t.Fatalf("GetContainerFirewallOptions: %v", err)
	}
	if got["enable"] != float64(1) {
		t.Errorf("enable: got %v, want 1", got["enable"])
	}
}

func TestGetContainerFirewallOptions_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv.URL).GetContainerFirewallOptions(context.Background(), "pve1", 9999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAddVMFirewallRule_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "want POST", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/nodes/pve1/qemu/100/firewall/rules" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, nil))
	}))
	defer srv.Close()

	req := &FirewallRuleRequest{Type: "in", Action: "ACCEPT", DPort: "22", Proto: "tcp"}
	if err := newTestClient(t, srv.URL).AddVMFirewallRule(context.Background(), "pve1", 100, req); err != nil {
		t.Fatalf("AddVMFirewallRule: %v", err)
	}
}

func TestAddVMFirewallRule_apiError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "firewall error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	req := &FirewallRuleRequest{Type: "in", Action: "ACCEPT"}
	err := newTestClient(t, srv.URL).AddVMFirewallRule(context.Background(), "pve1", 100, req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeleteVMFirewallRule_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "want DELETE", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/nodes/pve1/qemu/100/firewall/rules/0" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, nil))
	}))
	defer srv.Close()

	if err := newTestClient(t, srv.URL).DeleteVMFirewallRule(context.Background(), "pve1", 100, 0); err != nil {
		t.Fatalf("DeleteVMFirewallRule: %v", err)
	}
}

func TestDeleteVMFirewallRule_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	err := newTestClient(t, srv.URL).DeleteVMFirewallRule(context.Background(), "pve1", 9999, 0)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAddContainerFirewallRule_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "want POST", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/nodes/pve1/lxc/200/firewall/rules" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, nil))
	}))
	defer srv.Close()

	req := &FirewallRuleRequest{Type: "in", Action: "ACCEPT", DPort: "80", Proto: "tcp"}
	if err := newTestClient(t, srv.URL).AddContainerFirewallRule(context.Background(), "pve1", 200, req); err != nil {
		t.Fatalf("AddContainerFirewallRule: %v", err)
	}
}

func TestAddContainerFirewallRule_apiError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "firewall error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	req := &FirewallRuleRequest{Type: "in", Action: "ACCEPT"}
	err := newTestClient(t, srv.URL).AddContainerFirewallRule(context.Background(), "pve1", 200, req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeleteContainerFirewallRule_success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "want DELETE", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/nodes/pve1/lxc/200/firewall/rules/1" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonEnvelope(t, nil))
	}))
	defer srv.Close()

	if err := newTestClient(t, srv.URL).DeleteContainerFirewallRule(context.Background(), "pve1", 200, 1); err != nil {
		t.Fatalf("DeleteContainerFirewallRule: %v", err)
	}
}

func TestDeleteContainerFirewallRule_notFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	err := newTestClient(t, srv.URL).DeleteContainerFirewallRule(context.Background(), "pve1", 9999, 0)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
