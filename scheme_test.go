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
	url, err := scheme.Todo().Title("Buy groceries").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "things:///add?")
	assert.Contains(t, url, "title=Buy")
}

func TestTodoBuilder_TitleTooLong(t *testing.T) {
	scheme := NewScheme()
	longTitle := strings.Repeat("a", 4001)
	_, err := scheme.Todo().Title(longTitle).Build()
	assert.ErrorIs(t, err, ErrTitleTooLong)
}

func TestTodoBuilder_Notes(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Todo().Title("Test").Notes("Some notes").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "notes=Some")
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
		{WhenToday, "when=today"},
		{WhenTomorrow, "when=tomorrow"},
		{WhenEvening, "when=evening"},
		{WhenAnytime, "when=anytime"},
		{WhenSomeday, "when=someday"},
	}

	for _, tt := range tests {
		t.Run(string(tt.when), func(t *testing.T) {
			scheme := NewScheme()
			url, err := scheme.Todo().Title("Test").When(tt.when).Build()
			require.NoError(t, err)
			assert.Contains(t, url, tt.expected)
		})
	}
}

func TestTodoBuilder_WhenDate(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Todo().Title("Test").WhenDate(2025, time.December, 25).Build()
	require.NoError(t, err)
	assert.Contains(t, url, "when=2025-12-25")
}

func TestTodoBuilder_Deadline(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Todo().Title("Test").Deadline("2025-12-31").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "deadline=2025-12-31")
}

func TestTodoBuilder_Tags(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Todo().Title("Test").Tags("work", "urgent").Build()
	require.NoError(t, err)
	// Tags are comma-separated
	containsTags := strings.Contains(url, "tags=work%2Curgent") || strings.Contains(url, "tags=work,urgent")
	assert.True(t, containsTags, "URL should contain comma-separated tags")
}

func TestTodoBuilder_ChecklistItems(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Todo().Title("Test").ChecklistItems("Item 1", "Item 2").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "checklist-items=")
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
	url, err := scheme.Todo().Title("Test").List("My Project").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "list=My")
}

func TestTodoBuilder_ListID(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Todo().Title("Test").ListID("uuid-123").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "list-id=uuid-123")
}

func TestTodoBuilder_Completed(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Todo().Title("Test").Completed(true).Build()
	require.NoError(t, err)
	assert.Contains(t, url, "completed=true")
}

func TestTodoBuilder_Canceled(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Todo().Title("Test").Canceled(true).Build()
	require.NoError(t, err)
	assert.Contains(t, url, "canceled=true")
}

func TestTodoBuilder_ShowQuickEntry(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Todo().Title("Test").ShowQuickEntry(true).Build()
	require.NoError(t, err)
	assert.Contains(t, url, "show-quick-entry=true")
}

func TestTodoBuilder_Reveal(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Todo().Title("Test").Reveal(true).Build()
	require.NoError(t, err)
	assert.Contains(t, url, "reveal=true")
}

// TestTodoBuilder_Titles tests creating multiple to-dos at once
func TestTodoBuilder_Titles(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Todo().Titles("Task 1", "Task 2", "Task 3").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "titles=")
}

// TestTodoBuilder_TitlesTooLong tests validation for combined titles length
func TestTodoBuilder_TitlesTooLong(t *testing.T) {
	scheme := NewScheme()
	longTitle := strings.Repeat("a", 2000)
	_, err := scheme.Todo().Titles(longTitle, longTitle, longTitle).Build()
	assert.ErrorIs(t, err, ErrTitleTooLong)
}

// TestTodoBuilder_Heading tests placing a to-do under a project heading
func TestTodoBuilder_Heading(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Todo().
		Title("Subtask").
		List("My Project").
		Heading("Phase 1").
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "heading=Phase")
}

// TestTodoBuilder_HeadingID tests placing a to-do under a heading by UUID
func TestTodoBuilder_HeadingID(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Todo().
		Title("Subtask").
		ListID("project-uuid").
		HeadingID("heading-uuid").
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "heading-id=heading-uuid")
}

