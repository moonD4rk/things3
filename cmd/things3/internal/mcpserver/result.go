// Package mcpserver implements the things3 MCP server: verb-shaped tools that
// mirror the CLI's commands over the Model Context Protocol stdio transport.
// Reads run on the library's typed query builders and composed views; writes
// run resolve -> execute -> verify on the same internal/resolve and
// internal/verify infrastructure the CLI uses.
//
// The package is deliberately cobra-free so its handlers can be unit-tested
// directly against the thingstest fixture and over an in-memory MCP transport.
package mcpserver

// Domain error codes carried by ToolError. They are the machine-parseable
// vocabulary a model self-corrects against; transport failures (database I/O,
// canceled context) are surfaced as Go errors instead, never these codes.
const (
	codeInvalidInput    = "invalid_input"
	codeNotFound        = "not_found"
	codeAmbiguous       = "ambiguous"
	codeExecutionFailed = "execution_failed"
)

// ToolError is the structured failure every tool reports inside its normal
// output envelope, alongside success=false. Because it is part of the declared
// output schema, each failure is machine-parseable and self-correcting.
type ToolError struct {
	Code       string      `json:"code" jsonschema:"one of invalid_input, not_found, ambiguous, execution_failed"`
	Message    string      `json:"message" jsonschema:"human-readable explanation of the failure"`
	Hint       string      `json:"hint,omitempty" jsonschema:"an optional suggestion for how to retry successfully"`
	Candidates []Candidate `json:"candidates,omitempty" jsonschema:"for an ambiguous target, the matching items to disambiguate between"`
}

// Candidate is one row of an ambiguity report: a full UUID, its type, and title.
type Candidate struct {
	UUID  string `json:"uuid" jsonschema:"the full UUID to retry with"`
	Type  string `json:"type" jsonschema:"todo or project"`
	Title string `json:"title" jsonschema:"the item title"`
}

// Ref is an inline reference to a related item (project, area, or heading).
type Ref struct {
	UUID  string `json:"uuid"`
	Title string `json:"title"`
}

// ChecklistItem is a sub-item of a todo.
type ChecklistItem struct {
	UUID   string `json:"uuid"`
	Title  string `json:"title"`
	Status string `json:"status" jsonschema:"incomplete, completed, or canceled"`
}

// Item is the unified, type-discriminated shape for a todo or a project. Dates
// are YYYY-MM-DD strings, the reminder is HH:MM, and related items appear inline
// as Refs. It never nests another Item, keeping the output schema acyclic; a
// project's todos and headings ride on the enclosing GetResult instead.
type Item struct {
	Type   string `json:"type" jsonschema:"todo or project"`
	UUID   string `json:"uuid"`
	Title  string `json:"title"`
	Status string `json:"status" jsonschema:"incomplete, completed, or canceled"`
	Start  string `json:"start" jsonschema:"the start bucket: inbox, anytime, or someday"`
	Notes  string `json:"notes,omitempty"`
	// NotesTruncated marks a list or search item whose notes were shortened; the
	// full text is available through get. Detail views never truncate.
	NotesTruncated bool            `json:"notes_truncated,omitempty" jsonschema:"true when notes were shortened; use get for full text"`
	Project        *Ref            `json:"project,omitempty"`
	Area           *Ref            `json:"area,omitempty"`
	Heading        *Ref            `json:"heading,omitempty"`
	Tags           []string        `json:"tags,omitempty"`
	When           string          `json:"when,omitempty" jsonschema:"the scheduled start date, YYYY-MM-DD"`
	Deadline       string          `json:"deadline,omitempty" jsonschema:"the deadline date, YYYY-MM-DD"`
	Reminder       string          `json:"reminder,omitempty" jsonschema:"the reminder time, HH:MM"`
	Checklist      []ChecklistItem `json:"checklist,omitempty"`
	Evening        bool            `json:"evening,omitempty" jsonschema:"true when the todo is in the This Evening section of Today"`
	Repeating      bool            `json:"repeating,omitempty" jsonschema:"true when the item belongs to a repeating series"`
	CompletedAt    string          `json:"completed_at,omitempty" jsonschema:"when the item was completed or canceled, RFC 3339"`
}

// Area is the small shape for a Things area.
type Area struct {
	UUID  string   `json:"uuid"`
	Title string   `json:"title"`
	Tags  []string `json:"tags,omitempty"`
}

// Tag is the small shape for a Things tag.
type Tag struct {
	UUID     string `json:"uuid"`
	Title    string `json:"title"`
	Shortcut string `json:"shortcut,omitempty"`
}

// PageResult is the self-describing pagination envelope returned by every list
// and search tool. Items always encodes as a JSON array, never null, and
// total/page/pages describe the slice within the full result set.
type PageResult[T any] struct {
	Success bool       `json:"success"`
	Error   *ToolError `json:"error,omitempty"`
	Items   []T        `json:"items"`
	Total   int        `json:"total" jsonschema:"the total number of items before pagination"`
	Page    int        `json:"page" jsonschema:"the 1-based page number returned"`
	Pages   int        `json:"pages" jsonschema:"the total number of pages available"`
}

// GetResult is the output of the get tool. Item is the resolved todo or project;
// for a project, Todos carries its incomplete todos and Headings its headings.
type GetResult struct {
	Success  bool       `json:"success"`
	Error    *ToolError `json:"error,omitempty"`
	Item     *Item      `json:"item,omitempty"`
	Todos    []Item     `json:"todos,omitempty" jsonschema:"a project's incomplete todos"`
	Headings []Ref      `json:"headings,omitempty" jsonschema:"a project's headings"`
}

// WriteResult is the output of every write and navigation tool. Verified reports
// whether the fire-and-forget write was confirmed in the database; an
// unconfirmed send is still success=true, matching the CLI's exit-0 semantics.
type WriteResult struct {
	Success  bool       `json:"success"`
	Error    *ToolError `json:"error,omitempty"`
	Verified bool       `json:"verified" jsonschema:"whether the write was confirmed in the database"`
	Message  string     `json:"message,omitempty"`
	Item     *Item      `json:"item,omitempty" jsonschema:"the item as it stands after a confirmed write"`
}

// notFound builds a not_found ToolError for a query that matched no item.
func notFound(query string) *ToolError {
	return &ToolError{Code: codeNotFound, Message: "no item matches " + quote(query)}
}

// invalidInput builds an invalid_input ToolError from a parse failure message.
func invalidInput(msg string) *ToolError {
	return &ToolError{Code: codeInvalidInput, Message: msg}
}

// executionFailed builds an execution_failed ToolError for a URL-scheme send
// that could not run (for example, Things is not installed).
func executionFailed(err error) *ToolError {
	return &ToolError{
		Code:    codeExecutionFailed,
		Message: err.Error(),
		Hint:    "write tools require macOS with Things 3 installed and running",
	}
}

// quote wraps s in double quotes for interpolation into a message.
func quote(s string) string { return "\"" + s + "\"" }
