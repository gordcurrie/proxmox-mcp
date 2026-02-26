package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// CreateBackup creates a backup of a VM or container via vzdump. It returns the
// UPID of the asynchronous task. Use get_task_status to poll for completion.
// Mode defaults to "snapshot" and compress defaults to "zstd" when omitted.
func (c *Client) CreateBackup(ctx context.Context, node string, req *CreateBackupRequest) (string, error) {
	if req == nil {
		return "", errors.New("CreateBackup: req must not be nil")
	}
	var upid string
	path := "/nodes/" + url.PathEscape(node) + "/vzdump"
	if err := c.postWithBody(ctx, path, req, &upid); err != nil {
		return "", fmt.Errorf("creating backup for vmid %d on node %s: %w", req.VMID, node, err)
	}
	return upid, nil
}

// ListBackups returns all backup volumes stored in the given storage pool on
// the specified node. It is a convenience wrapper around ListStorageContent
// filtered to content type "backup".
func (c *Client) ListBackups(ctx context.Context, node, storage string) ([]StorageContent, error) {
	backups, err := c.ListStorageContent(ctx, node, storage, "backup")
	if err != nil {
		return nil, fmt.Errorf("listing backups on node %s storage %s: %w", node, storage, err)
	}
	return backups, nil
}
