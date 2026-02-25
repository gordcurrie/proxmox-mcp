package proxmox

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
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

// ListNodeStorage returns all storage pools available on a node.
func (c *Client) ListNodeStorage(ctx context.Context, node string) ([]map[string]any, error) {
	var storage []map[string]any
	if err := c.get(ctx, "/nodes/"+url.PathEscape(node)+"/storage", &storage); err != nil {
		return nil, fmt.Errorf("listing storage on node %s: %w", node, err)
	}
	return storage, nil
}

// ListNodeTasks returns recent tasks for a node. If limit is greater than
// zero, at most that many tasks are returned. A negative limit is an error.
func (c *Client) ListNodeTasks(ctx context.Context, node string, limit int) ([]map[string]any, error) {
	if limit < 0 {
		return nil, fmt.Errorf("limit must be >= 0, got %d", limit)
	}
	path := "/nodes/" + url.PathEscape(node) + "/tasks"
	if limit > 0 {
		path += "?limit=" + strconv.Itoa(limit)
	}
	var tasks []map[string]any
	if err := c.get(ctx, path, &tasks); err != nil {
		return nil, fmt.Errorf("listing tasks on node %s: %w", node, err)
	}
	return tasks, nil
}

// GetNodeDisks returns the list of physical disks detected on a node.
func (c *Client) GetNodeDisks(ctx context.Context, node string) ([]map[string]any, error) {
	var disks []map[string]any
	if err := c.get(ctx, "/nodes/"+url.PathEscape(node)+"/disks/list", &disks); err != nil {
		return nil, fmt.Errorf("getting disks for node %s: %w", node, err)
	}
	return disks, nil
}