// TestTodoBuilder_CreationDate tests backdating a to-do's creation
func TestTodoBuilder_CreationDate(t *testing.T) {
	scheme := NewScheme()
	pastDate := time.Date(2024, time.January, 15, 10, 30, 0, 0, time.UTC)
	url, err := scheme.Todo().
		Title("Historical task").
		CreationDate(pastDate).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "creation-date=")
}

// TestTodoBuilder_CompletionDate tests setting completion timestamp for imported tasks
func TestTodoBuilder_CompletionDate(t *testing.T) {
	scheme := NewScheme()
	completedAt := time.Date(2024, time.December, 1, 14, 0, 0, 0, time.UTC)
	url, err := scheme.Todo().
		Title("Imported completed task").
		Completed(true).
		CompletionDate(completedAt).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "completion-date=")
	assert.Contains(t, url, "completed=true")
}

func TestTodoBuilder_Chained(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Todo().
		Title("Buy groceries").
		Notes("Don't forget milk").
		When(WhenToday).
		Tags("shopping").
		Reveal(true).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "things:///add?")
	assert.Contains(t, url, "title=Buy")
	assert.Contains(t, url, "notes=Don")
	assert.Contains(t, url, "when=today")
	assert.Contains(t, url, "tags=shopping")
	assert.Contains(t, url, "reveal=true")
}

// =============================================================================
// ProjectBuilder Tests
// =============================================================================

func TestProjectBuilder_Title(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Project().Title("New Project").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "things:///add-project?")
	assert.Contains(t, url, "title=New")
}

func TestProjectBuilder_TitleTooLong(t *testing.T) {
	scheme := NewScheme()
	longTitle := strings.Repeat("a", 4001)
	_, err := scheme.Project().Title(longTitle).Build()
	assert.ErrorIs(t, err, ErrTitleTooLong)
}

func TestProjectBuilder_Area(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Project().Title("Test").Area("Work").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "area=Work")
}

func TestProjectBuilder_AreaID(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Project().Title("Test").AreaID("uuid-123").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "area-id=uuid-123")
}

func TestProjectBuilder_Todos(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Project().Title("Test").Todos("Task 1", "Task 2").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "to-dos=")
}

// TestProjectBuilder_Notes tests adding project description
func TestProjectBuilder_Notes(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Project().
		Title("Q1 Goals").
		Notes("Quarterly objectives and key results").
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "notes=Quarterly")
}

// TestProjectBuilder_When tests scheduling a project
func TestProjectBuilder_When(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Project().Title("Future Project").When(WhenSomeday).Build()
	require.NoError(t, err)
	assert.Contains(t, url, "when=someday")
}

// TestProjectBuilder_WhenDate tests scheduling a project for a specific date
func TestProjectBuilder_WhenDate(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Project().
		Title("Launch").
		WhenDate(2025, time.March, 1).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "when=2025-03-01")
}

// TestProjectBuilder_Deadline tests setting project deadline
func TestProjectBuilder_Deadline(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Project().
		Title("Release v2.0").
		Deadline("2025-06-30").
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "deadline=2025-06-30")
}

// TestProjectBuilder_Tags tests tagging a project
func TestProjectBuilder_Tags(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Project().
		Title("Website Redesign").
		Tags("work", "high-priority").
		Build()
	require.NoError(t, err)
	containsTags := strings.Contains(url, "tags=work%2Chigh-priority") ||
		strings.Contains(url, "tags=work,high-priority")
	assert.True(t, containsTags, "URL should contain comma-separated tags")
}

// TestProjectBuilder_Completed tests creating an already completed project (for imports)
func TestProjectBuilder_Completed(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Project().
		Title("Archived Project").
		Completed(true).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "completed=true")
}

