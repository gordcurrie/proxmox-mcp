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

// ListHAGroups returns all HA (High Availability) groups defined in the cluster.
func (c *Client) ListHAGroups(ctx context.Context) ([]map[string]any, error) {
	var groups []map[string]any
	if err := c.get(ctx, "/cluster/ha/groups", &groups); err != nil {
		return nil, fmt.Errorf("listing HA groups: %w", err)
	}
	return groups, nil
}

// ListHAResources returns all HA-managed resources (VMs and containers) in the cluster.
func (c *Client) ListHAResources(ctx context.Context) ([]map[string]any, error) {
	var resources []map[string]any
	if err := c.get(ctx, "/cluster/ha/resources", &resources); err != nil {
		return nil, fmt.Errorf("listing HA resources: %w", err)
	}
	return resources, nil
}

// GetHAStatus returns the current status of the HA manager, including quorum
// and per-node/per-resource state.
func (c *Client) GetHAStatus(ctx context.Context) ([]map[string]any, error) {
	var status []map[string]any
	if err := c.get(ctx, "/cluster/ha/status/current", &status); err != nil {
		return nil, fmt.Errorf("getting HA status: %w", err)
	}
	return status, nil
}

// ListClusterConfigNodes returns the corosync nodelist for the cluster
// (node names, IDs, and ring addresses).
func (c *Client) ListClusterConfigNodes(ctx context.Context) ([]map[string]any, error) {
	var nodes []map[string]any
	if err := c.get(ctx, "/cluster/config/nodes", &nodes); err != nil {
		return nil, fmt.Errorf("listing cluster config nodes: %w", err)
	}
	return nodes, nil
}
