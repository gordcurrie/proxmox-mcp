package tools

import (
	"context"
	"fmt"

	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerSnapshotTools adds snapshot MCP tools to the server.
func registerSnapshotTools(s *mcp.Server, client *proxmox.Client) {
	type vmSnapInput struct {
		Node     string `json:"node" jsonschema:"node the VM is on"`
		VMID     int    `json:"vmid" jsonschema:"numeric VM ID"`
		Snapname string `json:"snapname" jsonschema:"snapshot name"`
	}

	type ctSnapInput struct {
		Node     string `json:"node" jsonschema:"node the container is on"`
		VMID     int    `json:"vmid" jsonschema:"numeric container ID"`
		Snapname string `json:"snapname" jsonschema:"snapshot name"`
	}

	// --- VM snapshot tools ---

	type listVMSnapsInput struct {
		Node string `json:"node" jsonschema:"node the VM is on"`
		VMID int    `json:"vmid" jsonschema:"numeric VM ID"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_vm_snapshots",
		Description: "List all snapshots for a QEMU VM.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input listVMSnapsInput) (*mcp.CallToolResult, any, error) {
		snaps, err := client.ListVMSnapshots(ctx, input.Node, input.VMID)
		if err != nil {
			return nil, nil, fmt.Errorf("list_vm_snapshots: %w", err)
		}
		return jsonResult(snaps)
	})

	type createVMSnapInput struct {
		Node        string `json:"node"        jsonschema:"node the VM is on"`
		VMID        int    `json:"vmid"        jsonschema:"numeric VM ID"`
		Name        string `json:"name"        jsonschema:"snapshot name"`
		Description string `json:"description" jsonschema:"optional snapshot description"`
		IncludeRAM  bool   `json:"include_ram" jsonschema:"include RAM state in snapshot"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_vm_snapshot",
		Description: "Create a snapshot of a QEMU VM. Returns the async task ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input createVMSnapInput) (*mcp.CallToolResult, any, error) {
		vmstate := 0
		if input.IncludeRAM {
			vmstate = 1
		}
		req := proxmox.CreateVMSnapshotRequest{
			Snapname:    input.Name,
			Description: input.Description,
			VMState:     vmstate,
		}
		upid, err := client.CreateVMSnapshot(ctx, input.Node, input.VMID, req)
		if err != nil {
			return nil, nil, fmt.Errorf("create_vm_snapshot: %w", err)
		}
		return taskResult(upid)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rollback_vm_snapshot",
		Description: "Roll back a QEMU VM to a snapshot. Returns the async task ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input vmSnapInput) (*mcp.CallToolResult, any, error) {
		upid, err := client.RollbackVMSnapshot(ctx, input.Node, input.VMID, input.Snapname)
		if err != nil {
			return nil, nil, fmt.Errorf("rollback_vm_snapshot: %w", err)
		}
		return taskResult(upid)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_vm_snapshot",
		Description: "Delete a snapshot of a QEMU VM. Returns the async task ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input vmSnapInput) (*mcp.CallToolResult, any, error) {
		upid, err := client.DeleteVMSnapshot(ctx, input.Node, input.VMID, input.Snapname)
		if err != nil {
			return nil, nil, fmt.Errorf("delete_vm_snapshot: %w", err)
		}
		return taskResult(upid)
	})

	// --- Container snapshot tools ---

	type listCTSnapsInput struct {
		Node string `json:"node" jsonschema:"node the container is on"`
		VMID int    `json:"vmid" jsonschema:"numeric container ID"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_container_snapshots",
		Description: "List all snapshots for an LXC container.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input listCTSnapsInput) (*mcp.CallToolResult, any, error) {
		snaps, err := client.ListContainerSnapshots(ctx, input.Node, input.VMID)
		if err != nil {
			return nil, nil, fmt.Errorf("list_container_snapshots: %w", err)
		}
		return jsonResult(snaps)
	})

	type createCTSnapInput struct {
		Node        string `json:"node"        jsonschema:"node the container is on"`
		VMID        int    `json:"vmid"        jsonschema:"numeric container ID"`
		Name        string `json:"name"        jsonschema:"snapshot name"`
		Description string `json:"description" jsonschema:"optional snapshot description"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_container_snapshot",
		Description: "Create a snapshot of an LXC container. Returns the async task ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input createCTSnapInput) (*mcp.CallToolResult, any, error) {
		req := proxmox.CreateContainerSnapshotRequest{
			Snapname:    input.Name,
			Description: input.Description,
		}
		upid, err := client.CreateContainerSnapshot(ctx, input.Node, input.VMID, req)
		if err != nil {
			return nil, nil, fmt.Errorf("create_container_snapshot: %w", err)
		}
		return taskResult(upid)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rollback_container_snapshot",
		Description: "Roll back an LXC container to a snapshot. Returns the async task ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input ctSnapInput) (*mcp.CallToolResult, any, error) {
		upid, err := client.RollbackContainerSnapshot(ctx, input.Node, input.VMID, input.Snapname)
		if err != nil {
			return nil, nil, fmt.Errorf("rollback_container_snapshot: %w", err)
		}
		return taskResult(upid)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_container_snapshot",
		Description: "Delete a snapshot of an LXC container. Returns the async task ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input ctSnapInput) (*mcp.CallToolResult, any, error) {
		upid, err := client.DeleteContainerSnapshot(ctx, input.Node, input.VMID, input.Snapname)
		if err != nil {
			return nil, nil, fmt.Errorf("delete_container_snapshot: %w", err)
		}
		return taskResult(upid)
	})
}
