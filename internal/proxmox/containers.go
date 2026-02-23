package proxmox

import (
	"context"
	"fmt"
	"strconv"
)

// ListContainers returns all LXC containers on the specified node.
func (c *Client) ListContainers(ctx context.Context, node string) ([]Container, error) {
	var containers []Container
	if err := c.get(ctx, "/nodes/"+node+"/lxc", &containers); err != nil {
		return nil, fmt.Errorf("listing containers on node %s: %w", node, err)
	}
	return containers, nil
}

// GetContainerStatus returns detailed status and configuration for a specific
// LXC container. The returned map contains the full Proxmox API response for
// /nodes/{node}/lxc/{vmid}/status/current.
func (c *Client) GetContainerStatus(ctx context.Context, node string, vmid int) (map[string]any, error) {
	var status map[string]any
	path := "/nodes/" + node + "/lxc/" + strconv.Itoa(vmid) + "/status/current"
	if err := c.get(ctx, path, &status); err != nil {
		return nil, fmt.Errorf("getting status for container %d on node %s: %w", vmid, node, err)
	}
	return status, nil
}

// StartContainer starts an LXC container. It returns the UPID of the
// asynchronous task.
func (c *Client) StartContainer(ctx context.Context, node string, vmid int) (string, error) {
	var upid string
	path := "/nodes/" + node + "/lxc/" + strconv.Itoa(vmid) + "/status/start"
	if err := c.post(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("starting container %d on node %s: %w", vmid, node, err)
	}
	return upid, nil
}

// StopContainer stops an LXC container. It returns the UPID of the
// asynchronous task.
func (c *Client) StopContainer(ctx context.Context, node string, vmid int) (string, error) {
	var upid string
	path := "/nodes/" + node + "/lxc/" + strconv.Itoa(vmid) + "/status/stop"
	if err := c.post(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("stopping container %d on node %s: %w", vmid, node, err)
	}
	return upid, nil
}

// ShutdownContainer gracefully shuts down an LXC container via ACPI.
// It returns the UPID of the asynchronous task.
func (c *Client) ShutdownContainer(ctx context.Context, node string, vmid int) (string, error) {
	var upid string
	path := "/nodes/" + node + "/lxc/" + strconv.Itoa(vmid) + "/status/shutdown"
	if err := c.post(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("shutting down container %d on node %s: %w", vmid, node, err)
	}
	return upid, nil
}

// RebootContainer reboots an LXC container. It returns the UPID of the
// asynchronous task.
func (c *Client) RebootContainer(ctx context.Context, node string, vmid int) (string, error) {
	var upid string
	path := "/nodes/" + node + "/lxc/" + strconv.Itoa(vmid) + "/status/reboot"
	if err := c.post(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("rebooting container %d on node %s: %w", vmid, node, err)
	}
	return upid, nil
}
