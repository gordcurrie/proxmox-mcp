package tools

import (
	"context"
	"errors"
	"testing"
)

func TestDeleteVM(t *testing.T) {
	t.Run("returns task ID", func(t *testing.T) {
		const upid = "UPID:pve1:000015E3:00000000:60F4B3A7:qmdestroy:100:root@pam:"
		mock := &mockProxmoxClient{
			deleteVMFn: func(context.Context, string, int, bool) (string, error) {
				return upid, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "delete_vm", map[string]any{"node": "pve1", "vmid": 100, "confirmed": true})
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			deleteVMFn: func(context.Context, string, int, bool) (string, error) {
				return "", errors.New("VM is running")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "delete_vm", map[string]any{"node": "pve1", "vmid": 100, "confirmed": true})
		assertError(t, res, "VM is running")
	})
}

func TestDeleteContainer(t *testing.T) {
	t.Run("returns task ID", func(t *testing.T) {
		const upid = "UPID:pve1:000015E3:00000000:60F4B3A7:vzdestroy:200:root@pam:"
		mock := &mockProxmoxClient{
			deleteContainerFn: func(context.Context, string, int, bool) (string, error) {
				return upid, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "delete_container", map[string]any{"node": "pve1", "vmid": 200, "confirmed": true})
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			deleteContainerFn: func(context.Context, string, int, bool) (string, error) {
				return "", errors.New("container is running")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "delete_container", map[string]any{"node": "pve1", "vmid": 200, "confirmed": true})
		assertError(t, res, "container is running")
	})
}
