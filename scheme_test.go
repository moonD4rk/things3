package things3

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewScheme(t *testing.T) {
	scheme := NewScheme()
	assert.NotNil(t, scheme)
}

// =============================================================================
// TodoBuilder Tests
// =============================================================================

func TestTodoBuilder_Title(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Todo().Title("Buy groceries").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add", cmd)
	assertURLParam(t, params, "title", "Buy groceries")
	assertNoExtraParams(t, params, "title")
}

func TestTodoBuilder_TitleTooLong(t *testing.T) {
	scheme := NewScheme()
	longTitle := strings.Repeat("a", 4001)
	_, err := scheme.Todo().Title(longTitle).Build()
	assert.ErrorIs(t, err, ErrTitleTooLong)
}

func TestTodoBuilder_Notes(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Todo().Title("Test").Notes("Some notes").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add", cmd)
	assertURLParam(t, params, "title", "Test")
	assertURLParam(t, params, "notes", "Some notes")
	assertNoExtraParams(t, params, "title", "notes")
}

func TestTodoBuilder_NotesTooLong(t *testing.T) {
	scheme := NewScheme()
	longNotes := strings.Repeat("a", 10001)
	_, err := scheme.Todo().Title("Test").Notes(longNotes).Build()
	assert.ErrorIs(t, err, ErrNotesTooLong)
}

func TestTodoBuilder_When(t *testing.T) {
	tests := []struct {
		when     When
		expected string
	}{
		{WhenToday, "today"},
		{WhenTomorrow, "tomorrow"},
		{WhenEvening, "evening"},
		{WhenAnytime, "anytime"},
		{WhenSomeday, "someday"},
	}

	for _, tt := range tests {
		t.Run(string(tt.when), func(t *testing.T) {
			scheme := NewScheme()
			urlStr, err := scheme.Todo().Title("Test").When(tt.when).Build()
			require.NoError(t, err)

			cmd, params := parseThingsURL(t, urlStr)
			assert.Equal(t, "add", cmd)
			assertURLParam(t, params, "title", "Test")
			assertURLParam(t, params, "when", tt.expected)
			assertNoExtraParams(t, params, "title", "when")
		})
	}
}

func TestTodoBuilder_WhenDate(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Todo().Title("Test").WhenDate(2025, time.December, 25).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add", cmd)
	assertURLParam(t, params, "title", "Test")
	assertURLParam(t, params, "when", "2025-12-25")
	assertNoExtraParams(t, params, "title", "when")
}

func TestTodoBuilder_Deadline(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Todo().Title("Test").Deadline("2025-12-31").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add", cmd)
	assertURLParam(t, params, "title", "Test")
	assertURLParam(t, params, "deadline", "2025-12-31")
	assertNoExtraParams(t, params, "title", "deadline")
}

func TestTodoBuilder_Tags(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Todo().Title("Test").Tags("work", "urgent").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add", cmd)
	assertURLParam(t, params, "title", "Test")
	assertURLParam(t, params, "tags", "work,urgent")
	assertNoExtraParams(t, params, "title", "tags")
}

func TestTodoBuilder_ChecklistItems(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Todo().Title("Test").ChecklistItems("Item 1", "Item 2").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add", cmd)
	assertURLParam(t, params, "title", "Test")
	assertURLParam(t, params, "checklist-items", "Item 1\nItem 2")
	assertNoExtraParams(t, params, "title", "checklist-items")
}

func TestTodoBuilder_TooManyChecklistItems(t *testing.T) {
	scheme := NewScheme()
	items := make([]string, 101)
	for i := range items {
		items[i] = "checklist entry"
	}
	_, err := scheme.Todo().Title("Test").ChecklistItems(items...).Build()
	assert.ErrorIs(t, err, ErrTooManyChecklistItems)
}

func TestTodoBuilder_List(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Todo().Title("Test").List("My Project").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add", cmd)
	assertURLParam(t, params, "title", "Test")
	assertURLParam(t, params, "list", "My Project")
	assertNoExtraParams(t, params, "title", "list")
}

func TestTodoBuilder_ListID(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Todo().Title("Test").ListID("uuid-123").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add", cmd)
	assertURLParam(t, params, "title", "Test")
	assertURLParam(t, params, "list-id", "uuid-123")
	assertNoExtraParams(t, params, "title", "list-id")
}

