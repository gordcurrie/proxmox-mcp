package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
)

func TestListContainers(t *testing.T) {
	t.Run("returns containers as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listContainersFn: func(_ context.Context, _ string) ([]proxmox.Container, error) {
				return []proxmox.Container{{VMID: 200, Name: "debian-ct", Status: "running"}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_containers", map[string]any{"node": "pve1"})
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listContainersFn: func(context.Context, string) ([]proxmox.Container, error) {
				return nil, errors.New("node offline")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_containers", map[string]any{"node": "pve1"})
		assertError(t, res, "node offline")
	})
}

func TestStartContainer(t *testing.T) {
	t.Run("returns task ID", func(t *testing.T) {
		const upid = "UPID:pve1:000015E3:00000000:60F4B3A7:vzstart:200:root@pam:"
		mock := &mockProxmoxClient{
			startContainerFn: func(context.Context, string, int) (string, error) {
				return upid, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "start_container", map[string]any{"node": "pve1", "vmid": 200})
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			startContainerFn: func(context.Context, string, int) (string, error) {
				return "", errors.New("container already running")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "start_container", map[string]any{"node": "pve1", "vmid": 200})
		assertError(t, res, "container already running")
	})
}
