// Package things3 provides read-only access to the Things 3 macOS application's SQLite database.
//
// This package is a Go port of the Python things.py library, offering full API parity
// for querying tasks, projects, areas, and tags from the Things 3 app.
//
// # Basic Usage
//
// Create a database connection and query tasks:
//
//	db, err := things3.NewDB()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer db.Close()
//
//	// Get inbox tasks
//	inbox, err := db.Inbox(ctx)
//
//	// Get all incomplete to-dos
//	todos, err := db.Todos(ctx)
//
// # Query Builder
//
// For complex queries, use the fluent query builder:
//
//	tasks, err := db.Tasks().
//	    WithType(things3.TaskTypeTodo).
//	    WithStatus(things3.StatusIncomplete).
//	    InProject("project-uuid").
//	    WithDeadline(true).
//	    All(ctx)
//
// # Configuration
//
// Configure the database with functional options:
//
//	// Use custom database path
//	db, err := things3.NewDB(things3.WithDBPath("/path/to/main.sqlite"))
//
//	// Enable SQL query logging
//	db, err := things3.NewDB(things3.WithPrintSQL(true))
//
// # Database Discovery
//
// The database path is discovered in the following order:
//  1. Custom path provided via WithDBPath option
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
