# RFC 004: URL Scheme

Status: Draft
Author: @moond4rk

## Summary

This RFC defines the URL Scheme API layer of the things3 Go library. The `NewScheme()` entry point provides type-safe URL building and execution for Things 3 URL Scheme commands with compile-time enforcement of authentication requirements.

## Motivation

### Current Issues

1. **Type Safety**: Using `map[string]string` for URL parameters provides no compile-time safety
2. **Missing Commands**: `version` and `json` commands are not implemented
3. **Missing Enums**: No type-safe values for `when` keywords or built-in list IDs
4. **Token Confusion**: Unclear when authentication token is required
5. **Variable Naming**: Methods return "URL" but internal variables use `uri` (inconsistent)

### Goals

- Type-safe URL parameter construction via builder pattern
- Complete coverage of all Things URL Scheme commands
- Compile-time enforcement of token requirements via `WithToken()` pattern
- Enable IDE autocomplete for parameters and values
- Follow Go idioms and naming conventions

### Non-Goals

- x-callback-url response handling
- Rate limiting enforcement (document only, not enforce)

## Design

### Architecture Overview

```
NewScheme(opts...)  -> *Scheme  (URL building + execution)
    |
    +-- Options: WithForeground()
    |
    +-- [Execution Methods]
    |   +-- Show(ctx, uuid)      -> error  (execute directly)
    |   +-- Search(ctx, query)   -> error  (execute directly)
    |
    +-- [URL Building - No Auth Required]
    |   +-- AddTodo()     -> *AddTodoBuilder    -> Build() string
    |   +-- AddProject()  -> *AddProjectBuilder -> Build() string
    |   +-- ShowBuilder() -> *ShowBuilder       -> Build() string
    |   +-- Batch()       -> *BatchBuilder      -> Build() string (add only)
    |   +-- SearchURL(query) -> string
    |   +-- Version()     -> string
    |
    +-- WithToken(token)  -> *AuthScheme  (Authenticated operations)
        +-- UpdateTodo(id)    -> *UpdateTodoBuilder    -> Build() | Execute(ctx)
        +-- UpdateProject(id) -> *UpdateProjectBuilder -> Build() | Execute(ctx)
        +-- Batch()           -> *AuthBatchBuilder     -> Build() string
```

**Execution Behavior:**

Execution behavior differs by operation type:

| Operation Type | Default Behavior | Override Option |
|----------------|------------------|-----------------|
| Navigation (Show, Search, ShowBuilder) | Foreground | `WithBackground()` |
| Create/Update (Todo, Project, JSON, Update*) | Background | `WithForeground()` |

- **Navigation operations** run in foreground by default since the user intends to view content
- **Create/Update operations** run in background by default for silent operation without stealing focus

### Token Requirement Pattern

The design uses **compile-time type safety** to enforce token requirements:

| Entry Point | Available Methods | Token Required |
|-------------|-------------------|----------------|
| `scheme.` | AddTodo, AddProject, Show, Batch, Search, Version | No |
| `scheme.WithToken(token).` | UpdateTodo, UpdateProject, Batch | Yes (upfront) |

This ensures:
- Users cannot accidentally call update operations without a token
- IDE autocomplete only shows valid methods for each context
- No runtime surprises about missing tokens

### Entry Points

```go
// Scheme provides URL building and execution for Things URL Scheme.
type Scheme struct {
    foreground bool // For create/update operations: if true, bring Things to foreground
    background bool // For navigation operations: if true, run in background
}

// SchemeOption configures Scheme behavior.
type SchemeOption func(*Scheme)

// WithForeground configures the Scheme to bring Things to foreground
// when executing create/update operations (Todo, Project, JSON, Update*).
// By default, create/update operations run in background without stealing focus.
func WithForeground() SchemeOption

// WithBackground configures the Scheme to run navigation operations
// (Show, Search, ShowBuilder) in the background without stealing focus.
// By default, navigation operations bring Things to foreground.
func WithBackground() SchemeOption

// NewScheme creates a new URL Scheme builder.
// Options can be provided to configure execution behavior.
func NewScheme(opts ...SchemeOption) *Scheme
```

#### Scheme Methods

```go
// Execution methods (run URL scheme operations)
func (s *Scheme) Show(ctx context.Context, uuid string) error
func (s *Scheme) Search(ctx context.Context, query string) error

// URL building methods
func (s *Scheme) AddTodo() *AddTodoBuilder
func (s *Scheme) AddProject() *AddProjectBuilder
func (s *Scheme) ShowBuilder() *ShowBuilder  // For building show URLs
func (s *Scheme) Batch() *BatchBuilder       // Only AddTodo, AddProject available

// Simple URL methods
func (s *Scheme) SearchURL(query string) string  // Returns URL string
func (s *Scheme) Version() string

// Get authenticated scheme for update operations
func (s *Scheme) WithToken(token string) *AuthScheme
```

#### AuthScheme (Token Required)

```go
// AuthScheme provides URL building for authenticated operations.
// Obtained via Scheme.WithToken(token).
//
// AuthScheme uses pointer reference to Scheme (not embedding) to:
// - Share configuration (foreground setting)
// - Enable execution via scheme.execute()
// - Hide non-auth methods (only Update* methods visible)
type AuthScheme struct {
    scheme *Scheme  // Pointer reference for delegation
    token  string
}

// Update builders
func (a *AuthScheme) UpdateTodo(id string) *UpdateTodoBuilder
func (a *AuthScheme) UpdateProject(id string) *UpdateProjectBuilder
func (a *AuthScheme) Batch() *AuthBatchBuilder  // AddTodo, AddProject, UpdateTodo, UpdateProject available
```

