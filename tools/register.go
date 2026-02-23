// Package tools registers all MCP tools onto the server.
package tools

import (
	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterAll wires all Proxmox MCP tools onto the provided server.
func RegisterAll(s *mcp.Server, client *proxmox.Client) {
	registerNodeTools(s, client)
	registerVMTools(s, client)
	registerContainerTools(s, client)
	registerClusterTools(s, client)
}
