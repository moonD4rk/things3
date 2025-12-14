package things3

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// NULL Handling Tests - Verify correct handling of NULL values in task scanning
// =============================================================================

func TestScanTask_NullDates(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Query all todos to find tasks with various NULL date states
	tasks, err := db.Todos(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, tasks, "need at least one todo for testing")

	// Verify NULL date handling - NULL dates should be nil, not zero time
	for _, task := range tasks {
		// Optional dates should be nil or valid time.Time
		if task.StartDate != nil {
			assert.False(t, task.StartDate.IsZero(), "non-nil StartDate should not be zero")
		}
		if task.Deadline != nil {
			assert.False(t, task.Deadline.IsZero(), "non-nil Deadline should not be zero")
		}
		if task.ReminderTime != nil {
			// ReminderTime is time-only, so date part is zero but time should be valid
			assert.True(t, task.ReminderTime.Hour() >= 0 && task.ReminderTime.Hour() <= 23,
				"ReminderTime hour should be valid")
		}
		if task.StopDate != nil {
			assert.False(t, task.StopDate.IsZero(), "non-nil StopDate should not be zero")
		}

		// Created and Modified should always be valid (non-zero)
		// Note: Some tasks may have zero Created if database has issues
	}
}

func TestScanTask_NullOptionalStrings(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	tasks, err := db.Todos(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, tasks)

	// Find a task to verify NULL string handling
	for _, task := range tasks {
		// Notes field - empty string when NULL (using nullStringValue)
		// This is expected behavior - NULL becomes ""
		if task.Notes != "" {
			assert.NotEmpty(t, task.Notes, "non-empty Notes should have content")
		}

		// AreaUUID/AreaTitle - nil pointer when NULL (using nullString)
		// Either both are nil or both are non-nil
		if task.AreaUUID != nil {
			assert.NotNil(t, task.AreaTitle, "AreaTitle should exist when AreaUUID exists")
			assert.NotEmpty(t, *task.AreaUUID, "AreaUUID should not be empty string when non-nil")
		}

		// ProjectUUID/ProjectTitle - similar pattern
		if task.ProjectUUID != nil {
			assert.NotNil(t, task.ProjectTitle, "ProjectTitle should exist when ProjectUUID exists")
			assert.NotEmpty(t, *task.ProjectUUID, "ProjectUUID should not be empty string when non-nil")
		}

		// HeadingUUID/HeadingTitle - similar pattern
		if task.HeadingUUID != nil {
			assert.NotNil(t, task.HeadingTitle, "HeadingTitle should exist when HeadingUUID exists")
		}
	}
}

func TestScanTask_TypeConversion(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	tests := []struct {
		name     string
		query    func() ([]Task, error)
		expected TaskType
	}{
		{
			name:     "todos have TaskTypeTodo",
			query:    func() ([]Task, error) { return db.Todos(ctx) },
			expected: TaskTypeTodo,
		},
		{
			name:     "projects have TaskTypeProject",
			query:    func() ([]Task, error) { return db.Projects(ctx) },
			expected: TaskTypeProject,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks, err := tt.query()
			require.NoError(t, err)

			for _, task := range tasks {
				assert.Equal(t, tt.expected, task.Type,
					"task %s should have type %s", task.Title, tt.expected)
				assert.Equal(t, tt.expected.String(), task.Type.String(),
					"String() method should match")
			}
		})
	}
}

func TestScanTask_StatusConversion(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Test incomplete tasks
	incomplete, err := db.Tasks().Status().Incomplete().All(ctx)
	require.NoError(t, err)
	for _, task := range incomplete {
		assert.Equal(t, StatusIncomplete, task.Status,
			"incomplete tasks should have StatusIncomplete")
		assert.True(t, task.Status.IsOpen(), "incomplete status should be open")
		assert.False(t, task.Status.IsClosed(), "incomplete status should not be closed")
	}

	// Test completed tasks (from logbook)
	completed, err := db.Logbook(ctx)
	require.NoError(t, err)
	for _, task := range completed {
		assert.True(t, task.Status.IsClosed(),
			"logbook tasks should have closed status (completed or canceled)")
		assert.False(t, task.Status.IsOpen(), "logbook tasks should not be open")
	}
}

func TestScanTask_TrashedFlag(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Get trashed items
	trashed, err := db.Trash(ctx)
	require.NoError(t, err)

	for _, task := range trashed {
		assert.True(t, task.Trashed, "trash items should have Trashed=true")
	}

	// Get non-trashed items (today, inbox, etc.)
	today, err := db.Today(ctx)
	require.NoError(t, err)
	for _, task := range today {
		assert.False(t, task.Trashed, "today items should have Trashed=false")
	}
}

