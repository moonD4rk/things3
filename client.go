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
	database *db
	scheme   *scheme

	// Token management with mutex (not sync.Once to allow retry on transient failures)
	tokenMu    sync.Mutex
	tokenCache string
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
	var dbOpts []dbOption
	if options.databasePath != "" {
		dbOpts = append(dbOpts, withDBPath(options.databasePath))
	}
	if options.printSQL {
		dbOpts = append(dbOpts, withDBPrintSQL(options.printSQL))
	}

	// Build Scheme options
	var schemeOpts []schemeOption
	if options.foreground {
		schemeOpts = append(schemeOpts, withForeground())
	}
	if options.background {
		schemeOpts = append(schemeOpts, withBackground())
	}

	// Create DB connection
	database, err := newDB(dbOpts...)
	if err != nil {
		return nil, err
	}

	// Create Scheme
	s := newScheme(schemeOpts...)

	client := &Client{
		database: database,
		scheme:   s,
	}

	// Preload token if requested
	if options.preloadToken {
		if _, err := client.Token(context.Background()); err != nil {
			database.Close()
			return nil, err
		}
	}

	return client, nil
}

// Close closes the database connection.
func (c *Client) Close() error {
	if c.database != nil {
		return c.database.Close()
	}
	return nil
}

// ============================================================================
// Token Management
// ============================================================================

// ensureToken ensures the authentication token is loaded.
// Uses mutex for thread-safe lazy initialization.
// Unlike sync.Once, this allows retry on transient failures.
func (c *Client) ensureToken(ctx context.Context) (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	// Return cached token if available
	if c.tokenCache != "" {
		return c.tokenCache, nil
	}

	// Fetch token from database
	token, err := c.database.Token(ctx)
	if err != nil {
		return "", err
	}

	// Cache successful result
	c.tokenCache = token
	return token, nil
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
	return c.database.Inbox(ctx)
}

// Today returns tasks that would appear in Today view.
func (c *Client) Today(ctx context.Context) ([]Task, error) {
	return c.database.Today(ctx)
}

// Todos returns all incomplete to-do items.
func (c *Client) Todos(ctx context.Context) ([]Task, error) {
	return c.database.Todos(ctx)
}

// Projects returns all incomplete projects.
func (c *Client) Projects(ctx context.Context) ([]Task, error) {
	return c.database.Projects(ctx)
}

// Upcoming returns tasks scheduled for future dates.
func (c *Client) Upcoming(ctx context.Context) ([]Task, error) {
	return c.database.Upcoming(ctx)
}

// Anytime returns tasks in the Anytime list.
func (c *Client) Anytime(ctx context.Context) ([]Task, error) {
	return c.database.Anytime(ctx)
}

// Someday returns tasks in the Someday list.
func (c *Client) Someday(ctx context.Context) ([]Task, error) {
	return c.database.Someday(ctx)
}

// Logbook returns completed and canceled tasks.
func (c *Client) Logbook(ctx context.Context) ([]Task, error) {
	return c.database.Logbook(ctx)
}

// Trash returns trashed tasks.
func (c *Client) Trash(ctx context.Context) ([]Task, error) {
	return c.database.Trash(ctx)
}

// Completed returns completed tasks.
func (c *Client) Completed(ctx context.Context) ([]Task, error) {
	return c.database.Completed(ctx)
}

// Canceled returns canceled tasks.
func (c *Client) Canceled(ctx context.Context) ([]Task, error) {
	return c.database.Canceled(ctx)
}

// Deadlines returns tasks with deadlines, sorted by deadline.
func (c *Client) Deadlines(ctx context.Context) ([]Task, error) {
	return c.database.Deadlines(ctx)
}

// CreatedWithin returns tasks created after the specified time.
func (c *Client) CreatedWithin(ctx context.Context, since time.Time) ([]Task, error) {
	return c.database.CreatedWithin(ctx, since)
}

// ============================================================================
// Query Operations - Query Builders
// ============================================================================

// Tasks creates a new TaskQueryBuilder for querying tasks.
func (c *Client) Tasks() TaskQueryBuilder {
	return c.database.Tasks()
}

// Areas creates a new AreaQueryBuilder for querying areas.
func (c *Client) Areas() AreaQueryBuilder {
	return c.database.Areas()
}

