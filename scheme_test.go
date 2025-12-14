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
	thingsURL, err := scheme.Todo().Title("Buy groceries").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Buy groceries", params.Get("title"))
}

func TestTodoBuilder_TitleTooLong(t *testing.T) {
	scheme := NewScheme()
	longTitle := strings.Repeat("a", 4001)
	_, err := scheme.Todo().Title(longTitle).Build()
	assert.ErrorIs(t, err, ErrTitleTooLong)
}

func TestTodoBuilder_Notes(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Todo().Title("Test").Notes("Some notes").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "Some notes", params.Get("notes"))
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
			thingsURL, err := scheme.Todo().Title("Test").When(tt.when).Build()
			require.NoError(t, err)

			cmd, params := parseThingsURL(t, thingsURL)
			require.Equal(t, "add", cmd)
			require.Equal(t, "Test", params.Get("title"))
			require.Equal(t, tt.expected, params.Get("when"))
		})
	}
}

func TestTodoBuilder_WhenDate(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Todo().Title("Test").WhenDate(2025, time.December, 25).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "2025-12-25", params.Get("when"))
}

func TestTodoBuilder_Deadline(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Todo().Title("Test").Deadline("2025-12-31").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "2025-12-31", params.Get("deadline"))
}

func TestTodoBuilder_Tags(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Todo().Title("Test").Tags("work", "urgent").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "work,urgent", params.Get("tags"))
}

func TestTodoBuilder_ChecklistItems(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Todo().Title("Test").ChecklistItems("Item 1", "Item 2").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "Item 1\nItem 2", params.Get("checklist-items"))
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
	thingsURL, err := scheme.Todo().Title("Test").List("My Project").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "My Project", params.Get("list"))
}

func TestTodoBuilder_ListID(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Todo().Title("Test").ListID("uuid-123").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "uuid-123", params.Get("list-id"))
}

func TestTodoBuilder_Completed(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Todo().Title("Test").Completed(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "true", params.Get("completed"))
}

func TestTodoBuilder_Canceled(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Todo().Title("Test").Canceled(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "true", params.Get("canceled"))
}

func TestTodoBuilder_ShowQuickEntry(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Todo().Title("Test").ShowQuickEntry(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "true", params.Get("show-quick-entry"))
}

func TestTodoBuilder_Reveal(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Todo().Title("Test").Reveal(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "true", params.Get("reveal"))
}

func TestTodoBuilder_Titles(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Todo().Titles("Task 1", "Task 2", "Task 3").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Task 1\nTask 2\nTask 3", params.Get("titles"))
}

func TestTodoBuilder_TitlesTooLong(t *testing.T) {
	scheme := NewScheme()
	longTitle := strings.Repeat("a", 2000)
	_, err := scheme.Todo().Titles(longTitle, longTitle, longTitle).Build()
	assert.ErrorIs(t, err, ErrTitleTooLong)
}

func TestTodoBuilder_Heading(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Todo().
		Title("Subtask").
		List("My Project").
		Heading("Phase 1").
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Subtask", params.Get("title"))
	require.Equal(t, "My Project", params.Get("list"))
	require.Equal(t, "Phase 1", params.Get("heading"))
}

func TestTodoBuilder_HeadingID(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Todo().
		Title("Subtask").
		ListID("project-uuid").
		HeadingID("heading-uuid").
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Subtask", params.Get("title"))
	require.Equal(t, "project-uuid", params.Get("list-id"))
	require.Equal(t, "heading-uuid", params.Get("heading-id"))
}

func TestTodoBuilder_CreationDate(t *testing.T) {
	scheme := NewScheme()
	pastDate := time.Date(2024, time.January, 15, 10, 30, 0, 0, time.UTC)
	thingsURL, err := scheme.Todo().
		Title("Historical task").
		CreationDate(pastDate).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Historical task", params.Get("title"))
	assertDateParam(t, params, "creation-date", 2024, time.January, 15)
}

