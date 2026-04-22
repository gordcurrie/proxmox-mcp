package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
)

func TestCreateBackup(t *testing.T) {
	t.Run("returns task ID", func(t *testing.T) {
		const upid = "UPID:pve1:000015E3:00000000:60F4B3A7:vzdump:100:root@pam:"
		mock := &mockProxmoxClient{
			createBackupFn: func(_ context.Context, _ string, _ *proxmox.CreateBackupRequest) (string, error) {
				return upid, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "create_backup", map[string]any{
			"node":    "pve1",
			"vmid":    100,
			"storage": "local",
			"mode":    "snapshot",
		})
		assertSuccess(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			createBackupFn: func(context.Context, string, *proxmox.CreateBackupRequest) (string, error) {
				return "", errors.New("storage full")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "create_backup", map[string]any{
			"node":    "pve1",
			"vmid":    100,
			"storage": "local",
		})
		assertError(t, res, "storage full")
	})
}

func TestListBackups(t *testing.T) {
	t.Run("returns backups as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listBackupsFn: func(_ context.Context, _, _ string) ([]proxmox.StorageContent, error) {
				return []proxmox.StorageContent{{VolID: "local:backup/vzdump-qemu-100.vma.zst", Content: "backup"}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_backups", map[string]any{"node": "pve1", "storage": "local"})
		assertSuccess(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listBackupsFn: func(context.Context, string, string) ([]proxmox.StorageContent, error) {
				return nil, errors.New("storage offline")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_backups", map[string]any{"node": "pve1", "storage": "local"})
		assertError(t, res, "storage offline")
	})
}
