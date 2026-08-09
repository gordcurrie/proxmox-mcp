package tools

import (
	"context"

	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
)

// mockProxmoxClient is a test double for proxmoxClient. Each method delegates
// to the corresponding *Fn field if set; otherwise returns zero value + nil error.
// This allows individual tests to stub only the methods they care about.
type mockProxmoxClient struct {
	// Nodes
	listNodesFn       func(context.Context) ([]proxmox.Node, error)
	getNodeStatusFn   func(context.Context, string) (map[string]any, error)
	listNodeStorageFn func(context.Context, string) ([]map[string]any, error)
	listNodeTasksFn   func(context.Context, string, int) ([]map[string]any, error)
	getNodeDisksFn    func(context.Context, string) ([]map[string]any, error)
	getDiskSMARTFn    func(context.Context, string, string) (map[string]any, error)
	listZFSPoolsFn    func(context.Context, string) ([]map[string]any, error)
	getZFSPoolFn      func(context.Context, string, string) (map[string]any, error)
	getNodeJournalFn  func(context.Context, string, int64, int64, int) ([]string, error)
	nodeCommandFn     func(context.Context, string, string) error

	// VMs
	listVMsFn      func(context.Context, string) ([]proxmox.VM, error)
	getVMStatusFn  func(context.Context, string, int) (map[string]any, error)
	startVMFn      func(context.Context, string, int) (string, error)
	stopVMFn       func(context.Context, string, int) (string, error)
	shutdownVMFn   func(context.Context, string, int) (string, error)
	rebootVMFn     func(context.Context, string, int) (string, error)
	suspendVMFn    func(context.Context, string, int) (string, error)
	resumeVMFn     func(context.Context, string, int) (string, error)
	createVMFn     func(context.Context, string, *proxmox.CreateVMRequest) (string, error)
	cloneVMFn      func(context.Context, string, int, *proxmox.CloneVMRequest) (string, error)
	getVMConfigFn  func(context.Context, string, int) (map[string]any, error)
	setVMConfigFn  func(context.Context, string, int, *proxmox.SetVMConfigRequest) error
	resizeVMDiskFn func(context.Context, string, int, *proxmox.ResizeDiskRequest) (string, error)
	migrateVMFn    func(context.Context, string, int, *proxmox.MigrateVMRequest) (string, error)
	restoreVMFn    func(context.Context, string, *proxmox.RestoreVMRequest) (string, error)
	moveVMDiskFn   func(context.Context, string, int, *proxmox.MoveVMDiskRequest) (string, error)
	deleteVMFn     func(context.Context, string, int, bool) (string, error)

	// Containers
	listContainersFn      func(context.Context, string) ([]proxmox.Container, error)
	getContainerStatusFn  func(context.Context, string, int) (map[string]any, error)
	startContainerFn      func(context.Context, string, int) (string, error)
	stopContainerFn       func(context.Context, string, int) (string, error)
	shutdownContainerFn   func(context.Context, string, int) (string, error)
	rebootContainerFn     func(context.Context, string, int) (string, error)
	createContainerFn     func(context.Context, string, *proxmox.CreateContainerRequest) (string, error)
	cloneContainerFn      func(context.Context, string, int, *proxmox.CloneContainerRequest) (string, error)
	getContainerConfigFn  func(context.Context, string, int) (map[string]any, error)
	setContainerConfigFn  func(context.Context, string, int, *proxmox.SetContainerConfigRequest) error
	resizeContainerDiskFn func(context.Context, string, int, *proxmox.ResizeDiskRequest) (string, error)
	migrateContainerFn    func(context.Context, string, int, *proxmox.MigrateContainerRequest) (string, error)
	restoreContainerFn    func(context.Context, string, *proxmox.RestoreContainerRequest) (string, error)
	deleteContainerFn     func(context.Context, string, int, bool) (string, error)

	// Cluster
	listClusterResourcesFn func(context.Context, string) ([]proxmox.ClusterResource, error)
	getTaskStatusFn        func(context.Context, string, string) (*proxmox.TaskStatus, error)
	getClusterStatusFn     func(context.Context) ([]map[string]any, error)

	// Snapshots
	listVMSnapshotsFn           func(context.Context, string, int) ([]proxmox.Snapshot, error)
	createVMSnapshotFn          func(context.Context, string, int, proxmox.CreateVMSnapshotRequest) (string, error)
	rollbackVMSnapshotFn        func(context.Context, string, int, string) (string, error)
	deleteVMSnapshotFn          func(context.Context, string, int, string) (string, error)
	listContainerSnapshotsFn    func(context.Context, string, int) ([]proxmox.Snapshot, error)
	createContainerSnapshotFn   func(context.Context, string, int, proxmox.CreateContainerSnapshotRequest) (string, error)
	rollbackContainerSnapshotFn func(context.Context, string, int, string) (string, error)
	deleteContainerSnapshotFn   func(context.Context, string, int, string) (string, error)

	// Storage content
	listStorageContentFn    func(context.Context, string, string, string) ([]proxmox.StorageContent, error)
	getStorageContentInfoFn func(context.Context, string, string, string) (map[string]any, error)
	deleteStorageContentFn  func(context.Context, string, string, string) (string, error)

	// Storage definitions
	listStoragesFn  func(context.Context, string) ([]map[string]any, error)
	getStorageFn    func(context.Context, string) (map[string]any, error)
	addStorageFn    func(context.Context, *proxmox.AddStorageRequest) (map[string]any, error)
	updateStorageFn func(context.Context, string, *proxmox.UpdateStorageRequest) error
	removeStorageFn func(context.Context, string) error

	// Backups
	createBackupFn func(context.Context, string, *proxmox.CreateBackupRequest) (string, error)
	listBackupsFn  func(context.Context, string, string) ([]proxmox.StorageContent, error)

	// Network
	listNodeNetworkFn            func(context.Context, string, string) ([]map[string]any, error)
	getNodeNetworkInterfaceFn    func(context.Context, string, string) (map[string]any, error)
	createNodeNetworkInterfaceFn func(context.Context, string, string, *proxmox.NetworkInterfaceConfig) (map[string]any, error)
	updateNodeNetworkInterfaceFn func(context.Context, string, string, *proxmox.NetworkInterfaceConfig) error
	applyNodeNetworkChangesFn    func(context.Context, string) error
	deleteNodeNetworkInterfaceFn func(context.Context, string, string) error

	// Firewall
	listClusterFirewallRulesFn    func(context.Context) ([]map[string]any, error)
	getClusterFirewallOptionsFn   func(context.Context) (map[string]any, error)
	listVMFirewallRulesFn         func(context.Context, string, int) ([]map[string]any, error)
	getVMFirewallOptionsFn        func(context.Context, string, int) (map[string]any, error)
	listContainerFirewallRulesFn  func(context.Context, string, int) ([]map[string]any, error)
	getContainerFirewallOptionsFn func(context.Context, string, int) (map[string]any, error)
	addVMFirewallRuleFn           func(context.Context, string, int, *proxmox.FirewallRuleRequest) error
	deleteVMFirewallRuleFn        func(context.Context, string, int, int) error
	addContainerFirewallRuleFn    func(context.Context, string, int, *proxmox.FirewallRuleRequest) error
	deleteContainerFirewallRuleFn func(context.Context, string, int, int) error

	// Pools
	listPoolsFn  func(context.Context) ([]proxmox.Pool, error)
	getPoolFn    func(context.Context, string) (*proxmox.Pool, error)
	createPoolFn func(context.Context, *proxmox.CreatePoolRequest) error
	updatePoolFn func(context.Context, string, *proxmox.UpdatePoolRequest) error
	deletePoolFn func(context.Context, string) error

	// Access control
	listUsersFn      func(context.Context) ([]map[string]any, error)
	listUserTokensFn func(context.Context, string) ([]map[string]any, error)
}

