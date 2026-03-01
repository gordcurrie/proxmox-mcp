package proxmox

import (
	"context"
	"fmt"
	"net/url"
)

// ListStorages returns all storage definitions configured cluster-wide.
// The optional storageType parameter filters by type (e.g. "nfs", "pbs", "dir").
// Pass an empty string to return all storage definitions.
func (c *Client) ListStorages(ctx context.Context, storageType string) ([]map[string]any, error) {
	path := "/storage"
	if storageType != "" {
		path += "?type=" + url.QueryEscape(storageType)
	}

	var result []map[string]any
	if err := c.get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("listing storage definitions: %w", err)
	}

	return result, nil
}

// GetStorage returns the full configuration of a single storage definition.
func (c *Client) GetStorage(ctx context.Context, storage string) (map[string]any, error) {
	path := "/storage/" + url.PathEscape(storage)

	var result map[string]any
	if err := c.get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("getting storage definition %q: %w", storage, err)
	}

	return result, nil
}

// AddStorage creates a new cluster-wide storage definition and returns the
// resulting configuration.
func (c *Client) AddStorage(ctx context.Context, req *AddStorageRequest) (map[string]any, error) {
	if req == nil {
		return nil, fmt.Errorf("adding storage: request is nil")
	}

	var result map[string]any
	if err := c.postWithBody(ctx, "/storage", req, &result); err != nil {
		return nil, fmt.Errorf("adding storage %q: %w", req.Storage, err)
	}

	return result, nil
}

// UpdateStorage updates an existing cluster-wide storage definition.
// Only the fields set in req are changed; omitted fields are left as-is.
func (c *Client) UpdateStorage(ctx context.Context, storage string, req *UpdateStorageRequest) error {
	if req == nil {
		return fmt.Errorf("updating storage %q: request is nil", storage)
	}

	if err := c.put(ctx, "/storage/"+url.PathEscape(storage), req, nil); err != nil {
		return fmt.Errorf("updating storage %q: %w", storage, err)
	}

	return nil
}

// RemoveStorage permanently removes a storage definition from the cluster.
// This only removes the Proxmox storage entry — it does not affect the
// underlying storage server or any data stored on it.
func (c *Client) RemoveStorage(ctx context.Context, storage string) error {
	if err := c.delete(ctx, "/storage/"+url.PathEscape(storage), nil); err != nil {
		return fmt.Errorf("removing storage %q: %w", storage, err)
	}

	return nil
}