**Design Note:** AuthScheme uses pointer reference (not embedding) to Scheme. This is intentional:
- Embedding (`*Scheme`) would expose all Scheme methods on AuthScheme (e.g., `auth.AddTodo()`)
- Pointer reference only exposes Update methods, keeping the API clean and focused

## Type System

### Command Constants

```go
// Command represents Things URL Scheme commands.
type Command string

const (
    CommandShow          Command = "show"
    CommandAdd           Command = "add"
    CommandAddProject    Command = "add-project"
    CommandUpdate        Command = "update"
    CommandUpdateProject Command = "update-project"
    CommandSearch        Command = "search"
    CommandVersion       Command = "version"
    CommandJSON          Command = "json"
)
```

### Scheduling API

The scheduling API uses `time.Time` for all date operations, with dedicated methods for Things 3-specific concepts that cannot be expressed as dates.

```go
// Package-level convenience functions
func Today() time.Time     // returns today's date at midnight (00:00:00)
func Tomorrow() time.Time  // returns tomorrow's date at midnight (00:00:00)

// Builder methods for scheduling
func (b *TodoBuilder) When(t time.Time) *TodoBuilder  // schedule for specific date
func (b *TodoBuilder) WhenEvening() *TodoBuilder      // this evening
func (b *TodoBuilder) WhenAnytime() *TodoBuilder      // anytime (no specific date)
func (b *TodoBuilder) WhenSomeday() *TodoBuilder      // someday (indefinite future)
func (b *TodoBuilder) Deadline(t time.Time) *TodoBuilder  // deadline date
func (b *TodoBuilder) Reminder(hour, minute int) *TodoBuilder  // reminder time

// Internal type (private, not exported)
type when string

const (
    whenEvening when = "evening"
    whenAnytime when = "anytime"
    whenSomeday when = "someday"
)
```

**Usage Examples:**

```go
scheme.Todo().When(things3.Today())                      // today's date
scheme.Todo().When(things3.Tomorrow())                   // tomorrow's date
scheme.Todo().When(time.Now().AddDate(0, 0, 7))          // 7 days from now
scheme.Todo().WhenEvening()                              // this evening
scheme.Todo().WhenAnytime()                              // anytime
scheme.Todo().WhenSomeday()                              // someday
scheme.Todo().Deadline(time.Date(2025, 12, 31, 0, 0, 0, 0, time.Local))  // deadline
scheme.Todo().When(time.Now()).Reminder(14, 30)          // today at 14:30
```

### List IDs

```go
// ListID represents built-in Things list identifiers.
type ListID string

const (
    ListInbox          ListID = "inbox"
    ListToday          ListID = "today"
    ListAnytime        ListID = "anytime"
    ListUpcoming       ListID = "upcoming"
    ListSomeday        ListID = "someday"
    ListLogbook        ListID = "logbook"
    ListTomorrow       ListID = "tomorrow"
    ListDeadlines      ListID = "deadlines"
    ListRepeating      ListID = "repeating"
    ListAllProjects    ListID = "all-projects"
    ListLoggedProjects ListID = "logged-projects"
)
```

## Builders

### AddTodoBuilder (for add command)

```go
type AddTodoBuilder struct {
    params map[string]string
    errors []error
}

// Chainable methods
func (b *AddTodoBuilder) Title(title string) *AddTodoBuilder
func (b *AddTodoBuilder) Titles(titles ...string) *AddTodoBuilder
func (b *AddTodoBuilder) Notes(notes string) *AddTodoBuilder
func (b *AddTodoBuilder) When(t time.Time) *AddTodoBuilder      // schedule for specific date
func (b *AddTodoBuilder) WhenEvening() *AddTodoBuilder          // this evening
func (b *AddTodoBuilder) WhenAnytime() *AddTodoBuilder          // anytime
func (b *AddTodoBuilder) WhenSomeday() *AddTodoBuilder          // someday
func (b *AddTodoBuilder) Deadline(t time.Time) *AddTodoBuilder  // deadline date
func (b *AddTodoBuilder) Reminder(hour, minute int) *AddTodoBuilder
func (b *AddTodoBuilder) Tags(tags ...string) *AddTodoBuilder
func (b *AddTodoBuilder) ChecklistItems(items ...string) *AddTodoBuilder
func (b *AddTodoBuilder) List(name string) *AddTodoBuilder
func (b *AddTodoBuilder) ListID(id string) *AddTodoBuilder
func (b *AddTodoBuilder) Heading(name string) *AddTodoBuilder
func (b *AddTodoBuilder) HeadingID(id string) *AddTodoBuilder
func (b *AddTodoBuilder) Completed(completed bool) *AddTodoBuilder
func (b *AddTodoBuilder) Canceled(canceled bool) *AddTodoBuilder
func (b *AddTodoBuilder) Reveal(reveal bool) *AddTodoBuilder
func (b *AddTodoBuilder) CreationDate(date time.Time) *AddTodoBuilder
func (b *AddTodoBuilder) CompletionDate(date time.Time) *AddTodoBuilder

// Terminal method
func (b *AddTodoBuilder) Build() (string, error)
```

