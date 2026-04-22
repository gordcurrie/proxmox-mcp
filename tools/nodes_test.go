package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
)

func TestListNodes(t *testing.T) {
	t.Run("returns nodes as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listNodesFn: func(context.Context) ([]proxmox.Node, error) {
				return []proxmox.Node{{Node: "pve1", Status: "online"}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_nodes", nil)
		assertSuccess(t, res)
	})

	t.Run("propagates client error as isError", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listNodesFn: func(context.Context) ([]proxmox.Node, error) {
				return nil, errors.New("connection refused")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_nodes", nil)
		assertError(t, res, "connection refused")
	})
}

func TestGetNodeStatus(t *testing.T) {
	t.Run("returns status as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getNodeStatusFn: func(_ context.Context, node string) (map[string]any, error) {
				return map[string]any{"node": node, "uptime": 12345}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_node_status", map[string]any{"node": "pve1"})
		assertSuccess(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getNodeStatusFn: func(context.Context, string) (map[string]any, error) {
				return nil, errors.New("node not found")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_node_status", map[string]any{"node": "pve99"})
		assertError(t, res, "node not found")
	})
}

func TestListNodeTasks(t *testing.T) {
	t.Run("returns tasks as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listNodeTasksFn: func(_ context.Context, _ string, _ int) ([]map[string]any, error) {
				return []map[string]any{{"upid": "UPID:pve1:0001:qmstart:100:root@pam:"}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_node_tasks", map[string]any{"node": "pve1", "limit": 10})
		assertSuccess(t, res)
	})

	t.Run("rejects negative limit", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockProxmoxClient{})
		defer cleanup()

		res := callTool(t, cs, "list_node_tasks", map[string]any{"node": "pve1", "limit": -1})
		assertError(t, res, "limit must be >= 0")
	})
}

func TestGetNodeDisks(t *testing.T) {
	t.Run("returns disks as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getNodeDisksFn: func(context.Context, string) ([]map[string]any, error) {
				return []map[string]any{{"devpath": "/dev/sda", "size": 500107862016}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_node_disks", map[string]any{"node": "pve1"})
		assertSuccess(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getNodeDisksFn: func(context.Context, string) ([]map[string]any, error) {
				return nil, errors.New("disk enumeration failed")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_node_disks", map[string]any{"node": "pve1"})
		assertError(t, res, "disk enumeration failed")
	})
}
