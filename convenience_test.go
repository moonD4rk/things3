package things3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInbox(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	tasks, err := db.Inbox(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testInbox)
}

func TestToday(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test all Today items
	tasks, err := db.Today(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testToday)

	// Verify expected task titles in order
	expectedTitles := []string{
		"Upcoming To-Do in Today (yellow)",
		"Project in Today",
		"To-Do in Today",
		"Repeating To-Do",
		"Overdue Todo automatically shown in Today",
	}
	for i, title := range expectedTitles {
		if i < len(tasks) {
			assert.Equal(t, title, tasks[i].Title, "Today()[%d].Title mismatch", i)
		}
	}

	// Test Today projects only
	projects, err := db.Tasks().
		WithType(TaskTypeProject).
		WithStartDate(true).
		WithStart(StartAnytime).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, projects, testTodayProjects)
}

func TestUpcoming(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	tasks, err := db.Upcoming(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testUpcoming)
}

func TestAnytime(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	tasks, err := db.Anytime(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testAnytime)

	// Verify at least one task has area_title "Area 1"
	hasArea1 := false
	for _, task := range tasks {
		if task.AreaTitle != nil && *task.AreaTitle == "Area 1" {
			hasArea1 = true
			break
		}
	}
	assert.True(t, hasArea1, "Anytime() should contain a task with area_title 'Area 1'")
}

func TestSomeday(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	tasks, err := db.Someday(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testSomeday)
}

func TestLogbook(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	tasks, err := db.Logbook(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testLogbook)
}

func TestCompleted(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	tasks, err := db.Completed(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testCompleted)
}

func TestCanceled(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	tasks, err := db.Canceled(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testCanceled)
}

func TestTodos(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test incomplete todos
	todos, err := db.Todos(ctx)
	require.NoError(t, err)
	assert.Len(t, todos, testTodosIncomplete)

	// Test todos with start=Anytime
	todos, err = db.Tasks().
		WithType(TaskTypeTodo).
		WithStart(StartAnytime).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, todos, testTodosAnytime)

	// Test todos with start=Anytime, status=completed
	todos, err = db.Tasks().
		WithType(TaskTypeTodo).
		WithStart(StartAnytime).
		WithStatus(StatusCompleted).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, todos, testTodosAnytimeComplete)

	// Test todos with status=completed
	todos, err = db.Tasks().
		WithType(TaskTypeTodo).
		WithStatus(StatusCompleted).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, todos, testTodosComplete)

	// Test get todo by UUID - use a UUID that exists in the incomplete todos
	allTodos, err := db.Todos(ctx)
	require.NoError(t, err)
	if len(allTodos) > 0 {
		testUUID := allTodos[0].UUID
		todo, err := db.Tasks().WithUUID(testUUID).First(ctx)
		require.NoError(t, err)
		assert.Equal(t, testUUID, todo.UUID)
	}
}

func TestProjects(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test incomplete projects (not trashed)
	projects, err := db.Projects(ctx)
	require.NoError(t, err)
	assert.Len(t, projects, testProjectsNotTrashed)

	// Test projects with include_items
	projects, err = db.Tasks().
		WithType(TaskTypeProject).
		WithStatus(StatusIncomplete).
		IncludeItems(true).
		All(ctx)
	require.NoError(t, err)
	if len(projects) > 0 {
		assert.Len(t, projects[0].Items, testProjectItems)
	}
}

func TestAreas(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test all areas
	areas, err := db.Areas().All(ctx)
	require.NoError(t, err)
	assert.Len(t, areas, testAreas)

	// Test areas with include_items
	areas, err = db.Areas().IncludeItems(true).All(ctx)
	require.NoError(t, err)
	assert.Len(t, areas, testAreas)

	// Test area count
	count, err := db.Areas().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, testAreas, count)

	// Test get area by UUID
	area, err := db.Areas().WithUUID(testUUIDArea).First(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Area 3", area.Title)
}