### UpdateTodoBuilder (for update command)

```go
type UpdateTodoBuilder struct {
    scheme *Scheme  // For execution
    token  string   // Set by AuthScheme
    id     string
    params map[string]string
    errors []error
}

// All TodoBuilder methods plus:
func (b *UpdateTodoBuilder) PrependNotes(notes string) *UpdateTodoBuilder
func (b *UpdateTodoBuilder) AppendNotes(notes string) *UpdateTodoBuilder
func (b *UpdateTodoBuilder) AddTags(tags ...string) *UpdateTodoBuilder
func (b *UpdateTodoBuilder) PrependChecklistItems(items ...string) *UpdateTodoBuilder
func (b *UpdateTodoBuilder) AppendChecklistItems(items ...string) *UpdateTodoBuilder
func (b *UpdateTodoBuilder) Duplicate(duplicate bool) *UpdateTodoBuilder

// Terminal methods
func (b *UpdateTodoBuilder) Build() (string, error)      // Returns URL string
func (b *UpdateTodoBuilder) Execute(ctx context.Context) error  // Builds and executes
```

### AddProjectBuilder (for add-project command)

```go
type AddProjectBuilder struct {
    params map[string]string
    errors []error
}

func (b *AddProjectBuilder) Title(title string) *AddProjectBuilder
func (b *AddProjectBuilder) Notes(notes string) *AddProjectBuilder
func (b *AddProjectBuilder) When(t time.Time) *AddProjectBuilder      // schedule for specific date
func (b *AddProjectBuilder) WhenEvening() *AddProjectBuilder          // this evening
func (b *AddProjectBuilder) WhenAnytime() *AddProjectBuilder          // anytime
func (b *AddProjectBuilder) WhenSomeday() *AddProjectBuilder          // someday
func (b *AddProjectBuilder) Deadline(t time.Time) *AddProjectBuilder  // deadline date
func (b *AddProjectBuilder) Reminder(hour, minute int) *AddProjectBuilder
func (b *AddProjectBuilder) Tags(tags ...string) *AddProjectBuilder
func (b *AddProjectBuilder) Area(name string) *AddProjectBuilder
func (b *AddProjectBuilder) AreaID(id string) *AddProjectBuilder
func (b *AddProjectBuilder) Todos(titles ...string) *AddProjectBuilder
func (b *AddProjectBuilder) Completed(completed bool) *AddProjectBuilder
func (b *AddProjectBuilder) Canceled(canceled bool) *AddProjectBuilder
func (b *AddProjectBuilder) Reveal(reveal bool) *AddProjectBuilder

func (b *AddProjectBuilder) Build() (string, error)
```

### UpdateProjectBuilder (for update-project command)

```go
type UpdateProjectBuilder struct {
    scheme *Scheme  // For execution
    token  string   // Set by AuthScheme
    id     string
    params map[string]string
    errors []error
}

// All ProjectBuilder methods (except Todos) plus:
func (b *UpdateProjectBuilder) PrependNotes(notes string) *UpdateProjectBuilder
func (b *UpdateProjectBuilder) AppendNotes(notes string) *UpdateProjectBuilder
func (b *UpdateProjectBuilder) AddTags(tags ...string) *UpdateProjectBuilder

// Terminal methods
func (b *UpdateProjectBuilder) Build() (string, error)      // Returns URL string
func (b *UpdateProjectBuilder) Execute(ctx context.Context) error  // Builds and executes
```

### ShowBuilder (for show command)

```go
type ShowBuilder struct {
    params map[string]string
}

func (b *ShowBuilder) ID(id string) *ShowBuilder
func (b *ShowBuilder) List(id ListID) *ShowBuilder
func (b *ShowBuilder) Query(query string) *ShowBuilder
func (b *ShowBuilder) Filter(tags ...string) *ShowBuilder

func (b *ShowBuilder) Build() string
```

### BatchBuilder (for json command - create only)

```go
// BatchBuilder is for batch create operations (no token required).
// For update operations, use AuthScheme.Batch() to get AuthBatchBuilder.
type BatchBuilder struct {
    items  []JSONItem
    reveal bool
    errors []error
}

// Only create operations available
func (b *BatchBuilder) AddTodo(opts ...JSONOption) *BatchBuilder
func (b *BatchBuilder) AddProject(opts ...JSONOption) *BatchBuilder
func (b *BatchBuilder) Reveal(reveal bool) *BatchBuilder

// Terminal method (no token needed for create-only operations)
func (b *BatchBuilder) Build() (string, error)
```

### AuthBatchBuilder (for json command - create and update)

```go
// AuthBatchBuilder is for batch operations including updates (token required).
// Obtained via AuthScheme.Batch().
type AuthBatchBuilder struct {
    token  string
    items  []JSONItem
    reveal bool
    errors []error
}

// Create operations
func (b *AuthBatchBuilder) AddTodo(opts ...JSONOption) *AuthBatchBuilder
func (b *AuthBatchBuilder) AddProject(opts ...JSONOption) *AuthBatchBuilder

// Update operations (only available on AuthBatchBuilder)
func (b *AuthBatchBuilder) UpdateTodo(id string, opts ...JSONOption) *AuthBatchBuilder
func (b *AuthBatchBuilder) UpdateProject(id string, opts ...JSONOption) *AuthBatchBuilder
func (b *AuthBatchBuilder) Reveal(reveal bool) *AuthBatchBuilder

// Terminal method (token already set via AuthScheme)
func (b *AuthBatchBuilder) Build() (string, error)
```

