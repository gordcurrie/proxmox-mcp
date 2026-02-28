package tools

import (
	"context"
	"fmt"

	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerPoolTools adds pool management MCP tools to the server.
func registerPoolTools(s *mcp.Server, client *proxmox.Client) {
	type listPoolsInput struct{}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_pools",
		Description: "List all resource pools defined in the cluster.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ listPoolsInput) (*mcp.CallToolResult, any, error) {
		pools, err := client.ListPools(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("list_pools: %w", err)
		}

		return jsonResult(pools)
	})

	type getPoolInput struct {
		PoolID string `json:"poolid" jsonschema:"ID of the pool to retrieve"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_pool",
		Description: "Get the full details of a resource pool including its member VMs, containers, and storage.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input getPoolInput) (*mcp.CallToolResult, any, error) {
		pool, err := client.GetPool(ctx, input.PoolID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_pool: %w", err)
		}

		return jsonResult(pool)
	})

	type createPoolInput struct {
		PoolID  string `json:"poolid"            jsonschema:"unique ID for the new pool"`
		Comment string `json:"comment,omitempty" jsonschema:"optional description of the pool"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_pool",
		Description: "Create a new resource pool in the cluster.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input createPoolInput) (*mcp.CallToolResult, any, error) {
		req := &proxmox.CreatePoolRequest{
			PoolID:  input.PoolID,
			Comment: input.Comment,
		}
		if err := client.CreatePool(ctx, req); err != nil {
			return nil, nil, fmt.Errorf("create_pool: %w", err)
		}

		return textResult("pool created: " + input.PoolID)
	})

	type updatePoolInput struct {
		PoolID  string `json:"poolid"            jsonschema:"ID of the pool to update"`
		Comment string `json:"comment,omitempty" jsonschema:"new description for the pool"`
		VMs     string `json:"vms,omitempty"     jsonschema:"comma-separated numeric VM/CT IDs to add (or remove when delete=true)"`
		Storage string `json:"storage,omitempty" jsonschema:"comma-separated storage names to add (or remove when delete=true)"`
		Delete  bool   `json:"delete,omitempty"  jsonschema:"when true, remove the listed VMs/storage from the pool instead of adding them"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_pool",
		Description: "Update a resource pool: change its comment, or add/remove member VMs, containers, and storage. Set delete=true to remove the listed members instead of adding them.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input updatePoolInput) (*mcp.CallToolResult, any, error) {
		req := &proxmox.UpdatePoolRequest{
			Comment: input.Comment,
			VMs:     input.VMs,
			Storage: input.Storage,
		}
		if input.Delete {
			v := 1
			req.Delete = &v
		}
		if err := client.UpdatePool(ctx, input.PoolID, req); err != nil {
			return nil, nil, fmt.Errorf("update_pool: %w", err)
		}

		return textResult("pool updated: " + input.PoolID)
	})
}