func TestTodoBuilder_Completed(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Todo().Title("Test").Completed(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add", cmd)
	assertURLParam(t, params, "title", "Test")
	assertURLParam(t, params, "completed", "true")
	assertNoExtraParams(t, params, "title", "completed")
}

func TestTodoBuilder_Canceled(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Todo().Title("Test").Canceled(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add", cmd)
	assertURLParam(t, params, "title", "Test")
	assertURLParam(t, params, "canceled", "true")
	assertNoExtraParams(t, params, "title", "canceled")
}

func TestTodoBuilder_ShowQuickEntry(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Todo().Title("Test").ShowQuickEntry(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add", cmd)
	assertURLParam(t, params, "title", "Test")
	assertURLParam(t, params, "show-quick-entry", "true")
	assertNoExtraParams(t, params, "title", "show-quick-entry")
}

func TestTodoBuilder_Reveal(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Todo().Title("Test").Reveal(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add", cmd)
	assertURLParam(t, params, "title", "Test")
	assertURLParam(t, params, "reveal", "true")
	assertNoExtraParams(t, params, "title", "reveal")
}

func TestTodoBuilder_Titles(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Todo().Titles("Task 1", "Task 2", "Task 3").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add", cmd)
	assertURLParam(t, params, "titles", "Task 1\nTask 2\nTask 3")
	assertNoExtraParams(t, params, "titles")
}

func TestTodoBuilder_TitlesTooLong(t *testing.T) {
	scheme := NewScheme()
	longTitle := strings.Repeat("a", 2000)
	_, err := scheme.Todo().Titles(longTitle, longTitle, longTitle).Build()
	assert.ErrorIs(t, err, ErrTitleTooLong)
}

func TestTodoBuilder_Heading(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Todo().
		Title("Subtask").
		List("My Project").
		Heading("Phase 1").
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add", cmd)
	assertURLParam(t, params, "title", "Subtask")
	assertURLParam(t, params, "list", "My Project")
	assertURLParam(t, params, "heading", "Phase 1")
	assertNoExtraParams(t, params, "title", "list", "heading")
}

func TestTodoBuilder_HeadingID(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Todo().
		Title("Subtask").
		ListID("project-uuid").
		HeadingID("heading-uuid").
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add", cmd)
	assertURLParam(t, params, "title", "Subtask")
	assertURLParam(t, params, "list-id", "project-uuid")
	assertURLParam(t, params, "heading-id", "heading-uuid")
	assertNoExtraParams(t, params, "title", "list-id", "heading-id")
}

func TestTodoBuilder_CreationDate(t *testing.T) {
	scheme := NewScheme()
	pastDate := time.Date(2024, time.January, 15, 10, 30, 0, 0, time.UTC)
	urlStr, err := scheme.Todo().
		Title("Historical task").
		CreationDate(pastDate).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add", cmd)
	assertURLParam(t, params, "title", "Historical task")
	assertDateParam(t, params, "creation-date", 2024, time.January, 15)
	assertNoExtraParams(t, params, "title", "creation-date")
}

func TestTodoBuilder_CompletionDate(t *testing.T) {
	scheme := NewScheme()
	completedAt := time.Date(2024, time.December, 1, 14, 0, 0, 0, time.UTC)
	urlStr, err := scheme.Todo().
		Title("Imported completed task").
		Completed(true).
		CompletionDate(completedAt).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add", cmd)
	assertURLParam(t, params, "title", "Imported completed task")
	assertURLParam(t, params, "completed", "true")
	assertDateParam(t, params, "completion-date", 2024, time.December, 1)
	assertNoExtraParams(t, params, "title", "completed", "completion-date")
}

func TestTodoBuilder_Chained(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Todo().
		Title("Buy groceries").
		Notes("Don't forget milk").
		When(WhenToday).
		Tags("shopping").
		Reveal(true).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add", cmd)
	assertURLParam(t, params, "title", "Buy groceries")
	assertURLParam(t, params, "notes", "Don't forget milk")
	assertURLParam(t, params, "when", "today")
	assertURLParam(t, params, "tags", "shopping")
	assertURLParam(t, params, "reveal", "true")
	assertNoExtraParams(t, params, "title", "notes", "when", "tags", "reveal")
}

// =============================================================================
// ProjectBuilder Tests
// =============================================================================

func TestProjectBuilder_Title(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Project().Title("New Project").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add-project", cmd)
	assertURLParam(t, params, "title", "New Project")
	assertNoExtraParams(t, params, "title")
}

func TestProjectBuilder_TitleTooLong(t *testing.T) {
	scheme := NewScheme()
	longTitle := strings.Repeat("a", 4001)
	_, err := scheme.Project().Title(longTitle).Build()
	assert.ErrorIs(t, err, ErrTitleTooLong)
}

func TestProjectBuilder_Area(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Project().Title("Test").Area("Work").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add-project", cmd)
	assertURLParam(t, params, "title", "Test")
	assertURLParam(t, params, "area", "Work")
	assertNoExtraParams(t, params, "title", "area")
}

func TestProjectBuilder_AreaID(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Project().Title("Test").AreaID("uuid-123").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add-project", cmd)
	assertURLParam(t, params, "title", "Test")
	assertURLParam(t, params, "area-id", "uuid-123")
	assertNoExtraParams(t, params, "title", "area-id")
}

func TestProjectBuilder_Todos(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Project().Title("Test").Todos("Task 1", "Task 2").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add-project", cmd)
	assertURLParam(t, params, "title", "Test")
	assertURLParam(t, params, "to-dos", "Task 1\nTask 2")
	assertNoExtraParams(t, params, "title", "to-dos")
}

func TestProjectBuilder_Notes(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Project().
		Title("Q1 Goals").
		Notes("Quarterly objectives and key results").
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add-project", cmd)
	assertURLParam(t, params, "title", "Q1 Goals")
	assertURLParam(t, params, "notes", "Quarterly objectives and key results")
	assertNoExtraParams(t, params, "title", "notes")
}

func TestProjectBuilder_When(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Project().Title("Future Project").When(WhenSomeday).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add-project", cmd)
	assertURLParam(t, params, "title", "Future Project")
	assertURLParam(t, params, "when", "someday")
	assertNoExtraParams(t, params, "title", "when")
}

func TestProjectBuilder_WhenDate(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Project().
		Title("Launch").
		WhenDate(2025, time.March, 1).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add-project", cmd)
	assertURLParam(t, params, "title", "Launch")
	assertURLParam(t, params, "when", "2025-03-01")
	assertNoExtraParams(t, params, "title", "when")
}

func TestProjectBuilder_Deadline(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Project().
		Title("Release v2.0").
		Deadline("2025-06-30").
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add-project", cmd)
	assertURLParam(t, params, "title", "Release v2.0")
	assertURLParam(t, params, "deadline", "2025-06-30")
	assertNoExtraParams(t, params, "title", "deadline")
}

func TestProjectBuilder_Tags(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Project().
		Title("Website Redesign").
		Tags("work", "high-priority").
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add-project", cmd)
	assertURLParam(t, params, "title", "Website Redesign")
	assertURLParam(t, params, "tags", "work,high-priority")
	assertNoExtraParams(t, params, "title", "tags")
}

func TestProjectBuilder_Completed(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Project().
		Title("Archived Project").
		Completed(true).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add-project", cmd)
	assertURLParam(t, params, "title", "Archived Project")
	assertURLParam(t, params, "completed", "true")
	assertNoExtraParams(t, params, "title", "completed")
}

func TestProjectBuilder_Canceled(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Project().
		Title("Discontinued Project").
		Canceled(true).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add-project", cmd)
	assertURLParam(t, params, "title", "Discontinued Project")
	assertURLParam(t, params, "canceled", "true")
	assertNoExtraParams(t, params, "title", "canceled")
}

func TestProjectBuilder_Reveal(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Project().
		Title("New Project").
		Reveal(true).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add-project", cmd)
	assertURLParam(t, params, "title", "New Project")
	assertURLParam(t, params, "reveal", "true")
	assertNoExtraParams(t, params, "title", "reveal")
}

func TestProjectBuilder_CreationDate(t *testing.T) {
	scheme := NewScheme()
	created := time.Date(2024, time.June, 1, 9, 0, 0, 0, time.UTC)
	urlStr, err := scheme.Project().
		Title("Historical Project").
		CreationDate(created).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add-project", cmd)
	assertURLParam(t, params, "title", "Historical Project")
	assertDateParam(t, params, "creation-date", 2024, time.June, 1)
	assertNoExtraParams(t, params, "title", "creation-date")
}

func TestProjectBuilder_CompletionDate(t *testing.T) {
	scheme := NewScheme()
	completed := time.Date(2024, time.November, 15, 17, 0, 0, 0, time.UTC)
	urlStr, err := scheme.Project().
		Title("Imported Completed Project").
		Completed(true).
		CompletionDate(completed).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add-project", cmd)
	assertURLParam(t, params, "title", "Imported Completed Project")
	assertURLParam(t, params, "completed", "true")
	assertDateParam(t, params, "completion-date", 2024, time.November, 15)
	assertNoExtraParams(t, params, "title", "completed", "completion-date")
}

func TestProjectBuilder_FullProject(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.Project().
		Title("Product Launch").
		Notes("Launch plan for v2.0").
		Area("Work").
		Tags("priority").
		Deadline("2025-03-31").
		Todos("Design", "Development", "Testing", "Release").
		Reveal(true).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "add-project", cmd)
	assertURLParam(t, params, "title", "Product Launch")
	assertURLParam(t, params, "notes", "Launch plan for v2.0")
	assertURLParam(t, params, "area", "Work")
	assertURLParam(t, params, "tags", "priority")
	assertURLParam(t, params, "deadline", "2025-03-31")
	assertURLParam(t, params, "to-dos", "Design\nDevelopment\nTesting\nRelease")
	assertURLParam(t, params, "reveal", "true")
	assertNoExtraParams(t, params, "title", "notes", "area", "tags", "deadline", "to-dos", "reveal")
}

// =============================================================================
// ShowBuilder Tests
// =============================================================================

func TestShowBuilder_ID(t *testing.T) {
	scheme := NewScheme()
	urlStr := scheme.Show().ID("uuid-123").Build()

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "show", cmd)
	assertURLParam(t, params, "id", "uuid-123")
	assertNoExtraParams(t, params, "id")
}

func TestShowBuilder_List(t *testing.T) {
	tests := []struct {
		list     ListID
		expected string
	}{
		{ListInbox, "inbox"},
		{ListToday, "today"},
		{ListAnytime, "anytime"},
		{ListUpcoming, "upcoming"},
		{ListSomeday, "someday"},
		{ListLogbook, "logbook"},
		{ListTomorrow, "tomorrow"},
		{ListDeadlines, "deadlines"},
		{ListRepeating, "repeating"},
		{ListAllProjects, "all-projects"},
		{ListLoggedProjects, "logged-projects"},
	}

	for _, tt := range tests {
		t.Run(string(tt.list), func(t *testing.T) {
			scheme := NewScheme()
			urlStr := scheme.Show().List(tt.list).Build()

			cmd, params := parseThingsURL(t, urlStr)
			assert.Equal(t, "show", cmd)
			assertURLParam(t, params, "id", tt.expected)
			assertNoExtraParams(t, params, "id")
		})
	}
}

func TestShowBuilder_Query(t *testing.T) {
	scheme := NewScheme()
	urlStr := scheme.Show().Query("My Project").Build()

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "show", cmd)
	assertURLParam(t, params, "query", "My Project")
	assertNoExtraParams(t, params, "query")
}

func TestShowBuilder_Filter(t *testing.T) {
	scheme := NewScheme()
	urlStr := scheme.Show().List(ListToday).Filter("work", "urgent").Build()

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "show", cmd)
	assertURLParam(t, params, "id", "today")
	assertURLParam(t, params, "filter", "work,urgent")
	assertNoExtraParams(t, params, "id", "filter")
}

