package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// connectTestServer registers all tools on a new MCP server backed by the
// provided mock, then returns a connected client session. The cleanup function
// closes the session and must be deferred by the caller.
func connectTestServer(t *testing.T, mock *mockProxmoxClient) (session *mcp.ClientSession, cleanup func()) {
	t.Helper()

	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "v0.0.1"}, nil)
	RegisterAll(s, mock, Config{AllowDestructive: true})

	ct, st := mcp.NewInMemoryTransports()
	ctx := context.Background()

	if _, err := s.Connect(ctx, st, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}

	c := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "v0.0.1"}, nil)
	cs, err := c.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}

	return cs, func() {
		if err := cs.Close(); err != nil {
			t.Logf("client session close: %v", err)
		}
	}
}

// callTool is a test shorthand that calls the named tool and returns the result.
func callTool(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	ctx := context.Background()
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("CallTool %s: %v", name, err)
	}
	return res
}

// assertSuccess fails the test if result has IsError set.
func assertSuccess(t *testing.T, res *mcp.CallToolResult) {
	t.Helper()
	if res.IsError {
		var text string
		if len(res.Content) > 0 {
			if tc, ok := res.Content[0].(*mcp.TextContent); ok {
				text = tc.Text
			}
		}
		t.Errorf("expected success, got isError=true: %s", text)
	}
}

// assertError fails the test if result does not have IsError set, or if the
// error message does not contain the expected substring.
func assertError(t *testing.T, res *mcp.CallToolResult, wantContains string) {
	t.Helper()
	if !res.IsError {
		t.Errorf("expected isError=true, got success")
		return
	}
	if wantContains == "" {
		return
	}
	if len(res.Content) == 0 {
		t.Errorf("expected error content, got none")
		return
	}
	tc, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Errorf("expected TextContent, got %T", res.Content[0])
		return
	}
	if !strings.Contains(tc.Text, wantContains) {
		t.Errorf("error text %q does not contain %q", tc.Text, wantContains)
	}
}

// assertResultJSON fails the test if result has IsError set, has no content,
// or if the content is not valid non-empty JSON.
func assertResultJSON(t *testing.T, res *mcp.CallToolResult) {
	t.Helper()
	assertSuccess(t, res)
	if len(res.Content) == 0 {
		t.Fatalf("expected content, got none")
	}
	tc, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", res.Content[0])
	}
	if tc.Text == "" {
		t.Fatal("expected non-empty content text")
	}
	if !json.Valid([]byte(tc.Text)) {
		t.Fatalf("content is not valid JSON: %q", tc.Text)
	}
}