// TestProjectBuilder_Canceled tests creating a canceled project
func TestProjectBuilder_Canceled(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Project().
		Title("Discontinued Project").
		Canceled(true).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "canceled=true")
}

// TestProjectBuilder_Reveal tests navigating to the created project
func TestProjectBuilder_Reveal(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Project().
		Title("New Project").
		Reveal(true).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "reveal=true")
}

// TestProjectBuilder_CreationDate tests backdating project creation
func TestProjectBuilder_CreationDate(t *testing.T) {
	scheme := NewScheme()
	created := time.Date(2024, time.June, 1, 9, 0, 0, 0, time.UTC)
	url, err := scheme.Project().
		Title("Historical Project").
		CreationDate(created).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "creation-date=")
}

// TestProjectBuilder_CompletionDate tests setting completion time for imported projects
func TestProjectBuilder_CompletionDate(t *testing.T) {
	scheme := NewScheme()
	completed := time.Date(2024, time.November, 15, 17, 0, 0, 0, time.UTC)
	url, err := scheme.Project().
		Title("Imported Completed Project").
		Completed(true).
		CompletionDate(completed).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "completion-date=")
}

// TestProjectBuilder_FullProject tests creating a complete project with all attributes
func TestProjectBuilder_FullProject(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.Project().
		Title("Product Launch").
		Notes("Launch plan for v2.0").
		Area("Work").
		Tags("priority").
		Deadline("2025-03-31").
		Todos("Design", "Development", "Testing", "Release").
		Reveal(true).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "things:///add-project?")
	assert.Contains(t, url, "title=Product")
	assert.Contains(t, url, "area=Work")
	assert.Contains(t, url, "to-dos=")
}

// =============================================================================
// ShowBuilder Tests
// =============================================================================

func TestShowBuilder_ID(t *testing.T) {
	scheme := NewScheme()
	url := scheme.Show().ID("uuid-123").Build()
	assert.Equal(t, "things:///show?id=uuid-123", url)
}

func TestShowBuilder_List(t *testing.T) {
	tests := []struct {
		list     ListID
		expected string
	}{
		{ListInbox, "id=inbox"},
		{ListToday, "id=today"},
		{ListAnytime, "id=anytime"},
		{ListUpcoming, "id=upcoming"},
		{ListSomeday, "id=someday"},
		{ListLogbook, "id=logbook"},
		{ListTomorrow, "id=tomorrow"},
		{ListDeadlines, "id=deadlines"},
		{ListRepeating, "id=repeating"},
		{ListAllProjects, "id=all-projects"},
		{ListLoggedProjects, "id=logged-projects"},
	}

	for _, tt := range tests {
		t.Run(string(tt.list), func(t *testing.T) {
			scheme := NewScheme()
			url := scheme.Show().List(tt.list).Build()
			assert.Contains(t, url, tt.expected)
		})
	}
}

func TestShowBuilder_Query(t *testing.T) {
	scheme := NewScheme()
	url := scheme.Show().Query("My Project").Build()
	assert.Contains(t, url, "query=My")
}

func TestShowBuilder_Filter(t *testing.T) {
	scheme := NewScheme()
	url := scheme.Show().List(ListToday).Filter("work", "urgent").Build()
	assert.Contains(t, url, "id=today")
	// Tags are comma-separated
	containsFilter := strings.Contains(url, "filter=work%2Curgent") || strings.Contains(url, "filter=work,urgent")
	assert.True(t, containsFilter, "URL should contain comma-separated filter tags")
}

func TestShowBuilder_NoParams(t *testing.T) {
	scheme := NewScheme()
	url := scheme.Show().Build()
	assert.Equal(t, "things:///show", url)
}

// =============================================================================
// Search and Version Tests
// =============================================================================

func TestScheme_Search(t *testing.T) {
	scheme := NewScheme()
	url := scheme.Search("my query")
	assert.Contains(t, url, "things:///search?")
	containsQuery := strings.Contains(url, "query=my+query") || strings.Contains(url, "query=my%20query")
	assert.True(t, containsQuery, "URL should contain encoded query")
}

