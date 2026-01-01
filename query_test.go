package things3

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskQueryChaining(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test method chaining
	tasks, err := db.Tasks().
		Type().Todo().
		Status().Incomplete().
		Start().Anytime().
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testTodosAnytime)
}

func TestTaskQueryWithUUID(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// First get a valid task UUID from the database
	tasks, err := db.Tasks().Status().Incomplete().All(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, tasks, "No tasks in test database")

	targetUUID := tasks[0].UUID

	// Now test WithUUID with the valid UUID
	task, err := db.Tasks().WithUUID(targetUUID).First(ctx)
	require.NoError(t, err)
	assert.Equal(t, targetUUID, task.UUID)
}

func TestTaskQueryWithType(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test todos
	todos, err := db.Tasks().
		Type().Todo().
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range todos {
		assert.Equal(t, TaskTypeTodo, task.Type)
	}

	// Test projects
	projects, err := db.Tasks().
		Type().Project().
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range projects {
		assert.Equal(t, TaskTypeProject, task.Type)
	}
}

func TestTaskQueryWithStatus(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test incomplete
	tasks, err := db.Tasks().Status().Incomplete().All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Equal(t, StatusIncomplete, task.Status)
	}

	// Test completed
	tasks, err = db.Tasks().Status().Completed().All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Equal(t, StatusCompleted, task.Status)
	}

	// Test canceled
	tasks, err = db.Tasks().Status().Canceled().All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Equal(t, StatusCanceled, task.Status)
	}
}

func TestTaskQueryWithStart(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test Inbox
	tasks, err := db.Tasks().
		Start().Inbox().
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Equal(t, "Inbox", task.Start)
	}

	// Test Anytime
	tasks, err = db.Tasks().
		Start().Anytime().
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Equal(t, "Anytime", task.Start)
	}

	// Test Someday
	tasks, err = db.Tasks().
		Start().Someday().
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Equal(t, "Someday", task.Start)
	}
}

func TestTaskQueryInArea(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test with specific area
	tasks, err := db.Tasks().
		InArea(testUUIDArea).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		require.NotNil(t, task.AreaUUID)
		assert.Equal(t, testUUIDArea, *task.AreaUUID)
	}

	// Test with has area
	tasks, err = db.Tasks().
		HasArea(true).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.NotNil(t, task.AreaUUID, "HasArea(true) returned task without area")
	}

	// Test without area
	tasks, err = db.Tasks().
		HasArea(false).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Nil(t, task.AreaUUID, "HasArea(false) returned task with area")
	}
}

func TestTaskQueryInProject(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test with specific project
	tasks, err := db.Tasks().
		InProject(testUUIDProject).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testTasksInProject)
}

func TestTaskQueryInTag(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test with specific tag
	tasks, err := db.Tasks().
		InTag("Errand").
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)

	// Test with has tag
	tasks, err = db.Tasks().
		HasTag(true).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.NotEmpty(t, task.Tags, "HasTag(true) returned task without tags")
	}
}

func TestTaskQueryWithDeadline(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test has deadline
	tasks, err := db.Tasks().
		Deadline().Exists(true).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.NotNil(t, task.Deadline, "Deadline().Exists(true) returned task without deadline")
	}

	// Test no deadline
	tasks, err = db.Tasks().
		Deadline().Exists(false).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Nil(t, task.Deadline, "Deadline().Exists(false) returned task with deadline")
	}
}

func TestTaskQueryWithStartDate(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test has start date
	tasks, err := db.Tasks().
		StartDate().Exists(true).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.NotNil(t, task.StartDate, "StartDate().Exists(true) returned task without start date")
	}
}

func TestTaskQueryTrashed(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test trashed
	tasks, err := db.Tasks().
		Trashed(true).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.True(t, task.Trashed, "Trashed(true) returned non-trashed task")
	}

	// Test not trashed
	tasks, err = db.Tasks().
		Trashed(false).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.False(t, task.Trashed, "Trashed(false) returned trashed task")
	}
}

func TestTaskQuerySearch(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test search
	tasks, err := db.Tasks().
		Search("To-Do in Today").
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, tasks, "Search() returned no results")

	// Test search with no results
	tasks, err = db.Tasks().
		Search("xyznonexistent123").
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	assert.Empty(t, tasks)
}

