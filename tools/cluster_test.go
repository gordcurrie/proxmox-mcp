package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
)

func TestListClusterResources(t *testing.T) {
	t.Run("returns resources as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listClusterResourcesFn: func(_ context.Context, _ string) ([]proxmox.ClusterResource, error) {
				return []proxmox.ClusterResource{{ID: "qemu/100", Type: "qemu", Node: "pve1"}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_cluster_resources", nil)
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listClusterResourcesFn: func(context.Context, string) ([]proxmox.ClusterResource, error) {
				return nil, errors.New("cluster unreachable")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_cluster_resources", nil)
		assertError(t, res, "cluster unreachable")
	})
}

func TestGetTaskStatus(t *testing.T) {
	t.Run("returns task status as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getTaskStatusFn: func(_ context.Context, _, _ string) (*proxmox.TaskStatus, error) {
				return &proxmox.TaskStatus{Status: "OK", ExitStatus: "OK"}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_task_status", map[string]any{
			"node": "pve1",
			"upid": "UPID:pve1:000015E3:00000000:60F4B3A7:qmstart:100:root@pam:",
		})
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getTaskStatusFn: func(context.Context, string, string) (*proxmox.TaskStatus, error) {
				return nil, errors.New("task not found")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_task_status", map[string]any{"node": "pve1", "upid": "invalid"})
		assertError(t, res, "task not found")
	})
}

func TestGetClusterStatus(t *testing.T) {
	t.Run("returns cluster status as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getClusterStatusFn: func(context.Context) ([]map[string]any, error) {
				return []map[string]any{{"type": "cluster", "quorate": 1}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_cluster_status", nil)
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getClusterStatusFn: func(context.Context) ([]map[string]any, error) {
				return nil, errors.New("quorum lost")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_cluster_status", nil)
		assertError(t, res, "quorum lost")
	})
}

func TestListHAGroups(t *testing.T) {
	t.Run("returns groups as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listHAGroupsFn: func(context.Context) ([]map[string]any, error) {
				return []map[string]any{{"group": "no-pve3"}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_ha_groups", nil)
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listHAGroupsFn: func(context.Context) ([]map[string]any, error) {
				return nil, errors.New("API error")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_ha_groups", nil)
		assertError(t, res, "API error")
	})
}

func TestListHAResources(t *testing.T) {
	t.Run("returns resources as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listHAResourcesFn: func(context.Context) ([]map[string]any, error) {
				return []map[string]any{{"sid": "ct:104", "state": "started"}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_ha_resources", nil)
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listHAResourcesFn: func(context.Context) ([]map[string]any, error) {
				return nil, errors.New("API error")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_ha_resources", nil)
		assertError(t, res, "API error")
	})
}

func TestGetHAStatus(t *testing.T) {
	t.Run("returns status as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getHAStatusFn: func(context.Context) ([]map[string]any, error) {
				return []map[string]any{{"type": "quorum", "status": "OK"}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_ha_status", nil)
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getHAStatusFn: func(context.Context) ([]map[string]any, error) {
				return nil, errors.New("HA manager unreachable")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_ha_status", nil)
		assertError(t, res, "HA manager unreachable")
	})
}

func TestListClusterConfigNodes(t *testing.T) {
	t.Run("returns nodes as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listClusterConfigNodesFn: func(context.Context) ([]map[string]any, error) {
				return []map[string]any{{"name": "pve1", "nodeid": 1}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_cluster_config_nodes", nil)
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listClusterConfigNodesFn: func(context.Context) ([]map[string]any, error) {
				return nil, errors.New("API error")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_cluster_config_nodes", nil)
		assertError(t, res, "API error")
	})
}
