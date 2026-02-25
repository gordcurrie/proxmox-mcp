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

	type createContainerInput struct {
		Node       string                  `json:"node"                  jsonschema:"node to create the container on"`
		VMID       int                     `json:"vmid"                  jsonschema:"numeric container ID (must not already exist)"`
		OSTemplate string                  `json:"ostemplate"             jsonschema:"OS template in Proxmox format: storage:vztmpl/file.tar.zst"`
		Hostname   string                  `json:"hostname,omitempty"    jsonschema:"container hostname"`
		Memory     int                     `json:"memory,omitempty"      jsonschema:"memory in MB (e.g. 512)"`
		RootFS     string                  `json:"rootfs,omitempty"      jsonschema:"root filesystem in Proxmox format: storage:size_in_gb e.g. local-lvm:8"`
		Password   proxmox.SensitiveString `json:"password,omitempty"   jsonschema:"root password"`
		Net0       string                  `json:"net0,omitempty"        jsonschema:"network device e.g. name=eth0,bridge=vmbr0,dhcp=1"`
		Start      bool                    `json:"start,omitempty"       jsonschema:"start the container after creation"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_container",
		Description: "Create a new LXC container. Returns the async task ID. Use get_task_status to poll for completion.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input createContainerInput) (*mcp.CallToolResult, any, error) {
		start := 0
		if input.Start {
			start = 1
		}
		req := proxmox.CreateContainerRequest{
			VMID:       input.VMID,
			OSTemplate: input.OSTemplate,
			Hostname:   input.Hostname,
			Memory:     input.Memory,
			RootFS:     input.RootFS,
			Password:   input.Password,
			Net0:       input.Net0,
			Start:      start,
		}
		upid, err := client.CreateContainer(ctx, input.Node, &req)
		if err != nil {
			return nil, nil, fmt.Errorf("create_container: %w", err)
		}
		return taskResult(upid)
	})

	type cloneContainerInput struct {
		Node       string `json:"node"                  jsonschema:"node the source container is on"`
		VMID       int    `json:"vmid"                  jsonschema:"source container ID"`
		NewID      int    `json:"newid"                 jsonschema:"ID for the new container"`
		Hostname   string `json:"hostname,omitempty"    jsonschema:"hostname for the new container"`
		TargetNode string `json:"target_node,omitempty" jsonschema:"target node (defaults to source node)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "clone_container",
		Description: "Clone an LXC container to a new container ID. Returns the async task ID. Use get_task_status to poll for completion.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input cloneContainerInput) (*mcp.CallToolResult, any, error) {
		req := proxmox.CloneContainerRequest{
			NewID:    input.NewID,
			Hostname: input.Hostname,
			Target:   input.TargetNode,
		}
		upid, err := client.CloneContainer(ctx, input.Node, input.VMID, &req)
		if err != nil {
			return nil, nil, fmt.Errorf("clone_container: %w", err)
		}
		return taskResult(upid)
	})

	type containerConfigInput struct {
		Node string `json:"node" jsonschema:"node the container is on"`
		VMID int    `json:"vmid" jsonschema:"numeric container ID"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_container_config",
		Description: "Get the full configuration of an LXC container.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input containerConfigInput) (*mcp.CallToolResult, any, error) {
		config, err := client.GetContainerConfig(ctx, input.Node, input.VMID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_container_config: %w", err)
		}
		return jsonResult(config)
	})
}