// Nodes

func (m *mockProxmoxClient) ListNodes(ctx context.Context) ([]proxmox.Node, error) {
	if m.listNodesFn != nil {
		return m.listNodesFn(ctx)
	}
	return nil, nil
}

func (m *mockProxmoxClient) GetNodeStatus(ctx context.Context, node string) (map[string]any, error) {
	if m.getNodeStatusFn != nil {
		return m.getNodeStatusFn(ctx, node)
	}
	return nil, nil
}

func (m *mockProxmoxClient) ListNodeStorage(ctx context.Context, node string) ([]map[string]any, error) {
	if m.listNodeStorageFn != nil {
		return m.listNodeStorageFn(ctx, node)
	}
	return nil, nil
}

func (m *mockProxmoxClient) ListNodeTasks(ctx context.Context, node string, limit int) ([]map[string]any, error) {
	if m.listNodeTasksFn != nil {
		return m.listNodeTasksFn(ctx, node, limit)
	}
	return nil, nil
}

func (m *mockProxmoxClient) GetNodeDisks(ctx context.Context, node string) ([]map[string]any, error) {
	if m.getNodeDisksFn != nil {
		return m.getNodeDisksFn(ctx, node)
	}
	return nil, nil
}

func (m *mockProxmoxClient) GetDiskSMART(ctx context.Context, node, disk string) (map[string]any, error) {
	if m.getDiskSMARTFn != nil {
		return m.getDiskSMARTFn(ctx, node, disk)
	}
	return nil, nil
}

