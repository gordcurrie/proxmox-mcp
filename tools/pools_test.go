package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
)

func TestListPools(t *testing.T) {
	t.Run("returns pools as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listPoolsFn: func(context.Context) ([]proxmox.Pool, error) {
				return []proxmox.Pool{{PoolID: "dev", Comment: "dev pool"}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_pools", nil)
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listPoolsFn: func(context.Context) ([]proxmox.Pool, error) {
				return nil, errors.New("API error")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_pools", nil)
		assertError(t, res, "API error")
	})
}

func TestGetPool(t *testing.T) {
	t.Run("returns pool as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getPoolFn: func(_ context.Context, poolid string) (*proxmox.Pool, error) {
				return &proxmox.Pool{PoolID: poolid}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_pool", map[string]any{"poolid": "dev"})
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			getPoolFn: func(context.Context, string) (*proxmox.Pool, error) {
				return nil, errors.New("pool not found")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "get_pool", map[string]any{"poolid": "noexist"})
		assertError(t, res, "pool not found")
	})
}

func TestCreatePool(t *testing.T) {
	t.Run("succeeds on nil error", func(t *testing.T) {
		cs, cleanup := connectTestServer(t, &mockProxmoxClient{})
		defer cleanup()

		res := callTool(t, cs, "create_pool", map[string]any{"poolid": "prod", "comment": "prod pool"})
		assertSuccess(t, res) // create_pool returns textResult, not JSON
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			createPoolFn: func(context.Context, *proxmox.CreatePoolRequest) error {
				return errors.New("pool already exists")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "create_pool", map[string]any{"poolid": "dup"})
		assertError(t, res, "pool already exists")
	})
}
