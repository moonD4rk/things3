package things3

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_newScheme(t *testing.T) {
	scheme := newScheme()
	assert.NotNil(t, scheme)
}

func TestSchemeWithForeground(t *testing.T) {
	scheme := newScheme(withForeground())
	assert.True(t, scheme.foreground, "withForeground() should set foreground to true")

	schemeDefault := newScheme()
	assert.False(t, schemeDefault.foreground, "Default scheme should have foreground false")
}

// =============================================================================
// AddTodoBuilder Tests
// =============================================================================

func TestAddTodoBuilder_Title(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddTodo().Title("Buy groceries").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Buy groceries", params.Get("title"))
}

func TestAddTodoBuilder_TitleTooLong(t *testing.T) {
	scheme := newScheme()
	longTitle := strings.Repeat("a", 4001)
	_, err := scheme.AddTodo().Title(longTitle).Build()
	assert.ErrorIs(t, err, ErrTitleTooLong)
}

func TestAddTodoBuilder_Notes(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddTodo().Title("Test").Notes("Some notes").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "Some notes", params.Get("notes"))
}

func TestAddTodoBuilder_NotesTooLong(t *testing.T) {
	scheme := newScheme()
	longNotes := strings.Repeat("a", 10001)
	_, err := scheme.AddTodo().Title("Test").Notes(longNotes).Build()
	assert.ErrorIs(t, err, ErrNotesTooLong)
}

func TestAddTodoBuilder_When(t *testing.T) {
	scheme := newScheme()

	// Test with time.Time
	testDate := time.Date(2025, 6, 15, 0, 0, 0, 0, time.Local)
	thingsURL, err := scheme.AddTodo().Title("Test").When(testDate).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	require.Equal(t, "add", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "2025-06-15", params.Get("when"))
}

func TestAddTodoBuilder_WhenEvening(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddTodo().Title("Test").WhenEvening().Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	require.Equal(t, "add", cmd)
	require.Equal(t, "evening", params.Get("when"))
}

func TestAddTodoBuilder_WhenAnytime(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddTodo().Title("Test").WhenAnytime().Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	require.Equal(t, "add", cmd)
	require.Equal(t, "anytime", params.Get("when"))
}

func TestAddTodoBuilder_WhenSomeday(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddTodo().Title("Test").WhenSomeday().Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	require.Equal(t, "add", cmd)
	require.Equal(t, "someday", params.Get("when"))
}

func TestAddTodoBuilder_Deadline(t *testing.T) {
	scheme := newScheme()
	deadline := time.Date(2025, 12, 31, 0, 0, 0, 0, time.Local)
	thingsURL, err := scheme.AddTodo().Title("Test").Deadline(deadline).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "2025-12-31", params.Get("deadline"))
}

func TestAddTodoBuilder_Tags(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddTodo().Title("Test").Tags("work", "urgent").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "work,urgent", params.Get("tags"))
}

func TestAddTodoBuilder_ChecklistItems(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddTodo().Title("Test").ChecklistItems("Item 1", "Item 2").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "Item 1\nItem 2", params.Get("checklist-items"))
}

func TestAddTodoBuilder_TooManyChecklistItems(t *testing.T) {
	scheme := newScheme()
	items := make([]string, 101)
	for i := range items {
		items[i] = "checklist entry"
	}
	_, err := scheme.AddTodo().Title("Test").ChecklistItems(items...).Build()
	assert.ErrorIs(t, err, ErrTooManyChecklistItems)
}

func TestAddTodoBuilder_List(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddTodo().Title("Test").List("My Project").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "My Project", params.Get("list"))
}

func TestAddTodoBuilder_ListID(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddTodo().Title("Test").ListID("uuid-123").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "uuid-123", params.Get("list-id"))
}

func TestAddTodoBuilder_Completed(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddTodo().Title("Test").Completed(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "true", params.Get("completed"))
}

func TestAddTodoBuilder_Canceled(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddTodo().Title("Test").Canceled(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "true", params.Get("canceled"))
}

func TestAddTodoBuilder_ShowQuickEntry(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddTodo().Title("Test").ShowQuickEntry(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "true", params.Get("show-quick-entry"))
}

func TestAddTodoBuilder_Reveal(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddTodo().Title("Test").Reveal(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "true", params.Get("reveal"))
}

func TestAddTodoBuilder_Titles(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddTodo().Titles("Task 1", "Task 2", "Task 3").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Task 1\nTask 2\nTask 3", params.Get("titles"))
}

func TestAddTodoBuilder_TitlesTooLong(t *testing.T) {
	scheme := newScheme()
	longTitle := strings.Repeat("a", 2000)
	_, err := scheme.AddTodo().Titles(longTitle, longTitle, longTitle).Build()
	assert.ErrorIs(t, err, ErrTitleTooLong)
}

