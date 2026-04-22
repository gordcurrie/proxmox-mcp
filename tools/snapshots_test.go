package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
)

func TestListVMSnapshots(t *testing.T) {
	t.Run("returns snapshots as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listVMSnapshotsFn: func(_ context.Context, _ string, _ int) ([]proxmox.Snapshot, error) {
				return []proxmox.Snapshot{{Name: "pre-upgrade", Description: "before 6.0"}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_vm_snapshots", map[string]any{"node": "pve1", "vmid": 100})
		assertSuccess(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listVMSnapshotsFn: func(context.Context, string, int) ([]proxmox.Snapshot, error) {
				return nil, errors.New("VM not found")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_vm_snapshots", map[string]any{"node": "pve1", "vmid": 999})
		assertError(t, res, "VM not found")
	})
}

func TestCreateVMSnapshot(t *testing.T) {
	t.Run("returns task ID", func(t *testing.T) {
		const upid = "UPID:pve1:000015E3:00000000:60F4B3A7:qmsnapshot:100:root@pam:"
		mock := &mockProxmoxClient{
			createVMSnapshotFn: func(_ context.Context, _ string, _ int, _ proxmox.CreateVMSnapshotRequest) (string, error) {
				return upid, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "create_vm_snapshot", map[string]any{
			"node":     "pve1",
			"vmid":     100,
			"snapname": "pre-upgrade",
		})
		assertSuccess(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			createVMSnapshotFn: func(context.Context, string, int, proxmox.CreateVMSnapshotRequest) (string, error) {
				return "", errors.New("snapshot already exists")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "create_vm_snapshot", map[string]any{
			"node":     "pve1",
			"vmid":     100,
			"snapname": "dup",
		})
		assertError(t, res, "snapshot already exists")
	})
}

func TestListContainerSnapshots(t *testing.T) {
	t.Run("returns snapshots as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listContainerSnapshotsFn: func(context.Context, string, int) ([]proxmox.Snapshot, error) {
				return []proxmox.Snapshot{{Name: "clean"}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_container_snapshots", map[string]any{"node": "pve1", "vmid": 200})
		assertSuccess(t, res)
	})
}
