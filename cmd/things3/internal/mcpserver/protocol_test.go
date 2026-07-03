package mcpserver

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/thingstest"
)

// testClient opens a client against a writable copy of the shared fixture.
func testClient(t *testing.T) *things3.Client {
	t.Helper()
	client, err := things3.NewClient(things3.WithDatabasePath(thingstest.DatabasePath(t)))
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// newTestServer builds a server over the fixture with the given config.
func newTestServer(t *testing.T, cfg Config) *Server {
	t.Helper()
	srv, err := New(testClient(t), cfg)
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	return srv
}

// connect wires an in-memory client session to the server, returning the session
// and a context for calls.
func connect(t *testing.T, srv *Server) (*mcp.ClientSession, context.Context) {
	t.Helper()
	ctx := context.Background()
	t1, t2 := mcp.NewInMemoryTransports()
	if _, err := srv.mcp.Connect(ctx, t1, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	session, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session, ctx
}

// callTool invokes a tool expected to return (no transport error).
func callTool(t *testing.T, session *mcp.ClientSession, ctx context.Context, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("call %s: %v", name, err)
	}
	return res
}

// contentText returns the first text content block, which the SDK populates with
// the JSON serialization of a tool's structured output.
func contentText(res *mcp.CallToolResult) string {
	if len(res.Content) == 0 {
		return ""
	}
	if tc, ok := res.Content[0].(*mcp.TextContent); ok {
		return tc.Text
	}
	return ""
}

// structured decodes a tool result's output envelope into T, failing on a
// protocol-level tool error (IsError), which the envelope path never sets.
func structured[T any](t *testing.T, res *mcp.CallToolResult) T {
	t.Helper()
	if res.IsError {
		t.Fatalf("unexpected tool error: %s", contentText(res))
	}
	var v T
	if err := json.Unmarshal([]byte(contentText(res)), &v); err != nil {
		t.Fatalf("decode structured content: %v\n%s", err, contentText(res))
	}
	return v
}

// TestStructuredContent proves a successful read returns a decodable structured
// envelope through the in-memory transport.
func TestStructuredContent(t *testing.T) {
	srv := newTestServer(t, Config{Version: "test"})
	session, ctx := connect(t, srv)

	res := callTool(t, session, ctx, "list_todos", map[string]any{"view": "inbox"})
	page := structured[PageResult[Item]](t, res)
	if !page.Success {
		t.Errorf("want success=true")
	}
	if page.Total != thingstest.Inbox {
		t.Errorf("inbox total = %d, want %d", page.Total, thingstest.Inbox)
	}
	if len(page.Items) != thingstest.Inbox {
		t.Errorf("items = %d, want %d", len(page.Items), thingstest.Inbox)
	}
	for i := range page.Items {
		if page.Items[i].Type != typeTodo {
			t.Errorf("item %d type = %q, want todo", i, page.Items[i].Type)
		}
		if page.Items[i].UUID == "" {
			t.Errorf("item %d missing uuid", i)
		}
	}
}

// TestEnumRejection proves a value outside a schema enum is rejected by the SDK
// before the handler runs. It targets list_projects.status and search.type, whose
// handlers do NOT independently error on an unknown value (status silently defaults,
// an unknown search type matches neither and runs both branches), so the test fails if schema-level
// enum enforcement regresses rather than passing on an incidental handler error.
func TestEnumRejection(t *testing.T) {
	srv := newTestServer(t, Config{Version: "test"})
	session, ctx := connect(t, srv)

	cases := []struct {
		tool string
		args map[string]any
	}{
		{"list_projects", map[string]any{"status": "bogus"}},
		{"search", map[string]any{"query": "x", "type": "bogus"}},
		{"list_todos", map[string]any{"view": "bogus"}},
	}
	for _, tc := range cases {
		t.Run(tc.tool, func(t *testing.T) {
			res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: tc.tool, Arguments: tc.args})
			if err == nil && (res == nil || !res.IsError) {
				t.Fatalf("invalid enum should be rejected, got err=%v res=%+v", err, res)
			}
		})
	}
}

// TestFailureEnvelope proves a domain failure rides in the normal output
// envelope with success=false and a structured error, not as a tool error.
func TestFailureEnvelope(t *testing.T) {
	srv := newTestServer(t, Config{Version: "test"})
	session, ctx := connect(t, srv)

	res := callTool(t, session, ctx, "get", map[string]any{"id": "zzznope"})
	if res.IsError {
		t.Fatalf("domain not-found must not be a tool error: %s", contentText(res))
	}
	got := structured[GetResult](t, res)
	if got.Success {
		t.Errorf("want success=false")
	}
	if got.Error == nil || got.Error.Code != codeNotFound {
		t.Errorf("want not_found error, got %+v", got.Error)
	}
}

// TestToolSchemas proves every read tool is listed with both an input and an
// output schema (the latter inferred from the generic result types).
func TestToolSchemas(t *testing.T) {
	srv := newTestServer(t, Config{Version: "test"})
	session, ctx := connect(t, srv)

	var names []string
	for tool, err := range session.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("list tools: %v", err)
		}
		names = append(names, tool.Name)
		if tool.InputSchema == nil {
			t.Errorf("tool %s missing input schema", tool.Name)
		}
		if tool.OutputSchema == nil {
			t.Errorf("tool %s missing output schema", tool.Name)
		}
	}
	for _, want := range []string{"list_todos", "list_projects", "list_areas", "list_tags", "search", "get"} {
		if !slices.Contains(names, want) {
			t.Errorf("missing tool %q in %v", want, names)
		}
	}
}

// toolNames lists the tool names the server advertises over the transport.
func toolNames(t *testing.T, srv *Server) []string {
	t.Helper()
	session, ctx := connect(t, srv)
	var names []string
	for tool, err := range session.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("list tools: %v", err)
		}
		names = append(names, tool.Name)
	}
	return names
}

var (
	readToolNames  = []string{"list_todos", "list_projects", "list_areas", "list_tags", "search", "get"}
	writeToolNames = []string{"add_todo", "add_project", "complete", "schedule", "move", "edit", "open"}
)

// TestReadOnlyToolListing proves --read-only registers only the six read tools, so
// a read-only client cannot even list (let alone call) the write and nav tools --
// the write-safety boundary RFC 014 requires. A default server lists all thirteen.
func TestReadOnlyToolListing(t *testing.T) {
	t.Run("read-only omits write and nav tools", func(t *testing.T) {
		names := toolNames(t, newTestServer(t, Config{Version: "test", ReadOnly: true}))
		for _, r := range readToolNames {
			if !slices.Contains(names, r) {
				t.Errorf("read-only server missing read tool %q", r)
			}
		}
		for _, w := range writeToolNames {
			if slices.Contains(names, w) {
				t.Errorf("read-only server must not register %q", w)
			}
		}
	})

	t.Run("default registers write and nav tools", func(t *testing.T) {
		names := toolNames(t, newTestServer(t, Config{Version: "test"}))
		for _, w := range writeToolNames {
			if !slices.Contains(names, w) {
				t.Errorf("default server missing write tool %q", w)
			}
		}
	})
}
