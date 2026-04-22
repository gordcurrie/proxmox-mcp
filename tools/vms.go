package tools

import (
	"context"
	"fmt"

	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerVMTools adds QEMU VM MCP tools to the server.
func registerVMTools(s *mcp.Server, client proxmoxClient) {
	type listVMsInput struct {
		Node string `json:"node" jsonschema:"name of the node to list VMs on"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_vms",
		Description: "List all QEMU virtual machines on a Proxmox node.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input listVMsInput) (*mcp.CallToolResult, any, error) {
		vms, err := client.ListVMs(ctx, input.Node)
		if err != nil {
			return errorResult(fmt.Errorf("list_vms: %w", err))
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
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input vmInput) (*mcp.CallToolResult, any, error) {
		status, err := client.GetVMStatus(ctx, input.Node, input.VMID)
		if err != nil {
			return errorResult(fmt.Errorf("get_vm_status: %w", err))
		}
		return jsonResult(status)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "start_vm",
		Description: "Start a QEMU VM. Returns the async task ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input vmInput) (*mcp.CallToolResult, any, error) {
		upid, err := client.StartVM(ctx, input.Node, input.VMID)
		if err != nil {
			return errorResult(fmt.Errorf("start_vm: %w", err))
		}
		return taskResult(upid)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "stop_vm",
		Description: "Hard stop a QEMU VM (immediate power off). Returns the async task ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input vmInput) (*mcp.CallToolResult, any, error) {
		upid, err := client.StopVM(ctx, input.Node, input.VMID)
		if err != nil {
			return errorResult(fmt.Errorf("stop_vm: %w", err))
		}
		return taskResult(upid)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "shutdown_vm",
		Description: "Gracefully shut down a QEMU VM via ACPI. Returns the async task ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input vmInput) (*mcp.CallToolResult, any, error) {
		upid, err := client.ShutdownVM(ctx, input.Node, input.VMID)
		if err != nil {
			return errorResult(fmt.Errorf("shutdown_vm: %w", err))
		}
		return taskResult(upid)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "reboot_vm",
		Description: "Reboot a QEMU VM. Returns the async task ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input vmInput) (*mcp.CallToolResult, any, error) {
		upid, err := client.RebootVM(ctx, input.Node, input.VMID)
		if err != nil {
			return errorResult(fmt.Errorf("reboot_vm: %w", err))
		}
		return taskResult(upid)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "suspend_vm",
		Description: "Suspend a QEMU VM. Returns the async task ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input vmInput) (*mcp.CallToolResult, any, error) {
		upid, err := client.SuspendVM(ctx, input.Node, input.VMID)
		if err != nil {
			return errorResult(fmt.Errorf("suspend_vm: %w", err))
		}
		return taskResult(upid)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "resume_vm",
		Description: "Resume a suspended QEMU VM. Returns the async task ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input vmInput) (*mcp.CallToolResult, any, error) {
		upid, err := client.ResumeVM(ctx, input.Node, input.VMID)
		if err != nil {
			return errorResult(fmt.Errorf("resume_vm: %w", err))
		}
		return taskResult(upid)
	})

	type createVMInput struct {
		Node   string `json:"node"           jsonschema:"node to create the VM on"`
		VMID   int    `json:"vmid"           jsonschema:"numeric VM ID (must not already exist)"`
		Name   string `json:"name,omitempty" jsonschema:"VM name"`
		Memory int    `json:"memory,omitempty" jsonschema:"memory in MB (e.g. 512)"`
		Cores  int    `json:"cores,omitempty"  jsonschema:"number of CPU cores"`
		ISO    string `json:"iso,omitempty"    jsonschema:"ISO drive in Proxmox format: storage:iso/file.iso,media=cdrom"`
		Disk   string `json:"disk,omitempty"   jsonschema:"primary disk in Proxmox format: storage:size_in_gb e.g. local-lvm:32"`
		Net0   string `json:"net0,omitempty"   jsonschema:"network device e.g. virtio,bridge=vmbr0"`
		Start  bool   `json:"start,omitempty"  jsonschema:"start the VM after creation"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_vm",
		Description: "Create a new QEMU VM. Returns the async task ID. Use get_task_status to poll for completion.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input createVMInput) (*mcp.CallToolResult, any, error) {
		start := 0
		if input.Start {
			start = 1
		}
		req := proxmox.CreateVMRequest{
			VMID:   input.VMID,
			Name:   input.Name,
			Memory: input.Memory,
			Cores:  input.Cores,
			IDE2:   input.ISO,
			SCSI0:  input.Disk,
			Net0:   input.Net0,
			Start:  start,
		}
		upid, err := client.CreateVM(ctx, input.Node, &req)
		if err != nil {
			return errorResult(fmt.Errorf("create_vm: %w", err))
		}
		return taskResult(upid)
	})

	type cloneVMInput struct {
		Node       string `json:"node"                  jsonschema:"node the source VM is on"`
		VMID       int    `json:"vmid"                  jsonschema:"source VM ID"`
		NewID      int    `json:"newid"                 jsonschema:"ID for the new VM"`
		Name       string `json:"name,omitempty"        jsonschema:"name for the new VM"`
		TargetNode string `json:"target_node,omitempty" jsonschema:"target node (defaults to source node)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "clone_vm",
		Description: "Clone a QEMU VM to a new VM ID. Returns the async task ID. Use get_task_status to poll for completion.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input cloneVMInput) (*mcp.CallToolResult, any, error) {
		req := proxmox.CloneVMRequest{
			NewID:  input.NewID,
			Name:   input.Name,
			Target: input.TargetNode,
		}
		upid, err := client.CloneVM(ctx, input.Node, input.VMID, &req)
		if err != nil {
			return errorResult(fmt.Errorf("clone_vm: %w", err))
		}
		return taskResult(upid)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_vm_config",
		Description: "Get the full configuration of a QEMU VM.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input vmInput) (*mcp.CallToolResult, any, error) {
		config, err := client.GetVMConfig(ctx, input.Node, input.VMID)
		if err != nil {
			return errorResult(fmt.Errorf("get_vm_config: %w", err))
		}
		return jsonResult(config)
	})

	type setVMConfigInput struct {
		Node        string `json:"node"                    jsonschema:"node the VM is on"`
		VMID        int    `json:"vmid"                    jsonschema:"numeric VM ID"`
		Name        string `json:"name,omitempty"          jsonschema:"VM name; omit to leave unchanged"`
		Memory      int    `json:"memory,omitempty"        jsonschema:"memory in MB; omit to leave unchanged"`
		Cores       int    `json:"cores,omitempty"         jsonschema:"number of CPU cores; omit to leave unchanged"`
		OnBoot      *bool  `json:"onboot,omitempty"        jsonschema:"start VM at boot; omit to leave unchanged"`
		Description string `json:"description,omitempty"   jsonschema:"VM description; omit to leave unchanged"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "set_vm_config",
		Description: "Update the configuration of a QEMU VM. Only supplied fields are changed; omitted fields are left as-is.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input setVMConfigInput) (*mcp.CallToolResult, any, error) {
		var onboot *int
		if input.OnBoot != nil {
			v := 0
			if *input.OnBoot {
				v = 1
			}
			onboot = &v
		}
		req := proxmox.SetVMConfigRequest{
			Name:        input.Name,
			Memory:      input.Memory,
			Cores:       input.Cores,
			OnBoot:      onboot,
			Description: input.Description,
		}
		if err := client.SetVMConfig(ctx, input.Node, input.VMID, &req); err != nil {
			return errorResult(fmt.Errorf("set_vm_config: %w", err))
		}
		return jsonResult(map[string]string{"status": "ok"})
	})

	type resizeVMDiskInput struct {
		Node string `json:"node" jsonschema:"node the VM is on"`
		VMID int    `json:"vmid" jsonschema:"numeric VM ID"`
		Disk string `json:"disk" jsonschema:"disk to resize, e.g. scsi0"`
		Size string `json:"size" jsonschema:"new size: absolute (e.g. 50G) or increment (e.g. +10G)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "resize_vm_disk",
		Description: "Resize a disk attached to a QEMU VM. Returns the async task ID. Use get_task_status to poll for completion.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input resizeVMDiskInput) (*mcp.CallToolResult, any, error) {
		req := proxmox.ResizeDiskRequest{Disk: input.Disk, Size: input.Size}
		upid, err := client.ResizeVMDisk(ctx, input.Node, input.VMID, &req)
		if err != nil {
			return errorResult(fmt.Errorf("resize_vm_disk: %w", err))
		}
		return taskResult(upid)
	})

	type migrateVMInput struct {
		Node   string `json:"node"             jsonschema:"source node the VM is currently on"`
		VMID   int    `json:"vmid"             jsonschema:"numeric VM ID"`
		Target string `json:"target"           jsonschema:"destination node name"`
		Online bool   `json:"online,omitempty" jsonschema:"true for live migration (no guest downtime when QEMU guest agent is running)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "migrate_vm",
		Description: "Migrate a QEMU VM to another node. Returns the async task ID. Use get_task_status to poll for completion.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input migrateVMInput) (*mcp.CallToolResult, any, error) {
		var online *int
		if input.Online {
			v := 1
			online = &v
		}
		req := proxmox.MigrateVMRequest{Target: input.Target, Online: online}
		upid, err := client.MigrateVM(ctx, input.Node, input.VMID, &req)
		if err != nil {
			return errorResult(fmt.Errorf("migrate_vm: %w", err))
		}
		return taskResult(upid)
	})

	type restoreVMInput struct {
		Node    string `json:"node"             jsonschema:"node to restore the VM on"`
		VMID    int    `json:"vmid"             jsonschema:"numeric VM ID to assign to the restored VM"`
		Archive string `json:"archive"          jsonschema:"backup volume ID, e.g. local:backup/vzdump-qemu-100-2024_01_01-00_00_00.vma.zst"`
		Storage string `json:"storage,omitempty" jsonschema:"target storage pool for restored disk images"`
		Start   bool   `json:"start,omitempty"  jsonschema:"true to start the VM immediately after restore"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "restore_vm",
		Description: "Restore a QEMU VM from a vzdump backup archive. Returns the async task ID — use get_task_status to poll for completion. Use list_backups to find available archive volume IDs.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input restoreVMInput) (*mcp.CallToolResult, any, error) {
		start := 0
		if input.Start {
			start = 1
		}
		req := proxmox.RestoreVMRequest{
			VMID:    input.VMID,
			Archive: input.Archive,
			Storage: input.Storage,
			Start:   start,
		}
		upid, err := client.RestoreVM(ctx, input.Node, &req)
		if err != nil {
			return errorResult(fmt.Errorf("restore_vm: %w", err))
		}
		return taskResult(upid)
	})

	type moveVMDiskInput struct {
		Node         string `json:"node"                    jsonschema:"node the VM is on"`
		VMID         int    `json:"vmid"                    jsonschema:"numeric VM ID"`
		Disk         string `json:"disk"                    jsonschema:"disk to move, e.g. scsi0"`
		Storage      string `json:"storage"                 jsonschema:"destination storage pool"`
		DeleteSource bool   `json:"delete_source,omitempty" jsonschema:"true to delete the source volume after a successful move (default false)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "move_vm_disk",
		Description: "Move a QEMU VM disk to a different storage pool. Returns the async task ID — use get_task_status to poll for completion. Optionally set delete_source=true to remove the original volume after the move.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input moveVMDiskInput) (*mcp.CallToolResult, any, error) {
		req := proxmox.MoveVMDiskRequest{
			Disk:    input.Disk,
			Storage: input.Storage,
		}
		if input.DeleteSource {
			v := 1
			req.DeleteSource = &v
		}
		upid, err := client.MoveVMDisk(ctx, input.Node, input.VMID, &req)
		if err != nil {
			return errorResult(fmt.Errorf("move_vm_disk: %w", err))
		}
		return taskResult(upid)
	})
}
