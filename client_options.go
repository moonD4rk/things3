package things3

import "time"

// clientOptions holds the configuration options for the Client.
type clientOptions struct {
	// Database options
	databasePath string
	printSQL     bool
	clock        func() time.Time

	// Scheme options
	foreground bool // bring Things to foreground for create/update
	background bool // keep Things in background for navigation

	// Token options
	preloadToken bool // fetch token immediately during NewClient
}

// ClientOption is a functional option for configuring the Client.
type ClientOption func(*clientOptions)

// WithDatabasePath sets a custom path to the Things database.
// If not set, the database path is discovered automatically.
//
// Example:
//
//	client, err := things3.NewClient(things3.WithDatabasePath("/path/to/main.sqlite"))
func WithDatabasePath(path string) ClientOption {
	return func(opts *clientOptions) {
		opts.databasePath = path
	}
}

// WithPrintSQL enables SQL query logging to stdout.
// Useful for debugging and understanding the queries being executed.
//
// Example:
//
//	client, err := things3.NewClient(things3.WithPrintSQL(true))
func WithPrintSQL(enabled bool) ClientOption {
	return func(opts *clientOptions) {
		opts.printSQL = enabled
	}
}

// WithClock replaces the system clock behind every relative date decision: the
// Past and Future date filters, the composed views, and [Client.Now]. Use it to
// make a report reproducible or to pin "today" when testing against a fixed
// database.
//
// Example:
//
//	fixed := time.Date(2026, time.January, 1, 9, 0, 0, 0, time.Local)
//	client, err := things3.NewClient(things3.WithClock(func() time.Time { return fixed }))
func WithClock(now func() time.Time) ClientOption {
	return func(opts *clientOptions) {
		opts.clock = now
	}
}

// WithForegroundExecution configures the Client to bring Things to foreground
// when executing create/update operations (AddTodo, AddProject, UpdateTodo, etc.).
//
// By default, create/update operations run in background without stealing focus.
// Use this option when you want Things to become the active window after operations.
//
// Example:
//
//	client, err := things3.NewClient(things3.WithForegroundExecution())
//	client.AddTodo().Title("Buy milk").Execute(ctx)  // Things comes to foreground
func WithForegroundExecution() ClientOption {
	return func(opts *clientOptions) {
		opts.foreground = true
	}
}

// WithBackgroundNavigation configures the Client to run navigation operations
// (Show, ShowList, ShowSearch) in the background without stealing focus.
//
// By default, navigation operations bring Things to foreground since
// the user typically wants to view the content.
// Use this option for programmatic navigation where focus change is undesired.
//
// Example:
//
//	client, err := things3.NewClient(things3.WithBackgroundNavigation())
//	client.Show(ctx, "uuid")  // Things stays in background
func WithBackgroundNavigation() ClientOption {
	return func(opts *clientOptions) {
		opts.background = true
	}
}

// WithPreloadToken fetches the authentication token immediately during NewClient()
// instead of lazily on first update operation.
//
// Use this option when you know you will need authenticated operations
// and want to fail fast if the token cannot be retrieved.
//
// Example:
//
//	client, err := things3.NewClient(things3.WithPreloadToken())
//	if err != nil {
//	    // May fail due to token retrieval error
//	}
func WithPreloadToken() ClientOption {
	return func(opts *clientOptions) {
		opts.preloadToken = true
	}
}
