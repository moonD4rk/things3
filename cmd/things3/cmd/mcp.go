package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/cmd/things3/internal/mcpserver"
)

// mcp-only flags. Unlike the global surface these are local to the command,
// because they configure the server rather than shape list output.
const (
	flagReadOnly = "read-only"
	flagLogLevel = "log-level"
	flagMaxLimit = "max-limit"
)

func newMCPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mcp",
		Short:   "Serve the Model Context Protocol over stdio for AI assistants",
		GroupID: groupActions,
		Long: `mcp runs a local Model Context Protocol server over stdio, exposing the same
verbs as the CLI as MCP tools. Configure an MCP client to launch it with
{"command": "things3", "args": ["mcp"]}. Reads query the database; writes run
through the Things URL scheme and are verified. Logs go to stderr; stdout carries
the protocol. The server stops on stdin EOF or SIGINT/SIGTERM.`,
		Example: "  things3 mcp\n  things3 mcp --read-only\n  things3 mcp --max-limit 20\n  things3 mcp --log-level debug",
		Args:    cobra.NoArgs,
		RunE:    withClient(runMCP),
	}
	cmd.Flags().Bool(flagReadOnly, false, "register only the read tools; never execute the URL scheme")
	cmd.Flags().String(flagLogLevel, "info", "log level: debug, info, warn, or error")
	cmd.Flags().Int(flagMaxLimit, 0, "cap the list page size for the session; 0 uses the built-in maximum")
	return cmd
}

func runMCP(cmd *cobra.Command, _ []string, client *things3.Client) error {
	readOnly, _ := cmd.Flags().GetBool(flagReadOnly)
	levelName, _ := cmd.Flags().GetString(flagLogLevel)
	maxLimit, _ := cmd.Flags().GetInt(flagMaxLimit)
	if maxLimit < 0 {
		return fmt.Errorf("invalid --%s %d: use 0 for the built-in maximum, or a positive page size", flagMaxLimit, maxLimit)
	}
	level, err := parseLogLevel(levelName)
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewTextHandler(cmd.ErrOrStderr(), &slog.HandlerOptions{Level: level}))
	srv, err := mcpserver.New(client, mcpserver.Config{
		Version:  version,
		ReadOnly: readOnly,
		Logger:   logger,
		MaxLimit: maxLimit,
	})
	if err != nil {
		return err
	}

	// Both shutdown paths are normal terminations, not error exits: a SIGINT/SIGTERM
	// cancels main.go's signal context, and an MCP client stopping the server closes
	// stdin. Only a genuine failure returns non-nil.
	if err := srv.Run(cmd.Context()); !isCleanShutdown(err) {
		return err
	}
	return nil
}

// isCleanShutdown reports whether a server-run error is a normal stdio
// termination. A canceled context is SIGINT/SIGTERM; io.EOF or the SDK's
// "server is closing" jsonrpc2 error (which carries the EOF as text rather than a
// wrapped error) is the client closing stdin; a broken pipe is the client closing
// stdout. Anything else is a real failure.
func isCleanShutdown(err error) bool {
	switch {
	case err == nil,
		errors.Is(err, context.Canceled),
		errors.Is(err, io.EOF),
		errors.Is(err, syscall.EPIPE):
		return true
	default:
		return strings.Contains(err.Error(), "server is closing")
	}
}

// parseLogLevel maps a level name to a slog.Level, rejecting an unknown name.
func parseLogLevel(name string) (slog.Level, error) {
	switch strings.ToLower(name) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("invalid log level %q: use debug, info, warn, or error", name)
	}
}
