package scheme

// URL parameter keys for Things URL Scheme.
// These constants correspond to the official Things 3 URL Scheme documentation:
// https://culturedcode.com/things/support/articles/2803573/

// Common parameters (used across multiple commands).
const (
	// KeyTitle is the task or project name (max 4,000 chars).
	KeyTitle = "title"
	// KeyNotes is the description text (max 10,000 chars).
	KeyNotes = "notes"
	// KeyWhen is the scheduling date (today, tomorrow, evening, anytime, someday, or yyyy-mm-dd).
	KeyWhen = "when"
	// KeyDeadline is the due date in yyyy-mm-dd format.
	KeyDeadline = "deadline"
	// KeyTags is a comma-separated list of tags (must exist in Things).
	KeyTags = "tags"
	// KeyCompleted indicates completion status.
	KeyCompleted = "completed"
	// KeyCanceled indicates canceled status (overrides completed).
	KeyCanceled = "canceled"
	// KeyReveal navigates to the item after operation.
	KeyReveal = "reveal"
	// KeyCreationDate is the creation timestamp in ISO8601 format.
	KeyCreationDate = "creation-date"
	// KeyCompletionDate is the completion timestamp in ISO8601 format.
	KeyCompletionDate = "completion-date"
)

// Todo specific parameters.
const (
	// KeyTitles creates multiple tasks from newline-separated titles.
	KeyTitles = "titles"
	// KeyChecklistItems contains newline-separated sub-tasks (max 100 items).
	KeyChecklistItems = "checklist-items"
	// KeyList is the target project or area name.
	KeyList = "list"
	// KeyListID is the target project or area UUID (overrides list).
	KeyListID = "list-id"
	// KeyHeading is the section title within a project.
	KeyHeading = "heading"
	// KeyHeadingID is the section UUID (overrides heading).
	KeyHeadingID = "heading-id"
	// KeyShowQuickEntry displays quick entry dialog instead of adding directly.
	KeyShowQuickEntry = "show-quick-entry"
)

// Project specific parameters.
const (
	// KeyArea is the parent area name.
	KeyArea = "area"
	// KeyAreaID is the parent area UUID (overrides area).
	KeyAreaID = "area-id"
	// KeyTodos contains newline-separated child task titles.
	KeyTodos = "to-dos"
)

// Update-only parameters (require auth-token).
const (
	// KeyID is the target item UUID (required for update operations).
	KeyID = "id"
	// KeyAuthToken is the authorization token for update operations.
	KeyAuthToken = "auth-token"
	// KeyPrependNotes prepends text to existing notes.
	KeyPrependNotes = "prepend-notes"
	// KeyAppendNotes appends text to existing notes.
	KeyAppendNotes = "append-notes"
	// KeyAddTags adds tags without replacing existing ones.
	KeyAddTags = "add-tags"
	// KeyPrependChecklistItems prepends items to the checklist.
	KeyPrependChecklistItems = "prepend-checklist-items"
	// KeyAppendChecklistItems appends items to the checklist.
	KeyAppendChecklistItems = "append-checklist-items"
	// KeyDuplicate duplicates the item before updating.
	KeyDuplicate = "duplicate"
)

// Show command parameters.
const (
	// KeyQuery searches for area/project/tag by name.
	KeyQuery = "query"
	// KeyFilter applies comma-separated tag filter.
	KeyFilter = "filter"
)

// JSON command parameters.
const (
	// KeyData contains the URL-encoded JSON array.
	KeyData = "data"
)

// URL Scheme limits.
const (
	// MaxTitleLength is the maximum allowed length for titles in Things URL scheme.
	MaxTitleLength = 4000
	// MaxNotesLength is the maximum allowed length for notes in Things URL scheme.
	MaxNotesLength = 10000
	// MaxChecklistItems is the maximum allowed number of checklist items.
	MaxChecklistItems = 100
)

// When represents scheduling values for the "when" parameter in URL scheme.
// Use When(time.Time) for dates or WhenEvening, WhenAnytime, WhenSomeday
// for Things 3-specific concepts.
type When string

const (
	// WhenEvening schedules for this evening.
	WhenEvening When = "evening"
	// WhenAnytime schedules for anytime.
	WhenAnytime When = "anytime"
	// WhenSomeday schedules for someday.
	WhenSomeday When = "someday"
)
