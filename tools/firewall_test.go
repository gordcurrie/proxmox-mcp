package tools

import (
	"context"
	"errors"
	"testing"
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