### JSON Shared Types

```go
type JSONItem struct {
    Type       string         `json:"type"`
    Operation  string         `json:"operation,omitempty"`
    ID         string         `json:"id,omitempty"`
    Attributes map[string]any `json:"attributes,omitempty"`
    Items      []JSONItem     `json:"items,omitempty"`
}

// JSON Options (used by both JSONBuilder and AuthJSONBuilder)
type JSONOption func(*JSONItem)

func JSONTitle(title string) JSONOption
func JSONNotes(notes string) JSONOption
func JSONWhen(when When) JSONOption
func JSONDeadline(date string) JSONOption
func JSONTags(tags ...string) JSONOption
func JSONCompleted(completed bool) JSONOption
func JSONItems(items ...JSONItem) JSONOption
```

## Error Definitions

```go
var (
    ErrNotesTooLong          = errors.New("things3: notes exceed 10,000 character limit")
    ErrTitleTooLong          = errors.New("things3: title exceeds 4,000 character limit")
    ErrTooManyChecklistItems = errors.New("things3: checklist exceeds 100 item limit")
    ErrIDRequired            = errors.New("things3: id required for update operation")
    ErrEmptyToken            = errors.New("things3: empty token provided to WithToken")
)
```

Note: `ErrTokenRequired` is no longer needed because the type system now enforces token requirements at compile time.

## Usage Examples

### Creating Todos

```go
scheme := things3.NewScheme()

// Simple todo
url, err := scheme.AddTodo().
    Title("Buy groceries").
    When(things3.Today()).
    Build()

// Complex todo
url, err := scheme.AddTodo().
    Title("Review PR #123").
    Notes("Check the authentication changes").
    When(things3.Tomorrow()).
    Deadline(time.Date(2024, 12, 15, 0, 0, 0, 0, time.Local)).
    Tags("work", "urgent").
    ChecklistItems("Check tests", "Review security", "Add comments").
    ListID("project-uuid").
    Reveal(true).
    Build()
```

### Updating Todos

```go
// First get token from database (see RFC 003)
db, _ := things3.NewDB()
token, _ := db.Token(ctx)

// Get authenticated scheme with token upfront
scheme := things3.NewScheme()
auth := scheme.WithToken(token)

// Build URL only
url, err := auth.UpdateTodo("task-uuid").
    Completed(true).
    Build()

// Or build and execute directly (runs in background)
err := auth.UpdateTodo("task-uuid").
    AppendNotes("\n\nUpdate: Discussed with team").
    AddTags("reviewed", "approved").
    Execute(ctx)
```

### Executing URL Scheme Operations

```go
scheme := things3.NewScheme()

// Navigation operations: foreground by default (user wants to view content)
err := scheme.Show(ctx, "task-uuid")      // Opens Things in foreground
err := scheme.Search(ctx, "meeting notes") // Opens Things with search results

// Run navigation in background (for programmatic use)
bgScheme := things3.NewScheme(things3.WithBackground())
err := bgScheme.Show(ctx, "task-uuid")  // Things stays in background

// Create/Update operations: background by default (silent operation)
err := scheme.AddTodo().Title("Buy milk").Execute(ctx)  // Creates without focus change

// Bring Things to foreground for create/update operations
fgScheme := things3.NewScheme(things3.WithForeground())
err := fgScheme.AddTodo().Title("Buy milk").Execute(ctx)  // Things comes to foreground

// Update with foreground execution
auth := fgScheme.WithToken(token)
err := auth.UpdateTodo("task-uuid").
    Completed(true).
    Execute(ctx)  // Brings Things to foreground
```

### Creating Projects

```go
scheme := things3.NewScheme()

url, err := scheme.AddProject().
    Title("Q1 Planning").
    Notes("Quarterly planning for 2024").
    WhenAnytime().
    Tags("planning", "2024").
    Todos("Define goals", "Create timeline", "Assign owners").
    AreaID("area-uuid").
    Reveal(true).
    Build()
```

### Showing Lists and Items

```go
scheme := things3.NewScheme()

// Show built-in list
url := scheme.Show().List(things3.ListToday).Build()
url := scheme.Show().List(things3.ListInbox).Build()

// Show specific item
url := scheme.Show().ID("item-uuid").Build()

// Show with filter
url := scheme.Show().
    Query("My Project").
    Filter("urgent", "high-priority").
    Build()
```

### Batch Operations

```go
scheme := things3.NewScheme()

// Create multiple items (no token needed) - use scheme.Batch()
url, err := scheme.Batch().
    AddTodo(func(t *things3.BatchTodoBuilder) {
        t.Title("First task").When(things3.Today())
    }).
    AddTodo(func(t *things3.BatchTodoBuilder) {
        t.Title("Second task").When(things3.Tomorrow())
    }).
    AddProject(func(p *things3.BatchProjectBuilder) {
        p.Title("New Project").Notes("Project description")
    }).
    Reveal(true).
    Build()

// Mixed create and update (token required) - use auth.Batch()
db, _ := things3.NewDB()
token, _ := db.Token(ctx)
auth := scheme.WithToken(token)

url, err := auth.Batch().
    AddTodo(func(t *things3.BatchTodoBuilder) {
        t.Title("New task")
    }).
    UpdateTodo("existing-uuid", func(t *things3.BatchTodoBuilder) {
        t.Completed(true)
    }).
    Build()
```

