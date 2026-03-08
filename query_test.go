package things3

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// TodoQuery Tests
// =============================================================================

func TestTodoQueryChaining(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	todos, err := db.Todos().
		Status().Incomplete().
		Start().Anytime().
		All(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, testTodosAnytimeUUIDs, extractTodoUUIDs(todos))
}

func TestTodoQueryWithUUID(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	todo, err := db.Todos().WithUUID(testUUIDTodoInbox).First(ctx)
	require.NoError(t, err)
	assert.Equal(t, testUUIDTodoInbox, todo.UUID)
}

func TestTodoQueryWithStatus(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	// Test incomplete
	todos, err := db.Todos().Status().Incomplete().All(ctx)
	require.NoError(t, err)
	for _, todo := range todos {
		assert.Equal(t, StatusIncomplete, todo.Status)
	}

	// Test completed
	todos, err = db.Todos().Status().Completed().All(ctx)
	require.NoError(t, err)
	for _, todo := range todos {
		assert.Equal(t, StatusCompleted, todo.Status)
	}

	// Test canceled
	todos, err = db.Todos().Status().Canceled().All(ctx)
	require.NoError(t, err)
	for _, todo := range todos {
		assert.Equal(t, StatusCanceled, todo.Status)
	}
}

func TestTodoQueryWithStart(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	// Test Inbox
	todos, err := db.Todos().
		Start().Inbox().
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, todo := range todos {
		assert.Equal(t, StartInbox, todo.Start)
	}

	// Test Anytime
	todos, err = db.Todos().
		Start().Anytime().
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, todo := range todos {
		assert.Equal(t, StartAnytime, todo.Start)
	}

	// Test Someday
	todos, err = db.Todos().
		Start().Someday().
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, todo := range todos {
		assert.Equal(t, StartSomeday, todo.Start)
	}
}

func TestTodoQueryInArea(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	// Test with specific area
	todos, err := db.Todos().
		InArea(testUUIDArea3).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, todo := range todos {
		assert.Equal(t, testUUIDArea3, todo.AreaUUID)
	}

	// Test with has area
	todos, err = db.Todos().
		HasArea(true).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, todo := range todos {
		assert.NotEmpty(t, todo.AreaUUID, "HasArea(true) returned todo without area")
	}

	// Test without area
	todos, err = db.Todos().
		HasArea(false).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, todo := range todos {
		assert.Empty(t, todo.AreaUUID, "HasArea(false) returned todo with area")
	}
}

func TestTodoQueryInProject(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	todos, err := db.Todos().
		InProject(testUUIDProjectInArea1).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, todos, testTodosInProject)
}

func TestTodoQueryInTag(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	// Test with specific tag
	todos, err := db.Todos().
		InTag("Errand").
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	require.Len(t, todos, 1)
	assert.Equal(t, testUUIDTodoInArea1Tags, todos[0].UUID)

	// Test with has tag
	todos, err = db.Todos().
		HasTag(true).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, todo := range todos {
		assert.NotEmpty(t, todo.Tags, "HasTag(true) returned todo without tags")
	}
}

func TestTodoQueryWithDeadline(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	// Test has deadline
	todos, err := db.Todos().
		Deadline().Exists(true).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, todo := range todos {
		assert.NotNil(t, todo.Deadline, "Deadline().Exists(true) returned todo without deadline")
	}

	// Test no deadline
	todos, err = db.Todos().
		Deadline().Exists(false).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, todo := range todos {
		assert.Nil(t, todo.Deadline, "Deadline().Exists(false) returned todo with deadline")
	}
}

func TestTodoQueryWithStartDate(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	todos, err := db.Todos().
		StartDate().Exists(true).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, todo := range todos {
		assert.NotNil(t, todo.StartDate, "StartDate().Exists(true) returned todo without start date")
	}
}

func TestTodoQueryTrashed(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	// Test trashed
	todos, err := db.Todos().
		Trashed(true).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, todo := range todos {
		assert.True(t, todo.Trashed, "Trashed(true) returned non-trashed todo")
	}

	// Test not trashed
	todos, err = db.Todos().
		Trashed(false).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, todo := range todos {
		assert.False(t, todo.Trashed, "Trashed(false) returned trashed todo")
	}
}

func TestTodoQuerySearch(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	todos, err := db.Todos().
		Search("To-Do in Today").
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, todos, "Search() returned no results")

	// Non-matching search
	todos, err = db.Todos().
		Search("xyznonexistent123").
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	assert.Empty(t, todos)
}

