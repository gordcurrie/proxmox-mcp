package tools

import (
	"context"
	"fmt"

	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerNodeTools adds node MCP tools to the server.
func registerNodeTools(s *mcp.Server, client *proxmox.Client) {
	type listNodesInput struct{}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_nodes",
		Description: "List all nodes in the Proxmox VE cluster.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ listNodesInput) (*mcp.CallToolResult, any, error) {
		nodes, err := client.ListNodes(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("list_nodes: %w", err)
		}
		return jsonResult(nodes)
	})

	type getNodeStatusInput struct {
		Node string `json:"node" jsonschema:"name of the node (e.g. pve)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_node_status",
		Description: "Get detailed status of a specific Proxmox node.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input getNodeStatusInput) (*mcp.CallToolResult, any, error) {
		status, err := client.GetNodeStatus(ctx, input.Node)
		if err != nil {
			return nil, nil, fmt.Errorf("get_node_status: %w", err)
		}
		return jsonResult(status)
	})

	type nodeInput struct {
		Node string `json:"node" jsonschema:"name of the node"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_node_storage",
		Description: "List all storage pools available on a Proxmox node.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input nodeInput) (*mcp.CallToolResult, any, error) {
		storage, err := client.ListNodeStorage(ctx, input.Node)
		if err != nil {
			return nil, nil, fmt.Errorf("list_node_storage: %w", err)
		}
		return jsonResult(storage)
	})

	type listNodeTasksInput struct {
		Node  string `json:"node" jsonschema:"name of the node"`
		Limit int    `json:"limit,omitempty" jsonschema:"maximum number of tasks to return (0 = no limit)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_node_tasks",
		Description: "List recent tasks for a Proxmox node.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input listNodeTasksInput) (*mcp.CallToolResult, any, error) {
		if input.Limit < 0 {
			return nil, nil, fmt.Errorf("list_node_tasks: limit must be >= 0, got %d", input.Limit)
		}
		tasks, err := client.ListNodeTasks(ctx, input.Node, input.Limit)
		if err != nil {
			return nil, nil, fmt.Errorf("list_node_tasks: %w", err)
		}
		return jsonResult(tasks)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_node_disks",
		Description: "List physical disks detected on a Proxmox node.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input nodeInput) (*mcp.CallToolResult, any, error) {
		disks, err := client.GetNodeDisks(ctx, input.Node)
		if err != nil {
			return nil, nil, fmt.Errorf("get_node_disks: %w", err)
		}
		return jsonResult(disks)
	})
}
