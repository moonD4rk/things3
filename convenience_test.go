package things3

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInbox(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	tasks, err := db.Inbox(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, testInboxUUIDs, extractUUIDs(tasks))
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
		Type().Project().
		StartDate().Exists(true).
		Start().Anytime().
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, projects, testTodayProjects)
}

func TestUpcoming(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	tasks, err := db.Upcoming(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, testUpcomingUUIDs, extractUUIDs(tasks))
}

func TestAnytime(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	tasks, err := db.Anytime(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, testAnytimeUUIDs, extractUUIDs(tasks))

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
	assert.ElementsMatch(t, testSomedayUUIDs, extractUUIDs(tasks))
}

func TestLogbook(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	tasks, err := db.Logbook(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, append(testCompletedUUIDs, testCanceledUUIDs...), extractUUIDs(tasks))
}

func TestCompleted(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	tasks, err := db.Completed(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, testCompletedUUIDs, extractUUIDs(tasks))
}

func TestCanceled(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	tasks, err := db.Canceled(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, testCanceledUUIDs, extractUUIDs(tasks))
}

func TestTodos(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test incomplete todos
	todos, err := db.Todos(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, testTodosIncompleteUUIDs, extractUUIDs(todos))

	// Test todos with start=Anytime
	todos, err = db.Tasks().
		Type().Todo().
		Start().Anytime().
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, testTodosAnytimeUUIDs, extractUUIDs(todos))

	// Test todos with start=Anytime, status=completed
	todos, err = db.Tasks().
		Type().Todo().
		Start().Anytime().
		Status().Completed().
		All(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, testTodosAnytimeCompleteUUIDs, extractUUIDs(todos))

	// Test todos with status=completed
	todos, err = db.Tasks().
		Type().Todo().
		Status().Completed().
		All(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, testTodosCompleteUUIDs, extractUUIDs(todos))

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
	assert.ElementsMatch(t, testProjectsNotTrashedUUIDs, extractUUIDs(projects))

	// Test projects with include_items
	projects, err = db.Tasks().
		Type().Project().
		Status().Incomplete().
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
	assert.ElementsMatch(t, testAreaUUIDs, extractAreaUUIDs(areas))

	// Test areas with include_items
	areas, err = db.Areas().IncludeItems(true).All(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, testAreaUUIDs, extractAreaUUIDs(areas))

	// Test area count
	count, err := db.Areas().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, testAreas, count)

	// Test get area by UUID
	area, err := db.Areas().WithUUID(testUUIDArea3).First(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Area 3", area.Title)
}

func TestTags(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test all tags
	tags, err := db.Tags().All(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, testTagUUIDs, extractTagUUIDs(tags))

	// Test tags with include_items
	tags, err = db.Tags().IncludeItems(true).All(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, testTagUUIDs, extractTagUUIDs(tags))

	// Test get tag by title
	tag, err := db.Tags().WithTitle("Errand").First(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Errand", tag.Title)

	// Test tasks filtered by tag
	tasks, err := db.Tasks().InTag("Errand").All(ctx)
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, testUUIDTodoInArea1Tags, tasks[0].UUID)

	// Test tasks with tag "Home"
	tasks, err = db.Tasks().InTag("Home").All(ctx)
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, testUUIDTodoInArea1Tags, tasks[0].UUID)
}

func TestDeadlines(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test all deadlines
	tasks, err := db.Deadlines(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, testDeadlineUUIDs, extractUUIDs(tasks))

	// Test past deadlines
	tasks, err = db.Tasks().
		Deadline().Past().
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, testDeadlinePastUUIDs, extractUUIDs(tasks))

	// Test future deadlines
	tasks, err = db.Tasks().
		Deadline().Future().
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, testDeadlineFutureUUIDs, extractUUIDs(tasks))
}

func TestTrash(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test all trashed items
	tasks, err := db.Trash(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, testTrashedUUIDs, extractUUIDs(tasks))

	// Test trashed todos
	todos, err := db.Tasks().
		Type().Todo().
		Trashed(true).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, testTrashedTodoUUIDs, extractUUIDs(todos))

	// Test trashed projects
	projects, err := db.Tasks().
		Type().Project().
		Trashed(true).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, testTrashedProjectUUIDs, extractUUIDs(projects))
}

func TestChecklist(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test todo with checklist
	items, err := db.ChecklistItems(ctx, testUUIDTodoInboxChecklist)
	require.NoError(t, err)
	assert.ElementsMatch(t, testChecklistUUIDs, extractChecklistUUIDs(items))

	// Test todo without checklist
	items, err = db.ChecklistItems(ctx, testUUIDTodoRepeating)
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

	// Test search with SQL wildcard character - should not crash and should find match
	tasks, err = db.Tasks().
		Search("To-Do % Heading").
		All(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, tasks, "search with wildcard should find matching tasks")
}

func TestGetByUUID(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test invalid UUID
	result, err := db.Get(ctx, "invalid_uuid")
	require.NoError(t, err)
	assert.Nil(t, result)

	// Test get tag by UUID
	result, err = db.Get(ctx, testUUIDTagOffice)
	require.NoError(t, err)
	require.NotNil(t, result)
	tag, ok := result.(*Tag)
	require.True(t, ok, "expected *Tag, got %T", result)
	assert.Equal(t, testUUIDTagOffice, tag.UUID)
}

func TestCreatedWithin(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test zero time - should return error
	_, err := db.CreatedWithin(ctx, time.Time{})
	require.ErrorIs(t, err, ErrInvalidParameter)

	// Test many weeks ago - should return all filtered tasks
	allTasks, err := db.CreatedWithin(ctx, WeeksAgo(10000))
	require.NoError(t, err)
	assert.Len(t, allTasks, testTasksIncompleteFiltered)

	// Test months ago - recent filter should return subset, validate Created time
	threshold := MonthsAgo(1)
	recentTasks, err := db.CreatedWithin(ctx, threshold)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(recentTasks), len(allTasks))
	for _, task := range recentTasks {
		assert.True(t, task.Created.After(threshold),
			"Created %v should be after %v", task.Created, threshold)
	}

	// Test years ago - same validation
	threshold = YearsAgo(1)
	recentTasks, err = db.CreatedWithin(ctx, threshold)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(recentTasks), len(allTasks))
	for _, task := range recentTasks {
		assert.True(t, task.Created.After(threshold),
			"Created %v should be after %v", task.Created, threshold)
	}
}

func TestTasks(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test tasks count with filters
	count, err := db.Tasks().
		Status().Completed().
		CreatedAfter(YearsAgo(100)).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, testCompleted, count)

	// Test get task by UUID
	count, err = db.Tasks().WithUUID(testUUIDTodoInToday).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Test invalid UUID count
	count, err = db.Tasks().WithUUID("invalid_uuid").Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Test tasks in project (without status filter = all tasks)
	tasks, err := db.Tasks().InProject(testUUIDProjectInArea1).All(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testTasksInProjectAll)

	// Test tasks with tag and project
	tasks, err = db.Tasks().
		InTag("Home").
		InProject(testUUIDProjectInArea1).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
}

func TestDatabaseVersion(t *testing.T) {
	db := newTestDB(t)

	// Access internal db to get version
	version, err := getDatabaseVersion(db.sqlDB)
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

	task, err := db.Tasks().WithUUID(testUUIDTodoUpcoming).First(ctx)
	require.NoError(t, err)
	require.NotNil(t, task.ReminderTime)
	assert.Equal(t, 12, task.ReminderTime.Hour())
	assert.Equal(t, 34, task.ReminderTime.Minute())
}
