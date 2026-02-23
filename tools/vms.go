package tools

import (
	"context"
	"fmt"

	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerVMTools adds QEMU VM MCP tools to the server.
func registerVMTools(s *mcp.Server, client *proxmox.Client) {
	type listVMsInput struct {
		Node string `json:"node" jsonschema:"name of the node to list VMs on"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_vms",
		Description: "List all QEMU virtual machines on a Proxmox node.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input listVMsInput) (*mcp.CallToolResult, any, error) {
		vms, err := client.ListVMs(ctx, input.Node)
		if err != nil {
			return nil, nil, fmt.Errorf("list_vms: %w", err)
		}
		return jsonResult(vms)
	})

	type vmInput struct {
		Node string `json:"node" jsonschema:"node the VM is on"`
		VMID int    `json:"vmid" jsonschema:"numeric VM ID"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_vm_status",
		Description: "Get the current status and configuration of a QEMU VM.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input vmInput) (*mcp.CallToolResult, any, error) {
		status, err := client.GetVMStatus(ctx, input.Node, input.VMID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_vm_status: %w", err)
		}
		return jsonResult(status)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "start_vm",
		Description: "Start a QEMU VM. Returns the async task ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input vmInput) (*mcp.CallToolResult, any, error) {
		upid, err := client.StartVM(ctx, input.Node, input.VMID)
		if err != nil {
			return nil, nil, fmt.Errorf("start_vm: %w", err)
		}
		return taskResult(upid)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "stop_vm",
		Description: "Hard stop a QEMU VM (immediate power off). Returns the async task ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input vmInput) (*mcp.CallToolResult, any, error) {
		upid, err := client.StopVM(ctx, input.Node, input.VMID)
		if err != nil {
			return nil, nil, fmt.Errorf("stop_vm: %w", err)
		}
		return taskResult(upid)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "shutdown_vm",
		Description: "Gracefully shut down a QEMU VM via ACPI. Returns the async task ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input vmInput) (*mcp.CallToolResult, any, error) {
		upid, err := client.ShutdownVM(ctx, input.Node, input.VMID)
		if err != nil {
			return nil, nil, fmt.Errorf("shutdown_vm: %w", err)
		}
		return taskResult(upid)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "reboot_vm",
		Description: "Reboot a QEMU VM. Returns the async task ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input vmInput) (*mcp.CallToolResult, any, error) {
		upid, err := client.RebootVM(ctx, input.Node, input.VMID)
		if err != nil {
			return nil, nil, fmt.Errorf("reboot_vm: %w", err)
		}
		return taskResult(upid)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "suspend_vm",
		Description: "Suspend a QEMU VM. Returns the async task ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input vmInput) (*mcp.CallToolResult, any, error) {
		upid, err := client.SuspendVM(ctx, input.Node, input.VMID)
		if err != nil {
			return nil, nil, fmt.Errorf("suspend_vm: %w", err)
		}
		return taskResult(upid)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "resume_vm",
		Description: "Resume a suspended QEMU VM. Returns the async task ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input vmInput) (*mcp.CallToolResult, any, error) {
		upid, err := client.ResumeVM(ctx, input.Node, input.VMID)
		if err != nil {
			return nil, nil, fmt.Errorf("resume_vm: %w", err)
		}
		return taskResult(upid)
	})
}
