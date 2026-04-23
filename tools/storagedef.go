package tools

import (
	"context"
	"fmt"

	"github.com/gordcurrie/proxmox-mcp/internal/proxmox"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerStorageDefTools adds cluster-wide storage definition MCP tools to the server.
func registerStorageDefTools(s *mcp.Server, client proxmoxClient) {
	type listStoragesInput struct {
		Type string `json:"type,omitempty" jsonschema:"optional storage type filter: nfs, pbs, dir, cifs, zfspool, lvmthin, etc."`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_storages",
		Description: "List all cluster-wide storage definitions in Proxmox. Optionally filter by storage type (e.g. nfs, pbs, dir).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input listStoragesInput) (*mcp.CallToolResult, any, error) {
		storages, err := client.ListStorages(ctx, input.Type)
		if err != nil {
			return errorResult(fmt.Errorf("list_storages: %w", err))
		}
		return jsonResult(storages)
	})

	type getStorageInput struct {
		Storage string `json:"storage" jsonschema:"name of the storage definition (e.g. local, pbs-store)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_storage",
		Description: "Get the full configuration of a Proxmox storage definition.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input getStorageInput) (*mcp.CallToolResult, any, error) {
		storage, err := client.GetStorage(ctx, input.Storage)
		if err != nil {
			return errorResult(fmt.Errorf("get_storage: %w", err))
		}
		return jsonResult(storage)
	})

	type addStorageInput struct {
		Storage     string                  `json:"storage"                jsonschema:"unique name for the new storage (required)"`
		Type        string                  `json:"type"                   jsonschema:"storage type: nfs, pbs, dir, cifs, zfspool, lvmthin, etc. (required)"`
		Content     string                  `json:"content,omitempty"      jsonschema:"comma-separated content types to allow, e.g. backup,images,iso"`
		Nodes       string                  `json:"nodes,omitempty"        jsonschema:"comma-separated list of nodes that can use this storage; omit for all nodes"`
		Shared      bool                    `json:"shared,omitempty"       jsonschema:"true if the storage is accessible from all nodes simultaneously"`
		Server      string                  `json:"server,omitempty"       jsonschema:"hostname or IP of the NFS, PBS, or CIFS server"`
		Export      string                  `json:"export,omitempty"       jsonschema:"NFS export path on the server (NFS only)"`
		Path        string                  `json:"path,omitempty"         jsonschema:"local filesystem path (dir type only)"`
		Datastore   string                  `json:"datastore,omitempty"    jsonschema:"datastore name on the PBS server (pbs type only)"`
		Username    string                  `json:"username,omitempty"     jsonschema:"username for PBS (e.g. backup@pbs) or CIFS authentication"`
		Password    proxmox.SensitiveString `json:"password,omitempty"     jsonschema:"password for PBS or CIFS authentication"`
		Fingerprint string                  `json:"fingerprint,omitempty"  jsonschema:"TLS certificate fingerprint of the PBS server for verification"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "add_storage",
		Description: "Add a new cluster-wide storage definition to Proxmox. Supports nfs, pbs, dir, cifs, zfspool, and other types. For PBS: provide server, datastore, username, password. For NFS: provide server and export.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input addStorageInput) (*mcp.CallToolResult, any, error) {
		req := &proxmox.AddStorageRequest{
			Storage:     input.Storage,
			Type:        input.Type,
			Content:     input.Content,
			Nodes:       input.Nodes,
			Server:      input.Server,
			Export:      input.Export,
			Path:        input.Path,
			Datastore:   input.Datastore,
			Username:    input.Username,
			Password:    input.Password,
			Fingerprint: input.Fingerprint,
		}
		if input.Shared {
			one := 1
			req.Shared = &one
		}
		if _, err := client.AddStorage(ctx, req); err != nil {
			return errorResult(fmt.Errorf("add_storage: %w", err))
		}
		return jsonResult(map[string]string{"storage": input.Storage, "status": "created"})
	})

	type updateStorageInput struct {
		Storage     string                  `json:"storage"                jsonschema:"name of the storage to update (required)"`
		Content     string                  `json:"content,omitempty"      jsonschema:"comma-separated content types, e.g. backup,images"`
		Nodes       string                  `json:"nodes,omitempty"        jsonschema:"comma-separated node names to restrict storage to"`
		Shared      *bool                   `json:"shared,omitempty"       jsonschema:"true = accessible from all nodes, false = restrict"`
		Server      string                  `json:"server,omitempty"       jsonschema:"new server hostname or IP"`
		Export      string                  `json:"export,omitempty"       jsonschema:"new NFS export path"`
		Path        string                  `json:"path,omitempty"         jsonschema:"new local path (dir type)"`
		Datastore   string                  `json:"datastore,omitempty"    jsonschema:"new PBS datastore name"`
		Username    string                  `json:"username,omitempty"     jsonschema:"new PBS/CIFS username"`
		Password    proxmox.SensitiveString `json:"password,omitempty"     jsonschema:"new PBS/CIFS password"`
		Fingerprint string                  `json:"fingerprint,omitempty"  jsonschema:"new PBS TLS fingerprint"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_storage",
		Description: "Update the configuration of an existing Proxmox storage definition. Only supplied fields are changed; omitted fields are left as-is.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input updateStorageInput) (*mcp.CallToolResult, any, error) {
		req := &proxmox.UpdateStorageRequest{
			Content:     input.Content,
			Nodes:       input.Nodes,
			Server:      input.Server,
			Export:      input.Export,
			Path:        input.Path,
			Datastore:   input.Datastore,
			Username:    input.Username,
			Password:    input.Password,
			Fingerprint: input.Fingerprint,
		}
		if input.Shared != nil {
			if *input.Shared {
				one := 1
				req.Shared = &one
			} else {
				zero := 0
				req.Shared = &zero
			}
		}
		if err := client.UpdateStorage(ctx, input.Storage, req); err != nil {
			return errorResult(fmt.Errorf("update_storage: %w", err))
		}
		return jsonResult(map[string]string{"storage": input.Storage, "status": "updated"})
	})
}
