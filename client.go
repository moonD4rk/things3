package things3

import (
	"context"
	"sync"
	"time"
)

// Client provides unified access to Things 3 database and URL scheme operations.
// It combines read-only database access with URL scheme write operations,
// handling authentication token management automatically.
//
// Create a client using NewClient:
//
//	client, err := things3.NewClient()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
// Query operations read from the Things 3 database:
//
//	tasks, _ := client.Inbox(ctx)
//	tasks, _ := client.Tasks().Status().Incomplete().All(ctx)
//
// Add operations create new items via URL scheme:
//
//	client.AddTodo().Title("Buy milk").Execute(ctx)
//
// Update operations automatically manage authentication tokens:
//
//	client.UpdateTodo(uuid).Completed(true).Execute(ctx)
//
// Show operations display items in the Things app:
//
//	client.Show(ctx, uuid)
type Client struct {
	db     *DB
	scheme *Scheme

	// Token management
	tokenOnce  sync.Once
	tokenCache string
	tokenErr   error
}

// NewClient creates a new unified Things 3 client.
// It opens a database connection and initializes the URL scheme handler.
//
// Returns an error if the database cannot be found or opened.
//
// Example:
//
//	client, err := things3.NewClient()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	// With options
//	client, err := things3.NewClient(
//	    things3.WithDatabasePath("/custom/path"),
//	    things3.WithPrintSQL(true),
//	)
func NewClient(opts ...ClientOption) (*Client, error) {
	options := &clientOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Build DB options
	var dbOpts []DBOption
	if options.databasePath != "" {
		dbOpts = append(dbOpts, WithDBPath(options.databasePath))
	}
	if options.printSQL {
		dbOpts = append(dbOpts, withDBPrintSQL(options.printSQL))
	}

	// Build Scheme options
	var schemeOpts []SchemeOption
	if options.foreground {
		schemeOpts = append(schemeOpts, WithForeground())
	}
	if options.background {
		schemeOpts = append(schemeOpts, WithBackground())
	}

	// Create DB connection
	db, err := NewDB(dbOpts...)
	if err != nil {
		return nil, err
	}

	// Create Scheme
	scheme := NewScheme(schemeOpts...)

	client := &Client{
		db:     db,
		scheme: scheme,
	}

	// Preload token if requested
	if options.preloadToken {
		if _, err := client.Token(context.Background()); err != nil {
			db.Close()
			return nil, err
		}
	}

	return client, nil
}

// Close closes the database connection.
func (c *Client) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// ============================================================================
// Token Management
// ============================================================================

// ensureToken ensures the authentication token is loaded.
// Uses sync.Once for thread-safe lazy initialization.
func (c *Client) ensureToken(ctx context.Context) (string, error) {
	c.tokenOnce.Do(func() {
		c.tokenCache, c.tokenErr = c.db.Token(ctx)
	})
	return c.tokenCache, c.tokenErr
}

// Token returns the cached authentication token, fetching it if needed.
// Most users should not need this; use UpdateTodo/UpdateProject directly.
func (c *Client) Token(ctx context.Context) (string, error) {
	return c.ensureToken(ctx)
}

// ============================================================================
// Query Operations - Convenience Methods
// ============================================================================

// Inbox returns all tasks in the Inbox.
func (c *Client) Inbox(ctx context.Context) ([]Task, error) {
	return c.db.Inbox(ctx)
}

// Today returns tasks that would appear in Today view.
func (c *Client) Today(ctx context.Context) ([]Task, error) {
	return c.db.Today(ctx)
}

// Todos returns all incomplete to-do items.
func (c *Client) Todos(ctx context.Context) ([]Task, error) {
	return c.db.Todos(ctx)
}

// Projects returns all incomplete projects.
func (c *Client) Projects(ctx context.Context) ([]Task, error) {
	return c.db.Projects(ctx)
}

// Upcoming returns tasks scheduled for future dates.
func (c *Client) Upcoming(ctx context.Context) ([]Task, error) {
	return c.db.Upcoming(ctx)
}

// Anytime returns tasks in the Anytime list.
func (c *Client) Anytime(ctx context.Context) ([]Task, error) {
	return c.db.Anytime(ctx)
}

// Someday returns tasks in the Someday list.
func (c *Client) Someday(ctx context.Context) ([]Task, error) {
	return c.db.Someday(ctx)
}

// Logbook returns completed and canceled tasks.
func (c *Client) Logbook(ctx context.Context) ([]Task, error) {
	return c.db.Logbook(ctx)
}

// Trash returns trashed tasks.
func (c *Client) Trash(ctx context.Context) ([]Task, error) {
	return c.db.Trash(ctx)
}

// Completed returns completed tasks.
func (c *Client) Completed(ctx context.Context) ([]Task, error) {
	return c.db.Completed(ctx)
}

// Canceled returns canceled tasks.
func (c *Client) Canceled(ctx context.Context) ([]Task, error) {
	return c.db.Canceled(ctx)
}

