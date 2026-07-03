package mcpserver

import (
	"context"
	"errors"
	"fmt"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/cmd/things3/internal/resolve"
)

// candidatesOf renders up to maxCandidates matches as full-UUID candidates.
func candidatesOf(matches []resolve.Match) []Candidate {
	n := min(len(matches), maxCandidates)
	cs := make([]Candidate, n)
	for i := range n {
		cs[i] = Candidate{UUID: matches[i].UUID(), Type: string(matches[i].Kind), Title: matches[i].Title()}
	}
	return cs
}

// resolveError maps a resolve-package error into a structured ToolError. A
// transport failure (database I/O) is not a resolve error and is returned as the
// second value for the handler to surface as a Go error instead.
func resolveError(query string, err error) (*ToolError, error) {
	var nf *resolve.NotFoundError
	if errors.As(err, &nf) {
		return notFound(query), nil
	}
	var amb *resolve.AmbiguousError
	if errors.As(err, &amb) {
		return &ToolError{
			Code:       codeAmbiguous,
			Message:    fmt.Sprintf("%q matches %d items", query, len(amb.Matches)),
			Hint:       "retry with one of the candidate UUIDs, or a 4+ character prefix",
			Candidates: candidatesOf(amb.Matches),
		}, nil
	}
	return nil, err
}

// resolveTarget resolves a write target to exactly one todo or project.
func (s *Server) resolveTarget(ctx context.Context, query string) (resolve.Match, *ToolError, error) {
	m, err := resolve.ResolveOne(ctx, s.client, query)
	if err != nil {
		te, rerr := resolveError(query, err)
		return resolve.Match{}, te, rerr
	}
	return m, nil, nil
}

// resolveProject resolves a query to exactly one project.
func (s *Server) resolveProject(ctx context.Context, query string) (*things3.Project, *ToolError, error) {
	p, err := resolve.Project(ctx, s.client, query)
	if err != nil {
		te, rerr := resolveError(query, err)
		return nil, te, rerr
	}
	return p, nil, nil
}

// resolveArea resolves a query to exactly one area.
func (s *Server) resolveArea(ctx context.Context, query string) (*things3.Area, *ToolError, error) {
	a, err := resolve.Area(ctx, s.client, query)
	if err != nil {
		te, rerr := resolveError(query, err)
		return nil, te, rerr
	}
	return a, nil, nil
}

// resolveHeading resolves a query to exactly one heading within a project.
func (s *Server) resolveHeading(ctx context.Context, projectUUID, query string) (*things3.Heading, *ToolError, error) {
	h, err := resolve.Heading(ctx, s.client, projectUUID, query)
	if err != nil {
		te, rerr := resolveError(query, err)
		return nil, te, rerr
	}
	return h, nil, nil
}

// isNotFound reports whether err is a resolve not-found, used to fall back from
// one destination kind to another (project -> area).
func isNotFound(err error) bool {
	var nf *resolve.NotFoundError
	return errors.As(err, &nf)
}
