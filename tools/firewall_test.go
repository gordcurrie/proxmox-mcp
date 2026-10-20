package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
)

func TestListClusterFirewallRules(t *testing.T) {
	t.Run("returns rules as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listClusterFirewallRulesFn: func(context.Context) ([]map[string]any, error) {
				return []map[string]any{{"pos": 0, "action": "DROP", "type": "in"}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_cluster_firewall_rules", nil)
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listClusterFirewallRulesFn: func(context.Context) ([]map[string]any, error) {
				return nil, errors.New("firewall service down")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_cluster_firewall_rules", nil)
		assertError(t, res, "firewall service down")
	})
}

func TestListVMFirewallRules(t *testing.T) {
	t.Run("returns VM rules as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listVMFirewallRulesFn: func(context.Context, string, int) ([]map[string]any, error) {
				return []map[string]any{{"pos": 0, "action": "ACCEPT"}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_vm_firewall_rules", map[string]any{"node": "pve1", "vmid": 100})
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listVMFirewallRulesFn: func(context.Context, string, int) ([]map[string]any, error) {
				return nil, errors.New("VM not found")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_vm_firewall_rules", map[string]any{"node": "pve1", "vmid": 999})
		assertError(t, res, "VM not found")
	})
}

func TestAddFirewallRuleEnable(t *testing.T) {
	enabled, disabled := 1, 0
	tests := []struct {
		name   string
		enable any
		want   *int
	}{
		{name: "omitted defaults to enabled", enable: nil, want: &enabled},
		{name: "true enables", enable: true, want: &enabled},
		{name: "false disables", enable: false, want: &disabled},
	}
	for _, tool := range []string{"add_vm_firewall_rule", "add_container_firewall_rule"} {
		for _, tt := range tests {
			t.Run(tool+"/"+tt.name, func(t *testing.T) {
				var got *proxmox.FirewallRuleRequest
				capture := func(_ context.Context, _ string, _ int, req *proxmox.FirewallRuleRequest) error {
					got = req
					return nil
				}
				mock := &mockProxmoxClient{
					addVMFirewallRuleFn:        capture,
					addContainerFirewallRuleFn: capture,
				}
				cs, cleanup := connectTestServer(t, mock)
				defer cleanup()

				args := map[string]any{"node": "pve1", "vmid": 100, "type": "in", "action": "DROP"}
				if tt.enable != nil {
					args["enable"] = tt.enable
				}
				res := callTool(t, cs, tool, args)
				assertSuccess(t, res)
				if got == nil || got.Enable == nil {
					t.Fatalf("Enable not sent, want %d", *tt.want)
				}
				if *got.Enable != *tt.want {
					t.Errorf("Enable = %d, want %d", *got.Enable, *tt.want)
				}
			})
		}
	}
}

func TestAddFirewallRulePropagatesError(t *testing.T) {
	fail := func(context.Context, string, int, *proxmox.FirewallRuleRequest) error {
		return errors.New("rule rejected")
	}
	mock := &mockProxmoxClient{addVMFirewallRuleFn: fail, addContainerFirewallRuleFn: fail}
	for _, tool := range []string{"add_vm_firewall_rule", "add_container_firewall_rule"} {
		t.Run(tool, func(t *testing.T) {
			cs, cleanup := connectTestServer(t, mock)
			defer cleanup()

			res := callTool(t, cs, tool, map[string]any{"node": "pve1", "vmid": 100, "type": "in", "action": "DROP"})
			assertError(t, res, "rule rejected")
		})
	}
}
