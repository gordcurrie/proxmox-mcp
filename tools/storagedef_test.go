package tools

import (
	"context"
	"errors"
	"testing"
)

func TestListStorages(t *testing.T) {
	t.Run("returns storage definitions as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listStoragesFn: func(context.Context, string) ([]map[string]any, error) {
				return []map[string]any{{"storage": "local", "type": "dir"}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_storages", nil)
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listStoragesFn: func(context.Context, string) ([]map[string]any, error) {
				return nil, errors.New("API unavailable")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_storages", nil)
		assertError(t, res, "API unavailable")
	})
}

func TestGetStorage(t *testing.T) {
	t.Run("returns storage info as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getStorageFn: func(_ context.Context, storage string) (map[string]any, error) {
				return map[string]any{"storage": storage, "type": "dir", "path": "/var/lib/vz"}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_storage", map[string]any{"storage": "local"})
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getStorageFn: func(context.Context, string) (map[string]any, error) {
				return nil, errors.New("storage not found")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_storage", map[string]any{"storage": "noexist"})
		assertError(t, res, "storage not found")
	})
}
