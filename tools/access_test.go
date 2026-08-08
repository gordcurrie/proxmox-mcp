package tools

import (
	"context"
	"errors"
	"testing"
)

func TestListUsers(t *testing.T) {
	t.Run("returns users as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listUsersFn: func(context.Context) ([]map[string]any, error) {
				return []map[string]any{{"userid": "root@pam", "enable": 1}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_users", nil)
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listUsersFn: func(context.Context) ([]map[string]any, error) {
				return nil, errors.New("API error")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_users", nil)
		assertError(t, res, "API error")
	})
}

func TestListUserTokens(t *testing.T) {
	t.Run("returns tokens as JSON", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listUserTokensFn: func(_ context.Context, userid string) ([]map[string]any, error) {
				return []map[string]any{{"tokenid": "mcp", "userid": userid}}, nil
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_user_tokens", map[string]any{"userid": "root@pam"})
		assertResultJSON(t, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		mock := &mockProxmoxClient{
			listUserTokensFn: func(context.Context, string) ([]map[string]any, error) {
				return nil, errors.New("user not found")
			},
		}
		cs, cleanup := connectTestServer(t, mock)
		defer cleanup()

		res := callTool(t, cs, "list_user_tokens", map[string]any{"userid": "nobody@pam"})
		assertError(t, res, "user not found")
	})
}
