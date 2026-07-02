package scheme

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Scheme provides URL scheme execution for Things 3.
type Scheme struct {
	foreground bool // For create/update operations: if true, bring Things to foreground
	background bool // For navigation operations: if true, run in background
}

// New creates a new Scheme with the given options.
func New(opts ...Option) *Scheme {
	s := &Scheme{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// wrapExecError combines a command failure with its captured stderr output,
// so causes like AppleEvents permission denials remain distinguishable from
// malformed URLs. Returns nil when err is nil; the original error stays
// matchable via errors.Is/As.
func wrapExecError(err error, stderr []byte) error {
	if err == nil {
		return nil
	}
	if msg := strings.TrimSpace(string(stderr)); msg != "" {
		return fmt.Errorf("things3: URL scheme execution failed: %w: %s", err, msg)
	}
	return fmt.Errorf("things3: URL scheme execution failed: %w", err)
}

// run executes the command with stderr captured and wraps any failure.
func run(cmd *exec.Cmd) error {
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	return wrapExecError(cmd.Run(), stderr.Bytes())
}

// Execute opens a Things URL scheme for create/update operations.
func (s *Scheme) Execute(ctx context.Context, uri string) error {
	if s.foreground {
		return run(exec.CommandContext(ctx, "open", uri))
	}
	script := fmt.Sprintf(`tell application "Things3" to open location %q`, uri)
	return run(exec.CommandContext(ctx, "osascript", "-e", script))
}

// ExecuteNavigation opens a Things URL scheme for navigation operations.
func (s *Scheme) ExecuteNavigation(ctx context.Context, uri string) error {
	if !s.background {
		return run(exec.CommandContext(ctx, "open", uri))
	}
	script := fmt.Sprintf(`tell application "Things3" to open location %q`, uri)
	return run(exec.CommandContext(ctx, "osascript", "-e", script))
}
