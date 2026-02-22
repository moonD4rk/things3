package things3

import (
	"context"
	"strings"
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
	assert.ElementsMatch(t, testTodosAnytimeUUIDs, extractUUIDs(tasks))
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
		InArea(testUUIDArea3).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		require.NotNil(t, task.AreaUUID)
		assert.Equal(t, testUUIDArea3, *task.AreaUUID)
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
		InProject(testUUIDProjectInArea1).
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
	require.Len(t, tasks, 1)
	assert.Equal(t, testUUIDTodoInArea1Tags, tasks[0].UUID)

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
	allTasks, err := db.Tasks().
		CreatedAfter(YearsAgo(100)).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, allTasks, testTasksIncomplete, "YearsAgo(100) should include all test tasks")

	// Test weeks ago - test data was created years ago, so recent filters
	// should return fewer results. Validate each task's Created time.
	threshold := WeeksAgo(2)
	recentTasks, err := db.Tasks().
		CreatedAfter(threshold).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(recentTasks), len(allTasks))
	for _, task := range recentTasks {
		assert.True(t, task.Created.After(threshold),
			"Created %v should be after %v", task.Created, threshold)
	}

	// Test months ago
	threshold = MonthsAgo(1)
	recentTasks, err = db.Tasks().
		CreatedAfter(threshold).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(recentTasks), len(allTasks))
	for _, task := range recentTasks {
		assert.True(t, task.Created.After(threshold),
			"Created %v should be after %v", task.Created, threshold)
	}

	// Test days ago
	threshold = DaysAgo(7)
	recentTasks, err = db.Tasks().
		CreatedAfter(threshold).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(recentTasks), len(allTasks))
	for _, task := range recentTasks {
		assert.True(t, task.Created.After(threshold),
			"Created %v should be after %v", task.Created, threshold)
	}
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
		WithUUID(testUUIDTodoInboxChecklist).
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
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	t.Run("Deadline Past", func(t *testing.T) {
		tasks, err := db.Tasks().
			Deadline().Past().Status().Incomplete().All(ctx)
		require.NoError(t, err)
		require.Len(t, tasks, testDeadlinePast)
		for _, task := range tasks {
			if assert.NotNil(t, task.Deadline) {
				assert.False(t, task.Deadline.After(today),
					"Deadline %v should not be after today %v", task.Deadline, today)
			}
		}
	})

	t.Run("Deadline Future", func(t *testing.T) {
		tasks, err := db.Tasks().
			Deadline().Future().Status().Incomplete().All(ctx)
		require.NoError(t, err)
		require.Len(t, tasks, testDeadlineFuture)
		for _, task := range tasks {
			if assert.NotNil(t, task.Deadline) {
				assert.True(t, task.Deadline.After(today),
					"Deadline %v should be after today %v", task.Deadline, today)
			}
		}
	})

	t.Run("StartDate Past", func(t *testing.T) {
		tasks, err := db.Tasks().
			StartDate().Past().Status().Incomplete().All(ctx)
		require.NoError(t, err)
		for _, task := range tasks {
			if assert.NotNil(t, task.StartDate) {
				assert.False(t, task.StartDate.After(today),
					"StartDate %v should not be after today %v", task.StartDate, today)
			}
		}
	})

	t.Run("StartDate Future", func(t *testing.T) {
		tasks, err := db.Tasks().
			StartDate().Future().Status().Incomplete().All(ctx)
		require.NoError(t, err)
		for _, task := range tasks {
			if assert.NotNil(t, task.StartDate) {
				assert.True(t, task.StartDate.After(today),
					"StartDate %v should be after today %v", task.StartDate, today)
			}
		}
	})

	// Cross-validate: Past + Future == Exists(true) for Deadline
	t.Run("Deadline Past and Future partition", func(t *testing.T) {
		pastCount, err := db.Tasks().
			Deadline().Past().Status().Incomplete().Count(ctx)
		require.NoError(t, err)
		futureCount, err := db.Tasks().
			Deadline().Future().Status().Incomplete().Count(ctx)
		require.NoError(t, err)
		require.Equal(t, testDeadlines, pastCount+futureCount)
	})
}

