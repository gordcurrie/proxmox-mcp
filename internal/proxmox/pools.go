package proxmox

import (
	"context"
	"fmt"
	"net/url"
)

// ListPools returns all resource pools defined in the cluster.
// Only PoolID and Comment are populated; use GetPool to retrieve members.
func (c *Client) ListPools(ctx context.Context) ([]Pool, error) {
	var pools []Pool
	if err := c.get(ctx, "/pools", &pools); err != nil {
		return nil, fmt.Errorf("listing pools: %w", err)
	}

	return pools, nil
}

// GetPool returns the full details of a single resource pool including its
// member VMs, containers, and storage.
func (c *Client) GetPool(ctx context.Context, poolid string) (*Pool, error) {
	var pool Pool
	if err := c.get(ctx, "/pools/"+url.PathEscape(poolid), &pool); err != nil {
		return nil, fmt.Errorf("getting pool %q: %w", poolid, err)
	}

	return &pool, nil
}

// CreatePool creates a new resource pool with the given ID and optional comment.
func (c *Client) CreatePool(ctx context.Context, req *CreatePoolRequest) error {
	if err := c.postWithBody(ctx, "/pools", req, nil); err != nil {
		return fmt.Errorf("creating pool %q: %w", req.PoolID, err)
	}

	return nil
}

// UpdatePool updates the configuration of a resource pool.
// Set req.Delete to 1 to remove the listed VMs/storage instead of adding them.
func (c *Client) UpdatePool(ctx context.Context, poolid string, req *UpdatePoolRequest) error {
	if err := c.put(ctx, "/pools/"+url.PathEscape(poolid), req, nil); err != nil {
		return fmt.Errorf("updating pool %q: %w", poolid, err)
	}

	return nil
}

// DeletePool permanently removes a resource pool from the cluster.
// The pool must be empty before it can be deleted.
func (c *Client) DeletePool(ctx context.Context, poolid string) error {
	if err := c.delete(ctx, "/pools/"+url.PathEscape(poolid), nil); err != nil {
		return fmt.Errorf("deleting pool %q: %w", poolid, err)
	}

	return nil
}