func TestTodoBuilder_CompletionDate(t *testing.T) {
	scheme := NewScheme()
	completedAt := time.Date(2024, time.December, 1, 14, 0, 0, 0, time.UTC)
	thingsURL, err := scheme.Todo().
		Title("Imported completed task").
		Completed(true).
		CompletionDate(completedAt).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Imported completed task", params.Get("title"))
	require.Equal(t, "true", params.Get("completed"))
	assertDateParam(t, params, "completion-date", 2024, time.December, 1)
}

func TestTodoBuilder_Chained(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Todo().
		Title("Buy groceries").
		Notes("Don't forget milk").
		When(WhenToday).
		Tags("shopping").
		Reveal(true).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Buy groceries", params.Get("title"))
	require.Equal(t, "Don't forget milk", params.Get("notes"))
	require.Equal(t, "today", params.Get("when"))
	require.Equal(t, "shopping", params.Get("tags"))
	require.Equal(t, "true", params.Get("reveal"))
}

// =============================================================================
// ProjectBuilder Tests
// =============================================================================

func TestProjectBuilder_Title(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Project().Title("New Project").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "New Project", params.Get("title"))
}

func TestProjectBuilder_TitleTooLong(t *testing.T) {
	scheme := NewScheme()
	longTitle := strings.Repeat("a", 4001)
	_, err := scheme.Project().Title(longTitle).Build()
	assert.ErrorIs(t, err, ErrTitleTooLong)
}

func TestProjectBuilder_Area(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Project().Title("Test").Area("Work").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "Work", params.Get("area"))
}

func TestProjectBuilder_AreaID(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Project().Title("Test").AreaID("uuid-123").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "uuid-123", params.Get("area-id"))
}

func TestProjectBuilder_Todos(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Project().Title("Test").Todos("Task 1", "Task 2").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "Task 1\nTask 2", params.Get("to-dos"))
}

func TestProjectBuilder_Notes(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Project().
		Title("Q1 Goals").
		Notes("Quarterly objectives and key results").
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "Q1 Goals", params.Get("title"))
	require.Equal(t, "Quarterly objectives and key results", params.Get("notes"))
}

func TestProjectBuilder_When(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Project().Title("Future Project").When(WhenSomeday).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "Future Project", params.Get("title"))
	require.Equal(t, "someday", params.Get("when"))
}

func TestProjectBuilder_WhenDate(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Project().
		Title("Launch").
		WhenDate(2025, time.March, 1).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "Launch", params.Get("title"))
	require.Equal(t, "2025-03-01", params.Get("when"))
}

func TestProjectBuilder_Deadline(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Project().
		Title("Release v2.0").
		Deadline("2025-06-30").
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "Release v2.0", params.Get("title"))
	require.Equal(t, "2025-06-30", params.Get("deadline"))
}

func TestProjectBuilder_Tags(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Project().
		Title("Website Redesign").
		Tags("work", "high-priority").
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "Website Redesign", params.Get("title"))
	require.Equal(t, "work,high-priority", params.Get("tags"))
}

func TestProjectBuilder_Completed(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Project().
		Title("Archived Project").
		Completed(true).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "Archived Project", params.Get("title"))
	require.Equal(t, "true", params.Get("completed"))
}

func TestProjectBuilder_Canceled(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Project().
		Title("Discontinued Project").
		Canceled(true).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "Discontinued Project", params.Get("title"))
	require.Equal(t, "true", params.Get("canceled"))
}

func TestProjectBuilder_Reveal(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Project().
		Title("New Project").
		Reveal(true).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "New Project", params.Get("title"))
	require.Equal(t, "true", params.Get("reveal"))
}

func TestProjectBuilder_CreationDate(t *testing.T) {
	scheme := NewScheme()
	created := time.Date(2024, time.June, 1, 9, 0, 0, 0, time.UTC)
	thingsURL, err := scheme.Project().
		Title("Historical Project").
		CreationDate(created).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "Historical Project", params.Get("title"))
	assertDateParam(t, params, "creation-date", 2024, time.June, 1)
}

func TestProjectBuilder_CompletionDate(t *testing.T) {
	scheme := NewScheme()
	completed := time.Date(2024, time.November, 15, 17, 0, 0, 0, time.UTC)
	thingsURL, err := scheme.Project().
		Title("Imported Completed Project").
		Completed(true).
		CompletionDate(completed).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "Imported Completed Project", params.Get("title"))
	require.Equal(t, "true", params.Get("completed"))
	assertDateParam(t, params, "completion-date", 2024, time.November, 15)
}