func TestShowBuilder_NoParams(t *testing.T) {
	scheme := NewScheme()
	urlStr := scheme.Show().Build()
	assert.Equal(t, "things:///show", urlStr)
}

// =============================================================================
// Search and Version Tests
// =============================================================================

func TestScheme_Search(t *testing.T) {
	scheme := NewScheme()
	urlStr := scheme.Search("my query")

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "search", cmd)
	assertURLParam(t, params, "query", "my query")
	assertNoExtraParams(t, params, "query")
}

func TestScheme_Version(t *testing.T) {
	scheme := NewScheme()
	urlStr := scheme.Version()
	assert.Equal(t, "things:///version", urlStr)
}

// =============================================================================
// AuthScheme Tests
// =============================================================================

func TestAuthScheme_WithToken(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	assert.NotNil(t, auth)
}

func TestAuthScheme_EmptyToken(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("")
	_, err := auth.UpdateTodo("uuid").Completed(true).Build()
	assert.ErrorIs(t, err, ErrEmptyToken)
}

// =============================================================================
// UpdateTodoBuilder Tests
// =============================================================================

func TestUpdateTodoBuilder_Completed(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateTodo("uuid-123").Completed(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update", cmd)
	assertURLParam(t, params, "id", "uuid-123")
	assertURLParam(t, params, "auth-token", "test-token")
	assertURLParam(t, params, "completed", "true")
	assertNoExtraParams(t, params, "id", "auth-token", "completed")
}

func TestUpdateTodoBuilder_NoID(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	_, err := auth.UpdateTodo("").Completed(true).Build()
	assert.ErrorIs(t, err, ErrIDRequired)
}

func TestUpdateTodoBuilder_Title(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateTodo("uuid").Title("New Title").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update", cmd)
	assertURLParam(t, params, "id", "uuid")
	assertURLParam(t, params, "auth-token", "test-token")
	assertURLParam(t, params, "title", "New Title")
	assertNoExtraParams(t, params, "id", "auth-token", "title")
}

func TestUpdateTodoBuilder_PrependNotes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateTodo("uuid").PrependNotes("Prefix: ").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update", cmd)
	assertURLParam(t, params, "prepend-notes", "Prefix: ")
	assertNoExtraParams(t, params, "id", "auth-token", "prepend-notes")
}

func TestUpdateTodoBuilder_AppendNotes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateTodo("uuid").AppendNotes(" - Suffix").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update", cmd)
	assertURLParam(t, params, "append-notes", " - Suffix")
	assertNoExtraParams(t, params, "id", "auth-token", "append-notes")
}

