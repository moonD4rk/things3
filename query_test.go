package things3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskQueryChaining(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test method chaining
	tasks, err := db.Tasks().
		WithType(TaskTypeTodo).
		WithStatus(StatusIncomplete).
		WithStart(StartAnytime).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testTodosAnytime)
}

func TestTaskQueryWithUUID(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// First get a valid task UUID from the database
	tasks, err := db.Tasks().WithStatus(StatusIncomplete).All(ctx)
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
		WithType(TaskTypeTodo).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	for _, task := range todos {
		assert.Equal(t, TaskTypeTodo, task.Type)
	}

	// Test projects
	projects, err := db.Tasks().
		WithType(TaskTypeProject).
		WithStatus(StatusIncomplete).
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
	tasks, err := db.Tasks().WithStatus(StatusIncomplete).All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Equal(t, StatusIncomplete, task.Status)
	}

	// Test completed
	tasks, err = db.Tasks().WithStatus(StatusCompleted).All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Equal(t, StatusCompleted, task.Status)
	}

	// Test canceled
	tasks, err = db.Tasks().WithStatus(StatusCanceled).All(ctx)
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
		WithStart(StartInbox).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Equal(t, "Inbox", task.Start)
	}

	// Test Anytime
	tasks, err = db.Tasks().
		WithStart(StartAnytime).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Equal(t, "Anytime", task.Start)
	}

	// Test Someday
	tasks, err = db.Tasks().
		WithStart(StartSomeday).
		WithStatus(StatusIncomplete).
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
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		require.NotNil(t, task.AreaUUID)
		assert.Equal(t, testUUIDArea, *task.AreaUUID)
	}

	// Test with has area
	tasks, err = db.Tasks().
		InArea(true).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.NotNil(t, task.AreaUUID, "InArea(true) returned task without area")
	}

	// Test without area
	tasks, err = db.Tasks().
		InArea(false).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Nil(t, task.AreaUUID, "InArea(false) returned task with area")
	}
}

func TestTaskQueryInProject(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test with specific project
	tasks, err := db.Tasks().
		InProject(testUUIDProject).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testTasksInProject)
}

func TestTaskQueryWithTag(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test with specific tag
	tasks, err := db.Tasks().
		WithTag("Errand").
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)

	// Test with has tags
	tasks, err = db.Tasks().
		WithTag(true).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.NotEmpty(t, task.Tags, "WithTag(true) returned task without tags")
	}
}

func TestTaskQueryWithDeadline(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test has deadline
	tasks, err := db.Tasks().
		WithDeadline(true).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.NotNil(t, task.Deadline, "WithDeadline(true) returned task without deadline")
	}

	// Test no deadline
	tasks, err = db.Tasks().
		WithDeadline(false).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Nil(t, task.Deadline, "WithDeadline(false) returned task with deadline")
	}
}

func TestTaskQueryWithStartDate(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test has start date
	tasks, err := db.Tasks().
		WithStartDate(true).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.NotNil(t, task.StartDate, "WithStartDate(true) returned task without start date")
	}
}

func TestTaskQueryTrashed(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test trashed
	tasks, err := db.Tasks().
		Trashed(true).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.True(t, task.Trashed, "Trashed(true) returned non-trashed task")
	}

	// Test not trashed
	tasks, err = db.Tasks().
		Trashed(false).
		WithStatus(StatusIncomplete).
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
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, tasks, "Search() returned no results")

	// Test search with no results
	tasks, err = db.Tasks().
		Search("xyznonexistent123").
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	assert.Empty(t, tasks)
}

func TestTaskQueryLast(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test last 0 days
	tasks, err := db.Tasks().
		Last("0d").
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	assert.Empty(t, tasks)

	// Test last many years
	tasks, err = db.Tasks().
		Last("100y").
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, tasks, "Last(100y) returned no results")
}

func TestTaskQueryCount(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	count, err := db.Tasks().
		WithStatus(StatusIncomplete).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, testTasksIncomplete, count)
}

func TestTaskQueryFirst(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test First with results
	task, err := db.Tasks().
		WithStatus(StatusIncomplete).
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
		WithType(TaskTypeProject).
		WithStatus(StatusIncomplete).
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
		WithStartDate(true).
		WithStart(StartAnytime).
		WithStatus(StatusIncomplete).
		OrderByTodayIndex().
		All(ctx)
	require.NoError(t, err)

	// Verify tasks are ordered by today_index
	for i := 1; i < len(tasks); i++ {
		assert.GreaterOrEqual(t, tasks[i].TodayIndex, tasks[i-1].TodayIndex,
			"OrderByTodayIndex() not properly ordered at index %d", i)
	}
}
