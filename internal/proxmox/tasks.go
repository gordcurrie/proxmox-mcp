package proxmox

import (
	"context"
	"fmt"
	"net/url"
)

// GetTaskStatus returns the current status of an asynchronous Proxmox task
// identified by its UPID. Poll this until Status is "stopped".
func (c *Client) GetTaskStatus(ctx context.Context, node, upid string) (*TaskStatus, error) {
	var status TaskStatus
	path := "/nodes/" + url.PathEscape(node) + "/tasks/" + url.PathEscape(upid) + "/status"
	if err := c.get(ctx, path, &status); err != nil {
		return nil, fmt.Errorf("getting task status for %s on node %s: %w", upid, node, err)
	}
	return &status, nil
}
