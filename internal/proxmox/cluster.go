package proxmox

import (
	"context"
	"fmt"
	"net/url"
)

// ListClusterResources returns all resources across the cluster.
// resourceType is an optional filter; pass an empty string to get everything,
// or one of "vm", "storage", "node", "sdn" to filter by type.
func (c *Client) ListClusterResources(ctx context.Context, resourceType string) ([]ClusterResource, error) {
	path := "/cluster/resources"
	if resourceType != "" {
		path += "?type=" + url.QueryEscape(resourceType)
	}

	var resources []ClusterResource
	if err := c.get(ctx, path, &resources); err != nil {
		return nil, fmt.Errorf("listing cluster resources (type=%q): %w", resourceType, err)
	}
	return resources, nil
}

// GetClusterStatus returns the cluster status and quorum information.
// Each element in the returned slice describes a cluster member or quorum entry.
func (c *Client) GetClusterStatus(ctx context.Context) ([]map[string]any, error) {
	var status []map[string]any
	if err := c.get(ctx, "/cluster/status", &status); err != nil {
		return nil, fmt.Errorf("getting cluster status: %w", err)
	}
	return status, nil
}