// Deadlines returns tasks with deadlines, sorted by deadline.
func (c *Client) Deadlines(ctx context.Context) ([]Task, error) {
	return c.db.Deadlines(ctx)
}

// CreatedWithin returns tasks created after the specified time.
func (c *Client) CreatedWithin(ctx context.Context, since time.Time) ([]Task, error) {
	return c.db.CreatedWithin(ctx, since)
}

// ============================================================================
// Query Operations - Query Builders
// ============================================================================

// Tasks creates a new TaskQuery for querying tasks.
func (c *Client) Tasks() *TaskQuery {
	return c.db.Tasks()
}

// Areas creates a new AreaQuery for querying areas.
func (c *Client) Areas() *AreaQuery {
	return c.db.Areas()
}

// Tags creates a new TagQuery for querying tags.
func (c *Client) Tags() *TagQuery {
	return c.db.Tags()
}

// ============================================================================
// Query Operations - Utilities
// ============================================================================

// Get retrieves an object by UUID.
// Returns a Task, Area, or Tag depending on what is found.
// Returns nil if not found.
func (c *Client) Get(ctx context.Context, uuid string) (any, error) {
	return c.db.Get(ctx, uuid)
}

// Search searches for tasks matching the query.
func (c *Client) Search(ctx context.Context, query string) ([]Task, error) {
	return c.db.Search(ctx, query)
}

// ChecklistItems returns the checklist items for a to-do.
func (c *Client) ChecklistItems(ctx context.Context, todoUUID string) ([]ChecklistItem, error) {
	return c.db.ChecklistItems(ctx, todoUUID)
}

// ============================================================================
// Add Operations
// ============================================================================

// AddTodo returns an AddTodoBuilder for creating a new to-do.
//
// Example:
//
//	client.AddTodo().
//	    Title("Buy milk").
//	    Notes("From the grocery store").
//	    When(things3.Today()).
//	    Execute(ctx)
func (c *Client) AddTodo() *AddTodoBuilder {
	return c.scheme.AddTodo()
}

// AddProject returns an AddProjectBuilder for creating a new project.
//
// Example:
//
//	client.AddProject().
//	    Title("Home Renovation").
//	    Notes("Kitchen and bathroom").
//	    Execute(ctx)
func (c *Client) AddProject() *AddProjectBuilder {
	return c.scheme.AddProject()
}

// Batch returns a BatchBuilder for batch create operations.
//
// Example:
//
//	client.Batch().
//	    AddTodo(func(b *BatchTodoBuilder) {
//	        b.Title("Task 1")
//	    }).
//	    AddTodo(func(b *BatchTodoBuilder) {
//	        b.Title("Task 2")
//	    }).
//	    Execute(ctx)
func (c *Client) Batch() *BatchBuilder {
	return c.scheme.Batch()
}

// ============================================================================
// Update Operations
// ============================================================================

// UpdateTodo returns an UpdateTodoBuilder for modifying an existing to-do.
// The authentication token is fetched automatically on first use.
//
// Example:
//
//	client.UpdateTodo(uuid).
//	    Completed(true).
//	    Execute(ctx)
func (c *Client) UpdateTodo(id string) *UpdateTodoBuilder {
	return &UpdateTodoBuilder{
		scheme:    c.scheme,
		tokenFunc: c.ensureToken,
		id:        id,
		attrs:     urlAttrs{params: make(map[string]string)},
	}
}

// UpdateProject returns an UpdateProjectBuilder for modifying an existing project.
// The authentication token is fetched automatically on first use.
//
// Example:
//
//	client.UpdateProject(uuid).
//	    Title("Renamed Project").
//	    Execute(ctx)
func (c *Client) UpdateProject(id string) *UpdateProjectBuilder {
	return &UpdateProjectBuilder{
		scheme:    c.scheme,
		tokenFunc: c.ensureToken,
		id:        id,
		attrs:     urlAttrs{params: make(map[string]string)},
	}
}

// ============================================================================
// Show Operations
// ============================================================================

// Show opens Things and displays the item with the given UUID.
// By default, brings Things to foreground since the user wants to view the item.
// Use WithBackgroundNavigation() option to run in background without stealing focus.
func (c *Client) Show(ctx context.Context, uuid string) error {
	return c.scheme.Show(ctx, uuid)
}

// ShowList opens Things and displays the specified list.
// Use ListID constants like ListInbox, ListToday, etc.
//
// Example:
//
//	client.ShowList(ctx, things3.ListToday)
func (c *Client) ShowList(ctx context.Context, list ListID) error {
	uri := c.scheme.ShowBuilder().List(list).Build()
	return c.scheme.executeNavigation(ctx, uri)
}

// ShowSearch opens Things and performs a search for the given query.
// By default, brings Things to foreground since the user wants to view results.
func (c *Client) ShowSearch(ctx context.Context, query string) error {
	return c.scheme.Search(ctx, query)
}

// ShowBuilder returns a ShowBuilder for complex navigation operations.
func (c *Client) ShowBuilder() *ShowBuilder {
	return c.scheme.ShowBuilder()
}
