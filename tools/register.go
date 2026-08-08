// Package tools registers all MCP tools onto the server.
package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Config holds optional feature flags for tool registration.
type Config struct {
	// AllowDestructive enables destructive tools (delete_vm, delete_container).
	// Corresponds to the PROXMOX_ALLOW_DESTRUCTIVE environment variable.
	// Defaults to false — destructive tools are not registered unless explicitly opted in.
	AllowDestructive bool
}

// RegisterAll wires all Proxmox MCP tools onto the provided server.
func RegisterAll(s *mcp.Server, client proxmoxClient, cfg Config) {
	registerNodeTools(s, client)
	registerVMTools(s, client)
	registerContainerTools(s, client)
	registerClusterTools(s, client)
	registerSnapshotTools(s, client)
	registerStorageTools(s, client)
	registerBackupTools(s, client)
	registerNetworkTools(s, client)
	registerFirewallTools(s, client)
	registerPoolTools(s, client)
	registerStorageDefTools(s, client)
	registerAccessTools(s, client)
	if cfg.AllowDestructive {
		registerDestructiveTools(s, client)
	}
}