func (m *mockProxmoxClient) ListZFSPools(ctx context.Context, node string) ([]map[string]any, error) {
	if m.listZFSPoolsFn != nil {
		return m.listZFSPoolsFn(ctx, node)
	}
	return nil, nil
}

func (m *mockProxmoxClient) GetZFSPool(ctx context.Context, node, name string) (map[string]any, error) {
	if m.getZFSPoolFn != nil {
		return m.getZFSPoolFn(ctx, node, name)
	}
	return nil, nil
}

func (m *mockProxmoxClient) GetNodeJournal(ctx context.Context, node string, since, until int64, lastEntries int) ([]string, error) {
	if m.getNodeJournalFn != nil {
		return m.getNodeJournalFn(ctx, node, since, until, lastEntries)
	}
	return nil, nil
}

func (m *mockProxmoxClient) NodeCommand(ctx context.Context, node, command string) error {
	if m.nodeCommandFn != nil {
		return m.nodeCommandFn(ctx, node, command)
	}
	return nil
}

// VMs

func (m *mockProxmoxClient) ListVMs(ctx context.Context, node string) ([]proxmox.VM, error) {
	if m.listVMsFn != nil {
		return m.listVMsFn(ctx, node)
	}
	return nil, nil
}

func (m *mockProxmoxClient) GetVMStatus(ctx context.Context, node string, vmid int) (map[string]any, error) {
	if m.getVMStatusFn != nil {
		return m.getVMStatusFn(ctx, node, vmid)
	}
	return nil, nil
}

func (m *mockProxmoxClient) StartVM(ctx context.Context, node string, vmid int) (string, error) {
	if m.startVMFn != nil {
		return m.startVMFn(ctx, node, vmid)
	}
	return "", nil
}

func (m *mockProxmoxClient) StopVM(ctx context.Context, node string, vmid int) (string, error) {
	if m.stopVMFn != nil {
		return m.stopVMFn(ctx, node, vmid)
	}
	return "", nil
}

func (m *mockProxmoxClient) ShutdownVM(ctx context.Context, node string, vmid int) (string, error) {
	if m.shutdownVMFn != nil {
		return m.shutdownVMFn(ctx, node, vmid)
	}
	return "", nil
}

