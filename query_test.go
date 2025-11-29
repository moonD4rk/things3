package things3

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskQueryChaining(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Test method chaining
	tasks, err := client.Tasks().
		WithType(TaskTypeTodo).
		WithStatus(StatusIncomplete).
		WithStart(StartAnytime).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testTodosAnytime)
}

func TestTaskQueryWithUUID(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// First get a valid task UUID from the database
	tasks, err := client.Tasks().WithStatus(StatusIncomplete).All(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, tasks, "No tasks in test database")

	targetUUID := tasks[0].UUID

	// Now test WithUUID with the valid UUID
	task, err := client.Tasks().WithUUID(targetUUID).First(ctx)
	require.NoError(t, err)
	assert.Equal(t, targetUUID, task.UUID)
}

func TestTaskQueryWithType(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Test todos
	todos, err := client.Tasks().
		WithType(TaskTypeTodo).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	for _, task := range todos {
		assert.Equal(t, TaskTypeTodo, task.Type)
	}

	// Test projects
	projects, err := client.Tasks().
		WithType(TaskTypeProject).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	for _, task := range projects {
		assert.Equal(t, TaskTypeProject, task.Type)
	}
}

func TestTaskQueryWithStatus(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Test incomplete
	tasks, err := client.Tasks().WithStatus(StatusIncomplete).All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Equal(t, StatusIncomplete, task.Status)
	}

	// Test completed
	tasks, err = client.Tasks().WithStatus(StatusCompleted).All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Equal(t, StatusCompleted, task.Status)
	}

	// Test canceled
	tasks, err = client.Tasks().WithStatus(StatusCanceled).All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Equal(t, StatusCanceled, task.Status)
	}
}

func TestTaskQueryWithStart(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Test Inbox
	tasks, err := client.Tasks().
		WithStart(StartInbox).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Equal(t, "Inbox", task.Start)
	}

	// Test Anytime
	tasks, err = client.Tasks().
		WithStart(StartAnytime).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Equal(t, "Anytime", task.Start)
	}

	// Test Someday
	tasks, err = client.Tasks().
		WithStart(StartSomeday).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Equal(t, "Someday", task.Start)
	}
}

func TestTaskQueryInArea(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Test with specific area
	tasks, err := client.Tasks().
		InArea(testUUIDArea).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		require.NotNil(t, task.AreaUUID)
		assert.Equal(t, testUUIDArea, *task.AreaUUID)
	}

	// Test with has area
	tasks, err = client.Tasks().
		InArea(true).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.NotNil(t, task.AreaUUID, "InArea(true) returned task without area")
	}

	// Test without area
	tasks, err = client.Tasks().
		InArea(false).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Nil(t, task.AreaUUID, "InArea(false) returned task with area")
	}
}

func TestTaskQueryInProject(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Test with specific project
	tasks, err := client.Tasks().
		InProject(testUUIDProject).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, testTasksInProject)
}

func TestTaskQueryWithTag(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Test with specific tag
	tasks, err := client.Tasks().
		WithTag("Errand").
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)

	// Test with has tags
	tasks, err = client.Tasks().
		WithTag(true).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.NotEmpty(t, task.Tags, "WithTag(true) returned task without tags")
	}
}

func TestTaskQueryWithDeadline(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Test has deadline
	tasks, err := client.Tasks().
		WithDeadline(true).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.NotNil(t, task.Deadline, "WithDeadline(true) returned task without deadline")
	}

	// Test no deadline
	tasks, err = client.Tasks().
		WithDeadline(false).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Nil(t, task.Deadline, "WithDeadline(false) returned task with deadline")
	}
}

func TestTaskQueryWithStartDate(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Test has start date
	tasks, err := client.Tasks().
		WithStartDate(true).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.NotNil(t, task.StartDate, "WithStartDate(true) returned task without start date")
	}
}

func TestTaskQueryTrashed(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Test trashed
	tasks, err := client.Tasks().
		Trashed(true).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.True(t, task.Trashed, "Trashed(true) returned non-trashed task")
	}

	// Test not trashed
	tasks, err = client.Tasks().
		Trashed(false).
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.False(t, task.Trashed, "Trashed(false) returned trashed task")
	}
}

func TestTaskQuerySearch(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Test search
	tasks, err := client.Tasks().
		Search("To-Do in Today").
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, tasks, "Search() returned no results")

	// Test search with no results
	tasks, err = client.Tasks().
		Search("xyznonexistent123").
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	assert.Empty(t, tasks)
}

func TestTaskQueryLast(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Test last 0 days
	tasks, err := client.Tasks().
		Last("0d").
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	assert.Empty(t, tasks)

	// Test last many years
	tasks, err = client.Tasks().
		Last("100y").
		WithStatus(StatusIncomplete).
		All(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, tasks, "Last(100y) returned no results")
}

func TestTaskQueryCount(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	count, err := client.Tasks().
		WithStatus(StatusIncomplete).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, testTasksIncomplete, count)
}

func TestTaskQueryFirst(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Test First with results
	task, err := client.Tasks().
		WithStatus(StatusIncomplete).
		First(ctx)
	require.NoError(t, err)
	assert.NotNil(t, task)

	// Test First with no results (should return ErrTaskNotFound)
	_, err = client.Tasks().
		WithUUID("nonexistent-uuid").
		First(ctx)
	require.ErrorIs(t, err, ErrTaskNotFound)
}

func TestTaskQueryIncludeItems(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Test include items on project
	projects, err := client.Tasks().
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
	task, err := client.Tasks().
		WithUUID(testUUIDTodoChecklist).
		IncludeItems(true).
		First(ctx)
	require.NoError(t, err)
	assert.Len(t, task.Checklist, 3)
}

func TestTaskQueryOrderByTodayIndex(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	tasks, err := client.Tasks().
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