func TestAddTodoBuilder_Heading(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddTodo().
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

func TestAddTodoBuilder_HeadingID(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddTodo().
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

func TestAddTodoBuilder_CreationDate(t *testing.T) {
	scheme := newScheme()
	pastDate := time.Date(2024, time.January, 15, 10, 30, 0, 0, time.UTC)
	thingsURL, err := scheme.AddTodo().
		Title("Historical task").
		CreationDate(pastDate).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Historical task", params.Get("title"))
	require.Equal(t, pastDate.Format(time.RFC3339), params.Get("creation-date"))
}

func TestAddTodoBuilder_CompletionDate(t *testing.T) {
	scheme := newScheme()
	completedAt := time.Date(2024, time.December, 1, 14, 0, 0, 0, time.UTC)
	thingsURL, err := scheme.AddTodo().
		Title("Imported completed task").
		Completed(true).
		CompletionDate(completedAt).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Imported completed task", params.Get("title"))
	require.Equal(t, "true", params.Get("completed"))
	require.Equal(t, completedAt.Format(time.RFC3339), params.Get("completion-date"))
}

func TestAddTodoBuilder_Chained(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddTodo().
		Title("Buy groceries").
		Notes("Don't forget milk").
		When(Today()).
		Tags("shopping").
		Reveal(true).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	require.Equal(t, "Buy groceries", params.Get("title"))
	require.Equal(t, "Don't forget milk", params.Get("notes"))
	// Today() returns today's date in yyyy-mm-dd format
	require.NotEmpty(t, params.Get("when"))
	require.Equal(t, "shopping", params.Get("tags"))
	require.Equal(t, "true", params.Get("reveal"))
}

// =============================================================================
// AddProjectBuilder Tests
// =============================================================================

func TestAddProjectBuilder_Title(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddProject().Title("New Project").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "New Project", params.Get("title"))
}

func TestAddProjectBuilder_TitleTooLong(t *testing.T) {
	scheme := newScheme()
	longTitle := strings.Repeat("a", 4001)
	_, err := scheme.AddProject().Title(longTitle).Build()
	assert.ErrorIs(t, err, ErrTitleTooLong)
}

func TestAddProjectBuilder_Area(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddProject().Title("Test").Area("Work").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "Work", params.Get("area"))
}

func TestAddProjectBuilder_AreaID(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddProject().Title("Test").AreaID("uuid-123").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "uuid-123", params.Get("area-id"))
}

func TestAddProjectBuilder_Todos(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddProject().Title("Test").Todos("Task 1", "Task 2").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "Test", params.Get("title"))
	require.Equal(t, "Task 1\nTask 2", params.Get("to-dos"))
}

func TestAddProjectBuilder_Notes(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddProject().
		Title("Q1 Goals").
		Notes("Quarterly objectives and key results").
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "Q1 Goals", params.Get("title"))
	require.Equal(t, "Quarterly objectives and key results", params.Get("notes"))
}

func TestAddProjectBuilder_WhenSomeday(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddProject().Title("Future Project").WhenSomeday().Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "Future Project", params.Get("title"))
	require.Equal(t, "someday", params.Get("when"))
}

func TestAddProjectBuilder_When(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddProject().
		Title("Launch").
		When(time.Date(2025, time.March, 1, 0, 0, 0, 0, time.Local)).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "Launch", params.Get("title"))
	require.Equal(t, "2025-03-01", params.Get("when"))
}

func TestAddProjectBuilder_Deadline(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddProject().
		Title("Release v2.0").
		Deadline(time.Date(2025, time.June, 30, 0, 0, 0, 0, time.Local)).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "Release v2.0", params.Get("title"))
	require.Equal(t, "2025-06-30", params.Get("deadline"))
}

func TestAddProjectBuilder_Tags(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddProject().
		Title("Website Redesign").
		Tags("work", "high-priority").
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "Website Redesign", params.Get("title"))
	require.Equal(t, "work,high-priority", params.Get("tags"))
}

func TestAddProjectBuilder_Completed(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddProject().
		Title("Archived Project").
		Completed(true).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "Archived Project", params.Get("title"))
	require.Equal(t, "true", params.Get("completed"))
}

func TestAddProjectBuilder_Canceled(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddProject().
		Title("Discontinued Project").
		Canceled(true).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "Discontinued Project", params.Get("title"))
	require.Equal(t, "true", params.Get("canceled"))
}

func TestAddProjectBuilder_Reveal(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddProject().
		Title("New Project").
		Reveal(true).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "New Project", params.Get("title"))
	require.Equal(t, "true", params.Get("reveal"))
}

func TestAddProjectBuilder_CreationDate(t *testing.T) {
	scheme := newScheme()
	created := time.Date(2024, time.June, 1, 9, 0, 0, 0, time.UTC)
	thingsURL, err := scheme.AddProject().
		Title("Historical Project").
		CreationDate(created).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "Historical Project", params.Get("title"))
	require.Equal(t, created.Format(time.RFC3339), params.Get("creation-date"))
}

func TestAddProjectBuilder_CompletionDate(t *testing.T) {
	scheme := newScheme()
	completed := time.Date(2024, time.November, 15, 17, 0, 0, 0, time.UTC)
	thingsURL, err := scheme.AddProject().
		Title("Imported Completed Project").
		Completed(true).
		CompletionDate(completed).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	require.Equal(t, "Imported Completed Project", params.Get("title"))
	require.Equal(t, "true", params.Get("completed"))
	require.Equal(t, completed.Format(time.RFC3339), params.Get("completion-date"))
}

func TestAddProjectBuilder_FullProject(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddProject().
		Title("Product Launch").
		Notes("Launch plan for v2.0").
		Area("Work").
		Tags("priority").
		Deadline(time.Date(2025, time.March, 31, 0, 0, 0, 0, time.Local)).
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
	scheme := newScheme()
	thingsURL := scheme.ShowBuilder().ID("uuid-123").Build()

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
			scheme := newScheme()
			thingsURL := scheme.ShowBuilder().List(tt.list).Build()

			cmd, params := parseThingsURL(t, thingsURL)
			require.Equal(t, "show", cmd)
			require.Equal(t, tt.expected, params.Get("id"))
		})
	}
}

func TestShowBuilder_Query(t *testing.T) {
	scheme := newScheme()
	thingsURL := scheme.ShowBuilder().Query("My Project").Build()

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "show", cmd)
	require.Equal(t, "My Project", params.Get("query"))
}

func TestShowBuilder_Filter(t *testing.T) {
	scheme := newScheme()
	thingsURL := scheme.ShowBuilder().List(ListToday).Filter("work", "urgent").Build()

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "show", cmd)
	require.Equal(t, "today", params.Get("id"))
	require.Equal(t, "work,urgent", params.Get("filter"))
}

