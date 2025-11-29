package things3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInbox(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	tasks, err := client.Inbox(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testInbox)
}

func TestToday(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Test all Today items
	tasks, err := client.Today(ctx)
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
	projects, err := client.Tasks().
		WithType(TaskTypeProject).
		WithStartDate(true).
		WithStart(StartAnytime).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, projects, testTodayProjects)
}

func TestUpcoming(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	tasks, err := client.Upcoming(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testUpcoming)
}

func TestAnytime(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	tasks, err := client.Anytime(ctx)
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
	client := newTestClient(t)
	ctx := context.Background()

	tasks, err := client.Someday(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testSomeday)
}

func TestLogbook(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	tasks, err := client.Logbook(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testLogbook)
}

func TestCompleted(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	tasks, err := client.Completed(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testCompleted)
}

func TestCanceled(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	tasks, err := client.Canceled(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testCanceled)
}

func TestTodos(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Test incomplete todos
	todos, err := client.Todos(ctx)
	require.NoError(t, err)
	assert.Len(t, todos, testTodosIncomplete)

	// Test todos with start=Anytime
	todos, err = client.Tasks().
		WithType(TaskTypeTodo).
		WithStart(StartAnytime).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, todos, testTodosAnytime)

	// Test todos with start=Anytime, status=completed
	todos, err = client.Tasks().
		WithType(TaskTypeTodo).
		WithStart(StartAnytime).
		WithStatus(StatusCompleted).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, todos, testTodosAnytimeComplete)

	// Test todos with status=completed
	todos, err = client.Tasks().
		WithType(TaskTypeTodo).
		WithStatus(StatusCompleted).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, todos, testTodosComplete)

	// Test get todo by UUID - use a UUID that exists in the incomplete todos
	allTodos, err := client.Todos(ctx)
	require.NoError(t, err)
	if len(allTodos) > 0 {
		testUUID := allTodos[0].UUID
		todo, err := client.Tasks().WithUUID(testUUID).First(ctx)
		require.NoError(t, err)
		assert.Equal(t, testUUID, todo.UUID)
	}
}

func TestProjects(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Test incomplete projects (not trashed)
	projects, err := client.Projects(ctx)
	require.NoError(t, err)
	assert.Len(t, projects, testProjectsNotTrashed)

	// Test projects with include_items
	projects, err = client.Tasks().
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
	client := newTestClient(t)
	ctx := context.Background()

	// Test all areas
	areas, err := client.Areas().All(ctx)
	require.NoError(t, err)
	assert.Len(t, areas, testAreas)

	// Test areas with include_items
	areas, err = client.Areas().IncludeItems(true).All(ctx)
	require.NoError(t, err)
	assert.Len(t, areas, testAreas)

	// Test area count
	count, err := client.Areas().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, testAreas, count)

	// Test get area by UUID
	area, err := client.Areas().WithUUID(testUUIDArea).First(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Area 3", area.Title)
}

func TestTags(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Test all tags
	tags, err := client.Tags().All(ctx)
	require.NoError(t, err)
	assert.Len(t, tags, testTags)

	// Test tags with include_items
	tags, err = client.Tags().IncludeItems(true).All(ctx)
	require.NoError(t, err)
	assert.Len(t, tags, testTags)

	// Test get tag by title
	tag, err := client.Tags().WithTitle("Errand").First(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Errand", tag.Title)

	// Test tasks filtered by tag
	tasks, err := client.Tasks().WithTag("Errand").All(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)

	// Test tasks with tag "Home"
	tasks, err = client.Tasks().WithTag("Home").All(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
}

func TestDeadlines(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Test all deadlines
	tasks, err := client.Deadlines(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testDeadlines)

	// Test past deadlines
	tasks, err = client.Tasks().
		WithDeadline("past").
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testDeadlinePast)

	// Test future deadlines
	tasks, err = client.Tasks().
		WithDeadline("future").
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testDeadlineFuture)
}

func TestTrash(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Test all trashed items
	tasks, err := client.Trash(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testTrashed)

	// Test trashed todos
	todos, err := client.Tasks().
		WithType(TaskTypeTodo).
		Trashed(true).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, todos, testTrashedTodos)

	// Test trashed projects
	projects, err := client.Tasks().
		WithType(TaskTypeProject).
		Trashed(true).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, projects, testTrashedProjects)
}

func TestChecklist(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Test todo with checklist
	items, err := client.ChecklistItems(ctx, testUUIDTodoChecklist)
	require.NoError(t, err)
	assert.Len(t, items, 3)

	// Test todo without checklist
	items, err = client.ChecklistItems(ctx, testUUIDTodoNoChecklist)
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestSearch(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Test invalid query
	tasks, err := client.Search(ctx, "invalid_query")
	require.NoError(t, err)
	assert.Empty(t, tasks)

	// Test special characters
	tasks, err = client.Search(ctx, "'")
	require.NoError(t, err)
	assert.Empty(t, tasks)

	// Test search that finds results (search with status=nil to include all)
	_, err = client.Tasks().
		Search("To-Do % Heading").
		All(ctx)
	require.NoError(t, err)
	// Note: This searches incomplete tasks by default
}

func TestGetByUUID(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Test invalid UUID
	result, err := client.Get(ctx, "invalid_uuid")
	require.NoError(t, err)
	assert.Nil(t, result)

	// Test get tag by UUID
	result, err = client.Get(ctx, testUUIDTag)
	require.NoError(t, err)
	require.NotNil(t, result)
	tag, ok := result.(*Tag)
	require.True(t, ok, "expected *Tag, got %T", result)
	assert.Equal(t, testUUIDTag, tag.UUID)
}

func TestLast(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Test last 0 days
	tasks, err := client.Last(ctx, "0d")
	require.NoError(t, err)
	assert.Empty(t, tasks)

	// Test last 10000 weeks
	tasks, err = client.Last(ctx, "10000w")
	require.NoError(t, err)
	assert.Len(t, tasks, testTasksIncomplete)

	// Test invalid parameter
	_, err = client.Last(ctx, "")
	require.ErrorIs(t, err, ErrInvalidParameter)
}

func TestTasks(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Test tasks count with filters
	count, err := client.Tasks().
		WithStatus(StatusCompleted).
		Last("100y").
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, testCompleted, count)

	// Test get task by UUID
	count, err = client.Tasks().WithUUID(testUUIDTaskCount).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Test invalid UUID count
	count, err = client.Tasks().WithUUID("invalid_uuid").Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Test tasks in project (without status filter = all tasks)
	tasks, err := client.Tasks().InProject(testUUIDProject).All(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testTasksInProjectAll)

	// Test tasks with tag and project
	tasks, err = client.Tasks().
		WithTag("Home").
		InProject(testUUIDProject).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
}

func TestDatabaseVersion(t *testing.T) {
	client := newTestClient(t)

	// Access internal db to get version
	version, err := getDatabaseVersion(client.db)
	require.NoError(t, err)
	assert.Equal(t, testDatabaseVersion, version)
}

func TestDatabaseVersionMismatch(t *testing.T) {
	_, err := newTestClientOld(t)
	assert.Error(t, err, "expected error for old database version")
}

func TestToken(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	token, err := client.Token(ctx)
	require.NoError(t, err)
	assert.Equal(t, testAuthToken, token)
}

func TestReminderTime(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	task, err := client.Tasks().WithUUID(testUUIDTodoReminder).First(ctx)
	require.NoError(t, err)
	require.NotNil(t, task.ReminderTime)
	assert.Equal(t, "12:34", *task.ReminderTime)
}
