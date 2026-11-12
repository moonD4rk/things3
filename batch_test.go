package things3

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The json command is the only way the URL scheme can create a heading: a
// project's items list headings and todos in order, and a todo belongs to the
// heading that precedes it.
func TestBatchProjectHeadings(t *testing.T) {
	thingsURL, err := newScheme().Batch().
		AddProject(func(p BatchProjectConfigurator) {
			p.Title("Trip").
				Todos(func(todo BatchTodoConfigurator) { todo.Title("Loose") }).
				Heading("Before", func(todo BatchTodoConfigurator) { todo.Title("Book flights") }).
				Heading("Empty")
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, thingsURL)
	require.Len(t, items, 1)
	children, ok := items[0].Attributes["items"].([]any)
	require.True(t, ok, "a project's children ride in items")

	var shape []string
	for _, raw := range children {
		child := raw.(map[string]any)
		attrs := child["attributes"].(map[string]any)
		shape = append(shape, child["type"].(string)+":"+attrs["title"].(string))
	}
	assert.Equal(t, []string{"to-do:Loose", "heading:Before", "to-do:Book flights", "heading:Empty"}, shape)
}

func TestBatchReminder(t *testing.T) {
	thingsURL, err := newScheme().Batch().
		AddTodo(func(todo BatchTodoConfigurator) { todo.Title("Call").WhenEvening().Reminder(18, 5) }).
		AddTodo(func(todo BatchTodoConfigurator) { todo.Title("Wake").Reminder(7, 0) }).
		AddProject(func(p BatchProjectConfigurator) { p.Title("Launch").Reminder(9, 30) }).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, thingsURL)
	require.Len(t, items, 3)
	assert.Equal(t, "evening@18:05", items[0].Attributes["when"])
	assert.Equal(t, "today@07:00", items[1].Attributes["when"], "a reminder without a day lands on today")
	assert.Equal(t, "today@09:30", items[2].Attributes["when"])

	_, err = newScheme().Batch().
		AddTodo(func(todo BatchTodoConfigurator) { todo.Title("Later").WhenSomeday().Reminder(9, 0) }).
		Build()
	require.ErrorIs(t, err, ErrReminderNeedsDate)

	_, err = newScheme().Batch().
		AddProject(func(p BatchProjectConfigurator) {
			p.Title("Trip").Todos(func(todo BatchTodoConfigurator) { todo.Title("x").Reminder(25, 0) })
		}).
		Build()
	require.ErrorIs(t, err, ErrInvalidReminderTime, "a child todo's error fails the batch")
}

func TestBatchChecklistState(t *testing.T) {
	thingsURL, err := newScheme().Batch().
		AddTodo(func(todo BatchTodoConfigurator) {
			todo.Title("Pack").Checklist(
				ChecklistEntry{Title: "Passport", Completed: true},
				ChecklistEntry{Title: "Umbrella", Canceled: true},
				ChecklistEntry{Title: "Charger"},
			)
		}).
		Build()
	require.NoError(t, err)

	entries := parseJSONItems(t, thingsURL)[0].Attributes["checklist-items"].([]any)
	require.Len(t, entries, 3)
	attrsOf := func(i int) map[string]any { return entries[i].(map[string]any)["attributes"].(map[string]any) }
	assert.Equal(t, map[string]any{"title": "Passport", "completed": true}, attrsOf(0))
	assert.Equal(t, map[string]any{"title": "Umbrella", "canceled": true}, attrsOf(1))
	assert.Equal(t, map[string]any{"title": "Charger"}, attrsOf(2))
}

func TestAuthBatchUpdateChecklistAndHeading(t *testing.T) {
	thingsURL, err := newScheme().WithToken("test-token").Batch().
		UpdateTodo("uuid-1", func(todo BatchTodoConfigurator) {
			todo.AppendChecklistItems("Last").PrependChecklistItems("First").HeadingID("heading-1")
		}).
		Build()
	require.NoError(t, err)

	attrs := parseJSONItems(t, thingsURL)[0].Attributes
	assert.Equal(t, "heading-1", attrs["heading-id"])
	for key, title := range map[string]string{"append-checklist-items": "Last", "prepend-checklist-items": "First"} {
		entries, ok := attrs[key].([]any)
		require.Truef(t, ok && len(entries) == 1, "%s", key)
		assert.Equal(t, title, entries[0].(map[string]any)["attributes"].(map[string]any)["title"])
	}
}