func TestTodoQueryCreatedAfter(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	// YearsAgo(100) should include all test todos
	allTodos, err := db.Todos().
		CreatedAfter(YearsAgo(100)).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	assert.Len(t, allTodos, testTodosIncomplete)

	// Recent filter should return fewer results
	threshold := WeeksAgo(2)
	recentTodos, err := db.Todos().
		CreatedAfter(threshold).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(recentTodos), len(allTodos))
	for _, todo := range recentTodos {
		assert.True(t, todo.CreatedAt.After(threshold),
			"CreatedAt %v should be after %v", todo.CreatedAt, threshold)
	}
}

func TestTodoQueryCount(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	count, err := db.Todos().
		Status().Incomplete().
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, testTodosIncomplete, count)
}

func TestTodoQueryFirst(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	// Test First with results
	todo, err := db.Todos().
		Status().Incomplete().
		First(ctx)
	require.NoError(t, err)
	assert.NotNil(t, todo)

	// Test First with no results
	_, err = db.Todos().
		WithUUID("nonexistent-uuid").
		First(ctx)
	require.ErrorIs(t, err, ErrTodoNotFound)
}

func TestTodoQueryIncludeChecklist(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	// First() auto-loads checklist
	todo, err := db.Todos().
		WithUUID(testUUIDTodoInboxChecklist).
		First(ctx)
	require.NoError(t, err)
	assert.Len(t, todo.Checklist, 3)

	// All() without IncludeChecklist does not load checklist
	todos, err := db.Todos().
		WithUUID(testUUIDTodoInboxChecklist).
		All(ctx)
	require.NoError(t, err)
	require.Len(t, todos, 1)
	assert.Empty(t, todos[0].Checklist)

	// All() with IncludeChecklist loads checklist
	todos, err = db.Todos().
		WithUUID(testUUIDTodoInboxChecklist).
		IncludeChecklist().
		All(ctx)
	require.NoError(t, err)
	require.Len(t, todos, 1)
	assert.Len(t, todos[0].Checklist, 3)
}

func TestTodoQueryOrderByTodayIndex(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	_, err := db.Todos().
		StartDate().Exists(true).
		Start().Anytime().
		Status().Incomplete().
		OrderByTodayIndex().
		All(ctx)
	require.NoError(t, err)
}

func TestTodoQueryLimit(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	// Without limit - should return all incomplete todos
	all, err := db.Todos().Status().Incomplete().All(ctx)
	require.NoError(t, err)
	require.Greater(t, len(all), 3, "need enough todos to test limit")

	// With limit
	limited, err := db.Todos().Status().Incomplete().Limit(3).All(ctx)
	require.NoError(t, err)
	assert.Len(t, limited, 3)

	// Limit larger than result set returns all
	big, err := db.Todos().Status().Incomplete().Limit(1000).All(ctx)
	require.NoError(t, err)
	assert.Len(t, big, len(all))
}

// =============================================================================
// ProjectQuery Tests
// =============================================================================

func TestProjectQueryAll(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	projects, err := db.Projects().
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, projects)
}

func TestProjectQueryFirst(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	project, err := db.Projects().
		WithUUID(testUUIDProjectInArea1).
		First(ctx)
	require.NoError(t, err)
	assert.Equal(t, testUUIDProjectInArea1, project.UUID)

	// Not found
	_, err = db.Projects().
		WithUUID("nonexistent-uuid").
		First(ctx)
	require.ErrorIs(t, err, ErrProjectNotFound)
}

func TestProjectQueryInArea(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	projects, err := db.Projects().
		InArea(testUUIDArea1).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, project := range projects {
		assert.Equal(t, testUUIDArea1, project.AreaUUID)
	}
}

// =============================================================================
// HeadingQuery Tests
// =============================================================================

func TestHeadingQueryAll(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	headings, err := db.Headings().All(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, headings)
}

func TestHeadingQueryInProject(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	headings, err := db.Headings().
		InProject(testUUIDProjectInArea1).
		All(ctx)
	require.NoError(t, err)
	for _, heading := range headings {
		assert.Equal(t, testUUIDProjectInArea1, heading.ProjectUUID)
	}
}

func TestHeadingQueryFirst(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	// Not found
	_, err := db.Headings().
		WithUUID("nonexistent-uuid").
		First(ctx)
	require.ErrorIs(t, err, ErrHeadingNotFound)
}

// =============================================================================
// Date Filter Tests
// =============================================================================