func TestProjectBuilder_FullProject(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.Project().
		Title("Product Launch").
		Notes("Launch plan for v2.0").
		Area("Work").
		Tags("priority").
		Deadline("2025-03-31").
		Todos("Design", "Development", "Testing", "Release").
		Reveal(true).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "Product Launch", params.Get("title"))
	require.Equal(t, "Launch plan for v2.0", params.Get("notes"))
	require.Equal(t, "Work", params.Get("area"))
	require.Equal(t, "priority", params.Get("tags"))
	require.Equal(t, "2025-03-31", params.Get("deadline"))
	require.Equal(t, "Design\nDevelopment\nTesting\nRelease", params.Get("to-dos"))
	require.Equal(t, "true", params.Get("reveal"))
}

// =============================================================================
// ShowBuilder Tests
// =============================================================================

func TestShowBuilder_ID(t *testing.T) {
	scheme := NewScheme()
	thingsURL := scheme.Show().ID("uuid-123").Build()

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "show", cmd)
	require.Equal(t, "uuid-123", params.Get("id"))
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
			thingsURL := scheme.Show().List(tt.list).Build()

			cmd, params := parseThingsURL(t, thingsURL)
			require.Equal(t, "show", cmd)
			require.Equal(t, tt.expected, params.Get("id"))
		})
	}
}

func TestShowBuilder_Query(t *testing.T) {
	scheme := NewScheme()
	thingsURL := scheme.Show().Query("My Project").Build()

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "show", cmd)
	require.Equal(t, "My Project", params.Get("query"))
}

func TestShowBuilder_Filter(t *testing.T) {
	scheme := NewScheme()
	thingsURL := scheme.Show().List(ListToday).Filter("work", "urgent").Build()

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "show", cmd)
	require.Equal(t, "today", params.Get("id"))
	require.Equal(t, "work,urgent", params.Get("filter"))
}

func TestShowBuilder_NoParams(t *testing.T) {
	scheme := NewScheme()
	thingsURL := scheme.Show().Build()
	assert.Equal(t, "things:///show", thingsURL)
}

// =============================================================================
// Search and Version Tests
// =============================================================================

func TestScheme_Search(t *testing.T) {
	scheme := NewScheme()
	thingsURL := scheme.Search("my query")

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "search", cmd)
	require.Equal(t, "my query", params.Get("query"))
}

func TestScheme_Version(t *testing.T) {
	scheme := NewScheme()
	thingsURL := scheme.Version()
	assert.Equal(t, "things:///version", thingsURL)
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
	thingsURL, err := auth.UpdateTodo("uuid-123").Completed(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "uuid-123", params.Get("id"))
	require.Equal(t, "test-token", params.Get("auth-token"))
	require.Equal(t, "true", params.Get("completed"))
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
	thingsURL, err := auth.UpdateTodo("uuid").Title("New Title").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "uuid", params.Get("id"))
	require.Equal(t, "test-token", params.Get("auth-token"))
	require.Equal(t, "New Title", params.Get("title"))
}

func TestUpdateTodoBuilder_PrependNotes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").PrependNotes("Prefix: ").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "Prefix: ", params.Get("prepend-notes"))
}

func TestUpdateTodoBuilder_AppendNotes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").AppendNotes(" - Suffix").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, " - Suffix", params.Get("append-notes"))
}

func TestUpdateTodoBuilder_AddTags(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").AddTags("new-tag").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "new-tag", params.Get("add-tags"))
}

func TestUpdateTodoBuilder_ClearDeadline(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").ClearDeadline().Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Empty(t, params.Get("deadline"))
}

func TestUpdateTodoBuilder_Duplicate(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").Duplicate(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "true", params.Get("duplicate"))
}

func TestUpdateTodoBuilder_Notes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").Notes("New description").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "New description", params.Get("notes"))
}

func TestUpdateTodoBuilder_When(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").When(WhenTomorrow).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "tomorrow", params.Get("when"))
}

