// Package things3 provides read-only access to the Things 3 macOS application's SQLite database.
//
// This package is a Go port of the Python things.py library, offering full API parity
// for querying tasks, projects, areas, and tags from the Things 3 app.
//
// # Basic Usage
//
// Create a client and query tasks:
//
//	client, err := things3.New()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	// Get inbox tasks
//	inbox, err := client.Inbox(ctx)
//
//	// Get all incomplete to-dos
//	todos, err := client.Todos(ctx)
//
// # Query Builder
//
// For complex queries, use the fluent query builder:
//
//	tasks, err := client.Tasks().
//	    WithType(things3.TaskTypeTodo).
//	    WithStatus(things3.StatusIncomplete).
//	    InProject("project-uuid").
//	    WithDeadline(things3.DateOpExists).
//	    All(ctx)
//
// # Configuration
//
// Configure the client with functional options:
//
//	// Use custom database path
//	client, err := things3.New(things3.WithDatabasePath("/path/to/main.sqlite"))
//
//	// Enable SQL query logging
//	client, err := things3.New(things3.WithPrintSQL(true))
//
// # Database Discovery
//
// The database path is discovered in the following order:
//  1. Custom path provided via WithDatabasePath option
//  2. THINGSDB environment variable
//  3. Auto-discovery of default Things 3 database location
//
// # Type System
//
// The package uses integer-based enums that map directly to database values:
//   - TaskType: TaskTypeTodo (0), TaskTypeProject (1), TaskTypeHeading (2)
//   - Status: StatusIncomplete (0), StatusCanceled (2), StatusCompleted (3)
//   - StartBucket: StartInbox (0), StartAnytime (1), StartSomeday (2)
package things3