func TestDateFilterComparison(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Pivot date: 3 deadlines before 2025, 1 after (2040-11-04)
	pivot := time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local)

	t.Run("Deadline Before", func(t *testing.T) {
		tasks, err := db.Tasks().
			Deadline().Before(pivot).
			Status().Incomplete().
			All(ctx)
		require.NoError(t, err)
		require.Len(t, tasks, testDeadlinePast)
		for _, task := range tasks {
			if assert.NotNil(t, task.Deadline) {
				assert.True(t, task.Deadline.Before(pivot),
					"Deadline %v should be before %v", task.Deadline, pivot)
			}
		}
	})

	t.Run("Deadline OnOrAfter", func(t *testing.T) {
		tasks, err := db.Tasks().
			Deadline().OnOrAfter(pivot).
			Status().Incomplete().
			All(ctx)
		require.NoError(t, err)
		require.Len(t, tasks, testDeadlineFuture)
		for _, task := range tasks {
			if assert.NotNil(t, task.Deadline) {
				assert.False(t, task.Deadline.Before(pivot),
					"Deadline %v should be on or after %v", task.Deadline, pivot)
			}
		}
	})

	// Cross-validate: Before + OnOrAfter == Exists(true)
	t.Run("Before and OnOrAfter partition", func(t *testing.T) {
		beforeCount, err := db.Tasks().
			Deadline().Before(pivot).Status().Incomplete().Count(ctx)
		require.NoError(t, err)
		onOrAfterCount, err := db.Tasks().
			Deadline().OnOrAfter(pivot).Status().Incomplete().Count(ctx)
		require.NoError(t, err)
		totalCount, err := db.Tasks().
			Deadline().Exists(true).Status().Incomplete().Count(ctx)
		require.NoError(t, err)
		require.Equal(t, totalCount, beforeCount+onOrAfterCount)
	})

	t.Run("Deadline OnOrBefore", func(t *testing.T) {
		deadline := time.Date(2025, 12, 31, 0, 0, 0, 0, time.Local)
		tasks, err := db.Tasks().
			Deadline().OnOrBefore(deadline).
			Status().Incomplete().
			All(ctx)
		require.NoError(t, err)
		require.Len(t, tasks, testDeadlinePast)
		for _, task := range tasks {
			if assert.NotNil(t, task.Deadline) {
				assert.False(t, task.Deadline.After(deadline),
					"Deadline %v should be on or before %v", task.Deadline, deadline)
			}
		}
	})

	t.Run("Deadline After", func(t *testing.T) {
		afterDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
		tasks, err := db.Tasks().
			Deadline().After(afterDate).
			Status().Incomplete().
			All(ctx)
		require.NoError(t, err)
		require.Len(t, tasks, testDeadlineFuture)
		for _, task := range tasks {
			if assert.NotNil(t, task.Deadline) {
				assert.True(t, task.Deadline.After(afterDate),
					"Deadline %v should be after %v", task.Deadline, afterDate)
			}
		}
	})

	t.Run("StartDate On", func(t *testing.T) {
		// 2021-03-28: "To-Do in Today" has this start date
		onDate := time.Date(2021, 3, 28, 0, 0, 0, 0, time.Local)
		tasks, err := db.Tasks().
			StartDate().On(onDate).
			Status().Incomplete().
			All(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, tasks)
		for _, task := range tasks {
			if assert.NotNil(t, task.StartDate) {
				assert.Equal(t, onDate.Year(), task.StartDate.Year())
				assert.Equal(t, onDate.Month(), task.StartDate.Month())
				assert.Equal(t, onDate.Day(), task.StartDate.Day())
			}
		}
	})

	t.Run("StartDate OnOrAfter", func(t *testing.T) {
		threshold := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
		tasks, err := db.Tasks().
			StartDate().OnOrAfter(threshold).
			Status().Incomplete().
			All(ctx)
		require.NoError(t, err)
		for _, task := range tasks {
			if assert.NotNil(t, task.StartDate) {
				assert.False(t, task.StartDate.Before(threshold),
					"StartDate %v should be on or after %v", task.StartDate, threshold)
			}
		}
	})

	// Cross-validate StartDate: Before + OnOrAfter == Exists(true)
	t.Run("StartDate Before and OnOrAfter partition", func(t *testing.T) {
		startPivot := time.Date(2023, 1, 1, 0, 0, 0, 0, time.Local)
		beforeCount, err := db.Tasks().
			StartDate().Before(startPivot).Status().Incomplete().Count(ctx)
		require.NoError(t, err)
		onOrAfterCount, err := db.Tasks().
			StartDate().OnOrAfter(startPivot).Status().Incomplete().Count(ctx)
		require.NoError(t, err)
		totalCount, err := db.Tasks().
			StartDate().Exists(true).Status().Incomplete().Count(ctx)
		require.NoError(t, err)
		require.Equal(t, totalCount, beforeCount+onOrAfterCount)
	})
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
	assert.ElementsMatch(t, testTodosAnytimeUUIDs, extractUUIDs(tasks))

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
		if task.UUID == testUUIDTodoInDeletedProject {
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
		if task.UUID == testUUIDTodoInDeletedProject {
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
		WithUUID(testUUIDTodoInDeletedProject).
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

// =============================================================================
// 0% Coverage Query Method Tests
// =============================================================================

func TestWithUUIDPrefix(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// testUUIDTodoInbox = "DfYoiXcNLQssk9DkSoJV3Y" starts with "DfYo"
	tasks, err := db.Tasks().
		WithUUIDPrefix("DfYo").
		All(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, tasks)

	found := false
	for _, task := range tasks {
		if task.UUID == testUUIDTodoInbox {
			found = true
		}
		assert.True(t, strings.HasPrefix(task.UUID, "DfYo"),
			"WithUUIDPrefix returned task with non-matching UUID: %s", task.UUID)
	}
	assert.True(t, found, "WithUUIDPrefix should find testUUIDTodoInbox")

	// Non-matching prefix
	tasks, err = db.Tasks().
		WithUUIDPrefix("ZZZZZ").
		All(ctx)
	require.NoError(t, err)
	assert.Empty(t, tasks)
}

func TestStopDateFilter(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Completed tasks should have stop dates
	tasks, err := db.Tasks().
		StopDate().Exists(true).
		Status().Completed().
		All(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, tasks, "Completed tasks should have stop dates")
	for _, task := range tasks {
		assert.NotNil(t, task.StopDate, "StopDate().Exists(true) returned task without stop date")
	}

	// Incomplete tasks should not have stop dates
	tasks, err = db.Tasks().
		StopDate().Exists(false).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasks {
		assert.Nil(t, task.StopDate, "StopDate().Exists(false) returned task with stop date")
	}
}

func TestHasProject(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Tasks with a project
	tasksWithProject, err := db.Tasks().
		HasProject(true).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasksWithProject {
		hasProject := task.ProjectUUID != nil || task.HeadingUUID != nil
		assert.True(t, hasProject,
			"HasProject(true) returned task %q without project or heading context", task.UUID)
	}

	// Tasks without a project
	tasksWithoutProject, err := db.Tasks().
		HasProject(false).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, tasksWithoutProject)
}

func TestHasHeading(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Tasks in a heading
	tasksInHeading, err := db.Tasks().
		HasHeading(true).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, task := range tasksInHeading {
		assert.NotNil(t, task.HeadingUUID,
			"HasHeading(true) returned task %q without heading", task.UUID)
	}

	// Tasks not in a heading
	tasksNotInHeading, err := db.Tasks().
		HasHeading(false).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, tasksNotInHeading)
	for _, task := range tasksNotInHeading {
		assert.Nil(t, task.HeadingUUID,
			"HasHeading(false) returned task %q with heading", task.UUID)
	}
}