func TestUpdateTodoBuilder_WhenDate(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").WhenDate(2025, time.February, 14).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "2025-02-14", params.Get("when"))
}

func TestUpdateTodoBuilder_Deadline(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").Deadline("2025-01-31").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "2025-01-31", params.Get("deadline"))
}

func TestUpdateTodoBuilder_Tags(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").Tags("new-tag-1", "new-tag-2").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "new-tag-1,new-tag-2", params.Get("tags"))
}

func TestUpdateTodoBuilder_ChecklistItems(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").ChecklistItems("Step A", "Step B").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "Step A\nStep B", params.Get("checklist-items"))
}

func TestUpdateTodoBuilder_PrependChecklistItems(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").PrependChecklistItems("First step").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "First step", params.Get("prepend-checklist-items"))
}

func TestUpdateTodoBuilder_AppendChecklistItems(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").AppendChecklistItems("Final step").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "Final step", params.Get("append-checklist-items"))
}

func TestUpdateTodoBuilder_List(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").List("New Project").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "New Project", params.Get("list"))
}

func TestUpdateTodoBuilder_ListID(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").ListID("project-uuid").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "project-uuid", params.Get("list-id"))
}

func TestUpdateTodoBuilder_Heading(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").Heading("Phase 2").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "Phase 2", params.Get("heading"))
}

func TestUpdateTodoBuilder_HeadingID(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").HeadingID("heading-uuid").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "heading-uuid", params.Get("heading-id"))
}

func TestUpdateTodoBuilder_Canceled(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").Canceled(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "true", params.Get("canceled"))
}

func TestUpdateTodoBuilder_Reveal(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").Reveal(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "true", params.Get("reveal"))
}

func TestUpdateTodoBuilder_CreationDate(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	created := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	thingsURL, err := auth.UpdateTodo("uuid").CreationDate(created).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	assertDateParam(t, params, "creation-date", 2024, time.January, 1)
}

func TestUpdateTodoBuilder_CompletionDate(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	completed := time.Date(2024, time.December, 31, 23, 59, 0, 0, time.UTC)
	thingsURL, err := auth.UpdateTodo("uuid").Completed(true).CompletionDate(completed).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "true", params.Get("completed"))
	assertDateParam(t, params, "completion-date", 2024, time.December, 31)
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
	thingsURL, err := auth.UpdateProject("uuid").Title("New Project Title").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "uuid", params.Get("id"))
	require.Equal(t, "test-token", params.Get("auth-token"))
	require.Equal(t, "New Project Title", params.Get("title"))
}

func TestUpdateProjectBuilder_Completed(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").Completed(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "true", params.Get("completed"))
}

func TestUpdateProjectBuilder_Canceled(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").Canceled(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "true", params.Get("canceled"))
}

func TestUpdateProjectBuilder_Notes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").Notes("Updated project description").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "Updated project description", params.Get("notes"))
}

func TestUpdateProjectBuilder_PrependNotes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").PrependNotes("[UPDATE] ").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "[UPDATE] ", params.Get("prepend-notes"))
}

func TestUpdateProjectBuilder_AppendNotes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").AppendNotes("\n- Added new requirement").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "\n- Added new requirement", params.Get("append-notes"))
}

func TestUpdateProjectBuilder_When(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").When(WhenAnytime).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "anytime", params.Get("when"))
}

func TestUpdateProjectBuilder_WhenDate(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").WhenDate(2025, time.April, 1).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "2025-04-01", params.Get("when"))
}

func TestUpdateProjectBuilder_Deadline(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").Deadline("2025-12-31").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "2025-12-31", params.Get("deadline"))
}

func TestUpdateProjectBuilder_ClearDeadline(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").ClearDeadline().Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Empty(t, params.Get("deadline"))
}

func TestUpdateProjectBuilder_Tags(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").Tags("priority", "q1").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "priority,q1", params.Get("tags"))
}

func TestUpdateProjectBuilder_AddTags(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").AddTags("reviewed").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "reviewed", params.Get("add-tags"))
}

func TestUpdateProjectBuilder_Area(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").Area("Personal").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "Personal", params.Get("area"))
}