func TestDateFilterExists(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	// Test StartDate exists
	todos, err := db.Todos().
		StartDate().Exists(true).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, todo := range todos {
		assert.NotNil(t, todo.StartDate, "StartDate().Exists(true) returned todo without start date")
	}

	// Test StartDate not exists
	todos, err = db.Todos().
		StartDate().Exists(false).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, todo := range todos {
		assert.Nil(t, todo.StartDate, "StartDate().Exists(false) returned todo with start date")
	}

	// Test Deadline exists
	todos, err = db.Todos().
		Deadline().Exists(true).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, todo := range todos {
		assert.NotNil(t, todo.Deadline, "Deadline().Exists(true) returned todo without deadline")
	}

	// Test Deadline not exists
	todos, err = db.Todos().
		Deadline().Exists(false).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, todo := range todos {
		assert.Nil(t, todo.Deadline, "Deadline().Exists(false) returned todo with deadline")
	}
}

func TestDateFilterRelative(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	t.Run("Deadline Past", func(t *testing.T) {
		todos, err := db.Todos().
			Deadline().Past().Status().Incomplete().All(ctx)
		require.NoError(t, err)
		require.Len(t, todos, testDeadlinePast)
		for _, todo := range todos {
			if assert.NotNil(t, todo.Deadline) {
				assert.False(t, todo.Deadline.After(today),
					"Deadline %v should not be after today %v", todo.Deadline, today)
			}
		}
	})

	t.Run("Deadline Future", func(t *testing.T) {
		todos, err := db.Todos().
			Deadline().Future().Status().Incomplete().All(ctx)
		require.NoError(t, err)
		require.Len(t, todos, testDeadlineFuture)
		for _, todo := range todos {
			if assert.NotNil(t, todo.Deadline) {
				assert.True(t, todo.Deadline.After(today),
					"Deadline %v should be after today %v", todo.Deadline, today)
			}
		}
	})

	t.Run("StartDate Past", func(t *testing.T) {
		todos, err := db.Todos().
			StartDate().Past().Status().Incomplete().All(ctx)
		require.NoError(t, err)
		for _, todo := range todos {
			if assert.NotNil(t, todo.StartDate) {
				assert.False(t, todo.StartDate.After(today),
					"StartDate %v should not be after today %v", todo.StartDate, today)
			}
		}
	})

	t.Run("StartDate Future", func(t *testing.T) {
		todos, err := db.Todos().
			StartDate().Future().Status().Incomplete().All(ctx)
		require.NoError(t, err)
		for _, todo := range todos {
			if assert.NotNil(t, todo.StartDate) {
				assert.True(t, todo.StartDate.After(today),
					"StartDate %v should be after today %v", todo.StartDate, today)
			}
		}
	})

	// Cross-validate: Past + Future == Exists(true) for Deadline
	t.Run("Deadline Past and Future partition", func(t *testing.T) {
		pastCount, err := db.Todos().
			Deadline().Past().Status().Incomplete().Count(ctx)
		require.NoError(t, err)
		futureCount, err := db.Todos().
			Deadline().Future().Status().Incomplete().Count(ctx)
		require.NoError(t, err)
		require.Equal(t, testDeadlines, pastCount+futureCount)
	})
}