func TestTaskQueryCreatedAfter(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test many years ago - should return all tasks created after that time
	tasks, err := db.Tasks().
		CreatedAfter(YearsAgo(100)).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testTasksIncomplete, "YearsAgo(100) should include all test tasks")

	// Test weeks ago
	_, err = db.Tasks().
		CreatedAfter(WeeksAgo(2)).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)

	// Test months ago
	_, err = db.Tasks().
		CreatedAfter(MonthsAgo(1)).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)

	// Test days ago
	_, err = db.Tasks().
		CreatedAfter(DaysAgo(7)).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
}

func TestTaskQueryCount(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	count, err := db.Tasks().
		Status().Incomplete().
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, testTasksIncomplete, count)
}

func TestTaskQueryFirst(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test First with results
	task, err := db.Tasks().
		Status().Incomplete().
		First(ctx)
	require.NoError(t, err)
	assert.NotNil(t, task)

	// Test First with no results (should return ErrTaskNotFound)
	_, err = db.Tasks().
		WithUUID("nonexistent-uuid").
		First(ctx)
	require.ErrorIs(t, err, ErrTaskNotFound)
}

func TestTaskQueryIncludeItems(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test include items on project
	projects, err := db.Tasks().
		Type().Project().
		Status().Incomplete().
		IncludeItems(true).
		All(ctx)
	require.NoError(t, err)
	if len(projects) > 0 {
		// First project should have items
		assert.NotEmpty(t, projects[0].Items, "IncludeItems(project) returned project without items")
	}

	// Test include items on todo with checklist
	task, err := db.Tasks().
		WithUUID(testUUIDTodoChecklist).
		IncludeItems(true).
		First(ctx)
	require.NoError(t, err)
	assert.Len(t, task.Checklist, 3)
}

func TestTaskQueryOrderByTodayIndex(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	tasks, err := db.Tasks().
		StartDate().Exists(true).
		Start().Anytime().
		Status().Incomplete().
		OrderByTodayIndex().
		All(ctx)
	require.NoError(t, err)

	// Verify tasks are ordered by today_index
	for i := 1; i < len(tasks); i++ {
		assert.GreaterOrEqual(t, tasks[i].TodayIndex, tasks[i-1].TodayIndex,
			"OrderByTodayIndex() not properly ordered at index %d", i)
	}
}

// =============================================================================
// Type-Safe Sub-Builder Tests
// =============================================================================

func TestStatusFilter(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test Incomplete
	tasks, err := db.Tasks().
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Equal(t, StatusIncomplete, task.Status)
	}

	// Test Completed
	tasks, err = db.Tasks().
		Status().Completed().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Equal(t, StatusCompleted, task.Status)
	}

	// Test Canceled
	tasks, err = db.Tasks().
		Status().Canceled().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Equal(t, StatusCanceled, task.Status)
	}
}

func TestStartFilter(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test Inbox
	tasks, err := db.Tasks().
		Start().Inbox().
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Equal(t, "Inbox", task.Start)
	}

	// Test Anytime
	tasks, err = db.Tasks().
		Start().Anytime().
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Equal(t, "Anytime", task.Start)
	}

	// Test Someday
	tasks, err = db.Tasks().
		Start().Someday().
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Equal(t, "Someday", task.Start)
	}
}

func TestTypeFilter(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test Todo
	tasks, err := db.Tasks().
		Type().Todo().
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Equal(t, TaskTypeTodo, task.Type)
	}

	// Test Project
	tasks, err = db.Tasks().
		Type().Project().
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Equal(t, TaskTypeProject, task.Type)
	}

	// Test Heading
	tasks, err = db.Tasks().
		Type().Heading().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Equal(t, TaskTypeHeading, task.Type)
	}
}

func TestDateFilterExists(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test StartDate exists
	tasks, err := db.Tasks().
		StartDate().Exists(true).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.NotNil(t, task.StartDate, "StartDate().Exists(true) returned task without start date")
	}

	// Test StartDate not exists
	tasks, err = db.Tasks().
		StartDate().Exists(false).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Nil(t, task.StartDate, "StartDate().Exists(false) returned task with start date")
	}

	// Test Deadline exists
	tasks, err = db.Tasks().
		Deadline().Exists(true).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.NotNil(t, task.Deadline, "Deadline().Exists(true) returned task without deadline")
	}

	// Test Deadline not exists
	tasks, err = db.Tasks().
		Deadline().Exists(false).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Nil(t, task.Deadline, "Deadline().Exists(false) returned task with deadline")
	}
}

