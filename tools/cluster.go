package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerClusterTools adds cluster-wide and task MCP tools to the server.
func registerClusterTools(s *mcp.Server, client proxmoxClient) {
	type listClusterResourcesInput struct {
		Type string `json:"type,omitempty" jsonschema:"optional resource type filter: vm | storage | node | sdn"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_cluster_resources",
		Description: "List all resources across the Proxmox cluster. Optionally filter by type: vm, storage, node, sdn.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input listClusterResourcesInput) (*mcp.CallToolResult, any, error) {
		resources, err := client.ListClusterResources(ctx, input.Type)
		if err != nil {
			return errorResult(fmt.Errorf("list_cluster_resources: %w", err))
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
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input getTaskStatusInput) (*mcp.CallToolResult, any, error) {
		status, err := client.GetTaskStatus(ctx, input.Node, input.UPID)
		if err != nil {
			return errorResult(fmt.Errorf("get_task_status: %w", err))
		}
		return jsonResult(status)
	})

	type getClusterStatusInput struct{}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_cluster_status",
		Description: "Get the cluster status and quorum information.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ getClusterStatusInput) (*mcp.CallToolResult, any, error) {
		status, err := client.GetClusterStatus(ctx)
		if err != nil {
			return errorResult(fmt.Errorf("get_cluster_status: %w", err))
		}
		return jsonResult(status)
	})

	type noInput struct{}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_ha_groups",
		Description: "List all HA (High Availability) groups defined in the cluster.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ noInput) (*mcp.CallToolResult, any, error) {
		groups, err := client.ListHAGroups(ctx)
		if err != nil {
			return errorResult(fmt.Errorf("list_ha_groups: %w", err))
		}
		return jsonResult(groups)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_ha_resources",
		Description: "List all HA-managed resources (VMs and containers) in the cluster.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ noInput) (*mcp.CallToolResult, any, error) {
		resources, err := client.ListHAResources(ctx)
		if err != nil {
			return errorResult(fmt.Errorf("list_ha_resources: %w", err))
		}
		return jsonResult(resources)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_ha_status",
		Description: "Get the current status of the HA manager, including quorum and per-node/per-resource state.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ noInput) (*mcp.CallToolResult, any, error) {
		status, err := client.GetHAStatus(ctx)
		if err != nil {
			return errorResult(fmt.Errorf("get_ha_status: %w", err))
		}
		return jsonResult(status)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_cluster_config_nodes",
		Description: "List the corosync nodelist for the cluster (node names, IDs, and ring addresses).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ noInput) (*mcp.CallToolResult, any, error) {
		nodes, err := client.ListClusterConfigNodes(ctx)
		if err != nil {
			return errorResult(fmt.Errorf("list_cluster_config_nodes: %w", err))
		}
		return jsonResult(nodes)
	})
}
