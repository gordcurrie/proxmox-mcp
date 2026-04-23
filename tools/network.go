package tools

import (
	"context"
	"fmt"

	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerNetworkTools adds node network MCP tools to the server.
func registerNetworkTools(s *mcp.Server, client proxmoxClient) {
	type listNodeNetworkInput struct {
		Node string `json:"node" jsonschema:"name of the node (e.g. pve)"`
		Type string `json:"type,omitempty" jsonschema:"optional interface type filter: bridge, bond, eth, alias, vlan, OVSBridge, OVSBond, OVSPort, OVSIntPort, any_bridge"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_node_network",
		Description: "List network interfaces configured on a Proxmox node. Optionally filter by interface type (e.g. bridge, eth).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input listNodeNetworkInput) (*mcp.CallToolResult, any, error) {
		ifaces, err := client.ListNodeNetwork(ctx, input.Node, input.Type)
		if err != nil {
			return errorResult(fmt.Errorf("list_node_network: %w", err))
		}
		return jsonResult(ifaces)
	})

	type getNodeNetworkInterfaceInput struct {
		Node  string `json:"node"  jsonschema:"name of the node (e.g. pve)"`
		Iface string `json:"iface" jsonschema:"interface name (e.g. vmbr0)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_node_network_interface",
		Description: "Get the configuration of a specific network interface on a Proxmox node.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input getNodeNetworkInterfaceInput) (*mcp.CallToolResult, any, error) {
		iface, err := client.GetNodeNetworkInterface(ctx, input.Node, input.Iface)
		if err != nil {
			return errorResult(fmt.Errorf("get_node_network_interface: %w", err))
		}
		return jsonResult(iface)
	})

	type createNodeNetworkInterfaceInput struct {
		Node        string `json:"node"                    jsonschema:"required,name of the node (e.g. pve)"`
		Iface       string `json:"iface"                   jsonschema:"required,interface name to create (e.g. vmbr1)"`
		Type        string `json:"type"                    jsonschema:"required,interface type: bridge, bond, eth, alias, vlan, OVSBridge, OVSBond, OVSPort, OVSIntPort"`
		Address     string `json:"address,omitempty"       jsonschema:"IPv4 address or CIDR (e.g. 192.168.1.10 or 192.168.1.10/24)"`
		Netmask     string `json:"netmask,omitempty"       jsonschema:"IPv4 subnet mask"`
		Gateway     string `json:"gateway,omitempty"       jsonschema:"default IPv4 gateway address"`
		Address6    string `json:"address6,omitempty"      jsonschema:"IPv6 address or CIDR (e.g. 2001:db8::1 or 2001:db8::1/64)"`
		Gateway6    string `json:"gateway6,omitempty"      jsonschema:"default IPv6 gateway address"`
		MTU         int    `json:"mtu,omitempty"           jsonschema:"maximum transmission unit in bytes"`
		Autostart   *int   `json:"autostart,omitempty"     jsonschema:"1 to start on boot, 0 for manual; omit to leave unchanged"`
		BridgePorts string `json:"bridge_ports,omitempty"  jsonschema:"space-separated list of bridge member ports (bridge type only)"`
		BridgeSTP   string `json:"bridge_stp,omitempty"    jsonschema:"spanning tree protocol: on or off (bridge type only)"`
		BridgeFD    *int   `json:"bridge_fd,omitempty"     jsonschema:"bridge forward delay in seconds (bridge type only); omit to leave unchanged"`
		BondMode    string `json:"bond_mode,omitempty"     jsonschema:"bond mode: active-backup, 802.3ad, etc. (bond type only)"`
		Slaves      string `json:"slaves,omitempty"        jsonschema:"space-separated bond member ports (bond type only)"`
		Comments    string `json:"comments,omitempty"      jsonschema:"free-text comments; may include post-up/pre-down routing lines for static routes"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_node_network_interface",
		Description: "Create a new network interface on a Proxmox node. Changes are staged until apply_node_network_changes is called.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input createNodeNetworkInterfaceInput) (*mcp.CallToolResult, any, error) {
		cfg := &proxmox.NetworkInterfaceConfig{
			Type:        input.Type,
			Address:     input.Address,
			Netmask:     input.Netmask,
			Gateway:     input.Gateway,
			Address6:    input.Address6,
			Gateway6:    input.Gateway6,
			MTU:         input.MTU,
			Autostart:   input.Autostart,
			BridgePorts: input.BridgePorts,
			BridgeSTP:   input.BridgeSTP,
			BridgeFD:    input.BridgeFD,
			BondMode:    input.BondMode,
			Slaves:      input.Slaves,
			Comments:    input.Comments,
		}
		result, err := client.CreateNodeNetworkInterface(ctx, input.Node, input.Iface, cfg)
		if err != nil {
			return errorResult(fmt.Errorf("create_node_network_interface: %w", err))
		}
		return jsonResult(result)
	})

	type updateNodeNetworkInterfaceInput struct {
		Node        string `json:"node"                    jsonschema:"required,name of the node (e.g. pve)"`
		Iface       string `json:"iface"                   jsonschema:"required,interface name to update (e.g. vmbr0)"`
		Type        string `json:"type"                    jsonschema:"required,interface type: bridge, bond, eth, alias, vlan, OVSBridge, OVSBond, OVSPort, OVSIntPort"`
		Address     string `json:"address,omitempty"       jsonschema:"IPv4 address or CIDR (e.g. 192.168.1.10 or 192.168.1.10/24)"`
		Netmask     string `json:"netmask,omitempty"       jsonschema:"IPv4 subnet mask"`
		Gateway     string `json:"gateway,omitempty"       jsonschema:"default IPv4 gateway address"`
		Address6    string `json:"address6,omitempty"      jsonschema:"IPv6 address or CIDR (e.g. 2001:db8::1 or 2001:db8::1/64)"`
		Gateway6    string `json:"gateway6,omitempty"      jsonschema:"default IPv6 gateway address"`
		MTU         int    `json:"mtu,omitempty"           jsonschema:"maximum transmission unit in bytes"`
		Autostart   *int   `json:"autostart,omitempty"     jsonschema:"1 to start on boot, 0 for manual; omit to leave unchanged"`
		BridgePorts string `json:"bridge_ports,omitempty"  jsonschema:"space-separated list of bridge member ports (bridge type only)"`
		BridgeSTP   string `json:"bridge_stp,omitempty"    jsonschema:"spanning tree protocol: on or off (bridge type only)"`
		BridgeFD    *int   `json:"bridge_fd,omitempty"     jsonschema:"bridge forward delay in seconds (bridge type only); omit to leave unchanged"`
		BondMode    string `json:"bond_mode,omitempty"     jsonschema:"bond mode: active-backup, 802.3ad, etc. (bond type only)"`
		Slaves      string `json:"slaves,omitempty"        jsonschema:"space-separated bond member ports (bond type only)"`
		Comments    string `json:"comments,omitempty"      jsonschema:"free-text comments; may include post-up/pre-down routing lines for static routes"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_node_network_interface",
		Description: "Update the configuration of an existing network interface on a Proxmox node. Changes are staged until apply_node_network_changes is called.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input updateNodeNetworkInterfaceInput) (*mcp.CallToolResult, any, error) {
		cfg := &proxmox.NetworkInterfaceConfig{
			Type:        input.Type,
			Address:     input.Address,
			Netmask:     input.Netmask,
			Gateway:     input.Gateway,
			Address6:    input.Address6,
			Gateway6:    input.Gateway6,
			MTU:         input.MTU,
			Autostart:   input.Autostart,
			BridgePorts: input.BridgePorts,
			BridgeSTP:   input.BridgeSTP,
			BridgeFD:    input.BridgeFD,
			BondMode:    input.BondMode,
			Slaves:      input.Slaves,
			Comments:    input.Comments,
		}
		if err := client.UpdateNodeNetworkInterface(ctx, input.Node, input.Iface, cfg); err != nil {
			return errorResult(fmt.Errorf("update_node_network_interface: %w", err))
		}
		return jsonResult(map[string]string{"node": input.Node, "iface": input.Iface, "status": "updated"})
	})

	type applyNodeNetworkChangesInput struct {
		Node string `json:"node" jsonschema:"required,name of the node (e.g. pve)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "apply_node_network_changes",
		Description: "Apply all staged network configuration changes on a Proxmox node, reloading the network stack. Must be called after create_node_network_interface, update_node_network_interface, or delete_node_network_interface to make changes take effect.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input applyNodeNetworkChangesInput) (*mcp.CallToolResult, any, error) {
		if err := client.ApplyNodeNetworkChanges(ctx, input.Node); err != nil {
			return errorResult(fmt.Errorf("apply_node_network_changes: %w", err))
		}
		return textResult("network changes applied on node " + input.Node)
	})
}
