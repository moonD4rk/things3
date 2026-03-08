package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testAND = "\n            AND "

func ptr[T any](v T) *T { return &v }

func TestTaskFilter_buildWhere(t *testing.T) {
	and := testAND

	tests := []struct {
		name   string
		filter TaskFilter
		want   string
	}{
		{
			name:   "default (no filters)",
			filter: TaskFilter{},
			want:   "TASK.rt1_recurrenceRule IS NULL" + and + "TASK.trashed = 0",
		},
		{
			name:   "trashed true",
			filter: TaskFilter{Trashed: ptr(true)},
			want:   "TASK.rt1_recurrenceRule IS NULL" + and + "TASK.trashed = 1",
		},
		{
			name:   "trashed false explicit",
			filter: TaskFilter{Trashed: ptr(false)},
			want:   "TASK.rt1_recurrenceRule IS NULL" + and + "TASK.trashed = 0",
		},
		{
			name:   "context trashed false",
			filter: TaskFilter{ContextTrashed: ptr(false)},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"NOT IFNULL(PROJECT.trashed, 0)" + and +
				"NOT IFNULL(PROJECT_OF_HEADING.trashed, 0)",
		},
		{
			name:   "context trashed true",
			filter: TaskFilter{ContextTrashed: ptr(true)},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"PROJECT.trashed" + and +
				"PROJECT_OF_HEADING.trashed",
		},
		{
			name:   "task type",
			filter: TaskFilter{TaskType: ptr(0)},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"TASK.type = 0",
		},
		{
			name:   "status",
			filter: TaskFilter{Status: ptr(3)},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"TASK.status = 3",
		},
		{
			name:   "start bucket",
			filter: TaskFilter{Start: ptr(1)},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"TASK.start = 1",
		},
		{
			name:   "uuid",
			filter: TaskFilter{UUID: ptr("ABC-123")},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"TASK.uuid = 'ABC-123'",
		},
		{
			name:   "uuid prefix",
			filter: TaskFilter{UUIDPrefix: ptr("ABC")},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"TASK.uuid LIKE 'ABC%'",
		},
		{
			name:   "area uuid",
			filter: TaskFilter{AreaUUID: ptr("area-1")},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"TASK.area = 'area-1'",
		},
		{
			name:   "has area true",
			filter: TaskFilter{HasArea: ptr(true)},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"TASK.area IS NOT NULL",
		},
		{
			name:   "has area false",
			filter: TaskFilter{HasArea: ptr(false)},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"TASK.area IS NULL",
		},
		{
			name:   "area uuid takes precedence over has area",
			filter: TaskFilter{AreaUUID: ptr("area-1"), HasArea: ptr(true)},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"TASK.area = 'area-1'",
		},
		{
			name:   "project uuid",
			filter: TaskFilter{ProjectUUID: ptr("proj-1")},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"(TASK.project = 'proj-1' OR PROJECT_OF_HEADING.uuid = 'proj-1')",
		},
		{
			name:   "has project true",
			filter: TaskFilter{HasProject: ptr(true)},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"(TASK.project IS NOT NULL OR PROJECT_OF_HEADING.uuid IS NOT NULL)",
		},
		{
			name:   "has project false",
			filter: TaskFilter{HasProject: ptr(false)},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"(TASK.project IS NULL OR PROJECT_OF_HEADING.uuid IS NULL)",
		},
		{
			name:   "heading uuid",
			filter: TaskFilter{HeadingUUID: ptr("head-1")},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"TASK.heading = 'head-1'",
		},
		{
			name:   "has heading true",
			filter: TaskFilter{HasHeading: ptr(true)},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"TASK.heading IS NOT NULL",
		},
		{
			name:   "tag title",
			filter: TaskFilter{TagTitle: ptr("work")},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"TAG.title = 'work'",
		},
		{
			name:   "has tags true",
			filter: TaskFilter{HasTags: ptr(true)},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"TAG.title IS NOT NULL",
		},
		{
			name:   "deadline suppressed true",
			filter: TaskFilter{DeadlineSuppressed: ptr(true)},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"TASK.deadlineSuppressionDate IS NOT NULL",
		},
		{
			name:   "deadline suppressed false",
			filter: TaskFilter{DeadlineSuppressed: ptr(false)},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"TASK.deadlineSuppressionDate IS NULL",
		},
		{
			name:   "start date exists true",
			filter: TaskFilter{StartDateFilter: &DateFilterValue{HasDate: ptr(true)}},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"TASK.startDate IS NOT NULL",
		},
		{
			name:   "start date exists false",
			filter: TaskFilter{StartDateFilter: &DateFilterValue{HasDate: ptr(false)}},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"TASK.startDate IS NULL",
		},
		{
			name:   "start date future",
			filter: TaskFilter{StartDateFilter: &DateFilterValue{Relative: DateFuture}},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"TASK.startDate > " + todayThingsDateSQL(),
		},
		{
			name:   "start date past",
			filter: TaskFilter{StartDateFilter: &DateFilterValue{Relative: DatePast}},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"TASK.startDate <= " + todayThingsDateSQL(),
		},
		{
			name: "start date specific comparison",
			filter: TaskFilter{StartDateFilter: &DateFilterValue{
				Operator: ">=",
				Date:     ptr(time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)),
			}},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"TASK.startDate >= 132671360",
		},
		{
			name:   "stop date future (unix time)",
			filter: TaskFilter{StopDateFilter: &DateFilterValue{Relative: DateFuture}},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"date(TASK.stopDate, 'unixepoch', 'localtime') > date('now', 'localtime')",
		},
		{
			name:   "stop date past (unix time)",
			filter: TaskFilter{StopDateFilter: &DateFilterValue{Relative: DatePast}},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"date(TASK.stopDate, 'unixepoch', 'localtime') <= date('now', 'localtime')",
		},
		{
			name: "stop date specific comparison (unix time)",
			filter: TaskFilter{StopDateFilter: &DateFilterValue{
				Operator: "=",
				Date:     ptr(time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)),
			}},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"date(TASK.stopDate, 'unixepoch', 'localtime') = date('2024-06-15')",
		},
		{
			name:   "deadline future (things date)",
			filter: TaskFilter{DeadlineFilter: &DateFilterValue{Relative: DateFuture}},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"TASK.deadline > " + todayThingsDateSQL(),
		},
		{
			name:   "created after",
			filter: TaskFilter{CreatedAfter: ptr(time.Date(2024, 6, 15, 10, 30, 0, 0, time.Local))},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"datetime(TASK.creationDate, 'unixepoch', 'localtime') > '2024-06-15 10:30:00'",
		},
		{
			name:   "search query",
			filter: TaskFilter{SearchQuery: ptr("buy milk")},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"(TASK.title LIKE '%buy milk%' OR TASK.notes LIKE '%buy milk%' OR AREA.title LIKE '%buy milk%')",
		},
		{
			name:   "search with special chars",
			filter: TaskFilter{SearchQuery: ptr("it's")},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"(TASK.title LIKE '%it''s%' OR TASK.notes LIKE '%it''s%' OR AREA.title LIKE '%it''s%')",
		},
		{
			name: "complex filter combination",
			filter: TaskFilter{
				TaskType:       ptr(0),
				Status:         ptr(0),
				Start:          ptr(1),
				ContextTrashed: ptr(false),
				StartDateFilter: &DateFilterValue{
					HasDate: ptr(true),
				},
			},
			want: "TASK.rt1_recurrenceRule IS NULL" + and +
				"TASK.trashed = 0" + and +
				"NOT IFNULL(PROJECT.trashed, 0)" + and +
				"NOT IFNULL(PROJECT_OF_HEADING.trashed, 0)" + and +
				"TASK.type = 0" + and +
				"TASK.status = 0" + and +
				"TASK.start = 1" + and +
				"TASK.startDate IS NOT NULL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.buildWhere()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTaskFilter_buildOrder(t *testing.T) {
	tests := []struct {
		name   string
		filter TaskFilter
		want   string
	}{
		{"default", TaskFilter{}, `TASK."index"`},
		{"explicit default", TaskFilter{Index: IndexDefault}, `TASK."index"`},
		{"today index", TaskFilter{Index: IndexToday}, `TASK."todayIndex"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.filter.buildOrder())
		})
	}
}

func TestAreaFilter_buildWhere(t *testing.T) {
	and := testAND

	tests := []struct {
		name   string
		filter AreaFilter
		want   string
	}{
		{
			name:   "empty",
			filter: AreaFilter{},
			want:   "TRUE",
		},
		{
			name:   "uuid",
			filter: AreaFilter{UUID: ptr("area-1")},
			want:   "AREA.uuid = 'area-1'",
		},
		{
			name:   "title",
			filter: AreaFilter{Title: ptr("Work")},
			want:   "AREA.title = 'Work'",
		},
		{
			name:   "visible true",
			filter: AreaFilter{Visible: ptr(true)},
			want:   "AREA.visible",
		},
		{
			name:   "visible false",
			filter: AreaFilter{Visible: ptr(false)},
			want:   "NOT IFNULL(AREA.visible, 0)",
		},
		{
			name:   "tag title",
			filter: AreaFilter{TagTitle: ptr("important")},
			want:   "TAG.title = 'important'",
		},
		{
			name:   "has tag true",
			filter: AreaFilter{HasTag: ptr(true)},
			want:   "TAG.title IS NOT NULL",
		},
		{
			name:   "has tag false",
			filter: AreaFilter{HasTag: ptr(false)},
			want:   "TAG.title IS NULL",
		},
		{
			name:   "tag title takes precedence over has tag",
			filter: AreaFilter{TagTitle: ptr("work"), HasTag: ptr(true)},
			want:   "TAG.title = 'work'",
		},
		{
			name:   "multiple filters",
			filter: AreaFilter{UUID: ptr("area-1"), Visible: ptr(true)},
			want:   "AREA.uuid = 'area-1'" + and + "AREA.visible",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.buildWhere()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTagFilter_buildWhere(t *testing.T) {
	and := testAND

	tests := []struct {
		name   string
		filter TagFilter
		want   string
	}{
		{
			name:   "empty",
			filter: TagFilter{},
			want:   "TRUE",
		},
		{
			name:   "uuid",
			filter: TagFilter{UUID: ptr("tag-1")},
			want:   "uuid = 'tag-1'",
		},
		{
			name:   "title",
			filter: TagFilter{Title: ptr("work")},
			want:   "title = 'work'",
		},
		{
			name:   "parent uuid",
			filter: TagFilter{ParentUUID: ptr("parent-1")},
			want:   "parent = 'parent-1'",
		},
		{
			name:   "multiple",
			filter: TagFilter{UUID: ptr("tag-1"), Title: ptr("work")},
			want:   "uuid = 'tag-1'" + and + "title = 'work'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.buildWhere()
			assert.Equal(t, tt.want, got)
		})
	}
}