func TestUpdateTodoBuilder_AddTags(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateTodo("uuid").AddTags("new-tag").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update", cmd)
	assertURLParam(t, params, "add-tags", "new-tag")
	assertNoExtraParams(t, params, "id", "auth-token", "add-tags")
}

func TestUpdateTodoBuilder_ClearDeadline(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateTodo("uuid").ClearDeadline().Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update", cmd)
	assertURLParam(t, params, "deadline", "")
	assertNoExtraParams(t, params, "id", "auth-token", "deadline")
}

func TestUpdateTodoBuilder_Duplicate(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateTodo("uuid").Duplicate(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update", cmd)
	assertURLParam(t, params, "duplicate", "true")
	assertNoExtraParams(t, params, "id", "auth-token", "duplicate")
}

func TestUpdateTodoBuilder_Notes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateTodo("uuid").Notes("New description").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update", cmd)
	assertURLParam(t, params, "notes", "New description")
	assertNoExtraParams(t, params, "id", "auth-token", "notes")
}

func TestUpdateTodoBuilder_When(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateTodo("uuid").When(WhenTomorrow).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update", cmd)
	assertURLParam(t, params, "when", "tomorrow")
	assertNoExtraParams(t, params, "id", "auth-token", "when")
}

func TestUpdateTodoBuilder_WhenDate(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateTodo("uuid").WhenDate(2025, time.February, 14).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update", cmd)
	assertURLParam(t, params, "when", "2025-02-14")
	assertNoExtraParams(t, params, "id", "auth-token", "when")
}

func TestUpdateTodoBuilder_Deadline(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateTodo("uuid").Deadline("2025-01-31").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update", cmd)
	assertURLParam(t, params, "deadline", "2025-01-31")
	assertNoExtraParams(t, params, "id", "auth-token", "deadline")
}

func TestUpdateTodoBuilder_Tags(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateTodo("uuid").Tags("new-tag-1", "new-tag-2").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update", cmd)
	assertURLParam(t, params, "tags", "new-tag-1,new-tag-2")
	assertNoExtraParams(t, params, "id", "auth-token", "tags")
}

func TestUpdateTodoBuilder_ChecklistItems(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateTodo("uuid").ChecklistItems("Step A", "Step B").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update", cmd)
	assertURLParam(t, params, "checklist-items", "Step A\nStep B")
	assertNoExtraParams(t, params, "id", "auth-token", "checklist-items")
}

func TestUpdateTodoBuilder_PrependChecklistItems(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateTodo("uuid").PrependChecklistItems("First step").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update", cmd)
	assertURLParam(t, params, "prepend-checklist-items", "First step")
	assertNoExtraParams(t, params, "id", "auth-token", "prepend-checklist-items")
}

func TestUpdateTodoBuilder_AppendChecklistItems(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateTodo("uuid").AppendChecklistItems("Final step").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update", cmd)
	assertURLParam(t, params, "append-checklist-items", "Final step")
	assertNoExtraParams(t, params, "id", "auth-token", "append-checklist-items")
}

func TestUpdateTodoBuilder_List(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateTodo("uuid").List("New Project").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update", cmd)
	assertURLParam(t, params, "list", "New Project")
	assertNoExtraParams(t, params, "id", "auth-token", "list")
}

func TestUpdateTodoBuilder_ListID(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateTodo("uuid").ListID("project-uuid").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update", cmd)
	assertURLParam(t, params, "list-id", "project-uuid")
	assertNoExtraParams(t, params, "id", "auth-token", "list-id")
}

func TestUpdateTodoBuilder_Heading(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateTodo("uuid").Heading("Phase 2").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update", cmd)
	assertURLParam(t, params, "heading", "Phase 2")
	assertNoExtraParams(t, params, "id", "auth-token", "heading")
}

func TestUpdateTodoBuilder_HeadingID(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateTodo("uuid").HeadingID("heading-uuid").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update", cmd)
	assertURLParam(t, params, "heading-id", "heading-uuid")
	assertNoExtraParams(t, params, "id", "auth-token", "heading-id")
}