func TestUpdateProjectBuilder_AreaID(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").AreaID("area-uuid").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "area-uuid", params.Get("area-id"))
}

func TestUpdateProjectBuilder_Reveal(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").Reveal(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "true", params.Get("reveal"))
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
	thingsURL, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test Todo")
		}).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "json", cmd)
	require.True(t, params.Has("data"))
	require.False(t, params.Has("reveal"))

	items := parseJSONItems(t, thingsURL)
	require.Len(t, items, 1)
	assert.Equal(t, JSONItemTypeTodo, items[0].Type)
	assert.Equal(t, "Test Todo", items[0].Attributes["title"])
}

func TestJSONBuilder_AddProject(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test Project")
		}).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "json", cmd)
	require.True(t, params.Has("data"))

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeProject,
		Attributes: map[string]any{"title": "Test Project"},
	}}, parseJSONItems(t, thingsURL))
}

func TestJSONBuilder_Reveal(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test")
		}).
		Reveal(true).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "json", cmd)
	require.Equal(t, "true", params.Get("reveal"))
	require.True(t, params.Has("data"))
}

func TestJSONBuilder_NoItems(t *testing.T) {
	scheme := NewScheme()
	_, err := scheme.JSON().Build()
	assert.ErrorIs(t, err, ErrNoJSONItems)
}

func TestJSONBuilder_Multiple(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Todo 1")
		}).
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Todo 2")
		}).
		Build()
	require.NoError(t, err)

	cmd, _ := parseThingsURL(t, thingsURL)
	assert.Equal(t, "json", cmd)

	require.Equal(t, []JSONItem{
		{Type: JSONItemTypeTodo, Attributes: map[string]any{"title": "Todo 1"}},
		{Type: JSONItemTypeTodo, Attributes: map[string]any{"title": "Todo 2"}},
	}, parseJSONItems(t, thingsURL))
}

// =============================================================================
// AuthJSONBuilder Tests
// =============================================================================

func TestAuthJSONBuilder_UpdateTodo(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.JSON().
		UpdateTodo("uuid-123", func(todo *JSONTodoBuilder) {
			todo.Completed(true)
		}).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "json", cmd)
	require.Equal(t, "test-token", params.Get("auth-token"))
	require.True(t, params.Has("data"))

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeTodo,
		Operation:  JSONOperationUpdate,
		ID:         "uuid-123",
		Attributes: map[string]any{"completed": true},
	}}, parseJSONItems(t, thingsURL))
}

func TestAuthJSONBuilder_UpdateProject(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.JSON().
		UpdateProject("uuid-123", func(project *JSONProjectBuilder) {
			project.Completed(true)
		}).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "json", cmd)
	require.Equal(t, "test-token", params.Get("auth-token"))
	require.True(t, params.Has("data"))

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeProject,
		Operation:  JSONOperationUpdate,
		ID:         "uuid-123",
		Attributes: map[string]any{"completed": true},
	}}, parseJSONItems(t, thingsURL))
}

func TestAuthJSONBuilder_Mixed(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("New Todo")
		}).
		UpdateTodo("uuid-123", func(todo *JSONTodoBuilder) {
			todo.Completed(true)
		}).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "json", cmd)
	require.Equal(t, "test-token", params.Get("auth-token"))

	require.Equal(t, []JSONItem{
		{Type: JSONItemTypeTodo, Attributes: map[string]any{"title": "New Todo"}},
		{Type: JSONItemTypeTodo, Operation: JSONOperationUpdate, ID: "uuid-123", Attributes: map[string]any{"completed": true}},
	}, parseJSONItems(t, thingsURL))
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
	thingsURL, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").When(WhenToday)
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeTodo,
		Attributes: map[string]any{"title": "Test", "when": "today"},
	}}, parseJSONItems(t, thingsURL))
}

func TestJSONTodoBuilder_Tags(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").Tags("Risk", "Golang")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeTodo,
		Attributes: map[string]any{"title": "Test", "tags": []any{"Risk", "Golang"}},
	}}, parseJSONItems(t, thingsURL))
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
	thingsURL, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").Notes("Detailed description")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeTodo,
		Attributes: map[string]any{"title": "Test", "notes": "Detailed description"},
	}}, parseJSONItems(t, thingsURL))
}

