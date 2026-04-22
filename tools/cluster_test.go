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
		assertSuccess(t, res)
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
		assertSuccess(t, res)
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
		assertSuccess(t, res)
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
