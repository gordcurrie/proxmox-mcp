package tools

import (
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// jsonResult marshals v to compact JSON and returns it as a TextContent result.
func jsonResult(v any) (*mcp.CallToolResult, any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, nil, fmt.Errorf("marshalling result: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, nil, nil
}

type taskResponse struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
	Note   string `json:"note"`
}

// taskResult returns an async task response containing the UPID.
func taskResult(upid string) (*mcp.CallToolResult, any, error) {
	resp := taskResponse{
		TaskID: upid,
		Status: "dispatched",
		Note:   "Use get_task_status to poll for completion.",
	}

	return jsonResult(resp)
}

// textResult returns a plain text MCP tool result.
func textResult(msg string) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
	}, nil, nil
}

// errorResult returns an MCP-spec-compliant tool error result (isError: true).
// Per MCP spec §6, tool execution errors should use isError rather than
// propagating a protocol-level error.
func errorResult(err error) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: err.Error()},
		},
	}, nil, nil
}
