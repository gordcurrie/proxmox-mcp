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

	// ── Firewall write tools ──────────────────────────────────────────────────

	type addVMFirewallRuleInput struct {
		Node    string `json:"node"              jsonschema:"node the VM is on"`
		VMID    int    `json:"vmid"              jsonschema:"numeric VM ID"`
		Type    string `json:"type"              jsonschema:"rule direction: in or out"`
		Action  string `json:"action"            jsonschema:"rule action: ACCEPT, DROP, or REJECT"`
		Proto   string `json:"proto,omitempty"   jsonschema:"protocol: tcp, udp, icmp, etc."`
		DPort   string `json:"dport,omitempty"   jsonschema:"destination port or range, e.g. 22 or 80:443"`
		Sport   string `json:"sport,omitempty"   jsonschema:"source port or range"`
		Source  string `json:"source,omitempty"  jsonschema:"source IP/CIDR/IPSet/alias (in-rules)"`
		Dest    string `json:"dest,omitempty"    jsonschema:"destination IP/CIDR/IPSet/alias (out-rules)"`
		IFace   string `json:"iface,omitempty"   jsonschema:"restrict rule to this network interface name"`
		Comment string `json:"comment,omitempty" jsonschema:"human-readable description of the rule"`
		Enable  bool   `json:"enable,omitempty"  jsonschema:"set to true to explicitly mark the rule as enabled (default: Proxmox enables rules automatically)"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "add_vm_firewall_rule",
		Description: "Add a firewall rule to a specific QEMU VM. The rule takes effect immediately — no task UPID is returned.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input addVMFirewallRuleInput) (*mcp.CallToolResult, any, error) {
		req := &proxmox.FirewallRuleRequest{
			Type:    input.Type,
			Action:  input.Action,
			Proto:   input.Proto,
			DPort:   input.DPort,
			Sport:   input.Sport,
			Source:  input.Source,
			Dest:    input.Dest,
			IFace:   input.IFace,
			Comment: input.Comment,
		}
		if input.Enable {
			v := 1
			req.Enable = &v
		}
		if err := client.AddVMFirewallRule(ctx, input.Node, input.VMID, req); err != nil {
			return nil, nil, fmt.Errorf("add_vm_firewall_rule: %w", err)
		}
		return textResult("firewall rule added to VM")
	})

	type deleteVMFirewallRuleInput struct {
		Node string `json:"node" jsonschema:"node the VM is on"`
		VMID int    `json:"vmid" jsonschema:"numeric VM ID"`
		Pos  int    `json:"pos"  jsonschema:"zero-based position of the rule to delete (use list_vm_firewall_rules to find positions)"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_vm_firewall_rule",
		Description: "Delete a firewall rule from a specific QEMU VM by its position. Use list_vm_firewall_rules to find rule positions (zero-based).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input deleteVMFirewallRuleInput) (*mcp.CallToolResult, any, error) {
		if err := client.DeleteVMFirewallRule(ctx, input.Node, input.VMID, input.Pos); err != nil {
			return nil, nil, fmt.Errorf("delete_vm_firewall_rule: %w", err)
		}
		return textResult("firewall rule deleted from VM")
	})

	type addCTFirewallRuleInput struct {
		Node    string `json:"node"              jsonschema:"node the container is on"`
		VMID    int    `json:"vmid"              jsonschema:"numeric container ID"`
		Type    string `json:"type"              jsonschema:"rule direction: in or out"`
		Action  string `json:"action"            jsonschema:"rule action: ACCEPT, DROP, or REJECT"`
		Proto   string `json:"proto,omitempty"   jsonschema:"protocol: tcp, udp, icmp, etc."`
		DPort   string `json:"dport,omitempty"   jsonschema:"destination port or range, e.g. 22 or 80:443"`
		Sport   string `json:"sport,omitempty"   jsonschema:"source port or range"`
		Source  string `json:"source,omitempty"  jsonschema:"source IP/CIDR/IPSet/alias (in-rules)"`
		Dest    string `json:"dest,omitempty"    jsonschema:"destination IP/CIDR/IPSet/alias (out-rules)"`
		IFace   string `json:"iface,omitempty"   jsonschema:"restrict rule to this network interface name"`
		Comment string `json:"comment,omitempty" jsonschema:"human-readable description of the rule"`
		Enable  bool   `json:"enable,omitempty"  jsonschema:"set to true to explicitly mark the rule as enabled (default: Proxmox enables rules automatically)"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "add_container_firewall_rule",
		Description: "Add a firewall rule to a specific LXC container. The rule takes effect immediately — no task UPID is returned.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input addCTFirewallRuleInput) (*mcp.CallToolResult, any, error) {
		req := &proxmox.FirewallRuleRequest{
			Type:    input.Type,
			Action:  input.Action,
			Proto:   input.Proto,
			DPort:   input.DPort,
			Sport:   input.Sport,
			Source:  input.Source,
			Dest:    input.Dest,
			IFace:   input.IFace,
			Comment: input.Comment,
		}
		if input.Enable {
			v := 1
			req.Enable = &v
		}
		if err := client.AddContainerFirewallRule(ctx, input.Node, input.VMID, req); err != nil {
			return nil, nil, fmt.Errorf("add_container_firewall_rule: %w", err)
		}
		return textResult("firewall rule added to container")
	})

	type deleteCTFirewallRuleInput struct {
		Node string `json:"node" jsonschema:"node the container is on"`
		VMID int    `json:"vmid" jsonschema:"numeric container ID"`
		Pos  int    `json:"pos"  jsonschema:"zero-based position of the rule to delete (use list_container_firewall_rules to find positions)"`
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_container_firewall_rule",
		Description: "Delete a firewall rule from a specific LXC container by its position. Use list_container_firewall_rules to find rule positions (zero-based).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input deleteCTFirewallRuleInput) (*mcp.CallToolResult, any, error) {
		if err := client.DeleteContainerFirewallRule(ctx, input.Node, input.VMID, input.Pos); err != nil {
			return nil, nil, fmt.Errorf("delete_container_firewall_rule: %w", err)
		}
		return textResult("firewall rule deleted from container")
	})
}