func TestUpdateTodoBuilder_Canceled(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateTodo("uuid").Canceled(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update", cmd)
	assertURLParam(t, params, "canceled", "true")
	assertNoExtraParams(t, params, "id", "auth-token", "canceled")
}

func TestUpdateTodoBuilder_Reveal(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateTodo("uuid").Reveal(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update", cmd)
	assertURLParam(t, params, "reveal", "true")
	assertNoExtraParams(t, params, "id", "auth-token", "reveal")
}

func TestUpdateTodoBuilder_CreationDate(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	created := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	urlStr, err := auth.UpdateTodo("uuid").CreationDate(created).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update", cmd)
	assertDateParam(t, params, "creation-date", 2024, time.January, 1)
	assertNoExtraParams(t, params, "id", "auth-token", "creation-date")
}

func TestUpdateTodoBuilder_CompletionDate(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	completed := time.Date(2024, time.December, 31, 23, 59, 0, 0, time.UTC)
	urlStr, err := auth.UpdateTodo("uuid").Completed(true).CompletionDate(completed).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update", cmd)
	assertURLParam(t, params, "completed", "true")
	assertDateParam(t, params, "completion-date", 2024, time.December, 31)
	assertNoExtraParams(t, params, "id", "auth-token", "completed", "completion-date")
}

func TestUpdateTodoBuilder_ValidationError(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	longTitle := strings.Repeat("a", 4001)
	_, err := auth.UpdateTodo("uuid").Title(longTitle).Build()
	assert.ErrorIs(t, err, ErrTitleTooLong)
}

// =============================================================================
// UpdateProjectBuilder Tests
// =============================================================================

func TestUpdateProjectBuilder_Title(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateProject("uuid").Title("New Project Title").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update-project", cmd)
	assertURLParam(t, params, "id", "uuid")
	assertURLParam(t, params, "auth-token", "test-token")
	assertURLParam(t, params, "title", "New Project Title")
	assertNoExtraParams(t, params, "id", "auth-token", "title")
}

func TestUpdateProjectBuilder_Completed(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateProject("uuid").Completed(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update-project", cmd)
	assertURLParam(t, params, "completed", "true")
	assertNoExtraParams(t, params, "id", "auth-token", "completed")
}

func TestUpdateProjectBuilder_Canceled(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateProject("uuid").Canceled(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update-project", cmd)
	assertURLParam(t, params, "canceled", "true")
	assertNoExtraParams(t, params, "id", "auth-token", "canceled")
}

func TestUpdateProjectBuilder_Notes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateProject("uuid").Notes("Updated project description").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update-project", cmd)
	assertURLParam(t, params, "notes", "Updated project description")
	assertNoExtraParams(t, params, "id", "auth-token", "notes")
}

func TestUpdateProjectBuilder_PrependNotes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateProject("uuid").PrependNotes("[UPDATE] ").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update-project", cmd)
	assertURLParam(t, params, "prepend-notes", "[UPDATE] ")
	assertNoExtraParams(t, params, "id", "auth-token", "prepend-notes")
}

func TestUpdateProjectBuilder_AppendNotes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateProject("uuid").AppendNotes("\n- Added new requirement").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update-project", cmd)
	assertURLParam(t, params, "append-notes", "\n- Added new requirement")
	assertNoExtraParams(t, params, "id", "auth-token", "append-notes")
}

func TestUpdateProjectBuilder_When(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateProject("uuid").When(WhenAnytime).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update-project", cmd)
	assertURLParam(t, params, "when", "anytime")
	assertNoExtraParams(t, params, "id", "auth-token", "when")
}

func TestUpdateProjectBuilder_WhenDate(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateProject("uuid").WhenDate(2025, time.April, 1).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update-project", cmd)
	assertURLParam(t, params, "when", "2025-04-01")
	assertNoExtraParams(t, params, "id", "auth-token", "when")
}

func TestUpdateProjectBuilder_Deadline(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateProject("uuid").Deadline("2025-12-31").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update-project", cmd)
	assertURLParam(t, params, "deadline", "2025-12-31")
	assertNoExtraParams(t, params, "id", "auth-token", "deadline")
}

func TestUpdateProjectBuilder_ClearDeadline(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateProject("uuid").ClearDeadline().Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update-project", cmd)
	assertURLParam(t, params, "deadline", "")
	assertNoExtraParams(t, params, "id", "auth-token", "deadline")
}

func TestUpdateProjectBuilder_Tags(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateProject("uuid").Tags("priority", "q1").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update-project", cmd)
	assertURLParam(t, params, "tags", "priority,q1")
	assertNoExtraParams(t, params, "id", "auth-token", "tags")
}

func TestUpdateProjectBuilder_AddTags(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateProject("uuid").AddTags("reviewed").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update-project", cmd)
	assertURLParam(t, params, "add-tags", "reviewed")
	assertNoExtraParams(t, params, "id", "auth-token", "add-tags")
}

func TestUpdateProjectBuilder_Area(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateProject("uuid").Area("Personal").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update-project", cmd)
	assertURLParam(t, params, "area", "Personal")
	assertNoExtraParams(t, params, "id", "auth-token", "area")
}

func TestUpdateProjectBuilder_AreaID(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateProject("uuid").AreaID("area-uuid").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update-project", cmd)
	assertURLParam(t, params, "area-id", "area-uuid")
	assertNoExtraParams(t, params, "id", "auth-token", "area-id")
}

func TestUpdateProjectBuilder_Reveal(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.UpdateProject("uuid").Reveal(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "update-project", cmd)
	assertURLParam(t, params, "reveal", "true")
	assertNoExtraParams(t, params, "id", "auth-token", "reveal")
}

func TestUpdateProjectBuilder_NoID(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	_, err := auth.UpdateProject("").Completed(true).Build()
	assert.ErrorIs(t, err, ErrIDRequired)
}

func TestUpdateProjectBuilder_ValidationError(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	longTitle := strings.Repeat("a", 4001)
	_, err := auth.UpdateProject("uuid").Title(longTitle).Build()
	assert.ErrorIs(t, err, ErrTitleTooLong)
}

// =============================================================================
// JSONBuilder Tests
// =============================================================================

func TestJSONBuilder_AddTodo(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test Todo")
		}).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "json", cmd)
	assertURLParamExists(t, params, "data")
	assertURLParamNotExists(t, params, "reveal")

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeTodo, items[0].Type)
	assert.Equal(t, "Test Todo", items[0].Attributes["title"])
}

func TestJSONBuilder_AddProject(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test Project")
		}).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "json", cmd)
	assertURLParamExists(t, params, "data")

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeProject, items[0].Type)
	assert.Equal(t, "Test Project", items[0].Attributes["title"])
}

func TestJSONBuilder_Reveal(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test")
		}).
		Reveal(true).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "json", cmd)
	assertURLParam(t, params, "reveal", "true")
	assertURLParamExists(t, params, "data")
}

func TestJSONBuilder_NoItems(t *testing.T) {
	scheme := NewScheme()
	_, err := scheme.JSON().Build()
	assert.ErrorIs(t, err, ErrNoJSONItems)
}

func TestJSONBuilder_Multiple(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Todo 1")
		}).
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Todo 2")
		}).
		Build()
	require.NoError(t, err)

	cmd, _ := parseThingsURL(t, urlStr)
	assert.Equal(t, "json", cmd)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 2)

	assert.Equal(t, JSONItemTypeTodo, items[0].Type)
	assert.Equal(t, "Todo 1", items[0].Attributes["title"])

	assert.Equal(t, JSONItemTypeTodo, items[1].Type)
	assert.Equal(t, "Todo 2", items[1].Attributes["title"])
}

// =============================================================================
// AuthJSONBuilder Tests
// =============================================================================

func TestAuthJSONBuilder_UpdateTodo(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.JSON().
		UpdateTodo("uuid-123", func(todo *JSONTodoBuilder) {
			todo.Completed(true)
		}).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "json", cmd)
	assertURLParam(t, params, "auth-token", "test-token")
	assertURLParamExists(t, params, "data")

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeTodo, items[0].Type)
	assert.Equal(t, JSONOperationUpdate, items[0].Operation)
	assert.Equal(t, "uuid-123", items[0].ID)
	assert.Equal(t, true, items[0].Attributes["completed"])
}