func TestDateFilterRelative(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test StartDate future - should return tasks (may be empty depending on test data)
	_, err := db.Tasks().
		StartDate().Future().
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)

	// Test StartDate past - should return tasks (may be empty depending on test data)
	_, err = db.Tasks().
		StartDate().Past().
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)

	// Test Deadline future
	_, err = db.Tasks().
		Deadline().Future().
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)

	// Test Deadline past
	_, err = db.Tasks().
		Deadline().Past().
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
}

func TestDateFilterComparison(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test Deadline OnOrBefore
	deadline := time.Date(2025, 12, 31, 0, 0, 0, 0, time.Local)
	_, err := db.Tasks().
		Deadline().OnOrBefore(deadline).
		Status().Incomplete().
		Count(ctx)
	require.NoError(t, err)

	// Test Deadline After
	afterDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
	_, err = db.Tasks().
		Deadline().After(afterDate).
		Status().Incomplete().
		Count(ctx)
	require.NoError(t, err)

	// Test StartDate On
	onDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.Local)
	_, err = db.Tasks().
		StartDate().On(onDate).
		Status().Incomplete().
		Count(ctx)
	require.NoError(t, err)
}

func TestSubBuilderChaining(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test complex chaining with multiple sub-builders
	tasks, err := db.Tasks().
		Type().Todo().
		Status().Incomplete().
		Start().Anytime().
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testTodosAnytime)

	// Test chaining with date filters
	count, err := db.Tasks().
		Status().Incomplete().
		StartDate().Exists(true).
		Start().Anytime().
		Count(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 0)

	// Test multiple date filters
	_, err = db.Tasks().
		Status().Incomplete().
		StartDate().Exists(true).
		Deadline().Exists(true).
		All(ctx)
	require.NoError(t, err)
}

// =============================================================================
// Context Trashed Filter Tests
// =============================================================================

// testUUIDContextTrashedTask is a task where task.trashed=0 but project.trashed=1.
const testUUIDContextTrashedTask = "NoQLFamrMMooAELuBznao8"

func TestContextTrashedFilter(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test 1: Without ContextTrashed filter - should include context-trashed tasks
	tasksWithoutFilter, err := db.Tasks().
		Type().Todo().
		Start().Anytime().
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)

	// Find the context-trashed task in results
	foundWithoutFilter := false
	for _, task := range tasksWithoutFilter {
		if task.UUID == testUUIDContextTrashedTask {
			foundWithoutFilter = true
			break
		}
	}
	assert.True(t, foundWithoutFilter,
		"Without ContextTrashed filter, context-trashed task should be included")

	// Test 2: With ContextTrashed(false) - should exclude context-trashed tasks
	tasksWithFilter, err := db.Tasks().
		Type().Todo().
		Start().Anytime().
		Status().Incomplete().
		ContextTrashed(false).
		All(ctx)
	require.NoError(t, err)

	// Verify context-trashed task is NOT in results
	foundWithFilter := false
	for _, task := range tasksWithFilter {
		if task.UUID == testUUIDContextTrashedTask {
			foundWithFilter = true
			break
		}
	}
	assert.False(t, foundWithFilter,
		"With ContextTrashed(false), context-trashed task should be excluded")

	// Test 3: Verify the count difference
	// tasksWithFilter should have one less item (the context-trashed task)
	expectedCount := len(tasksWithoutFilter) - 1
	assert.Len(t, tasksWithFilter, expectedCount,
		"ContextTrashed(false) should filter out exactly 1 context-trashed task")
}

func TestContextTrashedTaskDetails(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Verify the test task exists and has expected properties
	task, err := db.Tasks().
		WithUUID(testUUIDContextTrashedTask).
		First(ctx)
	require.NoError(t, err)

	assert.Equal(t, "Task in Deleted Project", task.Title,
		"Test task should have expected title")
	assert.NotNil(t, task.ProjectTitle,
		"Test task should have a project")
	assert.Equal(t, "Deleted Project", *task.ProjectTitle,
		"Test task should be in 'Deleted Project'")

	// The task itself is not trashed (task.trashed = 0)
	assert.False(t, task.Trashed, "Task itself should not be trashed")
}
