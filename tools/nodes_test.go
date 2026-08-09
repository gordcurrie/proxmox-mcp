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
		assertResultJSON(t, res)
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
		assertResultJSON(t, res)
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
		assertResultJSON(t, res)
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
		assertResultJSON(t, res)
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

func TestGetDiskSMART(t *testing.T) {
	t.Run("returns SMART data as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getDiskSMARTFn: func(context.Context, string, string) (map[string]any, error) {
				return map[string]any{"health": "PASSED"}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_disk_smart", map[string]any{"node": "pve1", "disk": "/dev/sda"})
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getDiskSMARTFn: func(context.Context, string, string) (map[string]any, error) {
				return nil, errors.New("disk not found")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_disk_smart", map[string]any{"node": "pve1", "disk": "/dev/missing"})
		assertError(t, res, "disk not found")
	})
}

func TestListZFSPools(t *testing.T) {
	t.Run("returns pools as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listZFSPoolsFn: func(context.Context, string) ([]map[string]any, error) {
				return []map[string]any{{"name": "Storage", "health": "ONLINE"}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_zfs_pools", map[string]any{"node": "pve1"})
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listZFSPoolsFn: func(context.Context, string) ([]map[string]any, error) {
				return nil, errors.New("API error")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_zfs_pools", map[string]any{"node": "pve1"})
		assertError(t, res, "API error")
	})
}

func TestGetZFSPool(t *testing.T) {
	t.Run("returns pool as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getZFSPoolFn: func(_ context.Context, _, name string) (map[string]any, error) {
				return map[string]any{"name": name, "state": "DEGRADED"}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_zfs_pool", map[string]any{"node": "pve2", "name": "Storage"})
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getZFSPoolFn: func(context.Context, string, string) (map[string]any, error) {
				return nil, errors.New("pool not found")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_zfs_pool", map[string]any{"node": "pve2", "name": "NoSuch"})
		assertError(t, res, "pool not found")
	})
}

func TestGetNodeJournal(t *testing.T) {
	t.Run("returns journal lines as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getNodeJournalFn: func(context.Context, string, int64, int64, int) ([]string, error) {
				return []string{"Aug 07 11:30:38 pve1 sshd-session[26013]: Invalid user testuser"}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_node_journal", map[string]any{"node": "pve1", "last_entries": 50})
		assertResultJSON(t, res)
	})

	t.Run("rejects negative last_entries", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockProxmoxClient{})
		defer cleanup()

		res := callTool(t, cs, "get_node_journal", map[string]any{"node": "pve1", "last_entries": -1})
		assertError(t, res, "last_entries must be >= 0")
	})

	t.Run("rejects negative since", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockProxmoxClient{})
		defer cleanup()

		res := callTool(t, cs, "get_node_journal", map[string]any{"node": "pve1", "since": -1})
		assertError(t, res, "since must be >= 0")
	})

	t.Run("rejects negative until", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockProxmoxClient{})
		defer cleanup()

		res := callTool(t, cs, "get_node_journal", map[string]any{"node": "pve1", "until": -1})
		assertError(t, res, "until must be >= 0")
	})

	t.Run("rejects since after until", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockProxmoxClient{})
		defer cleanup()

		res := callTool(t, cs, "get_node_journal", map[string]any{"node": "pve1", "since": 200, "until": 100})
		assertError(t, res, "must be <= until")
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getNodeJournalFn: func(context.Context, string, int64, int64, int) ([]string, error) {
				return nil, errors.New("journal unavailable")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_node_journal", map[string]any{"node": "pve1"})
		assertError(t, res, "journal unavailable")
	})
}