func TestAuthJSONBuilder_UpdateProject(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.JSON().
		UpdateProject("uuid-123", func(project *JSONProjectBuilder) {
			project.Completed(true)
		}).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "json", cmd)
	assertURLParam(t, params, "auth-token", "test-token")
	assertURLParamExists(t, params, "data")

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeProject, items[0].Type)
	assert.Equal(t, JSONOperationUpdate, items[0].Operation)
	assert.Equal(t, "uuid-123", items[0].ID)
	assert.Equal(t, true, items[0].Attributes["completed"])
}

func TestAuthJSONBuilder_Mixed(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("New Todo")
		}).
		UpdateTodo("uuid-123", func(todo *JSONTodoBuilder) {
			todo.Completed(true)
		}).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "json", cmd)
	assertURLParam(t, params, "auth-token", "test-token")

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 2)

	// First item: create todo
	assert.Equal(t, JSONItemTypeTodo, items[0].Type)
	assert.Equal(t, "New Todo", items[0].Attributes["title"])

	// Second item: update todo
	assert.Equal(t, JSONItemTypeTodo, items[1].Type)
	assert.Equal(t, JSONOperationUpdate, items[1].Operation)
	assert.Equal(t, "uuid-123", items[1].ID)
	assert.Equal(t, true, items[1].Attributes["completed"])
}

func TestAuthJSONBuilder_EmptyToken(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("")
	_, err := auth.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test")
		}).
		Build()
	assert.ErrorIs(t, err, ErrEmptyToken)
}

func TestAuthJSONBuilder_NoItems(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	_, err := auth.JSON().Build()
	assert.ErrorIs(t, err, ErrNoJSONItems)
}

// =============================================================================
// JSONTodoBuilder Tests
// =============================================================================

func TestJSONTodoBuilder_When(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").When(WhenToday)
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, "Test", items[0].Attributes["title"])
	assert.Equal(t, "today", items[0].Attributes["when"])
}

func TestJSONTodoBuilder_Tags(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").Tags("Risk", "Golang")
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, "Test", items[0].Attributes["title"])
	assert.Equal(t, []any{"Risk", "Golang"}, items[0].Attributes["tags"])
}

func TestJSONTodoBuilder_TitleTooLong(t *testing.T) {
	scheme := NewScheme()
	longTitle := strings.Repeat("a", 4001)
	_, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title(longTitle)
		}).
		Build()
	assert.ErrorIs(t, err, ErrTitleTooLong)
}

// TestJSONTodoBuilder_Notes tests adding notes to a JSON todo
func TestJSONTodoBuilder_Notes(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").Notes("Detailed description")
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, "Test", items[0].Attributes["title"])
	assert.Equal(t, "Detailed description", items[0].Attributes["notes"])
}

// TestJSONTodoBuilder_WhenDate tests scheduling to a specific date
func TestJSONTodoBuilder_WhenDate(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").WhenDate(2025, time.March, 15)
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, "Test", items[0].Attributes["title"])
	assert.Equal(t, "2025-03-15", items[0].Attributes["when"])
}

// TestJSONTodoBuilder_Deadline tests setting a deadline
func TestJSONTodoBuilder_Deadline(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").Deadline("2025-06-30")
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, "Test", items[0].Attributes["title"])
	assert.Equal(t, "2025-06-30", items[0].Attributes["deadline"])
}

// TestJSONTodoBuilder_ChecklistItems tests adding a checklist
func TestJSONTodoBuilder_ChecklistItems(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").ChecklistItems("Step 1", "Step 2", "Step 3")
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeTodo, items[0].Type)
	assert.Equal(t, "Test", items[0].Attributes["title"])

	checklistItems := items[0].Attributes["checklist-items"].([]any)
	require.Len(t, checklistItems, 3)

	item0 := checklistItems[0].(map[string]any)
	assert.Equal(t, "checklist-item", item0["type"])
	item0Attrs := item0["attributes"].(map[string]any)
	assert.Equal(t, "Step 1", item0Attrs["title"])

	item1 := checklistItems[1].(map[string]any)
	item1Attrs := item1["attributes"].(map[string]any)
	assert.Equal(t, "Step 2", item1Attrs["title"])

	item2 := checklistItems[2].(map[string]any)
	item2Attrs := item2["attributes"].(map[string]any)
	assert.Equal(t, "Step 3", item2Attrs["title"])
}

// TestJSONTodoBuilder_ChecklistItemsTooMany tests the checklist limit
func TestJSONTodoBuilder_ChecklistItemsTooMany(t *testing.T) {
	scheme := NewScheme()
	items := make([]string, 101)
	for i := range items {
		items[i] = "json checklist entry"
	}
	_, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").ChecklistItems(items...)
		}).
		Build()
	assert.ErrorIs(t, err, ErrTooManyChecklistItems)
}

// TestJSONTodoBuilder_List tests placing todo in a project by name
func TestJSONTodoBuilder_List(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").List("My Project")
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeTodo, items[0].Type)
	assert.Equal(t, "Test", items[0].Attributes["title"])
	assert.Equal(t, "My Project", items[0].Attributes["list"])
}

// TestJSONTodoBuilder_ListID tests placing todo in a project by UUID
func TestJSONTodoBuilder_ListID(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").ListID("project-uuid-123")
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeTodo, items[0].Type)
	assert.Equal(t, "Test", items[0].Attributes["title"])
	assert.Equal(t, "project-uuid-123", items[0].Attributes["list-id"])
}

// TestJSONTodoBuilder_Heading tests placing todo under a heading
func TestJSONTodoBuilder_Heading(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").List("Project").Heading("Phase 1")
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeTodo, items[0].Type)
	assert.Equal(t, "Test", items[0].Attributes["title"])
	assert.Equal(t, "Project", items[0].Attributes["list"])
	assert.Equal(t, "Phase 1", items[0].Attributes["heading"])
}

// TestJSONTodoBuilder_Completed tests marking as completed
func TestJSONTodoBuilder_Completed(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").Completed(true)
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeTodo, items[0].Type)
	assert.Equal(t, "Test", items[0].Attributes["title"])
	assert.Equal(t, true, items[0].Attributes["completed"])
}

// TestJSONTodoBuilder_Canceled tests marking as canceled
func TestJSONTodoBuilder_Canceled(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").Canceled(true)
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeTodo, items[0].Type)
	assert.Equal(t, "Test", items[0].Attributes["title"])
	assert.Equal(t, true, items[0].Attributes["canceled"])
}

