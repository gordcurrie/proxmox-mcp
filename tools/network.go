package tools

import (
	"context"
	"fmt"

	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerNetworkTools adds node network MCP tools to the server.
func registerNetworkTools(s *mcp.Server, client *proxmox.Client) {
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
			return nil, nil, fmt.Errorf("list_node_network: %w", err)
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
			return nil, nil, fmt.Errorf("get_node_network_interface: %w", err)
		}
		return jsonResult(iface)
	})
}
