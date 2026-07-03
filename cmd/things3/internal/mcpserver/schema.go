package mcpserver

import (
	"reflect"

	"github.com/google/jsonschema-go/jsonschema"
)

// Named string types for the tool parameters that are constrained enums. The
// official SDK cannot express an enum through a struct tag, so each type is
// mapped to an explicit schema in enumSchemas and injected during input-schema
// inference. Keeping them as distinct Go types lets one map drive every tool.
type (
	// ViewName is the sidebar view selected by list_todos.
	ViewName string
	// StatusFilter is the status selector shared by list_projects and search.
	StatusFilter string
	// SearchType narrows search to todos, projects, or both.
	SearchType string
	// CompleteStatus is the required status flip for the complete tool.
	CompleteStatus string
	// OpenView is the built-in list open can reveal.
	OpenView string
)

// Enum value sets, also used by handlers to validate and default.
const (
	statusIncomplete = "incomplete"
	statusCompleted  = "completed"
	statusCanceled   = "canceled"
	statusAny        = "any"

	searchTodo    = "todo"
	searchProject = "project"
	searchAny     = "any"
)

// viewNames is the list_todos view enum, mirroring the CLI's eight view commands.
var viewNames = []string{nameInbox, nameToday, nameUpcoming, nameAnytime, nameSomeday, nameLogbook, nameDeadlines, nameTrash}

// openViews is the open.view enum: the navigable built-in lists plus projects.
// trash is excluded because the URL scheme has no list id for it.
var openViews = []string{nameInbox, nameToday, nameUpcoming, nameAnytime, nameSomeday, nameLogbook, nameDeadlines, nameProjects}

// enumSchemas maps each named enum type to its JSON Schema. jsonschema.For clones
// these in when it encounters the type during inference.
var enumSchemas = map[reflect.Type]*jsonschema.Schema{
	reflect.TypeFor[ViewName]():       enumSchema(viewNames...),
	reflect.TypeFor[StatusFilter]():   enumSchema(statusIncomplete, statusCompleted, statusCanceled, statusAny),
	reflect.TypeFor[SearchType]():     enumSchema(searchTodo, searchProject, searchAny),
	reflect.TypeFor[CompleteStatus](): enumSchema(statusCompleted, statusCanceled, statusIncomplete),
	reflect.TypeFor[OpenView]():       enumSchema(openViews...),
}

// enumSchema builds a string schema constrained to the given values.
func enumSchema(values ...string) *jsonschema.Schema {
	enum := make([]any, len(values))
	for i, v := range values {
		enum[i] = v
	}
	return &jsonschema.Schema{Type: "string", Enum: enum}
}

// inputSchemaFor infers an input schema for In, injecting the enum schemas so
// every constrained field is rejected by the SDK before a handler runs.
func inputSchemaFor[In any]() (*jsonschema.Schema, error) {
	return jsonschema.For[In](&jsonschema.ForOptions{TypeSchemas: enumSchemas})
}