// TestJSONTodoBuilder_CreationDate tests backdating creation
func TestJSONTodoBuilder_CreationDate(t *testing.T) {
	scheme := NewScheme()
	pastDate := time.Date(2024, time.June, 1, 10, 0, 0, 0, time.UTC)
	urlStr, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").CreationDate(pastDate)
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeTodo, items[0].Type)
	assert.Equal(t, "Test", items[0].Attributes["title"])
	assertJSONDateAttr(t, items[0].Attributes, "creation-date", 2024, time.June, 1)
}

// TestJSONTodoBuilder_CompletionDate tests setting completion timestamp
func TestJSONTodoBuilder_CompletionDate(t *testing.T) {
	scheme := NewScheme()
	completedDate := time.Date(2024, time.December, 15, 14, 30, 0, 0, time.UTC)
	urlStr, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").Completed(true).CompletionDate(completedDate)
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeTodo, items[0].Type)
	assert.Equal(t, "Test", items[0].Attributes["title"])
	assert.Equal(t, true, items[0].Attributes["completed"])
	assertJSONDateAttr(t, items[0].Attributes, "completion-date", 2024, time.December, 15)
}

// TestJSONTodoBuilder_UpdatePrependNotes tests prepending notes in update
func TestJSONTodoBuilder_UpdatePrependNotes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.JSON().
		UpdateTodo("uuid", func(todo *JSONTodoBuilder) {
			todo.PrependNotes("Important: ")
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeTodo, items[0].Type)
	assert.Equal(t, JSONOperationUpdate, items[0].Operation)
	assert.Equal(t, "uuid", items[0].ID)
	assert.Equal(t, "Important: ", items[0].Attributes["prepend-notes"])
}

// TestJSONTodoBuilder_UpdateAppendNotes tests appending notes in update
func TestJSONTodoBuilder_UpdateAppendNotes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.JSON().
		UpdateTodo("uuid", func(todo *JSONTodoBuilder) {
			todo.AppendNotes(" - Updated")
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeTodo, items[0].Type)
	assert.Equal(t, JSONOperationUpdate, items[0].Operation)
	assert.Equal(t, "uuid", items[0].ID)
	assert.Equal(t, " - Updated", items[0].Attributes["append-notes"])
}

// TestJSONTodoBuilder_UpdateAddTags tests adding tags without replacing
func TestJSONTodoBuilder_UpdateAddTags(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.JSON().
		UpdateTodo("uuid", func(todo *JSONTodoBuilder) {
			todo.AddTags("new-tag", "another-tag")
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeTodo, items[0].Type)
	assert.Equal(t, JSONOperationUpdate, items[0].Operation)
	addTags := items[0].Attributes["add-tags"].([]any)
	assert.Equal(t, []any{"new-tag", "another-tag"}, addTags)
}

// =============================================================================
// JSONProjectBuilder Tests
// =============================================================================

func TestJSONProjectBuilder_Area(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").Area("Work")
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeProject, items[0].Type)
	assert.Equal(t, "Test", items[0].Attributes["title"])
	assert.Equal(t, "Work", items[0].Attributes["area"])
}

func TestJSONProjectBuilder_Todos(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test Project").Todos(
				NewTodo().Title("Task 1"),
				NewTodo().Title("Task 2"),
			)
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeProject, items[0].Type)
	assert.Equal(t, "Test Project", items[0].Attributes["title"])

	todos := items[0].Attributes["items"].([]any)
	require.Len(t, todos, 2)

	todo0 := todos[0].(map[string]any)
	assert.Equal(t, "to-do", todo0["type"])
	todo0Attrs := todo0["attributes"].(map[string]any)
	assert.Equal(t, "Task 1", todo0Attrs["title"])

	todo1 := todos[1].(map[string]any)
	todo1Attrs := todo1["attributes"].(map[string]any)
	assert.Equal(t, "Task 2", todo1Attrs["title"])
}

// TestJSONProjectBuilder_Notes tests adding project notes
func TestJSONProjectBuilder_Notes(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").Notes("Project description")
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeProject, items[0].Type)
	assert.Equal(t, "Test", items[0].Attributes["title"])
	assert.Equal(t, "Project description", items[0].Attributes["notes"])
}

// TestJSONProjectBuilder_When tests scheduling project
func TestJSONProjectBuilder_When(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").When(WhenSomeday)
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeProject, items[0].Type)
	assert.Equal(t, "Test", items[0].Attributes["title"])
	assert.Equal(t, "someday", items[0].Attributes["when"])
}

// TestJSONProjectBuilder_WhenDate tests scheduling to specific date
func TestJSONProjectBuilder_WhenDate(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").WhenDate(2025, time.July, 1)
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeProject, items[0].Type)
	assert.Equal(t, "Test", items[0].Attributes["title"])
	assert.Equal(t, "2025-07-01", items[0].Attributes["when"])
}

// TestJSONProjectBuilder_Deadline tests setting project deadline
func TestJSONProjectBuilder_Deadline(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").Deadline("2025-12-31")
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeProject, items[0].Type)
	assert.Equal(t, "Test", items[0].Attributes["title"])
	assert.Equal(t, "2025-12-31", items[0].Attributes["deadline"])
}

// TestJSONProjectBuilder_Tags tests setting project tags
func TestJSONProjectBuilder_Tags(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").Tags("priority", "q1")
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeProject, items[0].Type)
	assert.Equal(t, "Test", items[0].Attributes["title"])
	tags := items[0].Attributes["tags"].([]any)
	assert.Equal(t, []any{"priority", "q1"}, tags)
}

// TestJSONProjectBuilder_AreaID tests placing project in area by UUID
func TestJSONProjectBuilder_AreaID(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").AreaID("area-uuid-456")
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeProject, items[0].Type)
	assert.Equal(t, "Test", items[0].Attributes["title"])
	assert.Equal(t, "area-uuid-456", items[0].Attributes["area-id"])
}

// TestJSONProjectBuilder_Completed tests marking project completed
func TestJSONProjectBuilder_Completed(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").Completed(true)
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeProject, items[0].Type)
	assert.Equal(t, "Test", items[0].Attributes["title"])
	assert.Equal(t, true, items[0].Attributes["completed"])
}