func TestDateFilterComparison(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	// Pivot date: 3 deadlines before 2025, 1 after (2040-11-04)
	pivot := time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local)

	t.Run("Deadline Before", func(t *testing.T) {
		todos, err := db.Todos().
			Deadline().Before(pivot).
			Status().Incomplete().
			All(ctx)
		require.NoError(t, err)
		require.Len(t, todos, testDeadlinePast)
		for _, todo := range todos {
			if assert.NotNil(t, todo.Deadline) {
				assert.True(t, todo.Deadline.Before(pivot),
					"Deadline %v should be before %v", todo.Deadline, pivot)
			}
		}
	})

	t.Run("Deadline OnOrAfter", func(t *testing.T) {
		todos, err := db.Todos().
			Deadline().OnOrAfter(pivot).
			Status().Incomplete().
			All(ctx)
		require.NoError(t, err)
		require.Len(t, todos, testDeadlineFuture)
		for _, todo := range todos {
			if assert.NotNil(t, todo.Deadline) {
				assert.False(t, todo.Deadline.Before(pivot),
					"Deadline %v should be on or after %v", todo.Deadline, pivot)
			}
		}
	})

	// Cross-validate: Before + OnOrAfter == Exists(true)
	t.Run("Before and OnOrAfter partition", func(t *testing.T) {
		beforeCount, err := db.Todos().
			Deadline().Before(pivot).Status().Incomplete().Count(ctx)
		require.NoError(t, err)
		onOrAfterCount, err := db.Todos().
			Deadline().OnOrAfter(pivot).Status().Incomplete().Count(ctx)
		require.NoError(t, err)
		totalCount, err := db.Todos().
			Deadline().Exists(true).Status().Incomplete().Count(ctx)
		require.NoError(t, err)
		require.Equal(t, totalCount, beforeCount+onOrAfterCount)
	})

	t.Run("Deadline OnOrBefore", func(t *testing.T) {
		deadline := time.Date(2025, 12, 31, 0, 0, 0, 0, time.Local)
		todos, err := db.Todos().
			Deadline().OnOrBefore(deadline).
			Status().Incomplete().
			All(ctx)
		require.NoError(t, err)
		require.Len(t, todos, testDeadlinePast)
		for _, todo := range todos {
			if assert.NotNil(t, todo.Deadline) {
				assert.False(t, todo.Deadline.After(deadline),
					"Deadline %v should be on or before %v", todo.Deadline, deadline)
			}
		}
	})

	t.Run("Deadline After", func(t *testing.T) {
		afterDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
		todos, err := db.Todos().
			Deadline().After(afterDate).
			Status().Incomplete().
			All(ctx)
		require.NoError(t, err)
		require.Len(t, todos, testDeadlineFuture)
		for _, todo := range todos {
			if assert.NotNil(t, todo.Deadline) {
				assert.True(t, todo.Deadline.After(afterDate),
					"Deadline %v should be after %v", todo.Deadline, afterDate)
			}
		}
	})

	t.Run("StartDate On", func(t *testing.T) {
		// 2021-03-28: "To-Do in Today" has this start date
		onDate := time.Date(2021, 3, 28, 0, 0, 0, 0, time.Local)
		todos, err := db.Todos().
			StartDate().On(onDate).
			Status().Incomplete().
			All(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, todos)
		for _, todo := range todos {
			if assert.NotNil(t, todo.StartDate) {
				assert.Equal(t, onDate.Year(), todo.StartDate.Year())
				assert.Equal(t, onDate.Month(), todo.StartDate.Month())
				assert.Equal(t, onDate.Day(), todo.StartDate.Day())
			}
		}
	})

	t.Run("StartDate OnOrAfter", func(t *testing.T) {
		threshold := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
		todos, err := db.Todos().
			StartDate().OnOrAfter(threshold).
			Status().Incomplete().
			All(ctx)
		require.NoError(t, err)
		for _, todo := range todos {
			if assert.NotNil(t, todo.StartDate) {
				assert.False(t, todo.StartDate.Before(threshold),
					"StartDate %v should be on or after %v", todo.StartDate, threshold)
			}
		}
	})

	// Cross-validate StartDate: Before + OnOrAfter == Exists(true)
	t.Run("StartDate Before and OnOrAfter partition", func(t *testing.T) {
		startPivot := time.Date(2023, 1, 1, 0, 0, 0, 0, time.Local)
		beforeCount, err := db.Todos().
			StartDate().Before(startPivot).Status().Incomplete().Count(ctx)
		require.NoError(t, err)
		onOrAfterCount, err := db.Todos().
			StartDate().OnOrAfter(startPivot).Status().Incomplete().Count(ctx)
		require.NoError(t, err)
		totalCount, err := db.Todos().
			StartDate().Exists(true).Status().Incomplete().Count(ctx)
		require.NoError(t, err)
		require.Equal(t, totalCount, beforeCount+onOrAfterCount)
	})
}

// =============================================================================
// Sub-Builder Chaining Tests
// =============================================================================

func TestSubBuilderChaining(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	// Test complex chaining with multiple sub-builders
	todos, err := db.Todos().
		Status().Incomplete().
		Start().Anytime().
		All(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, testTodosAnytimeUUIDs, extractTodoUUIDs(todos))

	// Test chaining with date filters
	count, err := db.Todos().
		Status().Incomplete().
		StartDate().Exists(true).
		Start().Anytime().
		Count(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 0)

	// Test multiple date filters
	_, err = db.Todos().
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
	ctx := t.Context()

	// Without ContextTrashed filter - should include context-trashed todos
	todosWithoutFilter, err := db.Todos().
		Start().Anytime().
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)

	foundWithoutFilter := false
	for _, todo := range todosWithoutFilter {
		if todo.UUID == testUUIDTodoInDeletedProject {
			foundWithoutFilter = true
			break
		}
	}
	assert.True(t, foundWithoutFilter,
		"Without ContextTrashed filter, context-trashed todo should be included")

	// With ContextTrashed(false) - should exclude context-trashed todos
	todosWithFilter, err := db.Todos().
		Start().Anytime().
		Status().Incomplete().
		ContextTrashed(false).
		All(ctx)
	require.NoError(t, err)

	foundWithFilter := false
	for _, todo := range todosWithFilter {
		if todo.UUID == testUUIDTodoInDeletedProject {
			foundWithFilter = true
			break
		}
	}
	assert.False(t, foundWithFilter,
		"With ContextTrashed(false), context-trashed todo should be excluded")

	expectedCount := len(todosWithoutFilter) - 1
	assert.Len(t, todosWithFilter, expectedCount,
		"ContextTrashed(false) should filter out exactly 1 context-trashed todo")
}

