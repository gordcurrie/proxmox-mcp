package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerAccessTools adds user/token audit MCP tools to the server.
func registerAccessTools(s *mcp.Server, client proxmoxClient) {
	type listUsersInput struct{}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_users",
		Description: "List all users configured in Proxmox access control, across all realms (pve, pam, ldap, etc). Use to audit for unrecognized or orphaned accounts.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ listUsersInput) (*mcp.CallToolResult, any, error) {
		users, err := client.ListUsers(ctx)
		if err != nil {
			return errorResult(fmt.Errorf("list_users: %w", err))
		}
		return jsonResult(users)
	})

	type listUserTokensInput struct {
		UserID string `json:"userid" jsonschema:"full user ID including realm, e.g. root@pam"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_user_tokens",
		Description: "List API tokens issued to a Proxmox user. Token secrets are never returned, only metadata (token ID, comment, expiry).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input listUserTokensInput) (*mcp.CallToolResult, any, error) {
		tokens, err := client.ListUserTokens(ctx, input.UserID)
		if err != nil {
			return errorResult(fmt.Errorf("list_user_tokens: %w", err))
		}
		return jsonResult(tokens)
	})
}