// TestJSONTodoBuilder_WhenDate tests scheduling to a specific date
func TestJSONTodoBuilder_WhenDate(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").WhenDate(2025, time.March, 15)
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeTodo,
		Attributes: map[string]any{"title": "Test", "when": "2025-03-15"},
	}}, parseJSONItems(t, thingsURL))
}

// TestJSONTodoBuilder_Deadline tests setting a deadline
func TestJSONTodoBuilder_Deadline(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").Deadline("2025-06-30")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeTodo,
		Attributes: map[string]any{"title": "Test", "deadline": "2025-06-30"},
	}}, parseJSONItems(t, thingsURL))
}

// TestJSONTodoBuilder_ChecklistItems tests adding a checklist
func TestJSONTodoBuilder_ChecklistItems(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").ChecklistItems("Step 1", "Step 2", "Step 3")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type: JSONItemTypeTodo,
		Attributes: map[string]any{
			"title": "Test",
			"checklist-items": []any{
				map[string]any{"type": "checklist-item", "attributes": map[string]any{"title": "Step 1"}},
				map[string]any{"type": "checklist-item", "attributes": map[string]any{"title": "Step 2"}},
				map[string]any{"type": "checklist-item", "attributes": map[string]any{"title": "Step 3"}},
			},
		},
	}}, parseJSONItems(t, thingsURL))
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
	thingsURL, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").List("My Project")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeTodo,
		Attributes: map[string]any{"title": "Test", "list": "My Project"},
	}}, parseJSONItems(t, thingsURL))
}

// TestJSONTodoBuilder_ListID tests placing todo in a project by UUID
func TestJSONTodoBuilder_ListID(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").ListID("project-uuid-123")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeTodo,
		Attributes: map[string]any{"title": "Test", "list-id": "project-uuid-123"},
	}}, parseJSONItems(t, thingsURL))
}

// TestJSONTodoBuilder_Heading tests placing todo under a heading
func TestJSONTodoBuilder_Heading(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").List("Project").Heading("Phase 1")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeTodo,
		Attributes: map[string]any{"title": "Test", "list": "Project", "heading": "Phase 1"},
	}}, parseJSONItems(t, thingsURL))
}

// TestJSONTodoBuilder_Completed tests marking as completed
func TestJSONTodoBuilder_Completed(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").Completed(true)
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeTodo,
		Attributes: map[string]any{"title": "Test", "completed": true},
	}}, parseJSONItems(t, thingsURL))
}

// TestJSONTodoBuilder_Canceled tests marking as canceled
func TestJSONTodoBuilder_Canceled(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").Canceled(true)
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeTodo,
		Attributes: map[string]any{"title": "Test", "canceled": true},
	}}, parseJSONItems(t, thingsURL))
}

// TestJSONTodoBuilder_CreationDate tests backdating creation
func TestJSONTodoBuilder_CreationDate(t *testing.T) {
	scheme := NewScheme()
	pastDate := time.Date(2024, time.June, 1, 10, 0, 0, 0, time.UTC)
	thingsURL, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").CreationDate(pastDate)
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type: JSONItemTypeTodo,
		Attributes: map[string]any{
			"title":         "Test",
			"creation-date": pastDate.Format(time.RFC3339),
		},
	}}, parseJSONItems(t, thingsURL))
}

// TestJSONTodoBuilder_CompletionDate tests setting completion timestamp
func TestJSONTodoBuilder_CompletionDate(t *testing.T) {
	scheme := NewScheme()
	completedDate := time.Date(2024, time.December, 15, 14, 30, 0, 0, time.UTC)
	thingsURL, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").Completed(true).CompletionDate(completedDate)
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type: JSONItemTypeTodo,
		Attributes: map[string]any{
			"title":           "Test",
			"completed":       true,
			"completion-date": completedDate.Format(time.RFC3339),
		},
	}}, parseJSONItems(t, thingsURL))
}

