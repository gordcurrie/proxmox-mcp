package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
)

func TestListVMs(t *testing.T) {
	t.Run("returns VMs as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listVMsFn: func(_ context.Context, _ string) ([]proxmox.VM, error) {
				return []proxmox.VM{{VMID: 100, Name: "debian12", Status: "running"}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_vms", map[string]any{"node": "pve1"})
		assertSuccess(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listVMsFn: func(context.Context, string) ([]proxmox.VM, error) {
				return nil, errors.New("node offline")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_vms", map[string]any{"node": "pve1"})
		assertError(t, res, "node offline")
	})
}

func TestGetVMStatus(t *testing.T) {
	t.Run("returns status as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getVMStatusFn: func(_ context.Context, _ string, _ int) (map[string]any, error) {
				return map[string]any{"status": "running", "vmid": 100}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_vm_status", map[string]any{"node": "pve1", "vmid": 100})
		assertSuccess(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getVMStatusFn: func(context.Context, string, int) (map[string]any, error) {
				return nil, errors.New("VM not found")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_vm_status", map[string]any{"node": "pve1", "vmid": 999})
		assertError(t, res, "VM not found")
	})
}

func TestStartVM(t *testing.T) {
	t.Run("returns task ID", func(t *testing.T) {
		const upid = "UPID:pve1:000015E3:00000000:60F4B3A7:qmstart:100:root@pam:"
		mock := &mockProxmoxClient{
			startVMFn: func(context.Context, string, int) (string, error) {
				return upid, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "start_vm", map[string]any{"node": "pve1", "vmid": 100})
		assertSuccess(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			startVMFn: func(context.Context, string, int) (string, error) {
				return "", errors.New("VM already running")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "start_vm", map[string]any{"node": "pve1", "vmid": 100})
		assertError(t, res, "VM already running")
	})
}

func TestGetVMConfig(t *testing.T) {
	t.Run("returns config as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getVMConfigFn: func(context.Context, string, int) (map[string]any, error) {
				return map[string]any{"cores": 2, "memory": 2048}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_vm_config", map[string]any{"node": "pve1", "vmid": 100})
		assertSuccess(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getVMConfigFn: func(context.Context, string, int) (map[string]any, error) {
				return nil, errors.New("config read failed")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_vm_config", map[string]any{"node": "pve1", "vmid": 100})
		assertError(t, res, "config read failed")
	})
}