func TestScheme_Version(t *testing.T) {
	scheme := NewScheme()
	url := scheme.Version()
	assert.Equal(t, "things:///version", url)
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
	url, err := auth.UpdateTodo("uuid-123").Completed(true).Build()
	require.NoError(t, err)
	assert.Contains(t, url, "things:///update?")
	assert.Contains(t, url, "id=uuid-123")
	assert.Contains(t, url, "auth-token=test-token")
	assert.Contains(t, url, "completed=true")
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
	url, err := auth.UpdateTodo("uuid").Title("New Title").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "title=New")
}

func TestUpdateTodoBuilder_PrependNotes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateTodo("uuid").PrependNotes("Prefix: ").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "prepend-notes=Prefix")
}

func TestUpdateTodoBuilder_AppendNotes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateTodo("uuid").AppendNotes(" - Suffix").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "append-notes=")
}

func TestUpdateTodoBuilder_AddTags(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateTodo("uuid").AddTags("new-tag").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "add-tags=new-tag")
}

func TestUpdateTodoBuilder_ClearDeadline(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateTodo("uuid").ClearDeadline().Build()
	require.NoError(t, err)
	assert.Contains(t, url, "deadline=")
}

func TestUpdateTodoBuilder_Duplicate(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateTodo("uuid").Duplicate(true).Build()
	require.NoError(t, err)
	assert.Contains(t, url, "duplicate=true")
}

// TestUpdateTodoBuilder_Notes tests replacing to-do notes
func TestUpdateTodoBuilder_Notes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateTodo("uuid").Notes("New description").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "notes=New")
}

// TestUpdateTodoBuilder_When tests rescheduling a to-do
func TestUpdateTodoBuilder_When(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateTodo("uuid").When(WhenTomorrow).Build()
	require.NoError(t, err)
	assert.Contains(t, url, "when=tomorrow")
}

// TestUpdateTodoBuilder_WhenDate tests scheduling to a specific date
func TestUpdateTodoBuilder_WhenDate(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateTodo("uuid").WhenDate(2025, time.February, 14).Build()
	require.NoError(t, err)
	assert.Contains(t, url, "when=2025-02-14")
}

// TestUpdateTodoBuilder_Deadline tests setting a deadline
func TestUpdateTodoBuilder_Deadline(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateTodo("uuid").Deadline("2025-01-31").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "deadline=2025-01-31")
}

// TestUpdateTodoBuilder_Tags tests replacing all tags
func TestUpdateTodoBuilder_Tags(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateTodo("uuid").Tags("new-tag-1", "new-tag-2").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "tags=")
}

// TestUpdateTodoBuilder_ChecklistItems tests replacing checklist
func TestUpdateTodoBuilder_ChecklistItems(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateTodo("uuid").ChecklistItems("Step A", "Step B").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "checklist-items=")
}

// TestUpdateTodoBuilder_PrependChecklistItems tests adding items at the beginning
func TestUpdateTodoBuilder_PrependChecklistItems(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateTodo("uuid").PrependChecklistItems("First step").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "prepend-checklist-items=")
}

// TestUpdateTodoBuilder_AppendChecklistItems tests adding items at the end
func TestUpdateTodoBuilder_AppendChecklistItems(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateTodo("uuid").AppendChecklistItems("Final step").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "append-checklist-items=")
}

// TestUpdateTodoBuilder_List tests moving to-do to another project
func TestUpdateTodoBuilder_List(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateTodo("uuid").List("New Project").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "list=New")
}

// TestUpdateTodoBuilder_ListID tests moving to-do by project UUID
func TestUpdateTodoBuilder_ListID(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateTodo("uuid").ListID("project-uuid").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "list-id=project-uuid")
}

