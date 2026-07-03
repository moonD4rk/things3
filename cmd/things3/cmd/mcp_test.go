package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"syscall"
	"testing"
)

// TestIsCleanShutdown pins the exit-code contract: SIGINT/SIGTERM, stdin EOF, the
// SDK's "server is closing" error, and a broken pipe are clean (exit 0); a genuine
// failure is not. Without this the mapping is mutation-insensitive.
func TestIsCleanShutdown(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, true},
		{"context canceled (signal)", context.Canceled, true},
		{"stdin EOF", io.EOF, true},
		{"wrapped stdin EOF", fmt.Errorf("read: %w", io.EOF), true},
		{"SDK server-closing on EOF", errors.New("server is closing: EOF"), true},
		{"broken pipe (client closed stdout)", &os.PathError{Op: "write", Path: "stdout", Err: syscall.EPIPE}, true},
		{"genuine failure", errors.New("database is locked"), false},
		{"startup schema failure", errors.New("mcpserver: input schema for get: boom"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isCleanShutdown(tc.err); got != tc.want {
				t.Errorf("isCleanShutdown(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

// TestParseLogLevel proves each level name maps correctly (case-insensitively) and
// an unknown name is rejected rather than silently defaulting.
func TestParseLogLevel(t *testing.T) {
	valid := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
		"INFO":  slog.LevelInfo,
	}
	for name, want := range valid {
		got, err := parseLogLevel(name)
		if err != nil || got != want {
			t.Errorf("parseLogLevel(%q) = %v, %v; want %v, nil", name, got, err, want)
		}
	}
	if _, err := parseLogLevel("verbose"); err == nil {
		t.Errorf("parseLogLevel(verbose) should reject an unknown level")
	}
}

// TestNewMCPCmd checks the command is in the Actions group with an example and that
// --read-only and --log-level are local flags, not persistent ones.
func TestNewMCPCmd(t *testing.T) {
	cmd := newMCPCmd()
	if cmd.GroupID != groupActions {
		t.Errorf("GroupID = %q, want %q", cmd.GroupID, groupActions)
	}
	if cmd.Example == "" {
		t.Errorf("mcp command should set an Example")
	}
	for _, name := range []string{flagReadOnly, flagLogLevel} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("--%s should be a local flag", name)
		}
		if cmd.PersistentFlags().Lookup(name) != nil {
			t.Errorf("--%s must not be persistent", name)
		}
	}
}
