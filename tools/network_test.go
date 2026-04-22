package tools

import (
	"context"
	"errors"
	"testing"
)

func TestListNodeNetwork(t *testing.T) {
	t.Run("returns interfaces as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listNodeNetworkFn: func(_ context.Context, _, _ string) ([]map[string]any, error) {
				return []map[string]any{{"iface": "vmbr0", "type": "bridge"}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_node_network", map[string]any{"node": "pve1"})
		assertSuccess(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listNodeNetworkFn: func(context.Context, string, string) ([]map[string]any, error) {
				return nil, errors.New("node unavailable")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_node_network", map[string]any{"node": "pve1"})
		assertError(t, res, "node unavailable")
	})
}

func TestGetNodeNetworkInterface(t *testing.T) {
	t.Run("returns interface config as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getNodeNetworkInterfaceFn: func(_ context.Context, _, iface string) (map[string]any, error) {
				return map[string]any{"iface": iface, "bridge_ports": "eth0"}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_node_network_interface", map[string]any{"node": "pve1", "iface": "vmbr0"})
		assertSuccess(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getNodeNetworkInterfaceFn: func(context.Context, string, string) (map[string]any, error) {
				return nil, errors.New("interface not found")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_node_network_interface", map[string]any{"node": "pve1", "iface": "noexist"})
		assertError(t, res, "interface not found")
	})
}
