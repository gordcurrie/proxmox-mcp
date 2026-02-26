package proxmox

import (
	"context"
	"fmt"
	"net/url"
)

// ListStorageContent returns the volumes stored in a storage pool on the
// specified node. The optional content parameter filters by type (e.g. "iso",
// "vztmpl", "backup", "images"). Pass an empty string to return all content.
func (c *Client) ListStorageContent(ctx context.Context, node, storage, content string) ([]StorageContent, error) {
	path := "/nodes/" + url.PathEscape(node) + "/storage/" + url.PathEscape(storage) + "/content"
	if content != "" {
		path += "?content=" + url.QueryEscape(content)
	}

	var result []StorageContent
	if err := c.get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("listing storage content on node %s storage %s: %w", node, storage, err)
	}

	return result, nil
}

// GetStorageContentInfo returns detailed information about a specific volume in
// a storage pool. volume is the full volid, e.g. "local:iso/debian-12.iso".
func (c *Client) GetStorageContentInfo(ctx context.Context, node, storage, volume string) (map[string]any, error) {
	path := "/nodes/" + url.PathEscape(node) + "/storage/" + url.PathEscape(storage) + "/content/" + url.PathEscape(volume)

	var result map[string]any
	if err := c.get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("getting storage content info for %s on node %s storage %s: %w", volume, node, storage, err)
	}

	return result, nil
}

// DeleteStorageContent deletes a volume from a storage pool. It returns the
// UPID of the asynchronous task. volume is the full volid,
// e.g. "local:iso/debian-12.iso".
func (c *Client) DeleteStorageContent(ctx context.Context, node, storage, volume string) (string, error) {
	path := "/nodes/" + url.PathEscape(node) + "/storage/" + url.PathEscape(storage) + "/content/" + url.PathEscape(volume)

	var upid string
	if err := c.delete(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("deleting storage content %s on node %s storage %s: %w", volume, node, storage, err)
	}

	return upid, nil
}
