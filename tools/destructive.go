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

	type removeStorageInput struct {
		Storage   string `json:"storage"   jsonschema:"name of the storage definition to remove (e.g. pbs-store)"`
		Confirmed bool   `json:"confirmed" jsonschema:"must be set to true to confirm removal"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "remove_storage",
		Description: "Remove a storage definition from the Proxmox cluster. This only removes the Proxmox entry — it does not affect the underlying server or data. Set confirmed=true to proceed.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: &destructiveHint,
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input removeStorageInput) (*mcp.CallToolResult, any, error) {
		if !input.Confirmed {
			return nil, nil, errors.New("remove_storage: confirmed must be true to proceed with removal")
		}
		if err := client.RemoveStorage(ctx, input.Storage); err != nil {
			return nil, nil, fmt.Errorf("remove_storage: %w", err)
		}
		return jsonResult(map[string]string{"storage": input.Storage, "status": "removed"})
	})

	type deleteStorageContentInput struct {
		Node      string `json:"node"      jsonschema:"node the storage pool is on"`
		Storage   string `json:"storage"   jsonschema:"name of the storage pool (e.g. local)"`
		Volume    string `json:"volume"    jsonschema:"volume ID to delete, e.g. local:iso/debian-12.iso"`
		Confirmed bool   `json:"confirmed" jsonschema:"must be set to true to confirm deletion"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_storage_content",
		Description: "Permanently delete a volume from a Proxmox storage pool. Set confirmed=true to proceed. Returns the async task ID.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: &destructiveHint,
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input deleteStorageContentInput) (*mcp.CallToolResult, any, error) {
		if !input.Confirmed {
			return nil, nil, errors.New("delete_storage_content: confirmed must be true to proceed with deletion")
		}
		upid, err := client.DeleteStorageContent(ctx, input.Node, input.Storage, input.Volume)
		if err != nil {
			return nil, nil, fmt.Errorf("delete_storage_content: %w", err)
		}
		return taskResult(upid)
	})

	type nodeCommandInput struct {
		Node      string `json:"node"      jsonschema:"name of the node (e.g. pve)"`
		Confirmed bool   `json:"confirmed" jsonschema:"must be set to true to confirm the operation"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "reboot_node",
		Description: "Reboot an entire Proxmox node. This takes down all VMs and containers on the node. Set confirmed=true to proceed.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: &destructiveHint,
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input nodeCommandInput) (*mcp.CallToolResult, any, error) {
		if !input.Confirmed {
			return nil, nil, errors.New("reboot_node: confirmed must be true to proceed")
		}
		if err := client.NodeCommand(ctx, input.Node, "reboot"); err != nil {
			return nil, nil, fmt.Errorf("reboot_node: %w", err)
		}
		return textResult("Reboot command sent to node " + input.Node)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "shutdown_node",
		Description: "Shut down an entire Proxmox node. This takes down all VMs and containers on the node. Set confirmed=true to proceed.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: &destructiveHint,
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input nodeCommandInput) (*mcp.CallToolResult, any, error) {
		if !input.Confirmed {
			return nil, nil, errors.New("shutdown_node: confirmed must be true to proceed")
		}
		if err := client.NodeCommand(ctx, input.Node, "shutdown"); err != nil {
			return nil, nil, fmt.Errorf("shutdown_node: %w", err)
		}
		return textResult("Shutdown command sent to node " + input.Node)
	})

	type deletePoolInput struct {
		PoolID    string `json:"poolid"    jsonschema:"ID of the pool to delete"`
		Confirmed bool   `json:"confirmed" jsonschema:"must be set to true to confirm deletion"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_pool",
		Description: "Permanently delete a resource pool from the cluster. The pool must be empty first. Set confirmed=true to proceed.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: &destructiveHint,
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input deletePoolInput) (*mcp.CallToolResult, any, error) {
		if !input.Confirmed {
			return nil, nil, errors.New("delete_pool: confirmed must be true to proceed with deletion")
		}
		if err := client.DeletePool(ctx, input.PoolID); err != nil {
			return nil, nil, fmt.Errorf("delete_pool: %w", err)
		}
		return textResult("pool deleted: " + input.PoolID)
	})

	type deleteNodeNetworkInterfaceInput struct {
		Node      string `json:"node"      jsonschema:"required,name of the node (e.g. pve)"`
		Iface     string `json:"iface"     jsonschema:"required,interface name to delete (e.g. vmbr1)"`
		Confirmed bool   `json:"confirmed" jsonschema:"must be set to true to confirm deletion"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_node_network_interface",
		Description: "Remove a network interface from a Proxmox node. Changes are staged until apply_node_network_changes is called. Set confirmed=true to proceed.",
		Annotations: &mcp.ToolAnnotations{
			DestructiveHint: &destructiveHint,
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input deleteNodeNetworkInterfaceInput) (*mcp.CallToolResult, any, error) {
		if !input.Confirmed {
			return nil, nil, errors.New("delete_node_network_interface: confirmed must be true to proceed with deletion")
		}
		if err := client.DeleteNodeNetworkInterface(ctx, input.Node, input.Iface); err != nil {
			return nil, nil, fmt.Errorf("delete_node_network_interface: %w", err)
		}
		return jsonResult(map[string]string{
			"node":   input.Node,
			"iface":  input.Iface,
			"status": "deleted (staged — call apply_node_network_changes to make permanent)",
		})
	})
}