func TestShowBuilder_NoParams(t *testing.T) {
	scheme := newScheme()
	thingsURL := scheme.ShowBuilder().Build()
	assert.Equal(t, "things:///show", thingsURL)
}

// =============================================================================
// Search and Version Tests
// =============================================================================

func TestScheme_Search(t *testing.T) {
	scheme := newScheme()
	thingsURL := scheme.SearchURL("my query")

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "search", cmd)
	require.Equal(t, "my query", params.Get("query"))
}

func TestScheme_Version(t *testing.T) {
	scheme := newScheme()
	thingsURL := scheme.Version()
	assert.Equal(t, "things:///version", thingsURL)
}

// =============================================================================
// AuthScheme Tests
// =============================================================================

func TestAuthScheme_WithToken(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	assert.NotNil(t, auth)
}

func TestAuthScheme_EmptyToken(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("")
	_, err := auth.UpdateTodo("uuid").Completed(true).Build()
	assert.ErrorIs(t, err, ErrEmptyToken)
}

// =============================================================================
// UpdateTodoBuilder Tests
// =============================================================================

func TestUpdateTodoBuilder_Completed(t *testing.T) {
	scheme := newScheme()
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
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	_, err := auth.UpdateTodo("").Completed(true).Build()
	assert.ErrorIs(t, err, ErrIDRequired)
}

func TestUpdateTodoBuilder_Title(t *testing.T) {
	scheme := newScheme()
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
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").PrependNotes("Prefix: ").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "Prefix: ", params.Get("prepend-notes"))
}

func TestUpdateTodoBuilder_AppendNotes(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").AppendNotes(" - Suffix").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, " - Suffix", params.Get("append-notes"))
}

func TestUpdateTodoBuilder_AddTags(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").AddTags("new-tag").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "new-tag", params.Get("add-tags"))
}

func TestUpdateTodoBuilder_ClearDeadline(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").ClearDeadline().Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Empty(t, params.Get("deadline"))
}

func TestUpdateTodoBuilder_Duplicate(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").Duplicate(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "true", params.Get("duplicate"))
}

func TestUpdateTodoBuilder_Notes(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").Notes("New description").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "New description", params.Get("notes"))
}

func TestUpdateTodoBuilder_When(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").When(time.Date(2025, time.January, 15, 0, 0, 0, 0, time.Local)).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "2025-01-15", params.Get("when"))
}

func TestUpdateTodoBuilder_WhenAnytime(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").WhenAnytime().Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "anytime", params.Get("when"))
}

func TestUpdateTodoBuilder_Deadline(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").Deadline(time.Date(2025, time.January, 31, 0, 0, 0, 0, time.Local)).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "2025-01-31", params.Get("deadline"))
}

func TestUpdateTodoBuilder_Tags(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").Tags("new-tag-1", "new-tag-2").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "new-tag-1,new-tag-2", params.Get("tags"))
}

func TestUpdateTodoBuilder_ChecklistItems(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").ChecklistItems("Step A", "Step B").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "Step A\nStep B", params.Get("checklist-items"))
}

func TestUpdateTodoBuilder_PrependChecklistItems(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").PrependChecklistItems("First step").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "First step", params.Get("prepend-checklist-items"))
}

func TestUpdateTodoBuilder_AppendChecklistItems(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").AppendChecklistItems("Final step").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "Final step", params.Get("append-checklist-items"))
}

func TestUpdateTodoBuilder_List(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").List("New Project").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "New Project", params.Get("list"))
}

func TestUpdateTodoBuilder_ListID(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").ListID("project-uuid").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "project-uuid", params.Get("list-id"))
}

func TestUpdateTodoBuilder_Heading(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").Heading("Phase 2").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "Phase 2", params.Get("heading"))
}

func TestUpdateTodoBuilder_HeadingID(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").HeadingID("heading-uuid").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "heading-uuid", params.Get("heading-id"))
}

func TestUpdateTodoBuilder_Canceled(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").Canceled(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "true", params.Get("canceled"))
}

func TestUpdateTodoBuilder_Reveal(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid").Reveal(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "true", params.Get("reveal"))
}

func TestUpdateTodoBuilder_CreationDate(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	created := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	thingsURL, err := auth.UpdateTodo("uuid").CreationDate(created).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, created.Format(time.RFC3339), params.Get("creation-date"))
}

func TestUpdateTodoBuilder_CompletionDate(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	completed := time.Date(2024, time.December, 31, 23, 59, 0, 0, time.UTC)
	thingsURL, err := auth.UpdateTodo("uuid").Completed(true).CompletionDate(completed).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	require.Equal(t, "true", params.Get("completed"))
	require.Equal(t, completed.Format(time.RFC3339), params.Get("completion-date"))
}

func TestUpdateTodoBuilder_ValidationError(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	longTitle := strings.Repeat("a", 4001)
	_, err := auth.UpdateTodo("uuid").Title(longTitle).Build()
	assert.ErrorIs(t, err, ErrTitleTooLong)
}

// =============================================================================
// UpdateProjectBuilder Tests
// =============================================================================

func TestUpdateProjectBuilder_Title(t *testing.T) {
	scheme := newScheme()
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
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").Completed(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "true", params.Get("completed"))
}

func TestUpdateProjectBuilder_Canceled(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").Canceled(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "true", params.Get("canceled"))
}

func TestUpdateProjectBuilder_Notes(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").Notes("Updated project description").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "Updated project description", params.Get("notes"))
}

func TestUpdateProjectBuilder_PrependNotes(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").PrependNotes("[UPDATE] ").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "[UPDATE] ", params.Get("prepend-notes"))
}