func (m *mockProxmoxClient) RebootVM(ctx context.Context, node string, vmid int) (string, error) {
	if m.rebootVMFn != nil {
		return m.rebootVMFn(ctx, node, vmid)
	}
	return "", nil
}

func (m *mockProxmoxClient) SuspendVM(ctx context.Context, node string, vmid int) (string, error) {
	if m.suspendVMFn != nil {
		return m.suspendVMFn(ctx, node, vmid)
	}
	return "", nil
}

func (m *mockProxmoxClient) ResumeVM(ctx context.Context, node string, vmid int) (string, error) {
	if m.resumeVMFn != nil {
		return m.resumeVMFn(ctx, node, vmid)
	}
	return "", nil
}

func (m *mockProxmoxClient) CreateVM(ctx context.Context, node string, req *proxmox.CreateVMRequest) (string, error) {
	if m.createVMFn != nil {
		return m.createVMFn(ctx, node, req)
	}
	return "", nil
}

func (m *mockProxmoxClient) CloneVM(ctx context.Context, node string, vmid int, req *proxmox.CloneVMRequest) (string, error) {
	if m.cloneVMFn != nil {
		return m.cloneVMFn(ctx, node, vmid, req)
	}
	return "", nil
}

func (m *mockProxmoxClient) GetVMConfig(ctx context.Context, node string, vmid int) (map[string]any, error) {
	if m.getVMConfigFn != nil {
		return m.getVMConfigFn(ctx, node, vmid)
	}
	return nil, nil
}

func (m *mockProxmoxClient) SetVMConfig(ctx context.Context, node string, vmid int, req *proxmox.SetVMConfigRequest) error {
	if m.setVMConfigFn != nil {
		return m.setVMConfigFn(ctx, node, vmid, req)
	}
	return nil
}

func (m *mockProxmoxClient) ResizeVMDisk(ctx context.Context, node string, vmid int, req *proxmox.ResizeDiskRequest) (string, error) {
	if m.resizeVMDiskFn != nil {
		return m.resizeVMDiskFn(ctx, node, vmid, req)
	}
	return "", nil
}

func (m *mockProxmoxClient) MigrateVM(ctx context.Context, node string, vmid int, req *proxmox.MigrateVMRequest) (string, error) {
	if m.migrateVMFn != nil {
		return m.migrateVMFn(ctx, node, vmid, req)
	}
	return "", nil
}

func (m *mockProxmoxClient) RestoreVM(ctx context.Context, node string, req *proxmox.RestoreVMRequest) (string, error) {
	if m.restoreVMFn != nil {
		return m.restoreVMFn(ctx, node, req)
	}
	return "", nil
}

func (m *mockProxmoxClient) MoveVMDisk(ctx context.Context, node string, vmid int, req *proxmox.MoveVMDiskRequest) (string, error) {
	if m.moveVMDiskFn != nil {
		return m.moveVMDiskFn(ctx, node, vmid, req)
	}
	return "", nil
}

func (m *mockProxmoxClient) DeleteVM(ctx context.Context, node string, vmid int, purge bool) (string, error) {
	if m.deleteVMFn != nil {
		return m.deleteVMFn(ctx, node, vmid, purge)
	}
	return "", nil
}

// Containers

func (m *mockProxmoxClient) ListContainers(ctx context.Context, node string) ([]proxmox.Container, error) {
	if m.listContainersFn != nil {
		return m.listContainersFn(ctx, node)
	}
	return nil, nil
}

func (m *mockProxmoxClient) GetContainerStatus(ctx context.Context, node string, vmid int) (map[string]any, error) {
	if m.getContainerStatusFn != nil {
		return m.getContainerStatusFn(ctx, node, vmid)
	}
	return nil, nil
}