func TestAreaWithTitle(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	area, err := db.Areas().WithTitle("Area 1").First(ctx)
	require.NoError(t, err)
	assert.Equal(t, "Area 1", area.Title)
	assert.Equal(t, testUUIDArea1, area.UUID)

	// Non-matching title
	_, err = db.Areas().WithTitle("Nonexistent Area").First(ctx)
	require.ErrorIs(t, err, ErrAreaNotFound)
}

func TestAreaVisible(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	visibleAreas, err := db.Areas().Visible(true).All(ctx)
	require.NoError(t, err)

	hiddenAreas, err := db.Areas().Visible(false).All(ctx)
	require.NoError(t, err)

	// Visible + hidden should partition all areas
	allAreas, err := db.Areas().All(ctx)
	require.NoError(t, err)
	require.Equal(t, len(allAreas), len(visibleAreas)+len(hiddenAreas))
}

func TestAreaHasTag(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	areasWithTag, err := db.Areas().HasTag(true).All(ctx)
	require.NoError(t, err)

	areasWithoutTag, err := db.Areas().HasTag(false).All(ctx)
	require.NoError(t, err)

	// Total should account for all areas
	assert.Equal(t, testAreas, len(areasWithTag)+len(areasWithoutTag))
}

func TestTagWithUUID(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	tag, err := db.Tags().WithUUID(testUUIDTagOffice).First(ctx)
	require.NoError(t, err)
	assert.Equal(t, testUUIDTagOffice, tag.UUID)
	assert.Equal(t, "Office", tag.Title)
}

func TestTagWithParent(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// No nested tags in test data, so result should be empty
	tags, err := db.Tags().WithParent("nonexistent").All(ctx)
	require.NoError(t, err)
	assert.Empty(t, tags)
}