func TestUpdateProjectBuilder_AppendNotes(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").AppendNotes("\n- Added new requirement").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "\n- Added new requirement", params.Get("append-notes"))
}

func TestUpdateProjectBuilder_WhenAnytime(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").WhenAnytime().Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "anytime", params.Get("when"))
}

func TestUpdateProjectBuilder_When(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").When(time.Date(2025, time.April, 1, 0, 0, 0, 0, time.Local)).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "2025-04-01", params.Get("when"))
}

func TestUpdateProjectBuilder_Deadline(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").Deadline(time.Date(2025, time.December, 31, 0, 0, 0, 0, time.Local)).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "2025-12-31", params.Get("deadline"))
}

func TestUpdateProjectBuilder_ClearDeadline(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").ClearDeadline().Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Empty(t, params.Get("deadline"))
}

func TestUpdateProjectBuilder_Tags(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").Tags("priority", "q1").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "priority,q1", params.Get("tags"))
}

func TestUpdateProjectBuilder_AddTags(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").AddTags("reviewed").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "reviewed", params.Get("add-tags"))
}

func TestUpdateProjectBuilder_Area(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").Area("Personal").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "Personal", params.Get("area"))
}

func TestUpdateProjectBuilder_AreaID(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").AreaID("area-uuid").Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "area-uuid", params.Get("area-id"))
}

func TestUpdateProjectBuilder_Reveal(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("uuid").Reveal(true).Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	require.Equal(t, "true", params.Get("reveal"))
}

func TestUpdateProjectBuilder_NoID(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	_, err := auth.UpdateProject("").Completed(true).Build()
	assert.ErrorIs(t, err, ErrIDRequired)
}

func TestUpdateProjectBuilder_ValidationError(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	longTitle := strings.Repeat("a", 4001)
	_, err := auth.UpdateProject("uuid").Title(longTitle).Build()
	assert.ErrorIs(t, err, ErrTitleTooLong)
}

// =============================================================================
// BatchBuilder Tests
// =============================================================================

