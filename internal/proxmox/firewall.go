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

// AddVMFirewallRule adds a firewall rule to the specified QEMU VM.
// The rule is applied synchronously — no task UPID is returned.
func (c *Client) AddVMFirewallRule(ctx context.Context, node string, vmid int, req *FirewallRuleRequest) error {
	path := fmt.Sprintf("/nodes/%s/qemu/%d/firewall/rules", url.PathEscape(node), vmid)
	if err := c.postWithBody(ctx, path, req, nil); err != nil {
		return fmt.Errorf("adding firewall rule to VM %d on node %s: %w", vmid, node, err)
	}

	return nil
}

// DeleteVMFirewallRule removes the firewall rule at position pos from the
// specified QEMU VM. Rule positions are zero-based.
func (c *Client) DeleteVMFirewallRule(ctx context.Context, node string, vmid, pos int) error {
	path := fmt.Sprintf("/nodes/%s/qemu/%d/firewall/rules/%d", url.PathEscape(node), vmid, pos)
	if err := c.delete(ctx, path, nil); err != nil {
		return fmt.Errorf("deleting firewall rule %d from VM %d on node %s: %w", pos, vmid, node, err)
	}

	return nil
}

// AddContainerFirewallRule adds a firewall rule to the specified LXC container.
// The rule is applied synchronously — no task UPID is returned.
func (c *Client) AddContainerFirewallRule(ctx context.Context, node string, vmid int, req *FirewallRuleRequest) error {
	path := fmt.Sprintf("/nodes/%s/lxc/%d/firewall/rules", url.PathEscape(node), vmid)
	if err := c.postWithBody(ctx, path, req, nil); err != nil {
		return fmt.Errorf("adding firewall rule to container %d on node %s: %w", vmid, node, err)
	}

	return nil
}

// DeleteContainerFirewallRule removes the firewall rule at position pos from
// the specified LXC container. Rule positions are zero-based.
func (c *Client) DeleteContainerFirewallRule(ctx context.Context, node string, vmid, pos int) error {
	path := fmt.Sprintf("/nodes/%s/lxc/%d/firewall/rules/%d", url.PathEscape(node), vmid, pos)
	if err := c.delete(ctx, path, nil); err != nil {
		return fmt.Errorf("deleting firewall rule %d from container %d on node %s: %w", pos, vmid, node, err)
	}

	return nil
}