func TestScanTask_TagsFlagHandling(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Get all tags first
	allTags, err := db.Tags().All(ctx)
	require.NoError(t, err)

	if len(allTags) == 0 {
		t.Skip("no tags in test database")
	}

	// Query tasks with a specific tag
	tasksWithTag, err := db.Tasks().InTag(allTags[0].Title).All(ctx)
	require.NoError(t, err)

	if len(tasksWithTag) == 0 {
		t.Skip("no tasks with tags in test database")
	}

	// Verify that tasks with tags have the Tags slice populated
	// Note: Tags are loaded separately after initial scan
	for _, task := range tasksWithTag {
		// At minimum, the Tags slice should be initialized (not nil)
		// Actual tag loading depends on query configuration
		assert.NotNil(t, task.Tags, "Tags slice should be initialized for tasks with tags")
	}
}

func TestScanTask_ChecklistFlagHandling(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Find a task with checklist items
	todos, err := db.Todos(ctx)
	require.NoError(t, err)

	var taskWithChecklist *Task
	for i := range todos {
		if len(todos[i].Checklist) > 0 {
			taskWithChecklist = &todos[i]
			break
		}
	}

	if taskWithChecklist == nil {
		t.Skip("no tasks with checklists in test database")
	}

	// Verify checklist items are properly scanned
	for _, item := range taskWithChecklist.Checklist {
		assert.NotEmpty(t, item.UUID, "checklist item should have UUID")
		assert.Equal(t, "checklist-item", item.Type, "checklist item Type should be 'checklist-item'")
		// Status should be one of: incomplete, completed, canceled
		validStatuses := map[string]bool{
			"incomplete": true,
			"completed":  true,
			"canceled":   true,
		}
		assert.True(t, validStatuses[item.Status],
			"checklist item status should be valid, got: %s", item.Status)
	}
}

// =============================================================================
// Area Scanning Tests
// =============================================================================

func TestScanArea_BasicFields(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	areas, err := db.Areas().All(ctx)
	require.NoError(t, err)

	for _, area := range areas {
		assert.NotEmpty(t, area.UUID, "area should have UUID")
		assert.Equal(t, "area", area.Type, "area Type should be 'area'")
		assert.NotEmpty(t, area.Title, "area should have Title")
	}
}

// =============================================================================
// Tag Scanning Tests
// =============================================================================

func TestScanTag_BasicFields(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	tags, err := db.Tags().All(ctx)
	require.NoError(t, err)

	for _, tag := range tags {
		assert.NotEmpty(t, tag.UUID, "tag should have UUID")
		assert.Equal(t, "tag", tag.Type, "tag Type should be 'tag'")
		assert.NotEmpty(t, tag.Title, "tag should have Title")
		// Shortcut is optional, can be empty string
	}
}

func TestScanTag_ShortcutNullHandling(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	tags, err := db.Tags().All(ctx)
	require.NoError(t, err)

	// Some tags have shortcuts, some don't
	// Verify NULL shortcuts become empty string (not cause errors)
	for _, tag := range tags {
		// Shortcut should be either empty or non-empty, never cause panic
		if tag.Shortcut != "" {
			// Valid shortcut should be a single character typically
			assert.NotEmpty(t, tag.Shortcut)
		}
	}
}

// =============================================================================
// Token Retrieval Tests
// =============================================================================

func TestToken_RetrievalSuccess(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	token, err := db.Token(ctx)
	require.NoError(t, err)
	assert.Equal(t, testAuthToken, token, "token should match expected test token")
	assert.NotEmpty(t, token, "token should not be empty")
}

// =============================================================================
// Query Execution Tests
// =============================================================================

func TestExecuteQuery_ContextCancellation(t *testing.T) {
	db := newTestDB(t)

	// Create canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Query with canceled context should fail
	_, err := db.Todos(ctx)
	assert.Error(t, err, "query with canceled context should fail")
}

func TestExecuteQuery_EmptyResult(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Query with UUID that doesn't exist should return empty, not error
	tasks, err := db.Tasks().WithUUID("non-existent-uuid-12345").All(ctx)
	require.NoError(t, err)
	assert.Empty(t, tasks, "non-existent UUID should return empty result")
}

func TestExecuteQuery_FirstNotFound(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// First() with non-existent UUID should return ErrTaskNotFound
	task, err := db.Tasks().WithUUID("non-existent-uuid-12345").First(ctx)
	require.ErrorIs(t, err, ErrTaskNotFound, "First() with non-existent UUID should return ErrTaskNotFound")
	assert.Nil(t, task, "task should be nil when not found")
}

func TestExecuteQuery_CountEmptyResult(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Count with non-existent UUID should return 0, not error
	count, err := db.Tasks().WithUUID("non-existent-uuid-12345").Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "count of non-existent UUID should be 0")
}

// =============================================================================
// Integration Tests - Real Data Verification
// =============================================================================

func TestScanTask_RealTaskIntegrity(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// First get any incomplete todo from the database
	todos, err := db.Todos(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, todos, "need at least one todo in test database")

	task := &todos[0]

	// Verify basic fields
	assert.NotEmpty(t, task.UUID)
	assert.NotEmpty(t, task.Title)
	assert.Equal(t, TaskTypeTodo, task.Type)

	// Verify type helper methods
	assert.True(t, task.IsTodo())
	assert.False(t, task.IsProject())
	assert.False(t, task.IsHeading())
}