### Simple URLs

```go
scheme := things3.NewScheme()

url := scheme.Search("meeting notes")
url := scheme.Version()
```

### Integration with go-things3-mcp

```go
// In go-things3-mcp

// 1. Get token from database
db, _ := things3.NewDB()
token, _ := db.Token(ctx)

// 2. Build URL using AuthScheme (token required upfront)
scheme := things3.NewScheme()
auth := scheme.WithToken(token)

url, err := auth.UpdateTodo(params.ID).
    Completed(params.Completed).
    Build()

// 3. Execute via AppleScript (mack library)
return mack.Tell("Things3", fmt.Sprintf(`open location "%s"`, url))
```

## File Organization

```text
things3/
├── scheme.go           # Scheme, AuthScheme types, NewScheme(), WithToken(), execution
├── scheme_options.go   # SchemeOption, WithForeground()
├── scheme_builder.go   # TodoBuilder, ProjectBuilder
├── scheme_update.go    # UpdateTodoBuilder, UpdateProjectBuilder with Execute()
├── scheme_show.go      # ShowBuilder
├── scheme_json.go      # JSONBuilder, AuthJSONBuilder, JSONItem, JSONOption
├── scheme_constants.go # Internal parameter keys
└── scheme_test.go      # Unit tests
```

| File | Responsibility |
|------|----------------|
| `scheme.go` | Entry points, NewScheme(), WithToken(), Show(), Search(), execute(), executeNavigation() |
| `scheme_options.go` | SchemeOption, WithForeground(), WithBackground() |
| `scheme_builder.go` | AddTodoBuilder, AddProjectBuilder for create operations |
| `scheme_update.go` | UpdateTodoBuilder, UpdateProjectBuilder with Execute() |
| `scheme_show.go` | ShowBuilder for navigation |
| `scheme_json.go` | Batch command builders (BatchBuilder, AuthBatchBuilder) and options |
| `scheme_constants.go` | URL parameter key constants |
| `scheme_test.go` | Unit tests for all builders |

## Data Type Reference

From Things 3 official documentation:

| Type | Format | Example |
|------|--------|---------|
| string | Percent-encoded, max 4000 chars | `Buy%20milk` |
| notes | Percent-encoded, max 10000 chars | Long text |
| date | `yyyy-mm-dd` or keyword | `2024-12-25`, `today` |
| time | `HH:MM` or `H:MMAM/PM` | `21:30`, `9:30PM` |
| datetime | `date@time` | `2024-12-25@14:00` |
| ISO8601 | RFC 3339 | `2024-12-25T14:30:00Z` |
| boolean | `true` or `false` | Case-sensitive |

### Rate Limits

- Maximum 250 items per 10 seconds for `add` command
- Library documents this but does not enforce

```go
// RateLimitItems is the maximum number of items Things accepts per 10 seconds.
const RateLimitItems = 250

// RateLimitWindow is the rate limit time window.
const RateLimitWindow = 10 * time.Second
```

## Testing Strategy

### Scheme Builder Tests (No Auth Required)

```go
func TestAddTodoBuilder(t *testing.T) {
    scheme := things3.NewScheme()

    tests := []struct {
        name     string
        build    func() (string, error)
        contains []string
        wantErr  error
    }{
        {
            name: "simple todo",
            build: func() (string, error) {
                return scheme.AddTodo().
                    Title("Test").
                    When(things3.Today()).
                    Build()
            },
            contains: []string{"things:///add?", "title=Test", "when="},
        },
        {
            name: "todo with all options",
            build: func() (string, error) {
                return scheme.AddTodo().
                    Title("Review PR").
                    Notes("Check changes").
                    Deadline(time.Date(2024, 12, 15, 0, 0, 0, 0, time.Local)).
                    Tags("work", "urgent").
                    ChecklistItems("Item 1", "Item 2").
                    Build()
            },
            contains: []string{
                "title=Review",
                "notes=Check",
                "deadline=2024-12-15",
                "tags=work%2Curgent",
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            url, err := tt.build()
            if tt.wantErr != nil {
                require.ErrorIs(t, err, tt.wantErr)
                return
            }
            require.NoError(t, err)
            for _, s := range tt.contains {
                assert.Contains(t, url, s)
            }
        })
    }
}
```

### AuthScheme Tests (Token Required)