// TestUpdateTodoBuilder_Heading tests moving to-do under a heading
func TestUpdateTodoBuilder_Heading(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateTodo("uuid").Heading("Phase 2").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "heading=Phase")
}

// TestUpdateTodoBuilder_HeadingID tests moving to-do under heading by UUID
func TestUpdateTodoBuilder_HeadingID(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateTodo("uuid").HeadingID("heading-uuid").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "heading-id=heading-uuid")
}

// TestUpdateTodoBuilder_Canceled tests canceling a to-do
func TestUpdateTodoBuilder_Canceled(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateTodo("uuid").Canceled(true).Build()
	require.NoError(t, err)
	assert.Contains(t, url, "canceled=true")
}

// TestUpdateTodoBuilder_Reveal tests navigating to the to-do after update
func TestUpdateTodoBuilder_Reveal(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateTodo("uuid").Reveal(true).Build()
	require.NoError(t, err)
	assert.Contains(t, url, "reveal=true")
}

// TestUpdateTodoBuilder_CreationDate tests modifying creation timestamp
func TestUpdateTodoBuilder_CreationDate(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	created := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	url, err := auth.UpdateTodo("uuid").CreationDate(created).Build()
	require.NoError(t, err)
	assert.Contains(t, url, "creation-date=")
}

// TestUpdateTodoBuilder_CompletionDate tests setting completion timestamp
func TestUpdateTodoBuilder_CompletionDate(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	completed := time.Date(2024, time.December, 31, 23, 59, 0, 0, time.UTC)
	url, err := auth.UpdateTodo("uuid").Completed(true).CompletionDate(completed).Build()
	require.NoError(t, err)
	assert.Contains(t, url, "completion-date=")
}

// TestUpdateTodoBuilder_ValidationError tests error propagation
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
	url, err := auth.UpdateProject("uuid").Title("New Project Title").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "things:///update-project?")
	assert.Contains(t, url, "id=uuid")
	assert.Contains(t, url, "auth-token=test-token")
	assert.Contains(t, url, "title=New")
}

func TestUpdateProjectBuilder_Completed(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateProject("uuid").Completed(true).Build()
	require.NoError(t, err)
	assert.Contains(t, url, "completed=true")
}

func TestUpdateProjectBuilder_Canceled(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateProject("uuid").Canceled(true).Build()
	require.NoError(t, err)
	assert.Contains(t, url, "canceled=true")
}

// TestUpdateProjectBuilder_Notes tests replacing project notes
func TestUpdateProjectBuilder_Notes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateProject("uuid").Notes("Updated project description").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "notes=Updated")
}

// TestUpdateProjectBuilder_PrependNotes tests prepending to project notes
func TestUpdateProjectBuilder_PrependNotes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateProject("uuid").PrependNotes("[UPDATE] ").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "prepend-notes=")
}

// TestUpdateProjectBuilder_AppendNotes tests appending to project notes
func TestUpdateProjectBuilder_AppendNotes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateProject("uuid").AppendNotes("\n- Added new requirement").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "append-notes=")
}

// TestUpdateProjectBuilder_When tests rescheduling a project
func TestUpdateProjectBuilder_When(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateProject("uuid").When(WhenAnytime).Build()
	require.NoError(t, err)
	assert.Contains(t, url, "when=anytime")
}

// TestUpdateProjectBuilder_WhenDate tests scheduling project to specific date
func TestUpdateProjectBuilder_WhenDate(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateProject("uuid").WhenDate(2025, time.April, 1).Build()
	require.NoError(t, err)
	assert.Contains(t, url, "when=2025-04-01")
}

// TestUpdateProjectBuilder_Deadline tests setting project deadline
func TestUpdateProjectBuilder_Deadline(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateProject("uuid").Deadline("2025-12-31").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "deadline=2025-12-31")
}

// TestUpdateProjectBuilder_ClearDeadline tests removing project deadline
func TestUpdateProjectBuilder_ClearDeadline(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateProject("uuid").ClearDeadline().Build()
	require.NoError(t, err)
	assert.Contains(t, url, "deadline=")
}

