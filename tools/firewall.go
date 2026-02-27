package tools

import (
	"context"
	"fmt"

	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerFirewallTools adds read-only firewall MCP tools to the server.
func registerFirewallTools(s *mcp.Server, client *proxmox.Client) {
	type listClusterFirewallRulesInput struct{}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_cluster_firewall_rules",
		Description: "List all firewall rules defined at the datacenter (cluster) level.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ listClusterFirewallRulesInput) (*mcp.CallToolResult, any, error) {
		rules, err := client.ListClusterFirewallRules(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("list_cluster_firewall_rules: %w", err)
		}
		return jsonResult(rules)
	})

	type getClusterFirewallOptionsInput struct{}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_cluster_firewall_options",
		Description: "Get the firewall policy options for the datacenter (cluster) level, including default input/output policies and logging settings.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ getClusterFirewallOptionsInput) (*mcp.CallToolResult, any, error) {
		opts, err := client.GetClusterFirewallOptions(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("get_cluster_firewall_options: %w", err)
		}
		return jsonResult(opts)
	})

	type vmFirewallInput struct {
		Node string `json:"node" jsonschema:"node the VM is on"`
		VMID int    `json:"vmid" jsonschema:"numeric VM ID"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_vm_firewall_rules",
		Description: "List all firewall rules for a specific QEMU VM.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input vmFirewallInput) (*mcp.CallToolResult, any, error) {
		rules, err := client.ListVMFirewallRules(ctx, input.Node, input.VMID)
		if err != nil {
			return nil, nil, fmt.Errorf("list_vm_firewall_rules: %w", err)
		}
		return jsonResult(rules)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_vm_firewall_options",
		Description: "Get the firewall policy options for a specific QEMU VM, including per-VM enable state and default input/output policies.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input vmFirewallInput) (*mcp.CallToolResult, any, error) {
		opts, err := client.GetVMFirewallOptions(ctx, input.Node, input.VMID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_vm_firewall_options: %w", err)
		}
		return jsonResult(opts)
	})

	type ctFirewallInput struct {
		Node string `json:"node" jsonschema:"node the container is on"`
		VMID int    `json:"vmid" jsonschema:"numeric container ID"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_container_firewall_rules",
		Description: "List all firewall rules for a specific LXC container.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input ctFirewallInput) (*mcp.CallToolResult, any, error) {
		rules, err := client.ListContainerFirewallRules(ctx, input.Node, input.VMID)
		if err != nil {
			return nil, nil, fmt.Errorf("list_container_firewall_rules: %w", err)
		}
		return jsonResult(rules)
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_container_firewall_options",
		Description: "Get the firewall policy options for a specific LXC container, including per-container enable state and default input/output policies.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input ctFirewallInput) (*mcp.CallToolResult, any, error) {
		opts, err := client.GetContainerFirewallOptions(ctx, input.Node, input.VMID)
		if err != nil {
			return nil, nil, fmt.Errorf("get_container_firewall_options: %w", err)
		}
		return jsonResult(opts)
	})
}
