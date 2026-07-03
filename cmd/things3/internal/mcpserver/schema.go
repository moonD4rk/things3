package mcpserver

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

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
// every constrained field is rejected by the SDK before a handler runs, then
// stamping the pagination bounds so limit and page carry machine-readable
// keywords rather than prose the model can ignore.
func inputSchemaFor[In any](maxLimit, defaultLimit int) (*jsonschema.Schema, error) {
	s, err := jsonschema.For[In](&jsonschema.ForOptions{TypeSchemas: enumSchemas})
	if err != nil {
		return nil, err
	}
	applyPageBounds(s, maxLimit, defaultLimit)
	return s, nil
}

// applyPageBounds stamps default/minimum/maximum onto the limit and page
// properties when a tool has them. The SDK runs ApplyDefaults then Validate on
// every call, so these are enforced, not merely advertised: an omitted limit
// arrives as defaultLimit and a limit above maxLimit is rejected before the
// handler runs. It is a no-op for tools without pagination.
func applyPageBounds(s *jsonschema.Schema, maxLimit, defaultLimit int) {
	if s == nil || s.Properties == nil {
		return
	}
	if lim := s.Properties["limit"]; lim != nil {
		lo, hi := 1.0, float64(maxLimit)
		lim.Minimum = &lo
		lim.Maximum = &hi
		lim.Default = json.RawMessage(strconv.Itoa(defaultLimit))
		lim.Description = fmt.Sprintf("page size; defaults to %d, capped at %d", defaultLimit, maxLimit)
	}
	if page := s.Properties["page"]; page != nil {
		lo := 1.0
		page.Minimum = &lo
		page.Default = json.RawMessage("1")
		page.Description = "1-based page number"
	}
}