// Tags creates a new TagQueryBuilder for querying tags.
func (c *Client) Tags() TagQueryBuilder {
	return c.database.Tags()
}

// ============================================================================
// Query Operations - Utilities
// ============================================================================

// Get retrieves an object by UUID.
// Returns a Task, Area, or Tag depending on what is found.
// Returns nil if not found.
func (c *Client) Get(ctx context.Context, uuid string) (any, error) {
	return c.database.Get(ctx, uuid)
}

// Search searches for tasks matching the query.
func (c *Client) Search(ctx context.Context, query string) ([]Task, error) {
	return c.database.Search(ctx, query)
}

// ChecklistItems returns the checklist items for a to-do.
func (c *Client) ChecklistItems(ctx context.Context, todoUUID string) ([]ChecklistItem, error) {
	return c.database.ChecklistItems(ctx, todoUUID)
}

// ============================================================================
// Add Operations
// ============================================================================

// AddTodo returns a TodoAdder for creating a new to-do.
//
// Example:
//
//	client.AddTodo().
//	    Title("Buy milk").
//	    Notes("From the grocery store").
//	    When(things3.Today()).
//	    Execute(ctx)
func (c *Client) AddTodo() TodoAdder {
	return c.scheme.AddTodo()
}

// AddProject returns a ProjectAdder for creating a new project.
//
// Example:
//
//	client.AddProject().
//	    Title("Home Renovation").
//	    Notes("Kitchen and bathroom").
//	    Execute(ctx)
func (c *Client) AddProject() ProjectAdder {
	return c.scheme.AddProject()
}

// Batch returns a BatchCreator for batch create operations.
//
// Example:
//
//	client.Batch().
//	    AddTodo(func(b BatchTodoConfigurator) {
//	        b.Title("Task 1")
//	    }).
//	    AddTodo(func(b BatchTodoConfigurator) {
//	        b.Title("Task 2")
//	    }).
//	    Execute(ctx)
func (c *Client) Batch() BatchCreator {
	return c.scheme.Batch()
}

// AuthBatch returns an AuthBatchCreator for batch operations including updates.
// The authentication token is fetched automatically on first use.
//
// Example:
//
//	client.AuthBatch().
//	    AddTodo(func(b BatchTodoConfigurator) {
//	        b.Title("New task")
//	    }).
//	    UpdateTodo("uuid", func(b BatchTodoConfigurator) {
//	        b.Completed(true)
//	    }).
//	    Execute(ctx)
func (c *Client) AuthBatch() AuthBatchCreator {
	return &authBatchBuilder{
		scheme:    c.scheme,
		tokenFunc: c.ensureToken,
		items:     make([]JSONItem, 0),
	}
}

// ============================================================================
// Update Operations
// ============================================================================

// UpdateTodo returns a TodoUpdater for modifying an existing to-do.
// The authentication token is fetched automatically on first use.
//
// Example:
//
//	client.UpdateTodo(uuid).
//	    Completed(true).
//	    Execute(ctx)
func (c *Client) UpdateTodo(id string) TodoUpdater {
	return &updateTodoBuilder{
		scheme:    c.scheme,
		tokenFunc: c.ensureToken,
		id:        id,
		attrs:     urlAttrs{params: make(map[string]string)},
	}
}

// UpdateProject returns a ProjectUpdater for modifying an existing project.
// The authentication token is fetched automatically on first use.
//
// Example:
//
//	client.UpdateProject(uuid).
//	    Title("Renamed Project").
//	    Execute(ctx)
func (c *Client) UpdateProject(id string) ProjectUpdater {
	return &updateProjectBuilder{
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
	uri, err := c.scheme.ShowBuilder().List(list).Build()
	if err != nil {
		return err
	}
	return c.scheme.executeNavigation(ctx, uri)
}

// ShowSearch opens Things and performs a search for the given query.
// By default, brings Things to foreground since the user wants to view results.
func (c *Client) ShowSearch(ctx context.Context, query string) error {
	return c.scheme.Search(ctx, query)
}

// ShowBuilder returns a ShowNavigator for complex navigation operations.
func (c *Client) ShowBuilder() ShowNavigator {
	return c.scheme.ShowBuilder()
}
