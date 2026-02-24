package proxmox

import (
	"context"
	"fmt"
	"net/url"
)

// ListNodes returns all nodes registered in the cluster.
func (c *Client) ListNodes(ctx context.Context) ([]Node, error) {
	var nodes []Node
	if err := c.get(ctx, "/nodes", &nodes); err != nil {
		return nil, fmt.Errorf("listing nodes: %w", err)
	}
	return nodes, nil
}

// GetNodeStatus returns detailed status information for the named node.
// The returned map contains the full Proxmox API response, including
// cpuinfo, memory, rootfs, loadavg, and more.
func (c *Client) GetNodeStatus(ctx context.Context, node string) (map[string]any, error) {
	var status map[string]any
	if err := c.get(ctx, "/nodes/"+url.PathEscape(node)+"/status", &status); err != nil {
		return nil, fmt.Errorf("getting status for node %s: %w", node, err)
	}
	return status, nil
}
