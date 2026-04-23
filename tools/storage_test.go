package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
)

func TestListStorageContent(t *testing.T) {
	t.Run("returns content list as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listStorageContentFn: func(_ context.Context, _, _, _ string) ([]proxmox.StorageContent, error) {
				return []proxmox.StorageContent{{VolID: "local:iso/debian.iso", Content: "iso"}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_storage_content", map[string]any{
			"node":    "pve1",
			"storage": "local",
		})
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listStorageContentFn: func(context.Context, string, string, string) ([]proxmox.StorageContent, error) {
				return nil, errors.New("storage not found")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_storage_content", map[string]any{"node": "pve1", "storage": "noexist"})
		assertError(t, res, "storage not found")
	})
}

func TestGetStorageContentInfo(t *testing.T) {
	t.Run("returns info as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getStorageContentInfoFn: func(context.Context, string, string, string) (map[string]any, error) {
				return map[string]any{"volid": "local:iso/debian.iso", "size": 500000000}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_storage_content_info", map[string]any{
			"node":    "pve1",
			"storage": "local",
			"volume":  "local:iso/debian.iso",
		})
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getStorageContentInfoFn: func(context.Context, string, string, string) (map[string]any, error) {
				return nil, errors.New("volume not found")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_storage_content_info", map[string]any{
			"node":    "pve1",
			"storage": "local",
			"volume":  "noexist",
		})
		assertError(t, res, "volume not found")
	})
}
