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

// GetNodeJournal returns raw systemd journal entries for a node — the same
// data backing the Proxmox web UI's Syslog viewer. Useful for auditing
// SSH/PAM authentication activity (grep client-side for "sshd",
// "Failed password", "Invalid user", etc) or general troubleshooting.
//
// since and until are Unix timestamps; pass 0 to omit either bound.
// lastEntries limits the result to the last N entries; pass 0 to use the
// Proxmox API default. A negative lastEntries is an error.
func (c *Client) GetNodeJournal(ctx context.Context, node string, since, until int64, lastEntries int) ([]string, error) {
	if lastEntries < 0 {
		return nil, fmt.Errorf("lastEntries must be >= 0, got %d", lastEntries)
	}

	q := url.Values{}
	if since > 0 {
		q.Set("since", strconv.FormatInt(since, 10))
	}
	if until > 0 {
		q.Set("until", strconv.FormatInt(until, 10))
	}
	if lastEntries > 0 {
		q.Set("lastentries", strconv.Itoa(lastEntries))
	}

	path := "/nodes/" + url.PathEscape(node) + "/journal"
	if enc := q.Encode(); enc != "" {
		path += "?" + enc
	}

	var lines []string
	if err := c.get(ctx, path, &lines); err != nil {
		return nil, fmt.Errorf("getting journal for node %s: %w", node, err)
	}
	return lines, nil
}

// NodeCommand sends a power management command to a node.
// command must be "reboot" or "shutdown". The operation is irreversible and
// takes down the entire node, so callers must validate before invoking this.
func (c *Client) NodeCommand(ctx context.Context, node, command string) error {
	req := &NodeCommandRequest{Command: command}
	var result any
	if err := c.postWithBody(ctx, "/nodes/"+url.PathEscape(node)+"/status", req, &result); err != nil {
		return fmt.Errorf("sending command %q to node %s: %w", command, node, err)
	}
	return nil
}
