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
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
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
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
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

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_container_config",
		Description: "Get the full configuration of an LXC container.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input containerInput) (*mcp.CallToolResult, any, error) {
		config, err := client.GetContainerConfig(ctx, input.Node, input.VMID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_container_config: %w", err)
		}
		return jsonResult(config)
	})

	type setContainerConfigInput struct {
		Node        string `json:"node"                  jsonschema:"node the container is on"`
		VMID        int    `json:"vmid"                  jsonschema:"numeric container ID"`
		Hostname    string `json:"hostname,omitempty"    jsonschema:"container hostname; omit to leave unchanged"`
		Memory      int    `json:"memory,omitempty"      jsonschema:"memory in MB; omit to leave unchanged"`
		Swap        *int   `json:"swap,omitempty"        jsonschema:"swap in MB; omit to leave unchanged; 0 disables swap"`
		OnBoot      *bool  `json:"onboot,omitempty"      jsonschema:"start container at boot; omit to leave unchanged"`
		Description string `json:"description,omitempty" jsonschema:"container description; omit to leave unchanged"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "set_container_config",
		Description: "Update the configuration of an LXC container. Only supplied fields are changed; omitted fields are left as-is.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input setContainerConfigInput) (*mcp.CallToolResult, any, error) {
		var onboot *int
		if input.OnBoot != nil {
			v := 0
			if *input.OnBoot {
				v = 1
			}
			onboot = &v
		}
		req := proxmox.SetContainerConfigRequest{
			Hostname:    input.Hostname,
			Memory:      input.Memory,
			Swap:        input.Swap,
			OnBoot:      onboot,
			Description: input.Description,
		}
		if err := client.SetContainerConfig(ctx, input.Node, input.VMID, &req); err != nil {
			return nil, nil, fmt.Errorf("set_container_config: %w", err)
		}
		return jsonResult(map[string]string{"status": "ok"})
	})

	type resizeContainerDiskInput struct {
		Node string `json:"node" jsonschema:"node the container is on"`
		VMID int    `json:"vmid" jsonschema:"numeric container ID"`
		Disk string `json:"disk" jsonschema:"disk to resize, e.g. rootfs"`
		Size string `json:"size" jsonschema:"new size: absolute (e.g. 10G) or increment (e.g. +5G)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "resize_container_disk",
		Description: "Resize a disk attached to an LXC container. Returns the async task ID. Use get_task_status to poll for completion.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input resizeContainerDiskInput) (*mcp.CallToolResult, any, error) {
		req := proxmox.ResizeDiskRequest{Disk: input.Disk, Size: input.Size}
		upid, err := client.ResizeContainerDisk(ctx, input.Node, input.VMID, &req)
		if err != nil {
			return nil, nil, fmt.Errorf("resize_container_disk: %w", err)
		}
		return taskResult(upid)
	})

	type migrateContainerInput struct {
		Node    string `json:"node"              jsonschema:"source node the container is currently on"`
		VMID    int    `json:"vmid"              jsonschema:"numeric container ID"`
		Target  string `json:"target"            jsonschema:"destination node name"`
		Restart bool   `json:"restart,omitempty" jsonschema:"true to stop, migrate, and restart the container on the target node"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "migrate_container",
		Description: "Migrate an LXC container to another node. Returns the async task ID. Use get_task_status to poll for completion.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input migrateContainerInput) (*mcp.CallToolResult, any, error) {
		var restart *int
		if input.Restart {
			v := 1
			restart = &v
		}
		req := proxmox.MigrateContainerRequest{Target: input.Target, Restart: restart}
		upid, err := client.MigrateContainer(ctx, input.Node, input.VMID, &req)
		if err != nil {
			return nil, nil, fmt.Errorf("migrate_container: %w", err)
		}
		return taskResult(upid)
	})

	type restoreContainerInput struct {
		Node     string `json:"node"              jsonschema:"node to restore the container on"`
		VMID     int    `json:"vmid"              jsonschema:"numeric container ID to assign to the restored container"`
		Archive  string `json:"archive"           jsonschema:"backup volume ID, e.g. local:backup/vzdump-lxc-200-2024_01_01-00_00_00.tar.zst"`
		Storage  string `json:"storage,omitempty" jsonschema:"target storage pool for the restored rootfs"`
		Hostname string `json:"hostname,omitempty" jsonschema:"override hostname after restore"`
		Start    bool   `json:"start,omitempty"   jsonschema:"true to start the container immediately after restore"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "restore_container",
		Description: "Restore an LXC container from a vzdump backup archive. Returns the async task ID — use get_task_status to poll for completion. Use list_backups to find available archive volume IDs.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input restoreContainerInput) (*mcp.CallToolResult, any, error) {
		start := 0
		if input.Start {
			start = 1
		}
		req := proxmox.RestoreContainerRequest{
			VMID:     input.VMID,
			Archive:  input.Archive,
			Storage:  input.Storage,
			Hostname: input.Hostname,
			Start:    start,
		}
		upid, err := client.RestoreContainer(ctx, input.Node, &req)
		if err != nil {
			return nil, nil, fmt.Errorf("restore_container: %w", err)
		}
		return taskResult(upid)
	})
}