func (m *mockProxmoxClient) StartContainer(ctx context.Context, node string, vmid int) (string, error) {
	if m.startContainerFn != nil {
		return m.startContainerFn(ctx, node, vmid)
	}
	return "", nil
}

func (m *mockProxmoxClient) StopContainer(ctx context.Context, node string, vmid int) (string, error) {
	if m.stopContainerFn != nil {
		return m.stopContainerFn(ctx, node, vmid)
	}
	return "", nil
}

func (m *mockProxmoxClient) ShutdownContainer(ctx context.Context, node string, vmid int) (string, error) {
	if m.shutdownContainerFn != nil {
		return m.shutdownContainerFn(ctx, node, vmid)
	}
	return "", nil
}

func (m *mockProxmoxClient) RebootContainer(ctx context.Context, node string, vmid int) (string, error) {
	if m.rebootContainerFn != nil {
		return m.rebootContainerFn(ctx, node, vmid)
	}
	return "", nil
}

func (m *mockProxmoxClient) CreateContainer(ctx context.Context, node string, req *proxmox.CreateContainerRequest) (string, error) {
	if m.createContainerFn != nil {
		return m.createContainerFn(ctx, node, req)
	}
	return "", nil
}

func (m *mockProxmoxClient) CloneContainer(ctx context.Context, node string, vmid int, req *proxmox.CloneContainerRequest) (string, error) {
	if m.cloneContainerFn != nil {
		return m.cloneContainerFn(ctx, node, vmid, req)
	}
	return "", nil
}

func (m *mockProxmoxClient) GetContainerConfig(ctx context.Context, node string, vmid int) (map[string]any, error) {
	if m.getContainerConfigFn != nil {
		return m.getContainerConfigFn(ctx, node, vmid)
	}
	return nil, nil
}

func (m *mockProxmoxClient) SetContainerConfig(ctx context.Context, node string, vmid int, req *proxmox.SetContainerConfigRequest) error {
	if m.setContainerConfigFn != nil {
		return m.setContainerConfigFn(ctx, node, vmid, req)
	}
	return nil
}

func (m *mockProxmoxClient) ResizeContainerDisk(ctx context.Context, node string, vmid int, req *proxmox.ResizeDiskRequest) (string, error) {
	if m.resizeContainerDiskFn != nil {
		return m.resizeContainerDiskFn(ctx, node, vmid, req)
	}
	return "", nil
}

func (m *mockProxmoxClient) MigrateContainer(ctx context.Context, node string, vmid int, req *proxmox.MigrateContainerRequest) (string, error) {
	if m.migrateContainerFn != nil {
		return m.migrateContainerFn(ctx, node, vmid, req)
	}
	return "", nil
}

func (m *mockProxmoxClient) RestoreContainer(ctx context.Context, node string, req *proxmox.RestoreContainerRequest) (string, error) {
	if m.restoreContainerFn != nil {
		return m.restoreContainerFn(ctx, node, req)
	}
	return "", nil
}

func (m *mockProxmoxClient) DeleteContainer(ctx context.Context, node string, vmid int, purge bool) (string, error) {
	if m.deleteContainerFn != nil {
		return m.deleteContainerFn(ctx, node, vmid, purge)
	}
	return "", nil
}

// Cluster

func (m *mockProxmoxClient) ListClusterResources(ctx context.Context, resourceType string) ([]proxmox.ClusterResource, error) {
	if m.listClusterResourcesFn != nil {
		return m.listClusterResourcesFn(ctx, resourceType)
	}
	return nil, nil
}

func (m *mockProxmoxClient) GetTaskStatus(ctx context.Context, node, upid string) (*proxmox.TaskStatus, error) {
	if m.getTaskStatusFn != nil {
		return m.getTaskStatusFn(ctx, node, upid)
	}
	return nil, nil
}

func (m *mockProxmoxClient) GetClusterStatus(ctx context.Context) ([]map[string]any, error) {
	if m.getClusterStatusFn != nil {
		return m.getClusterStatusFn(ctx)
	}
	return nil, nil
}

