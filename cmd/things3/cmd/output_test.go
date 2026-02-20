package cmd

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/moond4rk/things3"
)

func TestShortUUID(t *testing.T) {
	tests := []struct {
		name     string
		uuid     string
		expected string
	}{
		{"long UUID", "ABCDEFGH12345678", "ABCDEFGH"},
		{"exactly 8", "ABCDEFGH", "ABCDEFGH"},
		{"short", "ABC", "ABC"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, shortUUID(tt.uuid))
		})
	}
}

func TestFormatTaskType(t *testing.T) {
	tests := []struct {
		taskType things3.TaskType
		expected string
	}{
		{things3.TaskTypeTodo, "todo"},
		{things3.TaskTypeProject, "project"},
		{things3.TaskTypeHeading, "heading"},
		{things3.TaskType(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, formatTaskType(tt.taskType))
		})
	}
}

func TestFormatTags(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		expected string
	}{
		{"nil", nil, ""},
		{"empty", []string{}, ""},
		{"single", []string{"work"}, "#work"},
		{"multiple", []string{"work", "urgent"}, "#work #urgent"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, formatTags(tt.tags))
		})
	}
}

func TestGetRelevantDate(t *testing.T) {
	d := time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		task     things3.Task
		expected string
	}{
		{
			name:     "no dates",
			task:     things3.Task{},
			expected: "",
		},
		{
			name:     "start date only",
			task:     things3.Task{StartDate: &d},
			expected: "2025-03-15",
		},
		{
			name:     "deadline only",
			task:     things3.Task{Deadline: &d},
			expected: "due:2025-03-15",
		},
		{
			name:     "stop date only",
			task:     things3.Task{StopDate: &d},
			expected: "2025-03-15",
		},
		{
			name:     "stop date takes priority over deadline",
			task:     things3.Task{StopDate: &d, Deadline: &d},
			expected: "2025-03-15",
		},
		{
			name:     "deadline takes priority over start date",
			task:     things3.Task{Deadline: &d, StartDate: &d},
			expected: "due:2025-03-15",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, getRelevantDate(&tt.task))
		})
	}
}

func TestApplyLimit(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}

	assert.Equal(t, []int{1, 2, 3, 4, 5}, applyLimit(items, 0))
	assert.Equal(t, []int{1, 2, 3}, applyLimit(items, 3))
	assert.Equal(t, []int{1, 2, 3, 4, 5}, applyLimit(items, 10))
	assert.Equal(t, []int{1}, applyLimit(items, 1))
}

func TestFormatTaskLine(t *testing.T) {
	task := things3.Task{
		UUID:   "ABCDEFGH12345678",
		Type:   things3.TaskTypeTodo,
		Status: things3.StatusIncomplete,
		Title:  "Buy milk",
	}
	line := formatTaskLine(&task)
	assert.Contains(t, line, "[ ]")
	assert.Contains(t, line, "ABCDEFGH")
	assert.Contains(t, line, "todo")
	assert.Contains(t, line, "Buy milk")

	// Completed task
	task.Status = things3.StatusCompleted
	line = formatTaskLine(&task)
	assert.Contains(t, line, "[x]")

	// Canceled task
	task.Status = things3.StatusCanceled
	line = formatTaskLine(&task)
	assert.Contains(t, line, "[-]")

	// With tags
	task.Tags = []string{"shopping"}
	line = formatTaskLine(&task)
	assert.Contains(t, line, "#shopping")
}

func TestWriteTasks_Text(t *testing.T) {
	var buf bytes.Buffer
	tasks := []things3.Task{
		{UUID: "uuid1234abcd", Type: things3.TaskTypeTodo, Title: "Task 1"},
	}
	err := writeTasks(&buf, tasks, formatText)
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "STATUS")
	assert.Contains(t, output, "Task 1")
}

func TestWriteTasks_EmptyText(t *testing.T) {
	var buf bytes.Buffer
	err := writeTasks(&buf, nil, formatText)
	require.NoError(t, err)
	assert.Empty(t, buf.String())
}

func TestWriteTasks_JSON(t *testing.T) {
	var buf bytes.Buffer
	tasks := []things3.Task{
		{UUID: "uuid1234", Type: things3.TaskTypeTodo, Title: "Task 1"},
	}
	err := writeTasks(&buf, tasks, formatJSON)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), `"title": "Task 1"`)
}

func TestWriteAreas_Text(t *testing.T) {
	var buf bytes.Buffer
	areas := []things3.Area{
		{UUID: "a1", Title: "Work"},
		{UUID: "a2", Title: "Personal"},
	}
	err := writeAreas(&buf, areas, formatText)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "- Work\n")
	assert.Contains(t, buf.String(), "- Personal\n")
}

func TestWriteTags_Text(t *testing.T) {
	var buf bytes.Buffer
	tags := []things3.Tag{
		{UUID: "t1", Title: "urgent"},
	}
	err := writeTags(&buf, tags, formatText)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "- urgent\n")
}
