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
