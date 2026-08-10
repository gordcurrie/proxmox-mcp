package tools

import (
	"context"

	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
)

// proxmoxClient is the subset of proxmox.Client used by the tools layer.
// Declaring it as an interface allows tests to inject a mock without a real
// Proxmox server, mirroring the unifiClient pattern in unifi-mcp.
type proxmoxClient interface {
	// Nodes
	ListNodes(ctx context.Context) ([]proxmox.Node, error)
	GetNodeStatus(ctx context.Context, node string) (map[string]any, error)
	ListNodeStorage(ctx context.Context, node string) ([]map[string]any, error)
	ListNodeTasks(ctx context.Context, node string, limit int) ([]map[string]any, error)
	GetNodeDisks(ctx context.Context, node string) ([]map[string]any, error)
	GetDiskSMART(ctx context.Context, node, disk string) (map[string]any, error)
	ListZFSPools(ctx context.Context, node string) ([]map[string]any, error)
	GetZFSPool(ctx context.Context, node, name string) (map[string]any, error)
	GetNodeJournal(ctx context.Context, node string, since, until int64, lastEntries int) ([]string, error)
	NodeCommand(ctx context.Context, node, command string) error

	// VMs
	ListVMs(ctx context.Context, node string) ([]proxmox.VM, error)
	GetVMStatus(ctx context.Context, node string, vmid int) (map[string]any, error)
	StartVM(ctx context.Context, node string, vmid int) (string, error)
	StopVM(ctx context.Context, node string, vmid int) (string, error)
	ShutdownVM(ctx context.Context, node string, vmid int) (string, error)
	RebootVM(ctx context.Context, node string, vmid int) (string, error)
	SuspendVM(ctx context.Context, node string, vmid int) (string, error)
	ResumeVM(ctx context.Context, node string, vmid int) (string, error)
	CreateVM(ctx context.Context, node string, req *proxmox.CreateVMRequest) (string, error)
	CloneVM(ctx context.Context, node string, vmid int, req *proxmox.CloneVMRequest) (string, error)
	GetVMConfig(ctx context.Context, node string, vmid int) (map[string]any, error)
	SetVMConfig(ctx context.Context, node string, vmid int, req *proxmox.SetVMConfigRequest) error
	ResizeVMDisk(ctx context.Context, node string, vmid int, req *proxmox.ResizeDiskRequest) (string, error)
	MigrateVM(ctx context.Context, node string, vmid int, req *proxmox.MigrateVMRequest) (string, error)
	RestoreVM(ctx context.Context, node string, req *proxmox.RestoreVMRequest) (string, error)
	MoveVMDisk(ctx context.Context, node string, vmid int, req *proxmox.MoveVMDiskRequest) (string, error)
	DeleteVM(ctx context.Context, node string, vmid int, purge bool) (string, error)

	// Containers
	ListContainers(ctx context.Context, node string) ([]proxmox.Container, error)
	GetContainerStatus(ctx context.Context, node string, vmid int) (map[string]any, error)
	StartContainer(ctx context.Context, node string, vmid int) (string, error)
	StopContainer(ctx context.Context, node string, vmid int) (string, error)
	ShutdownContainer(ctx context.Context, node string, vmid int) (string, error)
	RebootContainer(ctx context.Context, node string, vmid int) (string, error)
	CreateContainer(ctx context.Context, node string, req *proxmox.CreateContainerRequest) (string, error)
	CloneContainer(ctx context.Context, node string, vmid int, req *proxmox.CloneContainerRequest) (string, error)
	GetContainerConfig(ctx context.Context, node string, vmid int) (map[string]any, error)
	SetContainerConfig(ctx context.Context, node string, vmid int, req *proxmox.SetContainerConfigRequest) error
	ResizeContainerDisk(ctx context.Context, node string, vmid int, req *proxmox.ResizeDiskRequest) (string, error)
	MigrateContainer(ctx context.Context, node string, vmid int, req *proxmox.MigrateContainerRequest) (string, error)
	RestoreContainer(ctx context.Context, node string, req *proxmox.RestoreContainerRequest) (string, error)
	DeleteContainer(ctx context.Context, node string, vmid int, purge bool) (string, error)

	// Cluster
	ListClusterResources(ctx context.Context, resourceType string) ([]proxmox.ClusterResource, error)
	GetTaskStatus(ctx context.Context, node, upid string) (*proxmox.TaskStatus, error)
	GetClusterStatus(ctx context.Context) ([]map[string]any, error)
	ListHAGroups(ctx context.Context) ([]map[string]any, error)
	ListHAResources(ctx context.Context) ([]map[string]any, error)
	GetHAStatus(ctx context.Context) ([]map[string]any, error)
	ListClusterConfigNodes(ctx context.Context) ([]map[string]any, error)

	// Snapshots
	ListVMSnapshots(ctx context.Context, node string, vmid int) ([]proxmox.Snapshot, error)
	CreateVMSnapshot(ctx context.Context, node string, vmid int, req proxmox.CreateVMSnapshotRequest) (string, error)
	RollbackVMSnapshot(ctx context.Context, node string, vmid int, snapname string) (string, error)
	DeleteVMSnapshot(ctx context.Context, node string, vmid int, snapname string) (string, error)
	ListContainerSnapshots(ctx context.Context, node string, vmid int) ([]proxmox.Snapshot, error)
	CreateContainerSnapshot(ctx context.Context, node string, vmid int, req proxmox.CreateContainerSnapshotRequest) (string, error)
	RollbackContainerSnapshot(ctx context.Context, node string, vmid int, snapname string) (string, error)
	DeleteContainerSnapshot(ctx context.Context, node string, vmid int, snapname string) (string, error)

	// Storage content
	ListStorageContent(ctx context.Context, node, storage, content string) ([]proxmox.StorageContent, error)
	GetStorageContentInfo(ctx context.Context, node, storage, volume string) (map[string]any, error)
	DeleteStorageContent(ctx context.Context, node, storage, volume string) (string, error)

	// Storage definitions
	ListStorages(ctx context.Context, storageType string) ([]map[string]any, error)
	GetStorage(ctx context.Context, storage string) (map[string]any, error)
	AddStorage(ctx context.Context, req *proxmox.AddStorageRequest) (map[string]any, error)
	UpdateStorage(ctx context.Context, storage string, req *proxmox.UpdateStorageRequest) error
	RemoveStorage(ctx context.Context, storage string) error

	// Backups
	CreateBackup(ctx context.Context, node string, req *proxmox.CreateBackupRequest) (string, error)
	ListBackups(ctx context.Context, node, storage string) ([]proxmox.StorageContent, error)

	// Network
	ListNodeNetwork(ctx context.Context, node, networkType string) ([]map[string]any, error)
	GetNodeNetworkInterface(ctx context.Context, node, iface string) (map[string]any, error)
	CreateNodeNetworkInterface(ctx context.Context, node, iface string, cfg *proxmox.NetworkInterfaceConfig) (map[string]any, error)
	UpdateNodeNetworkInterface(ctx context.Context, node, iface string, cfg *proxmox.NetworkInterfaceConfig) error
	ApplyNodeNetworkChanges(ctx context.Context, node string) error
	DeleteNodeNetworkInterface(ctx context.Context, node, iface string) error

	// Firewall
	ListClusterFirewallRules(ctx context.Context) ([]map[string]any, error)
	GetClusterFirewallOptions(ctx context.Context) (map[string]any, error)
	ListVMFirewallRules(ctx context.Context, node string, vmid int) ([]map[string]any, error)
	GetVMFirewallOptions(ctx context.Context, node string, vmid int) (map[string]any, error)
	ListContainerFirewallRules(ctx context.Context, node string, vmid int) ([]map[string]any, error)
	GetContainerFirewallOptions(ctx context.Context, node string, vmid int) (map[string]any, error)
	AddVMFirewallRule(ctx context.Context, node string, vmid int, req *proxmox.FirewallRuleRequest) error
	DeleteVMFirewallRule(ctx context.Context, node string, vmid, pos int) error
	AddContainerFirewallRule(ctx context.Context, node string, vmid int, req *proxmox.FirewallRuleRequest) error
	DeleteContainerFirewallRule(ctx context.Context, node string, vmid, pos int) error

	// Pools
	ListPools(ctx context.Context) ([]proxmox.Pool, error)
	GetPool(ctx context.Context, poolid string) (*proxmox.Pool, error)
	CreatePool(ctx context.Context, req *proxmox.CreatePoolRequest) error
	UpdatePool(ctx context.Context, poolid string, req *proxmox.UpdatePoolRequest) error
	DeletePool(ctx context.Context, poolid string) error

	// Access control
	ListUsers(ctx context.Context) ([]map[string]any, error)
	ListUserTokens(ctx context.Context, userid string) ([]map[string]any, error)
}
