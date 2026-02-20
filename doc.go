// Package things3 provides a Go library for Things 3 on macOS with read-only database
// access and full URL Scheme support for creating and updating tasks.
//
// # Features
//
// This package offers two main capabilities:
//   - Read-only access to the Things 3 SQLite database for querying tasks, projects, areas, and tags
//   - Full Things URL Scheme support for creating, updating, and navigating to items
//
// # Getting Started
//
// All operations go through a single Client:
//
//	client, err := things3.NewClient()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	// Convenience methods
//	inbox, _ := client.Inbox(ctx)
//	today, _ := client.Today(ctx)
//	todos, _ := client.Todos(ctx)
//
// # Fluent Query Builder
//
// For complex queries, use the type-safe fluent query builder:
//
//	tasks, _ := client.Tasks().
//	    Type().Todo().
//	    Status().Incomplete().
//	    StartDate().Future().
//	    All(ctx)
//
// # URL Scheme
//
// Create and update items via Things URL Scheme:
//
//	// Create a new to-do
//	client.AddTodo().
//	    Title("Buy groceries").
//	    When(things3.Today()).
//	    Tags("shopping").
//	    Execute(ctx)
//
//	// Update existing items (auth token managed automatically)
//	client.UpdateTodo("uuid").Completed(true).Execute(ctx)
//
// # Configuration
//
// Configure the client with functional options:
//
//	client, _ := things3.NewClient(things3.WithDatabasePath("/path/to/main.sqlite"))
//	client, _ := things3.NewClient(things3.WithPrintSQL(true))
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
