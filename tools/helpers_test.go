package tools

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestJsonResult_success(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	result, extra, err := jsonResult(payload{Name: "pve1", Value: 42})
	if err != nil {
		t.Fatalf("jsonResult returned error: %v", err)
	}
	if extra != nil {
		t.Errorf("expected nil extra, got %v", extra)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected *mcp.TextContent, got %T", result.Content[0])
	}

	var got payload
	if err := json.Unmarshal([]byte(tc.Text), &got); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if got.Name != "pve1" || got.Value != 42 {
		t.Errorf("got %+v, want {Name:pve1 Value:42}", got)
	}
	// Verify compact formatting — Marshal produces no newlines.
	if strings.Contains(tc.Text, "\n") {
		t.Errorf("expected compact JSON output (no newlines), got: %s", tc.Text)
	}
}

func TestJsonResult_nil(t *testing.T) {
	t.Parallel()

	result, _, err := jsonResult(nil)
	if err != nil {
		t.Fatalf("jsonResult(nil) returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected *mcp.TextContent, got %T", result.Content[0])
	}
	if tc.Text != "null" {
		t.Errorf("expected \"null\", got %q", tc.Text)
	}
}

func TestJsonResult_slice(t *testing.T) {
	t.Parallel()

	nodes := []string{"pve1", "pve2", "pve3"}
	result, _, err := jsonResult(nodes)
	if err != nil {
		t.Fatalf("jsonResult returned error: %v", err)
	}
	tc := result.Content[0].(*mcp.TextContent)
	var got []string
	if err := json.Unmarshal([]byte(tc.Text), &got); err != nil {
		t.Fatalf("unmarshal slice: %v", err)
	}
	if len(got) != 3 || got[0] != "pve1" || got[2] != "pve3" {
		t.Errorf("unexpected slice content: %v", got)
	}
}

func TestTaskResult(t *testing.T) {
	t.Parallel()

	const upid = "UPID:pve1:000015E3:00000000:60F4B3A7:qmstart:100:root@pam:"
	result, extra, err := taskResult(upid)
	if err != nil {
		t.Fatalf("taskResult returned error: %v", err)
	}
	if extra != nil {
		t.Errorf("expected nil extra, got %v", extra)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected *mcp.TextContent, got %T", result.Content[0])
	}

	// Output must be valid JSON.
	var got map[string]any
	if err := json.Unmarshal([]byte(tc.Text), &got); err != nil {
		t.Fatalf("unmarshal taskResult output: %v", err)
	}
	if got["task_id"] != upid {
		t.Errorf("task_id: got %q, want %q", got["task_id"], upid)
	}
	if got["status"] != "dispatched" {
		t.Errorf("status: got %q, want %q", got["status"], "dispatched")
	}
	if _, ok := got["note"]; !ok {
		t.Error("expected note field in task result")
	}
}

func TestTaskResult_specialChars(t *testing.T) {
	t.Parallel()

	// UPIDs contain colons, @ and ! — verify proper JSON escaping via %q.
	const upid = "UPID:pve2:00386A20:2FF79B77:699BB7BA:hastart:110:root@pam!mcp:"
	result, _, err := taskResult(upid)
	if err != nil {
		t.Fatalf("taskResult returned error: %v", err)
	}
	tc := result.Content[0].(*mcp.TextContent)
	var got map[string]any
	if err := json.Unmarshal([]byte(tc.Text), &got); err != nil {
		t.Fatalf("unmarshal with special chars: %v", err)
	}
	if got["task_id"] != upid {
		t.Errorf("task_id round-trip: got %q, want %q", got["task_id"], upid)
	}
}

func TestJsonResult_marshalError(t *testing.T) {
	t.Parallel()

	// Channels cannot be marshalled to JSON — exercises the error path.
	ch := make(chan int)
	_, _, err := jsonResult(ch)
	if err == nil {
		t.Fatal("expected marshal error for channel type, got nil")
	}
}

func TestErrorResult(t *testing.T) {
	t.Parallel()

	result, extra, err := errorResult(fmt.Errorf("something went wrong: %w", fmt.Errorf("root cause")))
	if err != nil {
		t.Fatalf("errorResult returned protocol error: %v", err)
	}
	if extra != nil {
		t.Errorf("expected nil extra, got %v", extra)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Error("expected IsError=true")
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected *mcp.TextContent, got %T", result.Content[0])
	}
	if !strings.Contains(tc.Text, "something went wrong") {
		t.Errorf("error text %q does not contain expected message", tc.Text)
	}
}
