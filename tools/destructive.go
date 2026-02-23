package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// destructiveHint is the *bool value used to set DestructiveHint on tools.
var destructiveHint = true

// registerDestructiveTools adds delete MCP tools to the server.
// These tools are only registered when PROXMOX_ALLOW_DESTRUCTIVE=true.
func registerDestructiveTools(s *mcp.Server, client *proxmox.Client) {
	type deleteVMInput struct {
		Node      string `json:"node"           jsonschema:"node the VM is on"`
		VMID      int    `json:"vmid"           jsonschema:"numeric VM ID"`
		Purge     bool   `json:"purge,omitempty" jsonschema:"also remove disk images and config (default false)"`
		Confirmed bool   `json:"confirmed"      jsonschema:"must be set to true to confirm deletion"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_vm",
		Description: "Permanently delete a QEMU VM. The VM must be stopped first. Set confirmed=true to proceed. Optionally set purge=true to also remove disk images. Returns the async task ID.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: &destructiveHint,
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input deleteVMInput) (*mcp.CallToolResult, any, error) {
		if !input.Confirmed {
			return nil, nil, errors.New("delete_vm: confirmed must be true to proceed with deletion")
		}
		upid, err := client.DeleteVM(ctx, input.Node, input.VMID, input.Purge)
		if err != nil {
			return nil, nil, fmt.Errorf("delete_vm: %w", err)
		}
		return taskResult(upid)
	})

	type deleteContainerInput struct {
		Node      string `json:"node"           jsonschema:"node the container is on"`
		VMID      int    `json:"vmid"           jsonschema:"numeric container ID"`
		Purge     bool   `json:"purge,omitempty" jsonschema:"also remove disk images and config (default false)"`
		Confirmed bool   `json:"confirmed"      jsonschema:"must be set to true to confirm deletion"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_container",
		Description: "Permanently delete an LXC container. The container must be stopped first. Set confirmed=true to proceed. Optionally set purge=true to also remove disk images. Returns the async task ID.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: &destructiveHint,
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input deleteContainerInput) (*mcp.CallToolResult, any, error) {
		if !input.Confirmed {
			return nil, nil, errors.New("delete_container: confirmed must be true to proceed with deletion")
		}
		upid, err := client.DeleteContainer(ctx, input.Node, input.VMID, input.Purge)
		if err != nil {
			return nil, nil, fmt.Errorf("delete_container: %w", err)
		}
		return taskResult(upid)
	})
}
