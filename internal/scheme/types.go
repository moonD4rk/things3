package scheme

// Command represents Things URL scheme commands.
type Command string

const (
	// CommandShow opens and shows an item.
	CommandShow Command = "show"
	// CommandAdd creates a new todo.
	CommandAdd Command = "add"
	// CommandAddProject creates a new project.
	CommandAddProject Command = "add-project"
	// CommandUpdate updates an existing item (requires auth token).
	CommandUpdate Command = "update"
	// CommandUpdateProject updates an existing project (requires auth token).
	CommandUpdateProject Command = "update-project"
	// CommandSearch performs a search.
	CommandSearch Command = "search"
	// CommandVersion returns Things version information.
	CommandVersion Command = "version"
	// CommandJSON enables advanced JSON-based operations.
	CommandJSON Command = "json"
)

// String returns the string representation of the Command.
func (c Command) String() string {
	return string(c)
}

// ListID represents built-in Things list identifiers for the show command.
type ListID string

const (
	// ListInbox is the Inbox list.
	ListInbox ListID = "inbox"
	// ListToday is the Today list.
	ListToday ListID = "today"
	// ListAnytime is the Anytime list.
	ListAnytime ListID = "anytime"
	// ListUpcoming is the Upcoming list.
	ListUpcoming ListID = "upcoming"
	// ListSomeday is the Someday list.
	ListSomeday ListID = "someday"
	// ListLogbook is the Logbook list.
	ListLogbook ListID = "logbook"
	// ListTomorrow is the Tomorrow list.
	ListTomorrow ListID = "tomorrow"
	// ListDeadlines is the Deadlines list.
	ListDeadlines ListID = "deadlines"
	// ListRepeating is the Repeating list.
	ListRepeating ListID = "repeating"
	// ListAllProjects is the All Projects list.
	ListAllProjects ListID = "all-projects"
	// ListLoggedProjects is the Logged Projects list.
	ListLoggedProjects ListID = "logged-projects"
)

// String returns the string representation of the ListID.
func (l ListID) String() string {
	return string(l)
}

// JSONOperation represents the operation type for a JSON item.
type JSONOperation string

const (
	// JSONOperationCreate creates a new item.
	JSONOperationCreate JSONOperation = "create"
	// JSONOperationUpdate updates an existing item.
	JSONOperationUpdate JSONOperation = "update"
)

// JSONItemType represents the type of item in a JSON operation.
type JSONItemType string

const (
	// JSONItemTypeTodo represents a todo item.
	JSONItemTypeTodo JSONItemType = "to-do"
	// JSONItemTypeProject represents a project item.
	JSONItemTypeProject JSONItemType = "project"
)

// JSONItem represents a single item in a JSON batch operation.
type JSONItem struct {
	Type       JSONItemType   `json:"type"`
	Operation  JSONOperation  `json:"operation,omitempty"`
	ID         string         `json:"id,omitempty"`
	Attributes map[string]any `json:"attributes,omitempty"`
}