func TestContextTrashedTodoDetails(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	todo, err := db.Todos().
		WithUUID(testUUIDTodoInDeletedProject).
		First(ctx)
	require.NoError(t, err)

	assert.Equal(t, "Task in Deleted Project", todo.Title)
	assert.NotEmpty(t, todo.ProjectTitle)
	assert.Equal(t, "Deleted Project", todo.ProjectTitle)
	assert.False(t, todo.Trashed, "Todo itself should not be trashed")
}

// =============================================================================
// StopDate Filter Tests
// =============================================================================

func TestStopDateFilter(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	// Completed todos should have stop dates (mapped to CompletedAt)
	todos, err := db.Todos().
		StopDate().Exists(true).
		Status().Completed().
		All(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, todos, "Completed todos should have stop dates")
	for _, todo := range todos {
		assert.NotNil(t, todo.CompletedAt, "Completed todo should have CompletedAt set")
	}

	// Incomplete todos should not have stop dates
	todos, err = db.Todos().
		StopDate().Exists(false).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, todo := range todos {
		assert.Nil(t, todo.CompletedAt, "Incomplete todo should not have CompletedAt")
		assert.Nil(t, todo.CanceledAt, "Incomplete todo should not have CanceledAt")
	}
}

// =============================================================================
// Relation Filter Tests
// =============================================================================

func TestHasProject(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	// Todos with a project
	todosWithProject, err := db.Todos().
		HasProject(true).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, todo := range todosWithProject {
		hasProject := todo.ProjectUUID != "" || todo.HeadingUUID != ""
		assert.True(t, hasProject,
			"HasProject(true) returned todo %q without project or heading context", todo.UUID)
	}

	// Todos without a project
	todosWithoutProject, err := db.Todos().
		HasProject(false).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, todosWithoutProject)
}

func TestHasHeading(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	// Todos in a heading
	todosInHeading, err := db.Todos().
		HasHeading(true).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	for _, todo := range todosInHeading {
		assert.NotEmpty(t, todo.HeadingUUID,
			"HasHeading(true) returned todo %q without heading", todo.UUID)
	}

	// Todos not in a heading
	todosNotInHeading, err := db.Todos().
		HasHeading(false).
		Status().Incomplete().
		All(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, todosNotInHeading)
	for _, todo := range todosNotInHeading {
		assert.Empty(t, todo.HeadingUUID,
			"HasHeading(false) returned todo %q with heading", todo.UUID)
	}
}

// =============================================================================
// Area Query Tests
// =============================================================================

func TestAreaWithTitle(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

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
	ctx := t.Context()

	visibleAreas, err := db.Areas().Visible(true).All(ctx)
	require.NoError(t, err)

	hiddenAreas, err := db.Areas().Visible(false).All(ctx)
	require.NoError(t, err)

	allAreas, err := db.Areas().All(ctx)
	require.NoError(t, err)
	require.Equal(t, len(allAreas), len(visibleAreas)+len(hiddenAreas))
}

func TestAreaHasTag(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	areasWithTag, err := db.Areas().HasTag(true).All(ctx)
	require.NoError(t, err)

	areasWithoutTag, err := db.Areas().HasTag(false).All(ctx)
	require.NoError(t, err)

	assert.Equal(t, testAreas, len(areasWithTag)+len(areasWithoutTag))
}

// =============================================================================
// Tag Query Tests
// =============================================================================

func TestTagWithUUID(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	tag, err := db.Tags().WithUUID(testUUIDTagOffice).First(ctx)
	require.NoError(t, err)
	assert.Equal(t, testUUIDTagOffice, tag.UUID)
	assert.Equal(t, "Office", tag.Title)
}

func TestTagWithParent(t *testing.T) {
	db := newTestDB(t)
	ctx := t.Context()

	tags, err := db.Tags().WithParent("nonexistent").All(ctx)
	require.NoError(t, err)
	assert.Empty(t, tags)
}
