package things3

import (
	"fmt"
	"net/url"
)

// Scheme provides URL building for Things URL Scheme.
// It is stateless and can be reused for multiple URL builds.
//
// Use NewScheme() to create a new instance:
//
//	scheme := things3.NewScheme()
//	url, _ := scheme.Todo().Title("Buy groceries").Build()
//
// For operations requiring authentication (update operations),
// use WithToken() to get an AuthScheme:
//
//	auth := scheme.WithToken(token)
//	url, _ := auth.UpdateTodo("uuid").Completed(true).Build()
type Scheme struct{}

// NewScheme creates a new URL Scheme builder.
func NewScheme() *Scheme {
	return &Scheme{}
}

// Todo returns a TodoBuilder for creating a new to-do.
func (s *Scheme) Todo() *TodoBuilder {
	return &TodoBuilder{
		params: make(map[string]string),
	}
}

// Project returns a ProjectBuilder for creating a new project.
func (s *Scheme) Project() *ProjectBuilder {
	return &ProjectBuilder{
		params: make(map[string]string),
	}
}

// Show returns a ShowBuilder for navigating to items or lists.
func (s *Scheme) Show() *ShowBuilder {
	return &ShowBuilder{
		params: make(map[string]string),
	}
}

// JSON returns a JSONBuilder for batch create operations.
// For operations including updates, use AuthScheme.JSON() instead.
func (s *Scheme) JSON() *JSONBuilder {
	return &JSONBuilder{
		items: make([]JSONItem, 0),
	}
}

// Search returns a URL to search for the given query in Things.
func (s *Scheme) Search(query string) string {
	q := url.Values{}
	q.Set("query", query)
	return fmt.Sprintf("things:///%s?%s", CommandSearch, q.Encode())
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
func (s *Scheme) WithToken(token string) *AuthScheme {
	return &AuthScheme{
		token: token,
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
	token string
}

// UpdateTodo returns an UpdateTodoBuilder for modifying an existing to-do.
func (a *AuthScheme) UpdateTodo(id string) *UpdateTodoBuilder {
	return &UpdateTodoBuilder{
		token:  a.token,
		id:     id,
		params: make(map[string]string),
	}
}

// UpdateProject returns an UpdateProjectBuilder for modifying an existing project.
func (a *AuthScheme) UpdateProject(id string) *UpdateProjectBuilder {
	return &UpdateProjectBuilder{
		token:  a.token,
		id:     id,
		params: make(map[string]string),
	}
}

// JSON returns an AuthJSONBuilder for batch operations including updates.
func (a *AuthScheme) JSON() *AuthJSONBuilder {
	return &AuthJSONBuilder{
		token: a.token,
		items: make([]JSONItem, 0),
	}
}