func TestBatchBuilder_AddTodo(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.Batch().
		AddTodo(func(todo BatchTodoConfigurator) {
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

func TestBatchBuilder_AddProject(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.Batch().
		AddProject(func(project BatchProjectConfigurator) {
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

func TestBatchBuilder_Reveal(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.Batch().
		AddTodo(func(todo BatchTodoConfigurator) {
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

func TestBatchBuilder_NoItems(t *testing.T) {
	scheme := newScheme()
	_, err := scheme.Batch().Build()
	assert.ErrorIs(t, err, ErrNoJSONItems)
}

func TestBatchBuilder_Multiple(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.Batch().
		AddTodo(func(todo BatchTodoConfigurator) {
			todo.Title("Todo 1")
		}).
		AddTodo(func(todo BatchTodoConfigurator) {
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
// AuthBatchBuilder Tests
// =============================================================================

func TestAuthBatchBuilder_UpdateTodo(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.Batch().
		UpdateTodo("uuid-123", func(todo BatchTodoConfigurator) {
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

func TestAuthBatchBuilder_UpdateProject(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.Batch().
		UpdateProject("uuid-123", func(project BatchProjectConfigurator) {
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

func TestAuthBatchBuilder_Mixed(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.Batch().
		AddTodo(func(todo BatchTodoConfigurator) {
			todo.Title("New Todo")
		}).
		UpdateTodo("uuid-123", func(todo BatchTodoConfigurator) {
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

func TestAuthBatchBuilder_EmptyToken(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("")
	_, err := auth.Batch().
		AddTodo(func(todo BatchTodoConfigurator) {
			todo.Title("Test")
		}).
		Build()
	assert.ErrorIs(t, err, ErrEmptyToken)
}

func TestAuthBatchBuilder_NoItems(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	_, err := auth.Batch().Build()
	assert.ErrorIs(t, err, ErrNoJSONItems)
}

// =============================================================================
// batchTodoBuilder Tests
// =============================================================================

func Test_batchTodoBuilder_When(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.Batch().
		AddTodo(func(todo BatchTodoConfigurator) {
			todo.Title("Test").When(time.Date(2025, time.January, 15, 0, 0, 0, 0, time.Local))
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeTodo,
		Attributes: map[string]any{"title": "Test", "when": "2025-01-15"},
	}}, parseJSONItems(t, thingsURL))
}

func Test_batchTodoBuilder_Tags(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.Batch().
		AddTodo(func(todo BatchTodoConfigurator) {
			todo.Title("Test").Tags("Risk", "Golang")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeTodo,
		Attributes: map[string]any{"title": "Test", "tags": []any{"Risk", "Golang"}},
	}}, parseJSONItems(t, thingsURL))
}

func Test_batchTodoBuilder_TitleTooLong(t *testing.T) {
	scheme := newScheme()
	longTitle := strings.Repeat("a", 4001)
	_, err := scheme.Batch().
		AddTodo(func(todo BatchTodoConfigurator) {
			todo.Title(longTitle)
		}).
		Build()
	assert.ErrorIs(t, err, ErrTitleTooLong)
}

// TestbatchTodoBuilder_Notes tests adding notes to a JSON todo
func Test_batchTodoBuilder_Notes(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.Batch().
		AddTodo(func(todo BatchTodoConfigurator) {
			todo.Title("Test").Notes("Detailed description")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeTodo,
		Attributes: map[string]any{"title": "Test", "notes": "Detailed description"},
	}}, parseJSONItems(t, thingsURL))
}

// TestbatchTodoBuilder_WhenDate tests scheduling to a specific date
func Test_batchTodoBuilder_WhenDate(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.Batch().
		AddTodo(func(todo BatchTodoConfigurator) {
			todo.Title("Test").When(time.Date(2025, time.March, 15, 0, 0, 0, 0, time.Local))
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeTodo,
		Attributes: map[string]any{"title": "Test", "when": "2025-03-15"},
	}}, parseJSONItems(t, thingsURL))
}

// TestbatchTodoBuilder_Deadline tests setting a deadline
func Test_batchTodoBuilder_Deadline(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.Batch().
		AddTodo(func(todo BatchTodoConfigurator) {
			todo.Title("Test").Deadline(time.Date(2025, time.June, 30, 0, 0, 0, 0, time.Local))
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeTodo,
		Attributes: map[string]any{"title": "Test", "deadline": "2025-06-30"},
	}}, parseJSONItems(t, thingsURL))
}

// TestbatchTodoBuilder_ChecklistItems tests adding a checklist
func Test_batchTodoBuilder_ChecklistItems(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.Batch().
		AddTodo(func(todo BatchTodoConfigurator) {
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

// TestbatchTodoBuilder_ChecklistItemsTooMany tests the checklist limit
func Test_batchTodoBuilder_ChecklistItemsTooMany(t *testing.T) {
	scheme := newScheme()
	items := make([]string, 101)
	for i := range items {
		items[i] = "json checklist entry"
	}
	_, err := scheme.Batch().
		AddTodo(func(todo BatchTodoConfigurator) {
			todo.Title("Test").ChecklistItems(items...)
		}).
		Build()
	assert.ErrorIs(t, err, ErrTooManyChecklistItems)
}

// TestbatchTodoBuilder_List tests placing todo in a project by name
func Test_batchTodoBuilder_List(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.Batch().
		AddTodo(func(todo BatchTodoConfigurator) {
			todo.Title("Test").List("My Project")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeTodo,
		Attributes: map[string]any{"title": "Test", "list": "My Project"},
	}}, parseJSONItems(t, thingsURL))
}

// TestbatchTodoBuilder_ListID tests placing todo in a project by UUID
func Test_batchTodoBuilder_ListID(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.Batch().
		AddTodo(func(todo BatchTodoConfigurator) {
			todo.Title("Test").ListID("project-uuid-123")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeTodo,
		Attributes: map[string]any{"title": "Test", "list-id": "project-uuid-123"},
	}}, parseJSONItems(t, thingsURL))
}

// TestbatchTodoBuilder_Heading tests placing todo under a heading
func Test_batchTodoBuilder_Heading(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.Batch().
		AddTodo(func(todo BatchTodoConfigurator) {
			todo.Title("Test").List("Project").Heading("Phase 1")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeTodo,
		Attributes: map[string]any{"title": "Test", "list": "Project", "heading": "Phase 1"},
	}}, parseJSONItems(t, thingsURL))
}

// TestbatchTodoBuilder_Completed tests marking as completed
func Test_batchTodoBuilder_Completed(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.Batch().
		AddTodo(func(todo BatchTodoConfigurator) {
			todo.Title("Test").Completed(true)
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeTodo,
		Attributes: map[string]any{"title": "Test", "completed": true},
	}}, parseJSONItems(t, thingsURL))
}

// TestbatchTodoBuilder_Canceled tests marking as canceled
func Test_batchTodoBuilder_Canceled(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.Batch().
		AddTodo(func(todo BatchTodoConfigurator) {
			todo.Title("Test").Canceled(true)
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeTodo,
		Attributes: map[string]any{"title": "Test", "canceled": true},
	}}, parseJSONItems(t, thingsURL))
}

// TestbatchTodoBuilder_CreationDate tests backdating creation
func Test_batchTodoBuilder_CreationDate(t *testing.T) {
	scheme := newScheme()
	pastDate := time.Date(2024, time.June, 1, 10, 0, 0, 0, time.UTC)
	thingsURL, err := scheme.Batch().
		AddTodo(func(todo BatchTodoConfigurator) {
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

// TestbatchTodoBuilder_CompletionDate tests setting completion timestamp
func Test_batchTodoBuilder_CompletionDate(t *testing.T) {
	completedDate := time.Date(2024, time.December, 15, 14, 30, 0, 0, time.UTC)
	thingsURL, err := newScheme().Batch().
		AddTodo(func(todo BatchTodoConfigurator) {
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

// TestbatchTodoBuilder_UpdatePrependNotes tests prepending notes in update
func Test_batchTodoBuilder_UpdatePrependNotes(t *testing.T) {
	auth := newScheme().WithToken("test-token")
	thingsURL, err := auth.Batch().
		UpdateTodo("uuid", func(todo BatchTodoConfigurator) {
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

// TestbatchTodoBuilder_UpdateAppendNotes tests appending notes in update
func Test_batchTodoBuilder_UpdateAppendNotes(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.Batch().
		UpdateTodo("uuid", func(todo BatchTodoConfigurator) {
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

// TestbatchTodoBuilder_UpdateAddTags tests adding tags without replacing
func Test_batchTodoBuilder_UpdateAddTags(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.Batch().
		UpdateTodo("uuid", func(todo BatchTodoConfigurator) {
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
// batchProjectBuilder Tests
// =============================================================================

func Test_batchProjectBuilder_Area(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.Batch().
		AddProject(func(project BatchProjectConfigurator) {
			project.Title("Test").Area("Work")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeProject,
		Attributes: map[string]any{"title": "Test", "area": "Work"},
	}}, parseJSONItems(t, thingsURL))
}

func Test_batchProjectBuilder_Todos(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.Batch().
		AddProject(func(project BatchProjectConfigurator) {
			project.Title("Test Project").Todos(
				func(todo BatchTodoConfigurator) { todo.Title("Task 1") },
				func(todo BatchTodoConfigurator) { todo.Title("Task 2") },
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

// TestbatchProjectBuilder_Notes tests adding project notes
func Test_batchProjectBuilder_Notes(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.Batch().
		AddProject(func(project BatchProjectConfigurator) {
			project.Title("Test").Notes("Project description")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeProject,
		Attributes: map[string]any{"title": "Test", "notes": "Project description"},
	}}, parseJSONItems(t, thingsURL))
}

// TestbatchProjectBuilder_WhenSomeday tests scheduling project for someday
func Test_batchProjectBuilder_WhenSomeday(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.Batch().
		AddProject(func(project BatchProjectConfigurator) {
			project.Title("Test").WhenSomeday()
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeProject,
		Attributes: map[string]any{"title": "Test", "when": "someday"},
	}}, parseJSONItems(t, thingsURL))
}

// TestbatchProjectBuilder_When tests scheduling to specific date
func Test_batchProjectBuilder_When(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.Batch().
		AddProject(func(project BatchProjectConfigurator) {
			project.Title("Test").When(time.Date(2025, time.July, 1, 0, 0, 0, 0, time.Local))
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeProject,
		Attributes: map[string]any{"title": "Test", "when": "2025-07-01"},
	}}, parseJSONItems(t, thingsURL))
}

// TestbatchProjectBuilder_Deadline tests setting project deadline
func Test_batchProjectBuilder_Deadline(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.Batch().
		AddProject(func(project BatchProjectConfigurator) {
			project.Title("Test").Deadline(time.Date(2025, time.December, 31, 0, 0, 0, 0, time.Local))
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeProject,
		Attributes: map[string]any{"title": "Test", "deadline": "2025-12-31"},
	}}, parseJSONItems(t, thingsURL))
}

// TestbatchProjectBuilder_Tags tests setting project tags
func Test_batchProjectBuilder_Tags(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.Batch().
		AddProject(func(project BatchProjectConfigurator) {
			project.Title("Test").Tags("priority", "q1")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeProject,
		Attributes: map[string]any{"title": "Test", "tags": []any{"priority", "q1"}},
	}}, parseJSONItems(t, thingsURL))
}

// TestbatchProjectBuilder_AreaID tests placing project in area by UUID
func Test_batchProjectBuilder_AreaID(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.Batch().
		AddProject(func(project BatchProjectConfigurator) {
			project.Title("Test").AreaID("area-uuid-456")
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeProject,
		Attributes: map[string]any{"title": "Test", "area-id": "area-uuid-456"},
	}}, parseJSONItems(t, thingsURL))
}

// TestbatchProjectBuilder_Completed tests marking project completed
func Test_batchProjectBuilder_Completed(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.Batch().
		AddProject(func(project BatchProjectConfigurator) {
			project.Title("Test").Completed(true)
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeProject,
		Attributes: map[string]any{"title": "Test", "completed": true},
	}}, parseJSONItems(t, thingsURL))
}

// TestbatchProjectBuilder_Canceled tests marking project canceled
func Test_batchProjectBuilder_Canceled(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.Batch().
		AddProject(func(project BatchProjectConfigurator) {
			project.Title("Test").Canceled(true)
		}).
		Build()
	require.NoError(t, err)

	require.Equal(t, []JSONItem{{
		Type:       JSONItemTypeProject,
		Attributes: map[string]any{"title": "Test", "canceled": true},
	}}, parseJSONItems(t, thingsURL))
}

// TestbatchProjectBuilder_CreationDate tests backdating project creation
func Test_batchProjectBuilder_CreationDate(t *testing.T) {
	scheme := newScheme()
	pastDate := time.Date(2024, time.January, 1, 9, 0, 0, 0, time.UTC)
	thingsURL, err := scheme.Batch().
		AddProject(func(project BatchProjectConfigurator) {
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

// TestbatchProjectBuilder_CompletionDate tests setting completion timestamp
func Test_batchProjectBuilder_CompletionDate(t *testing.T) {
	scheme := newScheme()
	completedDate := time.Date(2024, time.November, 30, 17, 0, 0, 0, time.UTC)
	thingsURL, err := scheme.Batch().
		AddProject(func(project BatchProjectConfigurator) {
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

// TestbatchProjectBuilder_UpdatePrependNotes tests prepending notes in update
func Test_batchProjectBuilder_UpdatePrependNotes(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.Batch().
		UpdateProject("uuid", func(project BatchProjectConfigurator) {
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

// TestbatchProjectBuilder_UpdateAppendNotes tests appending notes in update
func Test_batchProjectBuilder_UpdateAppendNotes(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.Batch().
		UpdateProject("uuid", func(project BatchProjectConfigurator) {
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

// TestbatchProjectBuilder_UpdateAddTags tests adding tags without replacing
func Test_batchProjectBuilder_UpdateAddTags(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.Batch().
		UpdateProject("uuid", func(project BatchProjectConfigurator) {
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

// TestbatchProjectBuilder_TodosWithError tests error propagation from child todos
func Test_batchProjectBuilder_TodosWithError(t *testing.T) {
	scheme := newScheme()
	longTitle := strings.Repeat("a", 4001)
	_, err := scheme.Batch().
		AddProject(func(project BatchProjectConfigurator) {
			project.Title("Test").Todos(
				func(todo BatchTodoConfigurator) { todo.Title(longTitle) },
			)
		}).
		Build()
	assert.ErrorIs(t, err, ErrTitleTooLong)
}

// TestAuthBatchBuilder_AddProject tests creating project via auth builder
func TestAuthBatchBuilder_AddProject(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.Batch().
		AddProject(func(project BatchProjectConfigurator) {
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

// TestAuthBatchBuilder_Reveal tests reveal option
func TestAuthBatchBuilder_Reveal(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.Batch().
		AddTodo(func(todo BatchTodoConfigurator) {
			todo.Title("Test")
		}).
		Reveal(true).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "json", cmd)
	require.Equal(t, "true", params.Get("reveal"))
}

// TestAuthBatchBuilder_CreateOnly tests create-only operations don't need auth token
func TestAuthBatchBuilder_CreateOnly(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.Batch().
		AddTodo(func(todo BatchTodoConfigurator) {
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

// TestWhen_Values verifies internal when constants format correctly.
// Note: When type is now private, so we test via builder methods.
func TestWhen_Values(t *testing.T) {
	scheme := newScheme()

	// Test WhenEvening
	url1, err := scheme.AddTodo().Title("Test").WhenEvening().Build()
	require.NoError(t, err)
	_, params := parseThingsURL(t, url1)
	assert.Equal(t, "evening", params.Get("when"))

	// Test WhenAnytime
	url2, err := scheme.AddTodo().Title("Test").WhenAnytime().Build()
	require.NoError(t, err)
	_, params = parseThingsURL(t, url2)
	assert.Equal(t, "anytime", params.Get("when"))

	// Test WhenSomeday
	url3, err := scheme.AddTodo().Title("Test").WhenSomeday().Build()
	require.NoError(t, err)
	_, params = parseThingsURL(t, url3)
	assert.Equal(t, "someday", params.Get("when"))
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

// =============================================================================
// URL Encoding Tests
// =============================================================================

// TestURLEncoding_SpacesAsPercent20 verifies that spaces are encoded as %20, not +.
// Things 3 URL Scheme expects %20 for spaces (RFC 3986), not + (HTML form encoding).
func TestURLEncoding_SpacesAsPercent20(t *testing.T) {
	tests := []struct {
		name     string
		buildURL func() (string, error)
	}{
		{
			name: "AddTodoBuilder with space in title",
			buildURL: func() (string, error) {
				return newScheme().AddTodo().Title("Buy groceries").Build()
			},
		},
		{
			name: "AddProjectBuilder with space in title",
			buildURL: func() (string, error) {
				return newScheme().AddProject().Title("My Project").Build()
			},
		},
		{
			name: "ShowBuilder with space in query",
			buildURL: func() (string, error) {
				return newScheme().ShowBuilder().Query("My Project").Build(), nil
			},
		},
		{
			name: "UpdateTodoBuilder with space in title",
			buildURL: func() (string, error) {
				return newScheme().WithToken("token").UpdateTodo("uuid").Title("New Title").Build()
			},
		},
		{
			name: "UpdateProjectBuilder with space in title",
			buildURL: func() (string, error) {
				return newScheme().WithToken("token").UpdateProject("uuid").Title("New Project").Build()
			},
		},
		{
			name: "BatchBuilder with space in title",
			buildURL: func() (string, error) {
				return newScheme().Batch().AddTodo(func(todo BatchTodoConfigurator) {
					todo.Title("Buy milk")
				}).Build()
			},
		},
		{
			name: "AuthBatchBuilder with space in title",
			buildURL: func() (string, error) {
				return newScheme().WithToken("token").Batch().UpdateTodo("uuid", func(todo BatchTodoConfigurator) {
					todo.Title("New Title")
				}).Build()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			thingsURL, err := tt.buildURL()
			require.NoError(t, err)

			// Verify spaces are encoded as %20, not +
			assert.NotContains(t, thingsURL, "+", "URL should not contain + for spaces")
			assert.Contains(t, thingsURL, "%20", "URL should contain %20 for spaces")
		})
	}
}

// TestURLEncoding_PlusCharacterPreserved verifies that original + characters
// in content are preserved as %2B (not confused with space encoding).
func TestURLEncoding_PlusCharacterPreserved(t *testing.T) {
	tests := []struct {
		name     string
		buildURL func() (string, error)
	}{
		{
			name: "AddTodoBuilder with plus in title",
			buildURL: func() (string, error) {
				return newScheme().AddTodo().Title("Learn C++").Build()
			},
		},
		{
			name: "AddTodoBuilder with plus in notes",
			buildURL: func() (string, error) {
				return newScheme().AddTodo().Title("Test").Notes("1+1=2").Build()
			},
		},
		{
			name: "AddProjectBuilder with plus in title",
			buildURL: func() (string, error) {
				return newScheme().AddProject().Title("C++ Project").Build()
			},
		},
		{
			name: "Search with plus sign",
			buildURL: func() (string, error) {
				return newScheme().SearchURL("C++"), nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			thingsURL, err := tt.buildURL()
			require.NoError(t, err)

			// Verify + is encoded as %2B
			assert.Contains(t, thingsURL, "%2B", "URL should contain %2B for + character")

			// Also verify the value decodes correctly
			cmd, params := parseThingsURL(t, thingsURL)
			assert.NotEmpty(t, cmd)

			// Check that the decoded values contain the original +
			foundPlus := false
			for _, values := range params {
				for _, v := range values {
					if strings.Contains(v, "+") {
						foundPlus = true
						break
					}
				}
			}
			assert.True(t, foundPlus, "Decoded URL should contain original + character")
		})
	}
}

// TestURLEncoding_SpaceAndPlusCombined verifies correct encoding when both
// spaces and + characters are present in the same string.
func TestURLEncoding_SpaceAndPlusCombined(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddTodo().Title("Learn C++ basics").Build()
	require.NoError(t, err)

	// Spaces should be %20
	assert.Contains(t, thingsURL, "%20", "URL should contain %20 for spaces")
	// Plus should be %2B
	assert.Contains(t, thingsURL, "%2B", "URL should contain %2B for + character")
	// No raw + (which would mean incorrectly encoded space)
	// Note: we check the query part only, not the scheme
	queryPart := strings.SplitN(thingsURL, "?", 2)
	require.Len(t, queryPart, 2)
	assert.NotContains(t, queryPart[1], "+", "Query should not contain + for spaces")

	// Verify decoding works correctly
	_, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "Learn C++ basics", params.Get("title"))
}

// TestEncodeQuery_DirectTest directly tests the encodeQuery helper function.
func TestEncodeQuery_DirectTest(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		contains []string
		excludes []string
	}{
		{
			name:     "space encoding",
			input:    map[string]string{"title": "Hello World"},
			contains: []string{"Hello%20World"},
			excludes: []string{"Hello+World"},
		},
		{
			name:     "plus encoding",
			input:    map[string]string{"title": "C++"},
			contains: []string{"C%2B%2B"},
			excludes: []string{},
		},
		{
			name:     "mixed space and plus",
			input:    map[string]string{"title": "C++ tutorial"},
			contains: []string{"%2B%2B", "%20"},
			excludes: []string{},
		},
		{
			name:     "special characters",
			input:    map[string]string{"title": "test@example.com"},
			contains: []string{"%40"}, // @ is encoded as %40
			excludes: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := url.Values{}
			for k, v := range tt.input {
				query.Set(k, v)
			}
			encoded := encodeQuery(query)

			for _, c := range tt.contains {
				assert.Contains(t, encoded, c, "encoded query should contain %s", c)
			}
			for _, e := range tt.excludes {
				assert.NotContains(t, encoded, e, "encoded query should not contain %s", e)
			}
		})
	}
}

// =============================================================================
// Reminder Tests
// =============================================================================

func TestAddTodoBuilder_Reminder_WithWhen(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddTodo().
		Title("Meeting").
		When(time.Date(2025, time.January, 2, 0, 0, 0, 0, time.Local)).
		Reminder(14, 30).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	assert.Equal(t, "Meeting", params.Get("title"))
	assert.Equal(t, "2025-01-02@14:30", params.Get("when"))
}

func TestAddTodoBuilder_Reminder_WithWhenTime(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddTodo().
		Title("Appointment").
		When(time.Date(2025, time.March, 15, 0, 0, 0, 0, time.Local)).
		Reminder(9, 0).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	assert.Equal(t, "Appointment", params.Get("title"))
	assert.Equal(t, "2025-03-15@09:00", params.Get("when"))
}

func TestAddTodoBuilder_Reminder_DefaultsToToday(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddTodo().
		Title("Call").
		Reminder(15, 0).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	assert.Equal(t, "Call", params.Get("title"))
	assert.Equal(t, "today@15:00", params.Get("when"))
}

func TestAddTodoBuilder_Reminder_WithEvening(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddTodo().
		Title("Dry cleaning").
		WhenEvening().
		Reminder(18, 0).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add", cmd)
	assert.Equal(t, "Dry cleaning", params.Get("title"))
	assert.Equal(t, "evening@18:00", params.Get("when"))
}

func TestAddTodoBuilder_Reminder_InvalidHour(t *testing.T) {
	scheme := newScheme()
	_, err := scheme.AddTodo().
		Title("Test").
		Reminder(24, 0).
		Build()
	require.ErrorIs(t, err, ErrInvalidReminderTime)

	_, err = scheme.AddTodo().
		Title("Test").
		Reminder(-1, 0).
		Build()
	require.ErrorIs(t, err, ErrInvalidReminderTime)
}

func TestAddTodoBuilder_Reminder_InvalidMinute(t *testing.T) {
	scheme := newScheme()
	_, err := scheme.AddTodo().
		Title("Test").
		Reminder(10, 60).
		Build()
	require.ErrorIs(t, err, ErrInvalidReminderTime)

	_, err = scheme.AddTodo().
		Title("Test").
		Reminder(10, -1).
		Build()
	require.ErrorIs(t, err, ErrInvalidReminderTime)
}

func TestAddProjectBuilder_Reminder(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddProject().
		Title("Project").
		When(time.Date(2025, time.June, 1, 0, 0, 0, 0, time.Local)).
		Reminder(10, 15).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	assert.Equal(t, "Project", params.Get("title"))
	assert.Equal(t, "2025-06-01@10:15", params.Get("when"))
}

func TestAddProjectBuilder_Reminder_DefaultsToToday(t *testing.T) {
	scheme := newScheme()
	thingsURL, err := scheme.AddProject().
		Title("Project").
		Reminder(8, 30).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "add-project", cmd)
	assert.Equal(t, "Project", params.Get("title"))
	assert.Equal(t, "today@08:30", params.Get("when"))
}

func TestUpdateTodoBuilder_Reminder(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid-123").
		When(time.Date(2025, time.January, 1, 0, 0, 0, 0, time.Local)).
		Reminder(16, 45).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	assert.Equal(t, "uuid-123", params.Get("id"))
	assert.Equal(t, "test-token", params.Get("auth-token"))
	assert.Equal(t, "2025-01-01@16:45", params.Get("when"))
}

func TestUpdateTodoBuilder_Reminder_DefaultsToToday(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateTodo("uuid-123").
		Reminder(12, 0).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update", cmd)
	assert.Equal(t, "uuid-123", params.Get("id"))
	assert.Equal(t, "today@12:00", params.Get("when"))
}

func TestUpdateProjectBuilder_Reminder(t *testing.T) {
	scheme := newScheme()
	auth := scheme.WithToken("test-token")
	thingsURL, err := auth.UpdateProject("project-uuid").
		When(time.Date(2025, time.July, 4, 0, 0, 0, 0, time.Local)).
		Reminder(9, 0).
		Build()
	require.NoError(t, err)

	cmd, params := parseThingsURL(t, thingsURL)
	assert.Equal(t, "update-project", cmd)
	assert.Equal(t, "project-uuid", params.Get("id"))
	assert.Equal(t, "test-token", params.Get("auth-token"))
	assert.Equal(t, "2025-07-04@09:00", params.Get("when"))
}