// TestUpdateProjectBuilder_Tags tests replacing all project tags
func TestUpdateProjectBuilder_Tags(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateProject("uuid").Tags("priority", "q1").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "tags=")
}

// TestUpdateProjectBuilder_AddTags tests adding tags to project
func TestUpdateProjectBuilder_AddTags(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateProject("uuid").AddTags("reviewed").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "add-tags=reviewed")
}

// TestUpdateProjectBuilder_Area tests moving project to an area
func TestUpdateProjectBuilder_Area(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateProject("uuid").Area("Personal").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "area=Personal")
}

// TestUpdateProjectBuilder_AreaID tests moving project to an area by UUID
func TestUpdateProjectBuilder_AreaID(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateProject("uuid").AreaID("area-uuid").Build()
	require.NoError(t, err)
	assert.Contains(t, url, "area-id=area-uuid")
}

// TestUpdateProjectBuilder_Reveal tests navigating to project after update
func TestUpdateProjectBuilder_Reveal(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.UpdateProject("uuid").Reveal(true).Build()
	require.NoError(t, err)
	assert.Contains(t, url, "reveal=true")
}

// TestUpdateProjectBuilder_NoID tests error when ID is missing
func TestUpdateProjectBuilder_NoID(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	_, err := auth.UpdateProject("").Completed(true).Build()
	assert.ErrorIs(t, err, ErrIDRequired)
}

// TestUpdateProjectBuilder_ValidationError tests error propagation
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
	url, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test Todo")
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "things:///json?")
	assert.Contains(t, url, "data=")
	assert.Contains(t, url, "to-do")
	assert.Contains(t, url, "Test+Todo")
}

func TestJSONBuilder_AddProject(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test Project")
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "things:///json?")
	assert.Contains(t, url, "project")
}

func TestJSONBuilder_Reveal(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test")
		}).
		Reveal(true).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "reveal=true")
}

func TestJSONBuilder_NoItems(t *testing.T) {
	scheme := NewScheme()
	_, err := scheme.JSON().Build()
	assert.ErrorIs(t, err, ErrNoJSONItems)
}

func TestJSONBuilder_Multiple(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Todo 1")
		}).
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Todo 2")
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "Todo+1")
	assert.Contains(t, url, "Todo+2")
}

// =============================================================================
// AuthJSONBuilder Tests
// =============================================================================

func TestAuthJSONBuilder_UpdateTodo(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.JSON().
		UpdateTodo("uuid-123", func(todo *JSONTodoBuilder) {
			todo.Completed(true)
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "things:///json?")
	assert.Contains(t, url, "auth-token=test-token")
	assert.Contains(t, url, "update")
	assert.Contains(t, url, "uuid-123")
}

func TestAuthJSONBuilder_UpdateProject(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.JSON().
		UpdateProject("uuid-123", func(project *JSONProjectBuilder) {
			project.Completed(true)
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "auth-token=test-token")
	assert.Contains(t, url, "update")
}

func TestAuthJSONBuilder_Mixed(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("New Todo")
		}).
		UpdateTodo("uuid-123", func(todo *JSONTodoBuilder) {
			todo.Completed(true)
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "auth-token=test-token")
	assert.Contains(t, url, "New+Todo")
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
	url, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").When(WhenToday)
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "today")
}

func TestJSONTodoBuilder_Tags(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").Tags("Risk", "Golang")
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "Risk")
	assert.Contains(t, url, "Golang")
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
	url, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").Notes("Detailed description")
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "Detailed")
}

// TestJSONTodoBuilder_WhenDate tests scheduling to a specific date
func TestJSONTodoBuilder_WhenDate(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").WhenDate(2025, time.March, 15)
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "2025-03-15")
}

// TestJSONTodoBuilder_Deadline tests setting a deadline
func TestJSONTodoBuilder_Deadline(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").Deadline("2025-06-30")
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "2025-06-30")
}

