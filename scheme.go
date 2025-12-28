package things3

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
)

// Scheme provides URL building and execution for Things URL Scheme.
//
// Use NewScheme() to create a new instance:
//
//	scheme := things3.NewScheme()
//	url, _ := scheme.Todo().Title("Buy groceries").Build()
//
// For executing URL scheme operations:
//
//	// Show a task (runs in background by default)
//	scheme.Show(ctx, "uuid")
//
//	// Show a task in foreground
//	things3.NewScheme(things3.WithForeground()).Show(ctx, "uuid")
//
// For operations requiring authentication (update operations),
// use WithToken() to get an AuthScheme:
//
//	token, _ := db.Token(ctx)
//	auth := scheme.WithToken(token)
//	auth.UpdateTodo("uuid").Completed(true).Execute(ctx)
type Scheme struct {
	foreground bool
}

// NewScheme creates a new URL Scheme builder.
// Options can be provided to configure execution behavior.
func NewScheme(opts ...SchemeOption) *Scheme {
	s := &Scheme{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Todo returns a TodoBuilder for creating a new to-do.
func (s *Scheme) Todo() *TodoBuilder {
	return &TodoBuilder{
		scheme: s,
		attrs:  urlAttrs{params: make(map[string]string)},
	}
}

// Project returns a ProjectBuilder for creating a new project.
func (s *Scheme) Project() *ProjectBuilder {
	return &ProjectBuilder{
		scheme: s,
		attrs:  urlAttrs{params: make(map[string]string)},
	}
}

// ShowBuilder returns a ShowBuilder for navigating to items or lists.
// For direct execution, use Show(ctx, uuid) instead.
func (s *Scheme) ShowBuilder() *ShowBuilder {
	return &ShowBuilder{
		scheme: s,
		params: make(map[string]string),
	}
}

// JSON returns a JSONBuilder for batch create operations.
// For operations including updates, use AuthScheme.JSON() instead.
func (s *Scheme) JSON() *JSONBuilder {
	return &JSONBuilder{
		scheme: s,
		items:  make([]JSONItem, 0),
	}
}

// SearchURL returns a URL to search for the given query in Things.
// For direct execution, use Search(ctx, query) instead.
func (s *Scheme) SearchURL(query string) string {
	q := url.Values{}
	q.Set("query", query)
	return fmt.Sprintf("things:///%s?%s", CommandSearch, encodeQuery(q))
}

// Version returns a URL to get Things version information.
func (s *Scheme) Version() string {
	return fmt.Sprintf("things:///%s", CommandVersion)
}

// WithToken returns an AuthScheme for authenticated operations.
// The token is required for update operations (UpdateTodo, UpdateProject).
//
// Get the token from the database:
//
//	db, _ := things3.NewDB()
//	token, _ := db.Token(ctx)
//	auth := scheme.WithToken(token)
//	auth.UpdateTodo("uuid").Completed(true).Execute(ctx)
func (s *Scheme) WithToken(token string) *AuthScheme {
	return &AuthScheme{
		scheme: s,
		token:  token,
	}
}

// AuthScheme provides URL building for authenticated operations.
// Obtained via Scheme.WithToken(token).
//
// AuthScheme exposes update methods that require authentication:
//   - UpdateTodo(id) - modify an existing to-do
//   - UpdateProject(id) - modify an existing project
//   - JSON() - batch operations including updates
type AuthScheme struct {
	scheme *Scheme
	token  string
}

// UpdateTodo returns an UpdateTodoBuilder for modifying an existing to-do.
func (a *AuthScheme) UpdateTodo(id string) *UpdateTodoBuilder {
	return &UpdateTodoBuilder{
		scheme: a.scheme,
		token:  a.token,
		id:     id,
		attrs:  urlAttrs{params: make(map[string]string)},
	}
}

// UpdateProject returns an UpdateProjectBuilder for modifying an existing project.
func (a *AuthScheme) UpdateProject(id string) *UpdateProjectBuilder {
	return &UpdateProjectBuilder{
		scheme: a.scheme,
		token:  a.token,
		id:     id,
		attrs:  urlAttrs{params: make(map[string]string)},
	}
}

// JSON returns an AuthJSONBuilder for batch operations including updates.
func (a *AuthScheme) JSON() *AuthJSONBuilder {
	return &AuthJSONBuilder{
		scheme: a.scheme,
		token:  a.token,
		items:  make([]JSONItem, 0),
	}
}

// execute opens a Things URL scheme.
// By default, uses osascript to run in background without stealing focus.
// If foreground is true, uses open command to bring Things to foreground.
func (s *Scheme) execute(ctx context.Context, uri string) error {
	if s.foreground {
		return exec.CommandContext(ctx, "open", uri).Run()
	}
	script := fmt.Sprintf("open location %q", uri)
	return exec.CommandContext(ctx, "osascript", "-e", script).Run()
}

// Show opens Things and shows the item with the given UUID.
// By default, runs in background without stealing focus.
// Use WithForeground() option to bring Things to foreground.
func (s *Scheme) Show(ctx context.Context, uuid string) error {
	uri := s.ShowBuilder().ID(uuid).Build()
	return s.execute(ctx, uri)
}

// Search opens Things and performs a search for the given query.
// By default, runs in background without stealing focus.
// Use WithForeground() option to bring Things to foreground.
func (s *Scheme) Search(ctx context.Context, query string) error {
	uri := s.SearchURL(query)
	return s.execute(ctx, uri)
}
