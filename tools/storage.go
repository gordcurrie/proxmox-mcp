package tools

import (
	"context"
	"fmt"

	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerStorageTools adds storage content MCP tools to the server.
func registerStorageTools(s *mcp.Server, client *proxmox.Client) {
	type listStorageContentInput struct {
		Node    string `json:"node"              jsonschema:"name of the node (e.g. pve)"`
		Storage string `json:"storage"           jsonschema:"name of the storage pool (e.g. local)"`
		Content string `json:"content,omitempty" jsonschema:"optional content type filter: iso, vztmpl, backup, images"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_storage_content",
		Description: "List the contents of a storage pool on a Proxmox node. Use the content filter to discover ISOs (iso), container templates (vztmpl), backups (backup), or disk images (images). Omit content to list all volumes.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input listStorageContentInput) (*mcp.CallToolResult, any, error) {
		items, err := client.ListStorageContent(ctx, input.Node, input.Storage, input.Content)
		if err != nil {
			return nil, nil, fmt.Errorf("list_storage_content: %w", err)
		}
		return jsonResult(items)
	})

	type getStorageContentInfoInput struct {
		Node    string `json:"node"    jsonschema:"name of the node (e.g. pve)"`
		Storage string `json:"storage" jsonschema:"name of the storage pool (e.g. local)"`
		Volume  string `json:"volume"  jsonschema:"volume ID, e.g. local:iso/debian-12.iso"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_storage_content_info",
		Description: "Get detailed information about a specific volume in a Proxmox storage pool. volume is the full volid, e.g. local:iso/debian-12.iso.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input getStorageContentInfoInput) (*mcp.CallToolResult, any, error) {
		info, err := client.GetStorageContentInfo(ctx, input.Node, input.Storage, input.Volume)
		if err != nil {
			return nil, nil, fmt.Errorf("get_storage_content_info: %w", err)
		}
		return jsonResult(info)
	})
}
