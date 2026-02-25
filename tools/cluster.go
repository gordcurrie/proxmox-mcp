package tools

import (
	"context"
	"fmt"

	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerClusterTools adds cluster-wide and task MCP tools to the server.
func registerClusterTools(s *mcp.Server, client *proxmox.Client) {
	type listClusterResourcesInput struct {
		Type string `json:"type,omitempty" jsonschema:"optional resource type filter: vm | storage | node | sdn"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_cluster_resources",
		Description: "List all resources across the Proxmox cluster. Optionally filter by type: vm, storage, node, sdn.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input listClusterResourcesInput) (*mcp.CallToolResult, any, error) {
		resources, err := client.ListClusterResources(ctx, input.Type)
		if err != nil {
			return nil, nil, fmt.Errorf("list_cluster_resources: %w", err)
		}
		return jsonResult(resources)
	})

	type getTaskStatusInput struct {
		Node string `json:"node" jsonschema:"node the task is running on"`
		UPID string `json:"upid" jsonschema:"unique task ID (UPID) returned by a lifecycle operation"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_task_status",
		Description: "Get the status of an async Proxmox task by UPID. Poll until status is stopped.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input getTaskStatusInput) (*mcp.CallToolResult, any, error) {
		status, err := client.GetTaskStatus(ctx, input.Node, input.UPID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_task_status: %w", err)
		}
		return jsonResult(status)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_cluster_status",
		Description: "Get the cluster status and quorum information.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		status, err := client.GetClusterStatus(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("get_cluster_status: %w", err)
		}
		return jsonResult(status)
	})
}
