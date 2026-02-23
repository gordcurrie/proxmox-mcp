package tools

import (
	"context"
	"fmt"

	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerContainerTools adds LXC container MCP tools to the server.
func registerContainerTools(s *mcp.Server, client *proxmox.Client) {
	type listContainersInput struct {
		Node string `json:"node" jsonschema:"name of the node to list containers on"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_containers",
		Description: "List all LXC containers on a Proxmox node.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input listContainersInput) (*mcp.CallToolResult, any, error) {
		containers, err := client.ListContainers(ctx, input.Node)
		if err != nil {
			return nil, nil, fmt.Errorf("list_containers: %w", err)
		}
		return jsonResult(containers)
	})

	type containerInput struct {
		Node string `json:"node" jsonschema:"node the container is on"`
		VMID int    `json:"vmid" jsonschema:"numeric container ID"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_container_status",
		Description: "Get the current status of an LXC container.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input containerInput) (*mcp.CallToolResult, any, error) {
		status, err := client.GetContainerStatus(ctx, input.Node, input.VMID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_container_status: %w", err)
		}
		return jsonResult(status)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "start_container",
		Description: "Start an LXC container. Returns the async task ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input containerInput) (*mcp.CallToolResult, any, error) {
		upid, err := client.StartContainer(ctx, input.Node, input.VMID)
		if err != nil {
			return nil, nil, fmt.Errorf("start_container: %w", err)
		}
		return taskResult(upid)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "stop_container",
		Description: "Stop an LXC container. Returns the async task ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input containerInput) (*mcp.CallToolResult, any, error) {
		upid, err := client.StopContainer(ctx, input.Node, input.VMID)
		if err != nil {
			return nil, nil, fmt.Errorf("stop_container: %w", err)
		}
		return taskResult(upid)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "shutdown_container",
		Description: "Gracefully shut down an LXC container via ACPI. Returns the async task ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input containerInput) (*mcp.CallToolResult, any, error) {
		upid, err := client.ShutdownContainer(ctx, input.Node, input.VMID)
		if err != nil {
			return nil, nil, fmt.Errorf("shutdown_container: %w", err)
		}
		return taskResult(upid)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "reboot_container",
		Description: "Reboot an LXC container. Returns the async task ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input containerInput) (*mcp.CallToolResult, any, error) {
		upid, err := client.RebootContainer(ctx, input.Node, input.VMID)
		if err != nil {
			return nil, nil, fmt.Errorf("reboot_container: %w", err)
		}
		return taskResult(upid)
	})
}