func TestTags(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test all tags
	tags, err := db.Tags().All(ctx)
	require.NoError(t, err)
	assert.Len(t, tags, testTags)

	// Test tags with include_items
	tags, err = db.Tags().IncludeItems(true).All(ctx)
	require.NoError(t, err)
	assert.Len(t, tags, testTags)

	// Test get tag by title
	tag, err := db.Tags().WithTitle("Errand").First(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Errand", tag.Title)

	// Test tasks filtered by tag
	tasks, err := db.Tasks().WithTag("Errand").All(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)

	// Test tasks with tag "Home"
	tasks, err = db.Tasks().WithTag("Home").All(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
}

func TestDeadlines(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test all deadlines
	tasks, err := db.Deadlines(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testDeadlines)

	// Test past deadlines
	tasks, err = db.Tasks().
		WithDeadline("past").
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testDeadlinePast)

	// Test future deadlines
	tasks, err = db.Tasks().
		WithDeadline("future").
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testDeadlineFuture)
}

func TestTrash(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test all trashed items
	tasks, err := db.Trash(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testTrashed)

	// Test trashed todos
	todos, err := db.Tasks().
		WithType(TaskTypeTodo).
		Trashed(true).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, todos, testTrashedTodos)

	// Test trashed projects
	projects, err := db.Tasks().
		WithType(TaskTypeProject).
		Trashed(true).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, projects, testTrashedProjects)
}

func TestChecklist(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test todo with checklist
	items, err := db.ChecklistItems(ctx, testUUIDTodoChecklist)
	require.NoError(t, err)
	assert.Len(t, items, 3)

	// Test todo without checklist
	items, err = db.ChecklistItems(ctx, testUUIDTodoNoChecklist)
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestSearch(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test invalid query
	tasks, err := db.Search(ctx, "invalid_query")
	require.NoError(t, err)
	assert.Empty(t, tasks)

	// Test special characters
	tasks, err = db.Search(ctx, "'")
	require.NoError(t, err)
	assert.Empty(t, tasks)

	// Test search that finds results (search with status=nil to include all)
	_, err = db.Tasks().
		Search("To-Do % Heading").
		All(ctx)
	require.NoError(t, err)
	// Note: This searches incomplete tasks by default
}

func TestGetByUUID(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test invalid UUID
	result, err := db.Get(ctx, "invalid_uuid")
	require.NoError(t, err)
	assert.Nil(t, result)

	// Test get tag by UUID
	result, err = db.Get(ctx, testUUIDTag)
	require.NoError(t, err)
	require.NotNil(t, result)
	tag, ok := result.(*Tag)
	require.True(t, ok, "expected *Tag, got %T", result)
	assert.Equal(t, testUUIDTag, tag.UUID)
}

func TestLast(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test last 0 days
	tasks, err := db.Last(ctx, "0d")
	require.NoError(t, err)
	assert.Empty(t, tasks)

	// Test last 10000 weeks
	tasks, err = db.Last(ctx, "10000w")
	require.NoError(t, err)
	assert.Len(t, tasks, testTasksIncomplete)

	// Test invalid parameter
	_, err = db.Last(ctx, "")
	require.ErrorIs(t, err, ErrInvalidParameter)
}

func TestTasks(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test tasks count with filters
	count, err := db.Tasks().
		WithStatus(StatusCompleted).
		Last("100y").
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, testCompleted, count)

	// Test get task by UUID
	count, err = db.Tasks().WithUUID(testUUIDTaskCount).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Test invalid UUID count
	count, err = db.Tasks().WithUUID("invalid_uuid").Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Test tasks in project (without status filter = all tasks)
	tasks, err := db.Tasks().InProject(testUUIDProject).All(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testTasksInProjectAll)

	// Test tasks with tag and project
	tasks, err = db.Tasks().
		WithTag("Home").
		InProject(testUUIDProject).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
}

func TestDatabaseVersion(t *testing.T) {
	db := newTestDB(t)

	// Access internal db to get version
	version, err := getDatabaseVersion(db.db)
	require.NoError(t, err)
	assert.Equal(t, testDatabaseVersion, version)
}

func TestDatabaseVersionMismatch(t *testing.T) {
	_, err := newTestDBOld(t)
	assert.Error(t, err, "expected error for old database version")
}

func TestToken(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	token, err := db.Token(ctx)
	require.NoError(t, err)
	assert.Equal(t, testAuthToken, token)
}

func TestReminderTime(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	task, err := db.Tasks().WithUUID(testUUIDTodoReminder).First(ctx)
	require.NoError(t, err)
	require.NotNil(t, task.ReminderTime)
	assert.Equal(t, "12:34", *task.ReminderTime)
}
