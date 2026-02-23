package proxmox

import (
	"context"
	"fmt"
	"strconv"
)

// ListVMSnapshots returns all snapshots for the specified QEMU VM.
func (c *Client) ListVMSnapshots(ctx context.Context, node string, vmid int) ([]Snapshot, error) {
	var snaps []Snapshot
	path := "/nodes/" + node + "/qemu/" + strconv.Itoa(vmid) + "/snapshot"
	if err := c.get(ctx, path, &snaps); err != nil {
		return nil, fmt.Errorf("listing snapshots for VM %d on node %s: %w", vmid, node, err)
	}
	return snaps, nil
}

// CreateVMSnapshot creates a snapshot of the specified QEMU VM.
// It returns the UPID of the asynchronous task.
func (c *Client) CreateVMSnapshot(ctx context.Context, node string, vmid int, req CreateVMSnapshotRequest) (string, error) {
	var upid string
	path := "/nodes/" + node + "/qemu/" + strconv.Itoa(vmid) + "/snapshot"
	if err := c.postWithBody(ctx, path, req, &upid); err != nil {
		return "", fmt.Errorf("creating snapshot %q for VM %d on node %s: %w", req.Snapname, vmid, node, err)
	}
	return upid, nil
}

// RollbackVMSnapshot rolls back a QEMU VM to the specified snapshot.
// It returns the UPID of the asynchronous task.
func (c *Client) RollbackVMSnapshot(ctx context.Context, node string, vmid int, snapname string) (string, error) {
	var upid string
	path := "/nodes/" + node + "/qemu/" + strconv.Itoa(vmid) + "/snapshot/" + snapname + "/rollback"
	if err := c.post(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("rolling back VM %d on node %s to snapshot %q: %w", vmid, node, snapname, err)
	}
	return upid, nil
}

// DeleteVMSnapshot deletes a snapshot of the specified QEMU VM.
// It returns the UPID of the asynchronous task.
func (c *Client) DeleteVMSnapshot(ctx context.Context, node string, vmid int, snapname string) (string, error) {
	var upid string
	path := "/nodes/" + node + "/qemu/" + strconv.Itoa(vmid) + "/snapshot/" + snapname
	if err := c.delete(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("deleting snapshot %q from VM %d on node %s: %w", snapname, vmid, node, err)
	}
	return upid, nil
}

// ListContainerSnapshots returns all snapshots for the specified LXC container.
func (c *Client) ListContainerSnapshots(ctx context.Context, node string, vmid int) ([]Snapshot, error) {
	var snaps []Snapshot
	path := "/nodes/" + node + "/lxc/" + strconv.Itoa(vmid) + "/snapshot"
	if err := c.get(ctx, path, &snaps); err != nil {
		return nil, fmt.Errorf("listing snapshots for container %d on node %s: %w", vmid, node, err)
	}
	return snaps, nil
}

// CreateContainerSnapshot creates a snapshot of the specified LXC container.
// It returns the UPID of the asynchronous task.
func (c *Client) CreateContainerSnapshot(ctx context.Context, node string, vmid int, req CreateContainerSnapshotRequest) (string, error) {
	var upid string
	path := "/nodes/" + node + "/lxc/" + strconv.Itoa(vmid) + "/snapshot"
	if err := c.postWithBody(ctx, path, req, &upid); err != nil {
		return "", fmt.Errorf("creating snapshot %q for container %d on node %s: %w", req.Snapname, vmid, node, err)
	}
	return upid, nil
}

// RollbackContainerSnapshot rolls back an LXC container to the specified snapshot.
// It returns the UPID of the asynchronous task.
func (c *Client) RollbackContainerSnapshot(ctx context.Context, node string, vmid int, snapname string) (string, error) {
	var upid string
	path := "/nodes/" + node + "/lxc/" + strconv.Itoa(vmid) + "/snapshot/" + snapname + "/rollback"
	if err := c.post(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("rolling back container %d on node %s to snapshot %q: %w", vmid, node, snapname, err)
	}
	return upid, nil
}

// DeleteContainerSnapshot deletes a snapshot of the specified LXC container.
// It returns the UPID of the asynchronous task.
func (c *Client) DeleteContainerSnapshot(ctx context.Context, node string, vmid int, snapname string) (string, error) {
	var upid string
	path := "/nodes/" + node + "/lxc/" + strconv.Itoa(vmid) + "/snapshot/" + snapname
	if err := c.delete(ctx, path, &upid); err != nil {
		return "", fmt.Errorf("deleting snapshot %q from container %d on node %s: %w", snapname, vmid, node, err)
	}
	return upid, nil
}