```go
func TestAuthScheme_UpdateTodo(t *testing.T) {
    scheme := things3.NewScheme()
    token := "test-token"
    auth := scheme.WithToken(token)

    url, err := auth.UpdateTodo("uuid-123").
        Completed(true).
        Build()

    require.NoError(t, err)
    assert.Contains(t, url, "things:///update?")
    assert.Contains(t, url, "id=uuid-123")
    assert.Contains(t, url, "auth-token=test-token")
    assert.Contains(t, url, "completed=true")
}

func TestAuthScheme_Batch(t *testing.T) {
    scheme := things3.NewScheme()
    auth := scheme.WithToken("test-token")

    // AuthBatchBuilder has UpdateTodo method
    url, err := auth.Batch().
        AddTodo(things3.JSONTitle("New task")).
        UpdateTodo("uuid-123", things3.JSONCompleted(true)).
        Build()

    require.NoError(t, err)
    assert.Contains(t, url, "things:///json?")
    assert.Contains(t, url, "auth-token=test-token")
}

func TestScheme_Batch_NoUpdateMethods(t *testing.T) {
    scheme := things3.NewScheme()

    // BatchBuilder (non-auth) only has AddTodo, AddProject - no UpdateTodo
    url, err := scheme.Batch().
        AddTodo(things3.JSONTitle("Task 1")).
        AddProject(things3.JSONTitle("Project 1")).
        Build()

    require.NoError(t, err)
    assert.Contains(t, url, "things:///json?")
    assert.NotContains(t, url, "auth-token")
}

func TestWithToken_EmptyToken(t *testing.T) {
    scheme := things3.NewScheme()

    // WithToken should validate token is not empty
    auth := scheme.WithToken("")
    _, err := auth.UpdateTodo("uuid").Build()

    require.ErrorIs(t, err, things3.ErrEmptyToken)
}
```

## Backward Compatibility

### Deprecated Methods (to be removed in v2.0)

The following methods on `Client` are deprecated in favor of `NewScheme()`:

```go
// Deprecated: Use things3.NewScheme().AddTodo().Build() instead
func (c *Client) AddTodoURL(params map[string]string) string

// Deprecated: Use things3.NewScheme().WithToken(token).UpdateTodo(id).Build()
func (c *Client) URL(ctx context.Context, cmd URLCommand, params map[string]string) (string, error)
```

### Migration Examples

```go
// Before (deprecated)
client, _ := things3.New()
url := client.AddTodoURL(map[string]string{"title": "Task"})

// After (recommended)
scheme := things3.NewScheme()
url, _ := scheme.AddTodo().Title("Task").Build()
```

```go
// Before (deprecated)
client, _ := things3.New()
url, _ := client.URL(ctx, things3.URLCommandUpdate, map[string]string{
    "id": "uuid",
    "completed": "true",
})

// After (recommended) - token required upfront via WithToken
db, _ := things3.NewDB()
token, _ := db.Token(ctx)

scheme := things3.NewScheme()
auth := scheme.WithToken(token)
url, _ := auth.UpdateTodo("uuid").Completed(true).Build()
```

## Design Principles

| Principle | Implementation |
|-----------|----------------|
| Separation of Concerns | URL building independent of database |
| Explicit Dependencies | Token required upfront via `WithToken()` |
| Compile-time Safety | `*AuthScheme` type only exposes update methods |
| Functional Options | `SchemeOption` for configurable behavior |
| Intent-Based Defaults | Navigation ops foreground (view), Create/Update ops background (silent) |
| Delegation Pattern | `AuthScheme` references `Scheme` for shared config |
| Builder Pattern | Chainable methods with terminal `.Build()` or `.Execute()` |
| Type Safety | Enums for `When`, `ListID`, `Command` |

## References

