package things3

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
)

// scheme provides URL building and execution for Things URL Scheme.
type scheme struct {
	foreground bool // For create/update operations: if true, bring Things to foreground
	background bool // For navigation operations: if true, run in background
}

// newScheme creates a new URL Scheme builder.
func newScheme(opts ...schemeOption) *scheme {
	s := &scheme{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// AddTodo returns an addTodoBuilder for creating a new to-do.
func (s *scheme) AddTodo() *addTodoBuilder {
	return &addTodoBuilder{
		scheme: s,
		attrs:  urlAttrs{params: make(map[string]string)},
	}
}

// AddProject returns an addProjectBuilder for creating a new project.
func (s *scheme) AddProject() *addProjectBuilder {
	return &addProjectBuilder{
		scheme: s,
		attrs:  urlAttrs{params: make(map[string]string)},
	}
}

// ShowBuilder returns a showBuilder for navigating to items or lists.
func (s *scheme) ShowBuilder() *showBuilder {
	return &showBuilder{
		scheme: s,
		params: make(map[string]string),
	}
}

// Batch returns a batchBuilder for batch create operations.
func (s *scheme) Batch() *batchBuilder {
	return &batchBuilder{
		scheme: s,
		items:  make([]JSONItem, 0),
	}
}

// SearchURL returns a URL to search for the given query in Things.
func (s *scheme) SearchURL(query string) string {
	q := url.Values{}
	q.Set("query", query)
	return fmt.Sprintf("things:///%s?%s", CommandSearch, encodeQuery(q))
}

// Version returns a URL to get Things version information.
func (s *scheme) Version() string {
	return fmt.Sprintf("things:///%s", CommandVersion)
}

// WithToken returns an authScheme for authenticated operations.
func (s *scheme) WithToken(token string) *authScheme {
	return &authScheme{
		scheme: s,
		token:  token,
	}
}

// authScheme provides URL building for authenticated operations.
type authScheme struct {
	scheme *scheme
	token  string
}

// UpdateTodo returns a TodoUpdater for modifying an existing to-do.
func (a *authScheme) UpdateTodo(id string) TodoUpdater {
	return &updateTodoBuilder{
		scheme: a.scheme,
		token:  a.token,
		id:     id,
		attrs:  urlAttrs{params: make(map[string]string)},
	}
}

// UpdateProject returns a ProjectUpdater for modifying an existing project.
func (a *authScheme) UpdateProject(id string) ProjectUpdater {
	return &updateProjectBuilder{
		scheme: a.scheme,
		token:  a.token,
		id:     id,
		attrs:  urlAttrs{params: make(map[string]string)},
	}
}

// Batch returns an AuthBatchCreator for batch operations including updates.
func (a *authScheme) Batch() AuthBatchCreator {
	return &authBatchBuilder{
		scheme: a.scheme,
		token:  a.token,
		items:  make([]JSONItem, 0),
	}
}

// execute opens a Things URL scheme for create/update operations.
func (s *scheme) execute(ctx context.Context, uri string) error {
	if s.foreground {
		return exec.CommandContext(ctx, "open", uri).Run()
	}
	script := fmt.Sprintf(`tell application "Things3" to open location %q`, uri)
	return exec.CommandContext(ctx, "osascript", "-e", script).Run()
}

// executeNavigation opens a Things URL scheme for navigation operations.
func (s *scheme) executeNavigation(ctx context.Context, uri string) error {
	if !s.background {
		return exec.CommandContext(ctx, "open", uri).Run()
	}
	script := fmt.Sprintf(`tell application "Things3" to open location %q`, uri)
	return exec.CommandContext(ctx, "osascript", "-e", script).Run()
}

// Show opens Things and shows the item with the given UUID.
func (s *scheme) Show(ctx context.Context, uuid string) error {
	uri, err := s.ShowBuilder().ID(uuid).Build()
	if err != nil {
		return err
	}
	return s.executeNavigation(ctx, uri)
}

// Search opens Things and performs a search for the given query.
func (s *scheme) Search(ctx context.Context, query string) error {
	uri := s.SearchURL(query)
	return s.executeNavigation(ctx, uri)
}
