package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testAND = "\n            AND "

func TestTaskFilter_buildWhere(t *testing.T) {
	and := testAND
	// defaultPrefix is the common prefix for all non-trashed queries:
	// recurring excluded + not trashed + parent not trashed
	defaultPrefix := "TASK.rt1_recurrenceRule IS NULL" + and +
		"TASK.trashed = 0" + and +
		"NOT IFNULL(PROJECT.trashed, 0)" + and +
		"NOT IFNULL(PROJECT_OF_HEADING.trashed, 0)"

	tests := []struct {
		name   string
		filter TaskFilter
		want   string
	}{
		{
			name:   "default excludes trashed and parent trashed",
			filter: TaskFilter{},
			want:   defaultPrefix,
		},
		{
			name:   "trashed true skips parent trashed check",
			filter: TaskFilter{Trashed: new(true)},
			want:   "TASK.rt1_recurrenceRule IS NULL" + and + "TASK.trashed = 1",
		},
		{
			name:   "trashed false explicit",
			filter: TaskFilter{Trashed: new(false)},
			want:   defaultPrefix,
		},
		{
			name:   "task type",
			filter: TaskFilter{TaskType: new(0)},
			want:   defaultPrefix + and + "TASK.type = 0",
		},
		{
			name:   "status",
			filter: TaskFilter{Status: new(3)},
			want:   defaultPrefix + and + "TASK.status = 3",
		},
		{
			name:   "start bucket",
			filter: TaskFilter{Start: new(1)},
			want:   defaultPrefix + and + "TASK.start = 1",
		},
		{
			name:   "uuid",
			filter: TaskFilter{UUID: new("ABC-123")},
			want:   defaultPrefix + and + "TASK.uuid = 'ABC-123'",
		},
		{
			name:   "uuid prefix",
			filter: TaskFilter{UUIDPrefix: new("ABC")},
			want:   defaultPrefix + and + `TASK.uuid LIKE 'ABC%' ESCAPE '\'`,
		},
		{
			name:   "uuid prefix escapes wildcards",
			filter: TaskFilter{UUIDPrefix: new("AB_C")},
			want:   defaultPrefix + and + `TASK.uuid LIKE 'AB\_C%' ESCAPE '\'`,
		},
		{
			name:   "title contains",
			filter: TaskFilter{Title: new("milk")},
			want:   defaultPrefix + and + `TASK.title LIKE '%milk%' ESCAPE '\'`,
		},
		{
			name:   "title escapes wildcards",
			filter: TaskFilter{Title: new("To_Do")},
			want:   defaultPrefix + and + `TASK.title LIKE '%To\_Do%' ESCAPE '\'`,
		},
		{
			name:   "area uuid",
			filter: TaskFilter{AreaUUID: new("area-1")},
			want:   defaultPrefix + and + "TASK.area = 'area-1'",
		},
		{
			name:   "has area true",
			filter: TaskFilter{HasArea: new(true)},
			want:   defaultPrefix + and + "TASK.area IS NOT NULL",
		},
		{
			name:   "has area false",
			filter: TaskFilter{HasArea: new(false)},
			want:   defaultPrefix + and + "TASK.area IS NULL",
		},
		{
			name:   "area uuid takes precedence over has area",
			filter: TaskFilter{AreaUUID: new("area-1"), HasArea: new(true)},
			want:   defaultPrefix + and + "TASK.area = 'area-1'",
		},
		{
			name:   "project uuid",
			filter: TaskFilter{ProjectUUID: new("proj-1")},
			want:   defaultPrefix + and + "(TASK.project = 'proj-1' OR PROJECT_OF_HEADING.uuid = 'proj-1')",
		},
		{
			name:   "has project true",
			filter: TaskFilter{HasProject: new(true)},
			want:   defaultPrefix + and + "(TASK.project IS NOT NULL OR PROJECT_OF_HEADING.uuid IS NOT NULL)",
		},
		{
			name:   "has project false requires both columns NULL",
			filter: TaskFilter{HasProject: new(false)},
			want:   defaultPrefix + and + "(TASK.project IS NULL AND PROJECT_OF_HEADING.uuid IS NULL)",
		},
		{
			name:   "heading uuid",
			filter: TaskFilter{HeadingUUID: new("head-1")},
			want:   defaultPrefix + and + "TASK.heading = 'head-1'",
		},
		{
			name:   "has heading true",
			filter: TaskFilter{HasHeading: new(true)},
			want:   defaultPrefix + and + "TASK.heading IS NOT NULL",
		},
		{
			name:   "tag title",
			filter: TaskFilter{TagTitle: new("work")},
			want:   defaultPrefix + and + "TAG.title = 'work'",
		},
		{
			name:   "has tags true",
			filter: TaskFilter{HasTags: new(true)},
			want:   defaultPrefix + and + "TAG.title IS NOT NULL",
		},
		{
			name:   "deadline suppressed true",
			filter: TaskFilter{DeadlineSuppressed: new(true)},
			want:   defaultPrefix + and + "TASK.deadlineSuppressionDate IS NOT NULL",
		},
		{
			name:   "deadline suppressed false",
			filter: TaskFilter{DeadlineSuppressed: new(false)},
			want:   defaultPrefix + and + "TASK.deadlineSuppressionDate IS NULL",
		},
		{
			name:   "start date exists true",
			filter: TaskFilter{StartDateFilter: &DateFilterValue{HasDate: new(true)}},
			want:   defaultPrefix + and + "TASK.startDate IS NOT NULL",
		},
		{
			name:   "start date exists false",
			filter: TaskFilter{StartDateFilter: &DateFilterValue{HasDate: new(false)}},
			want:   defaultPrefix + and + "TASK.startDate IS NULL",
		},
		{
			name:   "start date future",
			filter: TaskFilter{StartDateFilter: &DateFilterValue{Relative: DateFuture}},
			want:   defaultPrefix + and + "TASK.startDate > " + todayThingsDateSQL(),
		},
		{
			name:   "start date past",
			filter: TaskFilter{StartDateFilter: &DateFilterValue{Relative: DatePast}},
			want:   defaultPrefix + and + "TASK.startDate <= " + todayThingsDateSQL(),
		},
		{
			name: "start date specific comparison",
			filter: TaskFilter{StartDateFilter: &DateFilterValue{
				Operator: ">=",
				Date:     new(time.Date(2024, 6, 15, 0, 0, 0, 0, time.Local)),
			}},
			want: defaultPrefix + and + "TASK.startDate >= 132671360",
		},
		{
			name:   "stop date future (unix time)",
			filter: TaskFilter{StopDateFilter: &DateFilterValue{Relative: DateFuture}},
			want:   defaultPrefix + and + "date(TASK.stopDate, 'unixepoch', 'localtime') > date('now', 'localtime')",
		},
		{
			name:   "stop date past (unix time)",
			filter: TaskFilter{StopDateFilter: &DateFilterValue{Relative: DatePast}},
			want:   defaultPrefix + and + "date(TASK.stopDate, 'unixepoch', 'localtime') <= date('now', 'localtime')",
		},
		{
			name: "stop date specific comparison (unix time)",
			filter: TaskFilter{StopDateFilter: &DateFilterValue{
				Operator: "=",
				Date:     new(time.Date(2024, 6, 15, 0, 0, 0, 0, time.Local)),
			}},
			want: defaultPrefix + and + "date(TASK.stopDate, 'unixepoch', 'localtime') = date('2024-06-15')",
		},
		{
			name:   "deadline future (things date)",
			filter: TaskFilter{DeadlineFilter: &DateFilterValue{Relative: DateFuture}},
			want:   defaultPrefix + and + "TASK.deadline > " + todayThingsDateSQL(),
		},
		{
			name:   "created after",
			filter: TaskFilter{CreatedAfter: new(time.Date(2024, 6, 15, 10, 30, 0, 0, time.Local))},
			want:   defaultPrefix + and + "datetime(TASK.creationDate, 'unixepoch', 'localtime') > '2024-06-15 10:30:00'",
		},
		{
			name:   "search query",
			filter: TaskFilter{SearchQuery: new("buy milk")},
			want: defaultPrefix + and +
				`(TASK.title LIKE '%buy milk%' ESCAPE '\' OR TASK.notes LIKE '%buy milk%' ESCAPE '\' OR AREA.title LIKE '%buy milk%' ESCAPE '\')`,
		},
		{
			name:   "search with special chars",
			filter: TaskFilter{SearchQuery: new("it's")},
			want: defaultPrefix + and +
				`(TASK.title LIKE '%it''s%' ESCAPE '\' OR TASK.notes LIKE '%it''s%' ESCAPE '\' OR AREA.title LIKE '%it''s%' ESCAPE '\')`,
		},
		{
			name:   "search escapes like wildcards",
			filter: TaskFilter{SearchQuery: new("%")},
			want: defaultPrefix + and +
				`(TASK.title LIKE '%\%%' ESCAPE '\' OR TASK.notes LIKE '%\%%' ESCAPE '\' OR AREA.title LIKE '%\%%' ESCAPE '\')`,
		},
		{
			name: "complex filter combination",
			filter: TaskFilter{
				TaskType: new(0),
				Status:   new(0),
				Start:    new(1),
				StartDateFilter: &DateFilterValue{
					HasDate: new(true),
				},
			},
			want: defaultPrefix + and +
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
			filter: AreaFilter{UUID: new("area-1")},
			want:   "AREA.uuid = 'area-1'",
		},
		{
			name:   "title",
			filter: AreaFilter{Title: new("Work")},
			want:   "AREA.title = 'Work'",
		},
		{
			name:   "visible true treats NULL as visible",
			filter: AreaFilter{Visible: new(true)},
			want:   "IFNULL(AREA.visible, 1)",
		},
		{
			name:   "visible false treats NULL as visible",
			filter: AreaFilter{Visible: new(false)},
			want:   "NOT IFNULL(AREA.visible, 1)",
		},
		{
			name:   "tag title",
			filter: AreaFilter{TagTitle: new("important")},
			want:   "TAG.title = 'important'",
		},
		{
			name:   "has tag true",
			filter: AreaFilter{HasTag: new(true)},
			want:   "TAG.title IS NOT NULL",
		},
		{
			name:   "has tag false",
			filter: AreaFilter{HasTag: new(false)},
			want:   "TAG.title IS NULL",
		},
		{
			name:   "tag title takes precedence over has tag",
			filter: AreaFilter{TagTitle: new("work"), HasTag: new(true)},
			want:   "TAG.title = 'work'",
		},
		{
			name:   "multiple filters",
			filter: AreaFilter{UUID: new("area-1"), Visible: new(true)},
			want:   "AREA.uuid = 'area-1'" + and + "IFNULL(AREA.visible, 1)",
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
			filter: TagFilter{UUID: new("tag-1")},
			want:   "uuid = 'tag-1'",
		},
		{
			name:   "title",
			filter: TagFilter{Title: new("work")},
			want:   "title = 'work'",
		},
		{
			name:   "parent uuid",
			filter: TagFilter{ParentUUID: new("parent-1")},
			want:   "parent = 'parent-1'",
		},
		{
			name:   "multiple",
			filter: TagFilter{UUID: new("tag-1"), Title: new("work")},
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
