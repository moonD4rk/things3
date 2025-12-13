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
		items[i] = "item"
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

// =============================================================================
// JSONBuilder Tests
// =============================================================================

func TestJSONBuilder_AddTodo(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoItem) {
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
		AddProject(func(project *JSONProjectItem) {
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
		AddTodo(func(todo *JSONTodoItem) {
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
		AddTodo(func(todo *JSONTodoItem) {
			todo.Title("Todo 1")
		}).
		AddTodo(func(todo *JSONTodoItem) {
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
		UpdateTodo("uuid-123", func(todo *JSONTodoItem) {
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
		UpdateProject("uuid-123", func(project *JSONProjectItem) {
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
		AddTodo(func(todo *JSONTodoItem) {
			todo.Title("New Todo")
		}).
		UpdateTodo("uuid-123", func(todo *JSONTodoItem) {
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
		AddTodo(func(todo *JSONTodoItem) {
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
// JSONTodoItem Tests
// =============================================================================

func TestJSONTodoItem_When(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoItem) {
			todo.Title("Test").When(WhenToday)
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "today")
}

func TestJSONTodoItem_Tags(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoItem) {
			todo.Title("Test").Tags("work", "urgent")
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "work")
	assert.Contains(t, url, "urgent")
}

func TestJSONTodoItem_TitleTooLong(t *testing.T) {
	scheme := NewScheme()
	longTitle := strings.Repeat("a", 4001)
	_, err := scheme.JSON().
		AddTodo(func(todo *JSONTodoItem) {
			todo.Title(longTitle)
		}).
		Build()
	assert.ErrorIs(t, err, ErrTitleTooLong)
}

// =============================================================================
// JSONProjectItem Tests
// =============================================================================

func TestJSONProjectItem_Area(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.JSON().
		AddProject(func(project *JSONProjectItem) {
			project.Title("Test").Area("Work")
		}).
		Build()
	require.NoError(t, err)
	assert.Contains(t, url, "Work")
}

func TestJSONProjectItem_Todos(t *testing.T) {
	scheme := NewScheme()
	url, err := scheme.JSON().
		AddProject(func(project *JSONProjectItem) {
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