// TestJSONTodoBuilder_UpdatePrependNotes tests prepending notes in update
func TestJSONTodoBuilder_UpdatePrependNotes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.JSON().
		UpdateTodo("uuid", func(todo *JSONTodoBuilder) {
			todo.PrependNotes("Important: ")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeTodo,
		Operation:  JSONOperationUpdate,
		ID:         "uuid",
		Attributes: map[string]any{"prepend-notes": "Important: "},
	}}, parseJSONItems(t, thingsURL))
}

// TestJSONTodoBuilder_UpdateAppendNotes tests appending notes in update
func TestJSONTodoBuilder_UpdateAppendNotes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.JSON().
		UpdateTodo("uuid", func(todo *JSONTodoBuilder) {
			todo.AppendNotes(" - Updated")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeTodo,
		Operation:  JSONOperationUpdate,
		ID:         "uuid",
		Attributes: map[string]any{"append-notes": " - Updated"},
	}}, parseJSONItems(t, thingsURL))
}

// TestJSONTodoBuilder_UpdateAddTags tests adding tags without replacing
func TestJSONTodoBuilder_UpdateAddTags(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.JSON().
		UpdateTodo("uuid", func(todo *JSONTodoBuilder) {
			todo.AddTags("new-tag", "another-tag")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeTodo,
		Operation:  JSONOperationUpdate,
		ID:         "uuid",
		Attributes: map[string]any{"add-tags": []any{"new-tag", "another-tag"}},
	}}, parseJSONItems(t, thingsURL))
}

// =============================================================================
// JSONProjectBuilder Tests
// =============================================================================

func TestJSONProjectBuilder_Area(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").Area("Work")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeProject,
		Attributes: map[string]any{"title": "Test", "area": "Work"},
	}}, parseJSONItems(t, thingsURL))
}

func TestJSONProjectBuilder_Todos(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test Project").Todos(
				NewTodo().Title("Task 1"),
				NewTodo().Title("Task 2"),
			)
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type: JSONItemTypeProject,
		Attributes: map[string]any{
			"title": "Test Project",
			"items": []any{
				map[string]any{"type": "to-do", "attributes": map[string]any{"title": "Task 1"}},
				map[string]any{"type": "to-do", "attributes": map[string]any{"title": "Task 2"}},
			},
		},
	}}, parseJSONItems(t, thingsURL))
}

// TestJSONProjectBuilder_Notes tests adding project notes
func TestJSONProjectBuilder_Notes(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").Notes("Project description")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeProject,
		Attributes: map[string]any{"title": "Test", "notes": "Project description"},
	}}, parseJSONItems(t, thingsURL))
}

// TestJSONProjectBuilder_When tests scheduling project
func TestJSONProjectBuilder_When(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").When(WhenSomeday)
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeProject,
		Attributes: map[string]any{"title": "Test", "when": "someday"},
	}}, parseJSONItems(t, thingsURL))
}

// TestJSONProjectBuilder_WhenDate tests scheduling to specific date
func TestJSONProjectBuilder_WhenDate(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").WhenDate(2025, time.July, 1)
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeProject,
		Attributes: map[string]any{"title": "Test", "when": "2025-07-01"},
	}}, parseJSONItems(t, thingsURL))
}

// TestJSONProjectBuilder_Deadline tests setting project deadline
func TestJSONProjectBuilder_Deadline(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").Deadline("2025-12-31")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeProject,
		Attributes: map[string]any{"title": "Test", "deadline": "2025-12-31"},
	}}, parseJSONItems(t, thingsURL))
}

// TestJSONProjectBuilder_Tags tests setting project tags
func TestJSONProjectBuilder_Tags(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").Tags("priority", "q1")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeProject,
		Attributes: map[string]any{"title": "Test", "tags": []any{"priority", "q1"}},
	}}, parseJSONItems(t, thingsURL))
}

// TestJSONProjectBuilder_AreaID tests placing project in area by UUID
func TestJSONProjectBuilder_AreaID(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").AreaID("area-uuid-456")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeProject,
		Attributes: map[string]any{"title": "Test", "area-id": "area-uuid-456"},
	}}, parseJSONItems(t, thingsURL))
}