// TestJSONProjectBuilder_Canceled tests marking project canceled
func TestJSONProjectBuilder_Canceled(t *testing.T) {
	scheme := NewScheme()
	urlStr, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").Canceled(true)
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeProject, items[0].Type)
	assert.Equal(t, "Test", items[0].Attributes["title"])
	assert.Equal(t, true, items[0].Attributes["canceled"])
}

// TestJSONProjectBuilder_CreationDate tests backdating project creation
func TestJSONProjectBuilder_CreationDate(t *testing.T) {
	scheme := NewScheme()
	pastDate := time.Date(2024, time.January, 1, 9, 0, 0, 0, time.UTC)
	urlStr, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").CreationDate(pastDate)
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeProject, items[0].Type)
	assert.Equal(t, "Test", items[0].Attributes["title"])
	assertJSONDateAttr(t, items[0].Attributes, "creation-date", 2024, time.January, 1)
}

// TestJSONProjectBuilder_CompletionDate tests setting completion timestamp
func TestJSONProjectBuilder_CompletionDate(t *testing.T) {
	scheme := NewScheme()
	completedDate := time.Date(2024, time.November, 30, 17, 0, 0, 0, time.UTC)
	urlStr, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").Completed(true).CompletionDate(completedDate)
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeProject, items[0].Type)
	assert.Equal(t, "Test", items[0].Attributes["title"])
	assert.Equal(t, true, items[0].Attributes["completed"])
	assertJSONDateAttr(t, items[0].Attributes, "completion-date", 2024, time.November, 30)
}

// TestJSONProjectBuilder_UpdatePrependNotes tests prepending notes in update
func TestJSONProjectBuilder_UpdatePrependNotes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.JSON().
		UpdateProject("uuid", func(project *JSONProjectBuilder) {
			project.PrependNotes("Update: ")
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeProject, items[0].Type)
	assert.Equal(t, JSONOperationUpdate, items[0].Operation)
	assert.Equal(t, "uuid", items[0].ID)
	assert.Equal(t, "Update: ", items[0].Attributes["prepend-notes"])
}

// TestJSONProjectBuilder_UpdateAppendNotes tests appending notes in update
func TestJSONProjectBuilder_UpdateAppendNotes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.JSON().
		UpdateProject("uuid", func(project *JSONProjectBuilder) {
			project.AppendNotes(" - Reviewed")
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeProject, items[0].Type)
	assert.Equal(t, JSONOperationUpdate, items[0].Operation)
	assert.Equal(t, "uuid", items[0].ID)
	assert.Equal(t, " - Reviewed", items[0].Attributes["append-notes"])
}

// TestJSONProjectBuilder_UpdateAddTags tests adding tags without replacing
func TestJSONProjectBuilder_UpdateAddTags(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.JSON().
		UpdateProject("uuid", func(project *JSONProjectBuilder) {
			project.AddTags("reviewed", "approved")
		}).
		Build()
	require.NoError(t, err)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeProject, items[0].Type)
	assert.Equal(t, JSONOperationUpdate, items[0].Operation)
	addTags := items[0].Attributes["add-tags"].([]any)
	assert.Equal(t, []any{"reviewed", "approved"}, addTags)
}

// TestJSONProjectBuilder_TodosWithError tests error propagation from child todos
func TestJSONProjectBuilder_TodosWithError(t *testing.T) {
	scheme := NewScheme()
	longTitle := strings.Repeat("a", 4001)
	_, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").Todos(
				NewTodo().Title(longTitle),
			)
		}).
		Build()
	assert.ErrorIs(t, err, ErrTitleTooLong)
}

// TestAuthJSONBuilder_AddProject tests creating project via auth builder
func TestAuthJSONBuilder_AddProject(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("New Project").Area("Work")
		}).
		Build()
	require.NoError(t, err)

	cmd, _ := parseThingsURL(t, urlStr)
	assert.Equal(t, "json", cmd)

	items := parseJSONItems(t, urlStr)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeProject, items[0].Type)
	assert.Equal(t, "New Project", items[0].Attributes["title"])
	assert.Equal(t, "Work", items[0].Attributes["area"])
}

// TestAuthJSONBuilder_Reveal tests reveal option
func TestAuthJSONBuilder_Reveal(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test")
		}).
		Reveal(true).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "json", cmd)
	assertURLParam(t, params, "reveal", "true")
}

// TestAuthJSONBuilder_CreateOnly tests create-only operations don't need auth token
func TestAuthJSONBuilder_CreateOnly(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	urlStr, err := auth.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test")
		}).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, urlStr)
	assert.Equal(t, "json", cmd)
	// Create-only operations don't include auth-token in URL
	assertURLParamNotExists(t, params, "auth-token")
}

// =============================================================================
// Command and Type String Tests
// =============================================================================

func TestCommand_String(t *testing.T) {
	assert.Equal(t, "show", CommandShow.String())
	assert.Equal(t, "add", CommandAdd.String())
	assert.Equal(t, "add-project", CommandAddProject.String())
	assert.Equal(t, "update", CommandUpdate.String())
	assert.Equal(t, "update-project", CommandUpdateProject.String())
	assert.Equal(t, "search", CommandSearch.String())
	assert.Equal(t, "version", CommandVersion.String())
	assert.Equal(t, "json", CommandJSON.String())
}

func TestWhen_String(t *testing.T) {
	assert.Equal(t, "today", WhenToday.String())
	assert.Equal(t, "tomorrow", WhenTomorrow.String())
	assert.Equal(t, "evening", WhenEvening.String())
	assert.Equal(t, "anytime", WhenAnytime.String())
	assert.Equal(t, "someday", WhenSomeday.String())
}

func TestListID_String(t *testing.T) {
	assert.Equal(t, "inbox", ListInbox.String())
	assert.Equal(t, "today", ListToday.String())
	assert.Equal(t, "anytime", ListAnytime.String())
	assert.Equal(t, "upcoming", ListUpcoming.String())
	assert.Equal(t, "someday", ListSomeday.String())
	assert.Equal(t, "logbook", ListLogbook.String())
	assert.Equal(t, "tomorrow", ListTomorrow.String())
	assert.Equal(t, "deadlines", ListDeadlines.String())
	assert.Equal(t, "repeating", ListRepeating.String())
	assert.Equal(t, "all-projects", ListAllProjects.String())
	assert.Equal(t, "logged-projects", ListLoggedProjects.String())
}