// Snapshots

func (m *mockProxmoxClient) ListVMSnapshots(ctx context.Context, node string, vmid int) ([]proxmox.Snapshot, error) {
	if m.listVMSnapshotsFn != nil {
		return m.listVMSnapshotsFn(ctx, node, vmid)
	}
	return nil, nil
}

func (m *mockProxmoxClient) CreateVMSnapshot(ctx context.Context, node string, vmid int, req proxmox.CreateVMSnapshotRequest) (string, error) {
	if m.createVMSnapshotFn != nil {
		return m.createVMSnapshotFn(ctx, node, vmid, req)
	}
	return "", nil
}

func (m *mockProxmoxClient) RollbackVMSnapshot(ctx context.Context, node string, vmid int, snapname string) (string, error) {
	if m.rollbackVMSnapshotFn != nil {
		return m.rollbackVMSnapshotFn(ctx, node, vmid, snapname)
	}
	return "", nil
}

func (m *mockProxmoxClient) DeleteVMSnapshot(ctx context.Context, node string, vmid int, snapname string) (string, error) {
	if m.deleteVMSnapshotFn != nil {
		return m.deleteVMSnapshotFn(ctx, node, vmid, snapname)
	}
	return "", nil
}

func (m *mockProxmoxClient) ListContainerSnapshots(ctx context.Context, node string, vmid int) ([]proxmox.Snapshot, error) {
	if m.listContainerSnapshotsFn != nil {
		return m.listContainerSnapshotsFn(ctx, node, vmid)
	}
	return nil, nil
}

func (m *mockProxmoxClient) CreateContainerSnapshot(ctx context.Context, node string, vmid int, req proxmox.CreateContainerSnapshotRequest) (string, error) {
	if m.createContainerSnapshotFn != nil {
		return m.createContainerSnapshotFn(ctx, node, vmid, req)
	}
	return "", nil
}

func (m *mockProxmoxClient) RollbackContainerSnapshot(ctx context.Context, node string, vmid int, snapname string) (string, error) {
	if m.rollbackContainerSnapshotFn != nil {
		return m.rollbackContainerSnapshotFn(ctx, node, vmid, snapname)
	}
	return "", nil
}

func (m *mockProxmoxClient) DeleteContainerSnapshot(ctx context.Context, node string, vmid int, snapname string) (string, error) {
	if m.deleteContainerSnapshotFn != nil {
		return m.deleteContainerSnapshotFn(ctx, node, vmid, snapname)
	}
	return "", nil
}

// Storage content

func (m *mockProxmoxClient) ListStorageContent(ctx context.Context, node, storage, content string) ([]proxmox.StorageContent, error) {
	if m.listStorageContentFn != nil {
		return m.listStorageContentFn(ctx, node, storage, content)
	}
	return nil, nil
}

func (m *mockProxmoxClient) GetStorageContentInfo(ctx context.Context, node, storage, volume string) (map[string]any, error) {
	if m.getStorageContentInfoFn != nil {
		return m.getStorageContentInfoFn(ctx, node, storage, volume)
	}
	return nil, nil
}

func (m *mockProxmoxClient) DeleteStorageContent(ctx context.Context, node, storage, volume string) (string, error) {
	if m.deleteStorageContentFn != nil {
		return m.deleteStorageContentFn(ctx, node, storage, volume)
	}
	return "", nil
}

// Storage definitions

func (m *mockProxmoxClient) ListStorages(ctx context.Context, storageType string) ([]map[string]any, error) {
	if m.listStoragesFn != nil {
		return m.listStoragesFn(ctx, storageType)
	}
	return nil, nil
}

func (m *mockProxmoxClient) GetStorage(ctx context.Context, storage string) (map[string]any, error) {
	if m.getStorageFn != nil {
		return m.getStorageFn(ctx, storage)
	}
	return nil, nil
}