// TestJSONTodoBuilder_ChecklistItems tests adding a checklist
func TestJSONTodoBuilder_ChecklistItems(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").ChecklistItems("Step 1", "Step 2", "Step 3")
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "checklist-item")
	assert.Contains(t, url, "Step+1")
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
	url, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").List("My Project")
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "My+Project")
}

// TestJSONTodoBuilder_ListID tests placing todo in a project by UUID
func TestJSONTodoBuilder_ListID(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").ListID("project-uuid-123")
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "project-uuid-123")
}

// TestJSONTodoBuilder_Heading tests placing todo under a heading
func TestJSONTodoBuilder_Heading(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").List("Project").Heading("Phase 1")
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "Phase+1")
}

// TestJSONTodoBuilder_Completed tests marking as completed
func TestJSONTodoBuilder_Completed(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").Completed(true)
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "completed")
	assert.Contains(t, url, "true")
}

// TestJSONTodoBuilder_Canceled tests marking as canceled
func TestJSONTodoBuilder_Canceled(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").Canceled(true)
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "canceled")
	assert.Contains(t, url, "true")
}

// TestJSONTodoBuilder_CreationDate tests backdating creation
func TestJSONTodoBuilder_CreationDate(t *testing.T) {
	scheme := NewScheme()
	pastDate := time.Date(2024, time.June, 1, 10, 0, 0, 0, time.UTC)
	url, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").CreationDate(pastDate)
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "creation-date")
	assert.Contains(t, url, "2024-06-01")
}

// TestJSONTodoBuilder_CompletionDate tests setting completion timestamp
func TestJSONTodoBuilder_CompletionDate(t *testing.T) {
	scheme := NewScheme()
	completedDate := time.Date(2024, time.December, 15, 14, 30, 0, 0, time.UTC)
	url, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test").Completed(true).CompletionDate(completedDate)
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "completion-date")
	assert.Contains(t, url, "2024-12-15")
}

// TestJSONTodoBuilder_UpdatePrependNotes tests prepending notes in update
func TestJSONTodoBuilder_UpdatePrependNotes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.JSON().
		UpdateTodo("uuid", func(todo *JSONTodoBuilder) {
			todo.PrependNotes("Important: ")
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "prepend-notes")
	assert.Contains(t, url, "Important")
}

// TestJSONTodoBuilder_UpdateAppendNotes tests appending notes in update
func TestJSONTodoBuilder_UpdateAppendNotes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.JSON().
		UpdateTodo("uuid", func(todo *JSONTodoBuilder) {
			todo.AppendNotes(" - Updated")
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "append-notes")
	assert.Contains(t, url, "Updated")
}

// TestJSONTodoBuilder_UpdateAddTags tests adding tags without replacing
func TestJSONTodoBuilder_UpdateAddTags(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.JSON().
		UpdateTodo("uuid", func(todo *JSONTodoBuilder) {
			todo.AddTags("new-tag", "another-tag")
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "add-tags")
	assert.Contains(t, url, "new-tag")
}

// =============================================================================
// JSONProjectBuilder Tests
// =============================================================================

func TestJSONProjectBuilder_Area(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").Area("Work")
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "Work")
}

func TestJSONProjectBuilder_Todos(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test Project").Todos(
				NewTodo().Title("Task 1"),
				NewTodo().Title("Task 2"),
			)
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "Task+1")
	assert.Contains(t, url, "Task+2")
}

// TestJSONProjectBuilder_Notes tests adding project notes
func TestJSONProjectBuilder_Notes(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").Notes("Project description")
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "Project+description")
}

// TestJSONProjectBuilder_When tests scheduling project
func TestJSONProjectBuilder_When(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").When(WhenSomeday)
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "someday")
}

// TestJSONProjectBuilder_WhenDate tests scheduling to specific date
func TestJSONProjectBuilder_WhenDate(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").WhenDate(2025, time.July, 1)
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "2025-07-01")
}

