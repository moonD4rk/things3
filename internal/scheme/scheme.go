package scheme

import (
	"context"
	"fmt"
	"os/exec"
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

// Execute opens a Things URL scheme for create/update operations.
func (s *Scheme) Execute(ctx context.Context, uri string) error {
	if s.foreground {
		return exec.CommandContext(ctx, "open", uri).Run()
	}
	script := fmt.Sprintf(`tell application "Things3" to open location %q`, uri)
	return exec.CommandContext(ctx, "osascript", "-e", script).Run()
}

// ExecuteNavigation opens a Things URL scheme for navigation operations.
func (s *Scheme) ExecuteNavigation(ctx context.Context, uri string) error {
	if !s.background {
		return exec.CommandContext(ctx, "open", uri).Run()
	}
	script := fmt.Sprintf(`tell application "Things3" to open location %q`, uri)
	return exec.CommandContext(ctx, "osascript", "-e", script).Run()
}