func (m *mockProxmoxClient) AddStorage(ctx context.Context, req *proxmox.AddStorageRequest) (map[string]any, error) {
	if m.addStorageFn != nil {
		return m.addStorageFn(ctx, req)
	}
	return nil, nil
}

func (m *mockProxmoxClient) UpdateStorage(ctx context.Context, storage string, req *proxmox.UpdateStorageRequest) error {
	if m.updateStorageFn != nil {
		return m.updateStorageFn(ctx, storage, req)
	}
	return nil
}

func (m *mockProxmoxClient) RemoveStorage(ctx context.Context, storage string) error {
	if m.removeStorageFn != nil {
		return m.removeStorageFn(ctx, storage)
	}
	return nil
}

// Backups

func (m *mockProxmoxClient) CreateBackup(ctx context.Context, node string, req *proxmox.CreateBackupRequest) (string, error) {
	if m.createBackupFn != nil {
		return m.createBackupFn(ctx, node, req)
	}
	return "", nil
}

func (m *mockProxmoxClient) ListBackups(ctx context.Context, node, storage string) ([]proxmox.StorageContent, error) {
	if m.listBackupsFn != nil {
		return m.listBackupsFn(ctx, node, storage)
	}
	return nil, nil
}

// Network

func (m *mockProxmoxClient) ListNodeNetwork(ctx context.Context, node, networkType string) ([]map[string]any, error) {
	if m.listNodeNetworkFn != nil {
		return m.listNodeNetworkFn(ctx, node, networkType)
	}
	return nil, nil
}

func (m *mockProxmoxClient) GetNodeNetworkInterface(ctx context.Context, node, iface string) (map[string]any, error) {
	if m.getNodeNetworkInterfaceFn != nil {
		return m.getNodeNetworkInterfaceFn(ctx, node, iface)
	}
	return nil, nil
}

func (m *mockProxmoxClient) CreateNodeNetworkInterface(ctx context.Context, node, iface string, cfg *proxmox.NetworkInterfaceConfig) (map[string]any, error) {
	if m.createNodeNetworkInterfaceFn != nil {
		return m.createNodeNetworkInterfaceFn(ctx, node, iface, cfg)
	}
	return nil, nil
}

func (m *mockProxmoxClient) UpdateNodeNetworkInterface(ctx context.Context, node, iface string, cfg *proxmox.NetworkInterfaceConfig) error {
	if m.updateNodeNetworkInterfaceFn != nil {
		return m.updateNodeNetworkInterfaceFn(ctx, node, iface, cfg)
	}
	return nil
}

func (m *mockProxmoxClient) ApplyNodeNetworkChanges(ctx context.Context, node string) error {
	if m.applyNodeNetworkChangesFn != nil {
		return m.applyNodeNetworkChangesFn(ctx, node)
	}
	return nil
}

func (m *mockProxmoxClient) DeleteNodeNetworkInterface(ctx context.Context, node, iface string) error {
	if m.deleteNodeNetworkInterfaceFn != nil {
		return m.deleteNodeNetworkInterfaceFn(ctx, node, iface)
	}
	return nil
}

// Firewall

func (m *mockProxmoxClient) ListClusterFirewallRules(ctx context.Context) ([]map[string]any, error) {
	if m.listClusterFirewallRulesFn != nil {
		return m.listClusterFirewallRulesFn(ctx)
	}
	return nil, nil
}

func (m *mockProxmoxClient) GetClusterFirewallOptions(ctx context.Context) (map[string]any, error) {
	if m.getClusterFirewallOptionsFn != nil {
		return m.getClusterFirewallOptionsFn(ctx)
	}
	return nil, nil
}

func (m *mockProxmoxClient) ListVMFirewallRules(ctx context.Context, node string, vmid int) ([]map[string]any, error) {
	if m.listVMFirewallRulesFn != nil {
		return m.listVMFirewallRulesFn(ctx, node, vmid)
	}
	return nil, nil
}