- RFC 003: Database API - Provides `Token()` for authenticated operations
- [Things URL Scheme Documentation](https://culturedcode.com/things/support/articles/2803573/)
- [Go net/url Package](https://pkg.go.dev/net/url)
- [RFC 3986 - URI Generic Syntax](https://datatracker.ietf.org/doc/html/rfc3986/)
- [go-resty/resty](https://github.com/go-resty/resty) - API design inspiration
- [google/go-github](https://github.com/google/go-github) - Service grouping pattern

---

## Appendix A: Things 3 Official URL Scheme Reference

> **Source:** [https://culturedcode.com/things/support/articles/2803573/](https://culturedcode.com/things/support/articles/2803573/)
>
> **Last Updated:** 2024-12 (Document version for this RFC)
>
> This appendix contains the complete official documentation from Cultured Code. If the official documentation changes in the future, this section should be updated accordingly.

### A.1 Overview

Things uses custom URLs to enable external apps and scripts to interact with the task manager. The basic URL format is:

```
things:///commandName?parameter1=value1&parameter2=value2
```

The scheme supports x-callback-url conventions for `x-success`, `x-error`, and `x-cancel` callbacks.

### A.2 Commands

#### A.2.1 add - Create To-Dos

Creates individual tasks or multiple tasks at once.

| Parameter | Type | Description |
|-----------|------|-------------|
| `title` | string | Task name (max 4,000 chars) |
| `titles` | string | Multiple tasks, newline-separated |
| `notes` | string | Task description (max 10,000 chars) |
| `when` | date/time | Schedule: `today`, `tomorrow`, `evening`, `anytime`, `someday`, or `yyyy-mm-dd` |
| `deadline` | date string | Due date in `yyyy-mm-dd` format |
| `tags` | string | Comma-separated tags (must exist in Things) |
| `checklist-items` | string | Newline-separated sub-tasks (max 100 items) |
| `list` | string | Target project or area name |
| `list-id` | string | Target project or area UUID (overrides `list`) |
| `heading` | string | Section title within a project |
| `heading-id` | string | Section UUID (overrides `heading`) |
| `completed` | boolean | Task completion status |
| `canceled` | boolean | Task canceled status (overrides `completed`) |
| `show-quick-entry` | boolean | Display quick entry dialog instead of adding directly |
| `reveal` | boolean | Navigate to newly created task |
| `creation-date` | ISO8601 | Creation timestamp (future dates ignored) |
| `completion-date` | ISO8601 | Completion timestamp (future dates ignored) |

**Rate Limit:** Maximum 250 items per 10 seconds.

**Return Parameter:** `x-things-id` - Comma-separated IDs of created tasks.

#### A.2.2 add-project - Create Projects

Creates new projects with optional nested to-dos.

| Parameter | Type | Description |
|-----------|------|-------------|
| `title` | string | Project name |
| `notes` | string | Project description (max 10,000 chars) |
| `when` | date/time | Start date |
| `deadline` | date string | Due date |
| `tags` | string | Comma-separated tags |
| `area` | string | Parent area name |
| `area-id` | string | Parent area UUID (overrides `area`) |
| `to-dos` | string | Newline-separated child task titles |
| `completed` | boolean | Project completion status |
| `canceled` | boolean | Project canceled status |
| `reveal` | boolean | Navigate to new project |
| `creation-date` | ISO8601 | Creation timestamp |
| `completion-date` | ISO8601 | Completion timestamp |

**Return Parameter:** `x-things-id` - Project ID.

#### A.2.3 update - Modify Existing To-Dos

**Requires Authentication:** `auth-token` parameter is mandatory.

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string | **Required** - Target task UUID |
| `auth-token` | string | **Required** - Authorization token |
| `title` | string | Replace task name |
| `notes` | string | Replace notes |
| `prepend-notes` | string | Prepend to existing notes |
| `append-notes` | string | Append to existing notes |
| `when` | date/time | Reschedule task |
| `deadline` | date string | Set due date (empty value clears) |
| `tags` | string | Replace all tags |
| `add-tags` | string | Add tags without replacing |
| `checklist-items` | string | Replace checklist |
| `prepend-checklist-items` | string | Prepend checklist items |
| `append-checklist-items` | string | Append checklist items |
| `list` | string | Move to project/area by name |
| `list-id` | string | Move to project/area by UUID |
| `heading` | string | Move to section by name |
| `heading-id` | string | Move to section by UUID |
| `completed` | boolean | Completion status |
| `canceled` | boolean | Canceled status |
| `duplicate` | boolean | Duplicate before updating |
| `reveal` | boolean | Navigate to item |
| `creation-date` | ISO8601 | Creation timestamp |
| `completion-date` | ISO8601 | Completion timestamp |

**Clearing Values:** Include parameter with `=` but no value (e.g., `&deadline=`).

**Repeating Task Restrictions:** Cannot update `when`, `deadline`, `completed`, `canceled`, or `completion-date`; cannot duplicate.

**Return Parameter:** `x-things-id`

#### A.2.4 update-project - Modify Existing Projects

**Requires Authentication:** `auth-token` parameter is mandatory.

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string | **Required** - Project UUID |
| `auth-token` | string | **Required** - Authorization token |
| `title` | string | Replace project name |
| `notes` | string | Replace notes |
| `prepend-notes` | string | Prepend to notes |
| `append-notes` | string | Append to notes |
| `when` | date/time | Reschedule |
| `deadline` | date string | Set due date |
| `tags` | string | Replace all tags |
| `add-tags` | string | Add tags |
| `area` | string | Move to area by name |
| `area-id` | string | Move to area by UUID |
| `completed` | boolean | Completion status |
| `canceled` | boolean | Canceled status |
| `reveal` | boolean | Navigate to project |

**Project Completion Constraint:** Setting `completed=true` is ignored unless all child to-dos are completed or canceled and all child headings are archived.

**Repeating Project Restrictions:** Same as repeating tasks.

#### A.2.5 show - Navigate and Display

Opens specific lists, projects, areas, or to-dos.

| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | string | Target UUID or built-in list ID |
| `query` | string | Search for area/project/tag by name |
| `filter` | string | Comma-separated tag filter |

**Built-in List IDs:**

| ID | Description |
|----|-------------|
| `inbox` | Inbox |
| `today` | Today |
| `anytime` | Anytime |
| `upcoming` | Upcoming |
| `someday` | Someday |
| `logbook` | Logbook |
| `tomorrow` | Tomorrow |
| `deadlines` | Deadlines |
| `repeating` | Repeating |
| `all-projects` | All Projects |
| `logged-projects` | Logged Projects |

**Note:** Tasks cannot be shown using the `query` parameter; use `id` parameter or `search` command instead.

#### A.2.6 search - Invoke Search Interface

Displays search screen with optional query.

| Parameter | Type | Description |
|-----------|------|-------------|
| `query` | string | Search term (optional) |

#### A.2.7 version - Report Version Information

No parameters required.

**Return Parameters:**
- `x-things-scheme-version` - URL Scheme version
- `x-things-client-version` - Things app version

#### A.2.8 json - Advanced Data Import

Enables sophisticated multi-item creation and updates via JSON structures.

| Parameter | Type | Description |
|-----------|------|-------------|
| `data` | JSON string | URL-encoded JSON array |
| `auth-token` | string | Required for update operations |
| `reveal` | boolean | Navigate to first created item |

**Return Parameter:** `x-things-ids` - JSON array of created IDs.

### A.3 JSON Command Structure

#### A.3.1 Basic Structure

```json
[
  {
    "type": "to-do" | "project" | "heading" | "checklist-item",
    "operation": "create" | "update",
    "id": "UUID",
    "attributes": { /* type-specific properties */ }
  }
]
```

#### A.3.2 To-Do Attributes

| Attribute | Type | Create | Update |
|-----------|------|--------|--------|
| `title` | string | Yes | Yes |
| `notes` | string | Yes | Yes |
| `when` | date | Yes | Yes |
| `deadline` | date | Yes | Yes |
| `tags` | string[] | Yes | Yes |
| `checklist-items` | object[] | Yes | Yes |
| `list` / `list-id` | string | Yes | Yes |
| `heading` / `heading-id` | string | Yes | Yes |
| `completed` | boolean | Yes | Yes |
| `canceled` | boolean | Yes | Yes |
| `creation-date` | ISO8601 | Yes | Yes |
| `completion-date` | ISO8601 | Yes | Yes |
| `prepend-notes` | string | No | Yes |
| `append-notes` | string | No | Yes |
| `add-tags` | string[] | No | Yes |
| `prepend-checklist-items` | object[] | No | Yes |
| `append-checklist-items` | object[] | No | Yes |

#### A.3.3 Project Attributes

| Attribute | Type | Create | Update |
|-----------|------|--------|--------|
| `title` | string | Yes | Yes |
| `notes` | string | Yes | Yes |
| `when` | date | Yes | Yes |
| `deadline` | date | Yes | Yes |
| `tags` | string[] | Yes | Yes |
| `area` / `area-id` | string | Yes | Yes |
| `items` | object[] | Yes | No |
| `completed` | boolean | Yes | Yes |
| `canceled` | boolean | Yes | Yes |
| `creation-date` | ISO8601 | Yes | Yes |
| `completion-date` | ISO8601 | Yes | Yes |
| `prepend-notes` | string | No | Yes |
| `append-notes` | string | No | Yes |
| `add-tags` | string[] | No | Yes |

#### A.3.4 Heading Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `title` | string | Heading name |
| `archived` | boolean | Archive status (ignored unless all sub-tasks completed/canceled) |

#### A.3.5 Checklist Item Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `title` | string | Item text |
| `completed` | boolean | Completion status |
| `canceled` | boolean | Canceled status |

### A.4 Data Types

| Type | Format | Examples | Notes |
|------|--------|----------|-------|
| string | Percent-encoded | `Buy%20milk` | Max 4,000 chars (10,000 for notes) |
| date string | `yyyy-mm-dd` | `2024-12-25` | Or keywords: `today`, `tomorrow` |
| date string | Natural language | `in 3 days`, `next tuesday` | English only |
| time string | `HH:MM` (24-hour) | `14:30` | Local timezone |
| time string | `H:MMAM/PM` | `2:30PM` | 12-hour format |
| date time string | `yyyy-mm-dd@HH:MM` | `2024-12-25@14:30` | Date and time with `@` |
| ISO8601 | RFC 3339 | `2024-12-25T14:30:00Z` | For creation/completion dates |
| boolean | `true` / `false` | `true` | Case-sensitive |
| JSON string | Valid JSON | `[{"type":"to-do"}]` | Must be URL-encoded |

### A.5 Authentication

**Retrieving the Auth Token:**
- **Mac:** Things -> Settings -> General -> Enable Things URLs -> Manage
- **iOS:** Settings -> General -> Things URLs

**Commands Requiring Authentication:**
- `update`
- `update-project`
- `json` (when containing update operations)

### A.6 Getting Item IDs

**For To-Dos:**
- Mac: Control-click -> Share -> Copy Link
- iOS: Open task -> toolbar -> Share -> Copy Link

**For Lists/Projects/Areas:**
- Mac: Control-click in sidebar -> Share -> Copy Link
- iOS: Navigate to list -> top-right -> Share -> Copy Link

### A.7 Constraints and Limitations

| Constraint | Value | Notes |
|------------|-------|-------|
| String max length | 4,000 chars | Except notes |
| Notes max length | 10,000 chars | |
| Checklist items | 100 max | Per to-do |
| Add rate limit | 250 items | Per 10 seconds |
| Future dates | Rejected | For creation-date, completion-date |
| Repeating tasks | Limited | Cannot update when/deadline/completion |
| Project completion | Conditional | Requires all children completed |

### A.8 URL Encoding Requirements

All parameter values must be percent-encoded:

| Character | Encoding |
|-----------|----------|
| Space | `%20` |
| Newline | `%0a` |
| Ampersand | `%26` |
| Equals | `%3D` |
| Comma | `%2C` |

For JSON command, remove whitespace then URL-encode the entire JSON string.

### A.9 x-callback-url Support

All commands support x-callback-url convention:

| Callback | Description |
|----------|-------------|
| `x-success` | Called on successful completion with return parameters |
| `x-error` | Called on error with `errorMessage` parameter |
| `x-cancel` | Called if user cancels operation |

**Common Return Parameters:**
- `x-things-id` - Single ID or comma-separated IDs
- `x-things-ids` - JSON array of IDs (json command only)
- `x-things-scheme-version` - URL Scheme version (version command)
- `x-things-client-version` - App version (version command)
