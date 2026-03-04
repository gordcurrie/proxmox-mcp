package proxmox

import (
	"context"
	"fmt"
	"net/url"
)

// ListNodeNetwork returns the network interfaces configured on a node.
// The optional networkType parameter filters by interface type (e.g. "bridge",
// "bond", "eth", "alias", "vlan", "OVSBridge", "OVSBond", "OVSPort",
// "OVSIntPort", "any_bridge"). Pass an empty string to return all interfaces.
func (c *Client) ListNodeNetwork(ctx context.Context, node, networkType string) ([]map[string]any, error) {
	path := "/nodes/" + url.PathEscape(node) + "/network"
	if networkType != "" {
		path += "?type=" + url.QueryEscape(networkType)
	}

	var result []map[string]any
	if err := c.get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("listing network interfaces on node %s: %w", node, err)
	}

	return result, nil
}

// GetNodeNetworkInterface returns the configuration of a single network
// interface on the specified node.
func (c *Client) GetNodeNetworkInterface(ctx context.Context, node, iface string) (map[string]any, error) {
	path := "/nodes/" + url.PathEscape(node) + "/network/" + url.PathEscape(iface)

	var result map[string]any
	if err := c.get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("getting network interface %s on node %s: %w", iface, node, err)
	}

	return result, nil
}

// NetworkInterfaceConfig holds the configurable fields for a Proxmox node
// network interface. Used for both creating and updating interfaces.
// Not all fields apply to all interface types:
//   - bridge: BridgePorts, BridgeSTP, BridgeFD
//   - bond:   BondMode, Slaves
type NetworkInterfaceConfig struct {
	// Type is required. Valid values: bridge, bond, eth, alias, vlan,
	// OVSBridge, OVSBond, OVSPort, OVSIntPort.
	Type        string `json:"type"`
	Address     string `json:"address,omitempty"`      // IPv4 address in CIDR or dotted-decimal
	Netmask     string `json:"netmask,omitempty"`      // subnet mask (alternative to CIDR in address)
	Gateway     string `json:"gateway,omitempty"`      // default IPv4 gateway
	Address6    string `json:"address6,omitempty"`     // IPv6 address in CIDR
	Gateway6    string `json:"gateway6,omitempty"`     // default IPv6 gateway
	MTU         int    `json:"mtu,omitempty"`          // maximum transmission unit
	Autostart   *int   `json:"autostart,omitempty"`    // 1 = start on boot, 0 = manual; nil = omit
	BridgePorts string `json:"bridge_ports,omitempty"` // space-separated list of member ports
	BridgeSTP   string `json:"bridge_stp,omitempty"`   // spanning tree protocol: "on" or "off"
	BridgeFD    *int   `json:"bridge_fd,omitempty"`    // bridge forward delay (seconds); nil = omit
	BondMode    string `json:"bond_mode,omitempty"`    // active-backup, 802.3ad, etc.
	Slaves      string `json:"slaves,omitempty"`       // space-separated list of bond member ports
	Comments    string `json:"comments,omitempty"`     // free-text; may include post-up/pre-down lines
}

// createNetworkInterfaceBody is the internal wire body for POST /nodes/{node}/network.
// It combines the interface name (required in the POST body) with the config fields.
type createNetworkInterfaceBody struct {
	Iface string `json:"iface"`
	NetworkInterfaceConfig
}

// CreateNodeNetworkInterface creates a new network interface on the specified
// node. The changes are staged (written to /etc/network/interfaces.new) and
// must be applied with ApplyNodeNetworkChanges to take effect.
func (c *Client) CreateNodeNetworkInterface(ctx context.Context, node, iface string, cfg *NetworkInterfaceConfig) (map[string]any, error) {
	if node == "" {
		return nil, fmt.Errorf("creating network interface on node: node must not be empty")
	}
	if iface == "" {
		return nil, fmt.Errorf("creating network interface on node %s: iface must not be empty", node)
	}
	if cfg == nil {
		return nil, fmt.Errorf("creating network interface %s on node %s: config must not be nil", iface, node)
	}
	if cfg.Type == "" {
		return nil, fmt.Errorf("creating network interface %s on node %s: type must not be empty", iface, node)
	}

	body := createNetworkInterfaceBody{Iface: iface, NetworkInterfaceConfig: *cfg}
	var result map[string]any
	if err := c.postWithBody(ctx, "/nodes/"+url.PathEscape(node)+"/network", body, &result); err != nil {
		return nil, fmt.Errorf("creating network interface %s on node %s: %w", iface, node, err)
	}
	return result, nil
}

// UpdateNodeNetworkInterface modifies the configuration of an existing network
// interface on the specified node. The changes are staged and must be applied
// with ApplyNodeNetworkChanges to take effect.
func (c *Client) UpdateNodeNetworkInterface(ctx context.Context, node, iface string, cfg *NetworkInterfaceConfig) error {
	if node == "" {
		return fmt.Errorf("updating network interface on node: node must not be empty")
	}
	if iface == "" {
		return fmt.Errorf("updating network interface on node %s: iface must not be empty", node)
	}
	if cfg == nil {
		return fmt.Errorf("updating network interface %s on node %s: config must not be nil", iface, node)
	}
	if cfg.Type == "" {
		return fmt.Errorf("updating network interface %s on node %s: type must not be empty", iface, node)
	}

	path := "/nodes/" + url.PathEscape(node) + "/network/" + url.PathEscape(iface)
	if err := c.put(ctx, path, cfg, nil); err != nil {
		return fmt.Errorf("updating network interface %s on node %s: %w", iface, node, err)
	}
	return nil
}

// ApplyNodeNetworkChanges applies all staged network configuration changes on
// the specified node, reloading the network stack. This is equivalent to
// clicking "Apply Configuration" in the Proxmox web UI.
// Note: put always marshals the body, so the wire request carries {} as the
// JSON body. Proxmox accepts this form for the apply endpoint.
func (c *Client) ApplyNodeNetworkChanges(ctx context.Context, node string) error {
	if node == "" {
		return fmt.Errorf("applying network changes: node must not be empty")
	}
	if err := c.put(ctx, "/nodes/"+url.PathEscape(node)+"/network", struct{}{}, nil); err != nil {
		return fmt.Errorf("applying network changes on node %s: %w", node, err)
	}
	return nil
}

// DeleteNodeNetworkInterface removes an interface from the specified node's
// network configuration. Changes are staged and must be applied with
// ApplyNodeNetworkChanges to take effect.
func (c *Client) DeleteNodeNetworkInterface(ctx context.Context, node, iface string) error {
	if node == "" {
		return fmt.Errorf("deleting network interface on node: node must not be empty")
	}
	if iface == "" {
		return fmt.Errorf("deleting network interface on node %s: iface must not be empty", node)
	}
	path := "/nodes/" + url.PathEscape(node) + "/network/" + url.PathEscape(iface)
	if err := c.delete(ctx, path, nil); err != nil {
		return fmt.Errorf("deleting network interface %s on node %s: %w", iface, node, err)
	}
	return nil
}