func (m *mockProxmoxClient) GetVMFirewallOptions(ctx context.Context, node string, vmid int) (map[string]any, error) {
	if m.getVMFirewallOptionsFn != nil {
		return m.getVMFirewallOptionsFn(ctx, node, vmid)
	}
	return nil, nil
}

func (m *mockProxmoxClient) ListContainerFirewallRules(ctx context.Context, node string, vmid int) ([]map[string]any, error) {
	if m.listContainerFirewallRulesFn != nil {
		return m.listContainerFirewallRulesFn(ctx, node, vmid)
	}
	return nil, nil
}

func (m *mockProxmoxClient) GetContainerFirewallOptions(ctx context.Context, node string, vmid int) (map[string]any, error) {
	if m.getContainerFirewallOptionsFn != nil {
		return m.getContainerFirewallOptionsFn(ctx, node, vmid)
	}
	return nil, nil
}

func (m *mockProxmoxClient) AddVMFirewallRule(ctx context.Context, node string, vmid int, req *proxmox.FirewallRuleRequest) error {
	if m.addVMFirewallRuleFn != nil {
		return m.addVMFirewallRuleFn(ctx, node, vmid, req)
	}
	return nil
}

func (m *mockProxmoxClient) DeleteVMFirewallRule(ctx context.Context, node string, vmid, pos int) error {
	if m.deleteVMFirewallRuleFn != nil {
		return m.deleteVMFirewallRuleFn(ctx, node, vmid, pos)
	}
	return nil
}

func (m *mockProxmoxClient) AddContainerFirewallRule(ctx context.Context, node string, vmid int, req *proxmox.FirewallRuleRequest) error {
	if m.addContainerFirewallRuleFn != nil {
		return m.addContainerFirewallRuleFn(ctx, node, vmid, req)
	}
	return nil
}

func (m *mockProxmoxClient) DeleteContainerFirewallRule(ctx context.Context, node string, vmid, pos int) error {
	if m.deleteContainerFirewallRuleFn != nil {
		return m.deleteContainerFirewallRuleFn(ctx, node, vmid, pos)
	}
	return nil
}

// Pools

func (m *mockProxmoxClient) ListPools(ctx context.Context) ([]proxmox.Pool, error) {
	if m.listPoolsFn != nil {
		return m.listPoolsFn(ctx)
	}
	return nil, nil
}

func (m *mockProxmoxClient) GetPool(ctx context.Context, poolid string) (*proxmox.Pool, error) {
	if m.getPoolFn != nil {
		return m.getPoolFn(ctx, poolid)
	}
	return nil, nil
}

func (m *mockProxmoxClient) CreatePool(ctx context.Context, req *proxmox.CreatePoolRequest) error {
	if m.createPoolFn != nil {
		return m.createPoolFn(ctx, req)
	}
	return nil
}

func (m *mockProxmoxClient) UpdatePool(ctx context.Context, poolid string, req *proxmox.UpdatePoolRequest) error {
	if m.updatePoolFn != nil {
		return m.updatePoolFn(ctx, poolid, req)
	}
	return nil
}

func (m *mockProxmoxClient) DeletePool(ctx context.Context, poolid string) error {
	if m.deletePoolFn != nil {
		return m.deletePoolFn(ctx, poolid)
	}
	return nil
}

// Access control

func (m *mockProxmoxClient) ListUsers(ctx context.Context) ([]map[string]any, error) {
	if m.listUsersFn != nil {
		return m.listUsersFn(ctx)
	}
	return nil, nil
}

func (m *mockProxmoxClient) ListUserTokens(ctx context.Context, userid string) ([]map[string]any, error) {
	if m.listUserTokensFn != nil {
		return m.listUserTokensFn(ctx, userid)
	}
	return nil, nil
}
