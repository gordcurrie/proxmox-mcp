package tools

import (
	"context"
	"fmt"

	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerBackupTools adds backup MCP tools to the server.
func registerBackupTools(s *mcp.Server, client proxmoxClient) {
	type createBackupInput struct {
		Node     string `json:"node"              jsonschema:"name of the node the VM or container is on"`
		VMID     int    `json:"vmid"              jsonschema:"numeric VM or container ID to back up"`
		Storage  string `json:"storage,omitempty" jsonschema:"target storage pool for the backup"`
		Mode     string `json:"mode,omitempty"    jsonschema:"backup mode: snapshot (default), suspend, or stop"`
		Compress string `json:"compress,omitempty" jsonschema:"compression: zstd (default), gzip, lzo, or 0 (none)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_backup",
		Description: "Create a backup of a QEMU VM or LXC container via vzdump. Returns the async task ID — use get_task_status to poll for completion. mode defaults to snapshot (no guest downtime when QEMU guest agent is running). compress defaults to zstd.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input createBackupInput) (*mcp.CallToolResult, any, error) {
		req := &proxmox.CreateBackupRequest{
			VMID:     input.VMID,
			Storage:  input.Storage,
			Mode:     input.Mode,
			Compress: input.Compress,
		}
		upid, err := client.CreateBackup(ctx, input.Node, req)
		if err != nil {
			return errorResult(fmt.Errorf("create_backup: %w", err))
		}
		return taskResult(upid)
	})

	type listBackupsInput struct {
		Node    string `json:"node"    jsonschema:"name of the node"`
		Storage string `json:"storage" jsonschema:"name of the storage pool to list backups from"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_backups",
		Description: "List all backup volumes stored in a Proxmox storage pool on a node.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input listBackupsInput) (*mcp.CallToolResult, any, error) {
		backups, err := client.ListBackups(ctx, input.Node, input.Storage)
		if err != nil {
			return errorResult(fmt.Errorf("list_backups: %w", err))
		}
		return jsonResult(backups)
	})
}
