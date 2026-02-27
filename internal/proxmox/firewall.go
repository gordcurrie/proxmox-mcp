package proxmox

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// ListClusterFirewallRules returns all firewall rules defined at the datacenter
// (cluster) level.
func (c *Client) ListClusterFirewallRules(ctx context.Context) ([]map[string]any, error) {
	var result []map[string]any
	if err := c.get(ctx, "/cluster/firewall/rules", &result); err != nil {
		return nil, fmt.Errorf("listing cluster firewall rules: %w", err)
	}

	return result, nil
}

// GetClusterFirewallOptions returns the firewall policy options for the
// datacenter (cluster) level.
func (c *Client) GetClusterFirewallOptions(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	if err := c.get(ctx, "/cluster/firewall/options", &result); err != nil {
		return nil, fmt.Errorf("getting cluster firewall options: %w", err)
	}

	return result, nil
}

// ListVMFirewallRules returns all firewall rules for the specified QEMU VM.
func (c *Client) ListVMFirewallRules(ctx context.Context, node string, vmid int) ([]map[string]any, error) {
	path := "/nodes/" + url.PathEscape(node) + "/qemu/" + strconv.Itoa(vmid) + "/firewall/rules"

	var result []map[string]any
	if err := c.get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("listing firewall rules for VM %d on node %s: %w", vmid, node, err)
	}

	return result, nil
}

// GetVMFirewallOptions returns the firewall policy options for the specified
// QEMU VM.
func (c *Client) GetVMFirewallOptions(ctx context.Context, node string, vmid int) (map[string]any, error) {
	path := "/nodes/" + url.PathEscape(node) + "/qemu/" + strconv.Itoa(vmid) + "/firewall/options"

	var result map[string]any
	if err := c.get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("getting firewall options for VM %d on node %s: %w", vmid, node, err)
	}

	return result, nil
}

// ListContainerFirewallRules returns all firewall rules for the specified LXC
// container.
func (c *Client) ListContainerFirewallRules(ctx context.Context, node string, vmid int) ([]map[string]any, error) {
	path := "/nodes/" + url.PathEscape(node) + "/lxc/" + strconv.Itoa(vmid) + "/firewall/rules"

	var result []map[string]any
	if err := c.get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("listing firewall rules for container %d on node %s: %w", vmid, node, err)
	}

	return result, nil
}

// GetContainerFirewallOptions returns the firewall policy options for the
// specified LXC container.
func (c *Client) GetContainerFirewallOptions(ctx context.Context, node string, vmid int) (map[string]any, error) {
	path := "/nodes/" + url.PathEscape(node) + "/lxc/" + strconv.Itoa(vmid) + "/firewall/options"

	var result map[string]any
	if err := c.get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("getting firewall options for container %d on node %s: %w", vmid, node, err)
	}

	return result, nil
}
