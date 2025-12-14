package things3

// URL parameter keys for Things URL Scheme.
// These constants correspond to the official Things 3 URL Scheme documentation:
// https://culturedcode.com/things/support/articles/2803573/
//
// Using constants ensures type safety and makes refactoring easier.
// They also serve as documentation for the supported parameters.

// Common parameters (used across multiple commands).
const (
	// keyTitle is the task or project name (max 4,000 chars).
	keyTitle = "title"
	// keyNotes is the description text (max 10,000 chars).
	keyNotes = "notes"
	// keyWhen is the scheduling date (today, tomorrow, evening, anytime, someday, or yyyy-mm-dd).
	keyWhen = "when"
	// keyDeadline is the due date in yyyy-mm-dd format.
	keyDeadline = "deadline"
	// keyTags is a comma-separated list of tags (must exist in Things).
	keyTags = "tags"
	// keyCompleted indicates completion status.
	keyCompleted = "completed"
	// keyCanceled indicates canceled status (overrides completed).
	keyCanceled = "canceled"
	// keyReveal navigates to the item after operation.
	keyReveal = "reveal"
	// keyCreationDate is the creation timestamp in ISO8601 format.
	keyCreationDate = "creation-date"
	// keyCompletionDate is the completion timestamp in ISO8601 format.
	keyCompletionDate = "completion-date"
)

// To-do specific parameters.
const (
	// keyTitles creates multiple tasks from newline-separated titles.
	keyTitles = "titles"
	// keyChecklistItems contains newline-separated sub-tasks (max 100 items).
	keyChecklistItems = "checklist-items"
	// keyList is the target project or area name.
	keyList = "list"
	// keyListID is the target project or area UUID (overrides list).
	keyListID = "list-id"
	// keyHeading is the section title within a project.
	keyHeading = "heading"
	// keyHeadingID is the section UUID (overrides heading).
	keyHeadingID = "heading-id"
	// keyShowQuickEntry displays quick entry dialog instead of adding directly.
	keyShowQuickEntry = "show-quick-entry"
)

// Project specific parameters.
const (
	// keyArea is the parent area name.
	keyArea = "area"
	// keyAreaID is the parent area UUID (overrides area).
	keyAreaID = "area-id"
	// keyTodos contains newline-separated child task titles.
	keyTodos = "to-dos"
)

// Update-only parameters (require auth-token).
const (
	// keyID is the target item UUID (required for update operations).
	keyID = "id"
	// keyAuthToken is the authorization token for update operations.
	keyAuthToken = "auth-token"
	// keyPrependNotes prepends text to existing notes.
	keyPrependNotes = "prepend-notes"
	// keyAppendNotes appends text to existing notes.
	keyAppendNotes = "append-notes"
	// keyAddTags adds tags without replacing existing ones.
	keyAddTags = "add-tags"
	// keyPrependChecklistItems prepends items to the checklist.
	keyPrependChecklistItems = "prepend-checklist-items"
	// keyAppendChecklistItems appends items to the checklist.
	keyAppendChecklistItems = "append-checklist-items"
	// keyDuplicate duplicates the item before updating.
	keyDuplicate = "duplicate"
)

// Show command parameters.
const (
	// keyQuery searches for area/project/tag by name.
	keyQuery = "query"
	// keyFilter applies comma-separated tag filter.
	keyFilter = "filter"
)

// JSON command parameters.
const (
	// keyData contains the URL-encoded JSON array.
	keyData = "data"
)