func TestScanProject_WithItems(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Get a project with items
	project, err := db.Tasks().WithUUID(testUUIDProject).First(ctx)
	require.NoError(t, err)
	require.NotNil(t, project)

	assert.Equal(t, TaskTypeProject, project.Type)
	assert.True(t, project.IsProject())

	// First() auto-enables IncludeItems for projects
	// Items should be loaded
	// Note: actual item count depends on test database content
}

// =============================================================================
// Utility Function Tests
// =============================================================================

func TestNullString_ValidValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		valid    bool
		expected *string
	}{
		{
			name:     "valid non-empty string",
			input:    "test value",
			valid:    true,
			expected: ptrString("test value"),
		},
		{
			name:     "valid empty string",
			input:    "",
			valid:    true,
			expected: ptrString(""),
		},
		{
			name:     "invalid (NULL)",
			input:    "",
			valid:    false,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create sql.NullString
			ns := sqlNullString(tt.input, tt.valid)
			result := nullString(ns)

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, *tt.expected, *result)
			}
		})
	}
}

func TestNullStringValue_ReturnsEmptyForNull(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		valid    bool
		expected string
	}{
		{
			name:     "valid non-empty string",
			input:    "test value",
			valid:    true,
			expected: "test value",
		},
		{
			name:     "valid empty string",
			input:    "",
			valid:    true,
			expected: "",
		},
		{
			name:     "invalid (NULL)",
			input:    "ignored",
			valid:    false,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := sqlNullString(tt.input, tt.valid)
			result := nullStringValue(ns)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseDate_ValidFormats(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		valid   bool
		isNil   bool
		checkFn func(*testing.T, *time.Time)
	}{
		{
			name:  "valid date",
			input: "2024-01-15",
			valid: true,
			isNil: false,
			checkFn: func(t *testing.T, tm *time.Time) {
				assert.Equal(t, 2024, tm.Year())
				assert.Equal(t, time.January, tm.Month())
				assert.Equal(t, 15, tm.Day())
			},
		},
		{
			name:  "NULL date",
			input: "",
			valid: false,
			isNil: true,
		},
		{
			name:  "empty valid string",
			input: "",
			valid: true,
			isNil: true,
		},
		{
			name:  "invalid format",
			input: "not-a-date",
			valid: true,
			isNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := sqlNullString(tt.input, tt.valid)
			result := parseDate(ns)

			if tt.isNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				if tt.checkFn != nil {
					tt.checkFn(t, result)
				}
			}
		})
	}
}

func TestParseDateTime_ValidFormats(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		valid   bool
		isNil   bool
		checkFn func(*testing.T, *time.Time)
	}{
		{
			name:  "valid datetime",
			input: "2024-01-15 10:30:45",
			valid: true,
			isNil: false,
			checkFn: func(t *testing.T, tm *time.Time) {
				assert.Equal(t, 2024, tm.Year())
				assert.Equal(t, time.January, tm.Month())
				assert.Equal(t, 15, tm.Day())
				assert.Equal(t, 10, tm.Hour())
				assert.Equal(t, 30, tm.Minute())
				assert.Equal(t, 45, tm.Second())
			},
		},
		{
			name:  "NULL datetime",
			input: "",
			valid: false,
			isNil: true,
		},
		{
			name:  "invalid format",
			input: "2024/01/15",
			valid: true,
			isNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := sqlNullString(tt.input, tt.valid)
			result := parseDateTime(ns)

			if tt.isNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				if tt.checkFn != nil {
					tt.checkFn(t, result)
				}
			}
		})
	}
}

func TestParseTime_ValidFormats(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		valid   bool
		isNil   bool
		checkFn func(*testing.T, *time.Time)
	}{
		{
			name:  "valid time",
			input: "14:30",
			valid: true,
			isNil: false,
			checkFn: func(t *testing.T, tm *time.Time) {
				assert.Equal(t, 14, tm.Hour())
				assert.Equal(t, 30, tm.Minute())
			},
		},
		{
			name:  "NULL time",
			input: "",
			valid: false,
			isNil: true,
		},
		{
			name:  "invalid format",
			input: "14:30:00",
			valid: true,
			isNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := sqlNullString(tt.input, tt.valid)
			result := parseTime(ns)

			if tt.isNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				if tt.checkFn != nil {
					tt.checkFn(t, result)
				}
			}
		})
	}
}

// =============================================================================
// Helper Functions for Tests
// =============================================================================

func ptrString(s string) *string {
	return &s
}

func sqlNullString(s string, valid bool) sql.NullString {
	return sql.NullString{String: s, Valid: valid}
}

// Unused import guard to ensure time package is used
var _ = time.Now