// TestJSONProjectBuilder_Completed tests marking project completed
func TestJSONProjectBuilder_Completed(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").Completed(true)
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeProject,
		Attributes: map[string]any{"title": "Test", "completed": true},
	}}, parseJSONItems(t, thingsURL))
}

// TestJSONProjectBuilder_Canceled tests marking project canceled
func TestJSONProjectBuilder_Canceled(t *testing.T) {
	scheme := NewScheme()
	thingsURL, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").Canceled(true)
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeProject,
		Attributes: map[string]any{"title": "Test", "canceled": true},
	}}, parseJSONItems(t, thingsURL))
}

// TestJSONProjectBuilder_CreationDate tests backdating project creation
func TestJSONProjectBuilder_CreationDate(t *testing.T) {
	scheme := NewScheme()
	pastDate := time.Date(2024, time.January, 1, 9, 0, 0, 0, time.UTC)
	thingsURL, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").CreationDate(pastDate)
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type: JSONItemTypeProject,
		Attributes: map[string]any{
			"title":         "Test",
			"creation-date": pastDate.Format(time.RFC3339),
		},
	}}, parseJSONItems(t, thingsURL))
}

// TestJSONProjectBuilder_CompletionDate tests setting completion timestamp
func TestJSONProjectBuilder_CompletionDate(t *testing.T) {
	scheme := NewScheme()
	completedDate := time.Date(2024, time.November, 30, 17, 0, 0, 0, time.UTC)
	thingsURL, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").Completed(true).CompletionDate(completedDate)
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type: JSONItemTypeProject,
		Attributes: map[string]any{
			"title":           "Test",
			"completed":       true,
			"completion-date": completedDate.Format(time.RFC3339),
		},
	}}, parseJSONItems(t, thingsURL))
}

// TestJSONProjectBuilder_UpdatePrependNotes tests prepending notes in update
func TestJSONProjectBuilder_UpdatePrependNotes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.JSON().
		UpdateProject("uuid", func(project *JSONProjectBuilder) {
			project.PrependNotes("Update: ")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeProject,
		Operation:  JSONOperationUpdate,
		ID:         "uuid",
		Attributes: map[string]any{"prepend-notes": "Update: "},
	}}, parseJSONItems(t, thingsURL))
}

// TestJSONProjectBuilder_UpdateAppendNotes tests appending notes in update
func TestJSONProjectBuilder_UpdateAppendNotes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.JSON().
		UpdateProject("uuid", func(project *JSONProjectBuilder) {
			project.AppendNotes(" - Reviewed")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeProject,
		Operation:  JSONOperationUpdate,
		ID:         "uuid",
		Attributes: map[string]any{"append-notes": " - Reviewed"},
	}}, parseJSONItems(t, thingsURL))
}

// TestJSONProjectBuilder_UpdateAddTags tests adding tags without replacing
func TestJSONProjectBuilder_UpdateAddTags(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.JSON().
		UpdateProject("uuid", func(project *JSONProjectBuilder) {
			project.AddTags("reviewed", "approved")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeProject,
		Operation:  JSONOperationUpdate,
		ID:         "uuid",
		Attributes: map[string]any{"add-tags": []any{"reviewed", "approved"}},
	}}, parseJSONItems(t, thingsURL))
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
	thingsURL, err := auth.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("New Project").Area("Work")
		}).
		Build()
	require.NoError(t, err)

	cmd, _ := parseThingsURL(t, thingsURL)
	require.Equal(t, "json", cmd)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeProject,
		Attributes: map[string]any{"title": "New Project", "area": "Work"},
	}}, parseJSONItems(t, thingsURL))
}

// TestAuthJSONBuilder_Reveal tests reveal option
func TestAuthJSONBuilder_Reveal(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test")
		}).
		Reveal(true).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "json", cmd)
	require.Equal(t, "true", params.Get("reveal"))
}

// TestAuthJSONBuilder_CreateOnly tests create-only operations don't need auth token
func TestAuthJSONBuilder_CreateOnly(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test")
		}).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "json", cmd)
	// Create-only operations don't include auth-token in URL
	require.False(t, params.Has("auth-token"))
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
