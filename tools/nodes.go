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
}
