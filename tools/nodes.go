package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerNodeTools adds node MCP tools to the server.
func registerNodeTools(s *mcp.Server, client proxmoxClient) {
	type listNodesInput struct{}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_nodes",
		Description: "List all nodes in the Proxmox VE cluster.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ listNodesInput) (*mcp.CallToolResult, any, error) {
		nodes, err := client.ListNodes(ctx)
		if err != nil {
			return errorResult(fmt.Errorf("list_nodes: %w", err))
		}
		return jsonResult(nodes)
	})

	type nodeInput struct {
		Node string `json:"node" jsonschema:"name of the node (e.g. pve)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_node_status",
		Description: "Get detailed status of a specific Proxmox node.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input nodeInput) (*mcp.CallToolResult, any, error) {
		status, err := client.GetNodeStatus(ctx, input.Node)
		if err != nil {
			return errorResult(fmt.Errorf("get_node_status: %w", err))
		}
		return jsonResult(status)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_node_storage",
		Description: "List all storage pools available on a Proxmox node.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input nodeInput) (*mcp.CallToolResult, any, error) {
		storage, err := client.ListNodeStorage(ctx, input.Node)
		if err != nil {
			return errorResult(fmt.Errorf("list_node_storage: %w", err))
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
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input listNodeTasksInput) (*mcp.CallToolResult, any, error) {
		if input.Limit < 0 {
			return errorResult(fmt.Errorf("list_node_tasks: limit must be >= 0, got %d", input.Limit))
		}
		tasks, err := client.ListNodeTasks(ctx, input.Node, input.Limit)
		if err != nil {
			return errorResult(fmt.Errorf("list_node_tasks: %w", err))
		}
		return jsonResult(tasks)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_node_disks",
		Description: "List physical disks detected on a Proxmox node.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input nodeInput) (*mcp.CallToolResult, any, error) {
		disks, err := client.GetNodeDisks(ctx, input.Node)
		if err != nil {
			return errorResult(fmt.Errorf("get_node_disks: %w", err))
		}
		return jsonResult(disks)
	})

	type getNodeJournalInput struct {
		Node        string `json:"node"                    jsonschema:"name of the node"`
		Since       int64  `json:"since,omitempty"         jsonschema:"only return entries at or after this Unix timestamp (0 = no lower bound)"`
		Until       int64  `json:"until,omitempty"         jsonschema:"only return entries at or before this Unix timestamp (0 = no upper bound)"`
		LastEntries int    `json:"last_entries,omitempty"  jsonschema:"limit to the last N entries (0 = Proxmox API default)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_node_journal",
		Description: "Get raw systemd journal entries for a Proxmox node — the same data behind the web UI's Syslog viewer. Useful for auditing SSH/PAM authentication activity (look for 'sshd', 'Failed password', 'Invalid user', etc) or general troubleshooting.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input getNodeJournalInput) (*mcp.CallToolResult, any, error) {
		if input.LastEntries < 0 {
			return errorResult(fmt.Errorf("get_node_journal: last_entries must be >= 0, got %d", input.LastEntries))
		}
		lines, err := client.GetNodeJournal(ctx, input.Node, input.Since, input.Until, input.LastEntries)
		if err != nil {
			return errorResult(fmt.Errorf("get_node_journal: %w", err))
		}
		return jsonResult(lines)
	})
}