// TestJSONProjectBuilder_Deadline tests setting project deadline
func TestJSONProjectBuilder_Deadline(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").Deadline("2025-12-31")
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "2025-12-31")
}

// TestJSONProjectBuilder_Tags tests setting project tags
func TestJSONProjectBuilder_Tags(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").Tags("priority", "q1")
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "priority")
	assert.Contains(t, url, "q1")
}

// TestJSONProjectBuilder_AreaID tests placing project in area by UUID
func TestJSONProjectBuilder_AreaID(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").AreaID("area-uuid-456")
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "area-uuid-456")
}

// TestJSONProjectBuilder_Completed tests marking project completed
func TestJSONProjectBuilder_Completed(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").Completed(true)
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "completed")
	assert.Contains(t, url, "true")
}

// TestJSONProjectBuilder_Canceled tests marking project canceled
func TestJSONProjectBuilder_Canceled(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").Canceled(true)
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "canceled")
	assert.Contains(t, url, "true")
}

// TestJSONProjectBuilder_CreationDate tests backdating project creation
func TestJSONProjectBuilder_CreationDate(t *testing.T) {
	scheme := NewScheme()
	pastDate := time.Date(2024, time.January, 1, 9, 0, 0, 0, time.UTC)
	url, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").CreationDate(pastDate)
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "creation-date")
	assert.Contains(t, url, "2024-01-01")
}

// TestJSONProjectBuilder_CompletionDate tests setting completion timestamp
func TestJSONProjectBuilder_CompletionDate(t *testing.T) {
	scheme := NewScheme()
	completedDate := time.Date(2024, time.November, 30, 17, 0, 0, 0, time.UTC)
	url, err := scheme.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("Test").Completed(true).CompletionDate(completedDate)
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "completion-date")
	assert.Contains(t, url, "2024-11-30")
}

// TestJSONProjectBuilder_UpdatePrependNotes tests prepending notes in update
func TestJSONProjectBuilder_UpdatePrependNotes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.JSON().
		UpdateProject("uuid", func(project *JSONProjectBuilder) {
			project.PrependNotes("Update: ")
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "prepend-notes")
	assert.Contains(t, url, "Update")
}

// TestJSONProjectBuilder_UpdateAppendNotes tests appending notes in update
func TestJSONProjectBuilder_UpdateAppendNotes(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.JSON().
		UpdateProject("uuid", func(project *JSONProjectBuilder) {
			project.AppendNotes(" - Reviewed")
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "append-notes")
	assert.Contains(t, url, "Reviewed")
}

// TestJSONProjectBuilder_UpdateAddTags tests adding tags without replacing
func TestJSONProjectBuilder_UpdateAddTags(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.JSON().
		UpdateProject("uuid", func(project *JSONProjectBuilder) {
			project.AddTags("reviewed", "approved")
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "add-tags")
	assert.Contains(t, url, "reviewed")
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
	url, err := auth.JSON().
		AddProject(func(project *JSONProjectBuilder) {
			project.Title("New Project").Area("Work")
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "things:///json?")
	assert.Contains(t, url, "New+Project")
	assert.Contains(t, url, "Work")
}

// TestAuthJSONBuilder_Reveal tests reveal option
func TestAuthJSONBuilder_Reveal(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test")
		}).
		Reveal(true).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "reveal=true")
}

// TestAuthJSONBuilder_CreateOnly tests create-only operations don't need auth token
func TestAuthJSONBuilder_CreateOnly(t *testing.T) {
	scheme := NewScheme()
	auth := scheme.WithToken("test-token")
	url, err := auth.JSON().
		AddTodo(func(todo *JSONTodoBuilder) {
			todo.Title("Test")
		}).
		Build()
	require.NoError(t, err)
	// Create-only operations don't include auth-token in URL
	assert.NotContains(t, url, "auth-token")
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
