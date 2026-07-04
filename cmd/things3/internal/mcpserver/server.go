package mcpserver

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/cmd/things3/internal/verify"
)

// URLBuilder is the common Build/Execute surface of every scheme builder the
// write tools drive. It matches the CLI's urlBuilder so the two layers share the
// same fire-and-forget contract.
type URLBuilder interface {
	Build() (string, error)
	Execute(ctx context.Context) error
}

// ExecuteFunc runs a scheme builder. The production default calls Execute;
// tests inject a recorder that builds the URL and captures it without touching
// osascript.
type ExecuteFunc func(ctx context.Context, b URLBuilder) error

// Config carries the server's construction seams. The zero value is usable: a
// nil Logger discards, a nil Execute runs the real URL scheme, and a zero Verify
// uses the default polling bounds.
type Config struct {
	// Version is the server version reported to clients (the CLI's ldflags value).
	Version string
	// ReadOnly registers only the six read tools; writes and open are omitted.
	ReadOnly bool
	// Logger receives structured server activity. stdout belongs to the protocol,
	// so a real logger must write elsewhere (the CLI points it at stderr).
	Logger *slog.Logger
	// Verify tunes write confirmation; tests inject an instant clock.
	Verify verify.Options
	// Execute runs a scheme builder; tests inject a URL recorder.
	Execute ExecuteFunc
	// MaxLimit caps the list page size for the whole session; zero uses the
	// built-in maximum. The effective default page size is clamped to this cap so
	// a low cap never leaves default above maximum.
	MaxLimit int
}

// Server is a things3 MCP server bound to one database client for its whole
// session. Reads run fully concurrent; writeMu serializes execute-plus-verify
// pairs so concurrent tool calls cannot cross-confirm same-title creates.
type Server struct {
	client  *things3.Client
	cfg     Config
	mcp     *mcp.Server
	writeMu sync.Mutex
	// maxLimit and defaultLimit are the effective page-size bounds for the
	// session, resolved once at construction and threaded into both the input
	// schemas and the pagination clamp so the advertised cap and the enforced cap
	// never diverge.
	maxLimit     int
	defaultLimit int
}

// New builds a things3 MCP server over the given client and configuration. It
// registers the tool surface (six read tools always; the write tools and open
// unless ReadOnly) and returns an error if any tool schema fails to build.
func New(client *things3.Client, cfg Config) (*Server, error) {
	if cfg.Logger == nil {
		cfg.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	if cfg.Execute == nil {
		cfg.Execute = func(ctx context.Context, b URLBuilder) error { return b.Execute(ctx) }
	}

	s := &Server{client: client, cfg: cfg}
	s.defaultLimit, s.maxLimit = resolveLimits(cfg)
	s.mcp = mcp.NewServer(
		&mcp.Implementation{Name: serverName, Version: cfg.Version},
		&mcp.ServerOptions{Instructions: instructions, Logger: cfg.Logger},
	)
	if err := s.register(); err != nil {
		return nil, err
	}
	return s, nil
}

// Run serves the MCP protocol over stdio until stdin reaches EOF or ctx is
// canceled (SIGINT/SIGTERM via the caller's signal context).
func (s *Server) Run(ctx context.Context) error {
	return s.mcp.Run(ctx, &mcp.StdioTransport{})
}

// resolveLimits fixes the session's effective page-size bounds once. A positive
// Config.MaxLimit lowers the cap (the --max-limit flag); it can only tighten,
// never raise the cap above the built-in MaxLimit, so a model can never request
// a larger page than the built-in ceiling. Zero uses the built-in maximum. The
// default page size is clamped to the cap, because the SDK rejects a schema
// whose default exceeds its maximum and would panic at registration.
func resolveLimits(cfg Config) (defaultLimit, maxLimit int) {
	maxLimit = MaxLimit
	if cfg.MaxLimit > 0 {
		maxLimit = min(cfg.MaxLimit, MaxLimit)
	}
	return min(DefaultLimit, maxLimit), maxLimit
}

// register wires every tool. Read tools are always present; write tools and open
// are registered only when the server is not read-only, so a read-only client
// cannot even list them.
func (s *Server) register() error {
	r := &registrar{srv: s.mcp, maxLimit: s.maxLimit, defaultLimit: s.defaultLimit}
	s.registerRead(r)
	if !s.cfg.ReadOnly {
		s.registerWrite(r)
	}
	return r.err
}

// registrar accumulates tool registrations, short-circuiting on the first schema
// error so register reports it from a single place. It carries the page-size
// bounds so every list tool's schema advertises the same cap the handler enforces.
type registrar struct {
	srv          *mcp.Server
	err          error
	maxLimit     int
	defaultLimit int
}

// regTool builds In's input schema (injecting enum constraints and pagination
// bounds) and adds the tool. It is a no-op once a prior registration has failed.
func regTool[In, Out any](r *registrar, name, desc string, h mcp.ToolHandlerFor[In, Out]) {
	if r.err != nil {
		return
	}
	schema, err := inputSchemaFor[In](r.maxLimit, r.defaultLimit)
	if err != nil {
		r.err = fmt.Errorf("mcpserver: input schema for %s: %w", name, err)
		return
	}
	mcp.AddTool(r.srv, &mcp.Tool{Name: name, Description: desc, InputSchema: schema}, h)
}

// lockWrites takes the write mutex and returns its unlock, serializing every
// execute-plus-verify pair so concurrent tool calls cannot cross-confirm
// same-title creates. The caller holds the lock across its own verification.
func (s *Server) lockWrites() func() {
	s.writeMu.Lock()
	return s.writeMu.Unlock
}
