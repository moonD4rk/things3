// Package things3 provides a Go library for Things 3 on macOS with read-only database
// access and full URL Scheme support for creating and updating tasks.
//
// # Features
//
// This package offers two main capabilities:
//   - Read-only access to the Things 3 SQLite database for querying tasks, projects, areas, and tags
//   - Full Things URL Scheme support for creating, updating, and navigating to items
//
// # Database Access
//
// Create a database connection and query tasks:
//
//	db, err := things3.NewDB()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer db.Close()
//
//	// Convenience methods
//	inbox, _ := db.Inbox(ctx)
//	today, _ := db.Today(ctx)
//	todos, _ := db.Todos(ctx)
//
// # Fluent Query Builder
//
// For complex queries, use the type-safe fluent query builder:
//
//	tasks, _ := db.Tasks().
//	    Type().Todo().
//	    Status().Incomplete().
//	    StartDate().Future().
//	    All(ctx)
//
// # URL Scheme
//
// Create Things URLs for automation and integration:
//
//	scheme := things3.NewScheme()
//
//	// Create a new todo
//	url, _ := scheme.Todo().
//	    Title("Buy groceries").
//	    When(things3.WhenToday).
//	    Tags("shopping").
//	    Build()
//
//	// Update existing items (requires auth token)
//	token, _ := db.Token(ctx)
//	auth := scheme.WithToken(token)
//	url, _ := auth.UpdateTodo("uuid").Completed(true).Build()
//
// # Configuration
//
// Configure the database with functional options:
//
//	db, _ := things3.NewDB(things3.WithDatabasePath("/path/to/main.sqlite"))
//	db, _ := things3.NewDB(things3.WithPrintSQL(true))
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
//
// # References
//
// For the official Things URL Scheme documentation, see:
// https://culturedcode.com/things/support/articles/2803573/
package things3
