package database

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testAND = "\n            AND "

func TestTaskFilter_buildWhere(t *testing.T) {
	and := testAND
	today := time.Date(2024, 6, 15, 9, 0, 0, 0, time.Local)
	const todayPacked = "132671360"

	liveOnly := "TASK.trashed = 0" + and +
		"NOT IFNULL(PROJECT.trashed, 0)" + and +
		"NOT IFNULL(PROJECT_OF_HEADING.trashed, 0)"
	// defaultPrefix is what every unscoped query starts with: no repeating
	// templates, nothing trashed, nothing inside a trashed project.
	defaultPrefix := "TASK.rt1_recurrenceRule IS NULL" + and + liveOnly
	templatePrefix := "TASK.rt1_recurrenceRule IS NOT NULL" + and + liveOnly
	startDate := "(" + sqlStartDate + ")"
	createdAfter := time.Date(2024, 6, 15, 10, 30, 0, 0, time.Local)

	tests := []struct {
		name   string
		filter TaskFilter
		want   string
	}{
		{
			name:   "default excludes templates, trashed, and parent trashed",
			filter: TaskFilter{},
			want:   defaultPrefix,
		},
		{
			name:   "trash only skips the parent check",
			filter: TaskFilter{Trash: TrashOnly},
			want:   "TASK.rt1_recurrenceRule IS NULL" + and + "TASK.trashed = 1",
		},
		{
			name:   "trash any drops every trash predicate",
			filter: TaskFilter{Trash: TrashAny},
			want:   "TASK.rt1_recurrenceRule IS NULL",
		},
		{
			name:   "templates only inverts the recurrence exclusion",
			filter: TaskFilter{Templates: TemplatesOnly},
			want:   templatePrefix,
		},
		{
			name:   "templates any drops the recurrence predicate",
			filter: TaskFilter{Templates: TemplatesAny},
			want:   liveOnly,
		},
		{
			name:   "identity scope has no default predicates",
			filter: TaskFilter{Trash: TrashAny, Templates: TemplatesAny, UUID: new("ABC-123")},
			want:   "TASK.uuid = 'ABC-123'",
		},
		{
			name:   "task type",
			filter: TaskFilter{Types: []int{0}},
			want:   defaultPrefix + and + "TASK.type = 0",
		},
		{
			name:   "several task types",
			filter: TaskFilter{Types: []int{0, 1}},
			want:   defaultPrefix + and + "TASK.type IN (0, 1)",
		},
		{
			name:   "closed means completed or canceled",
			filter: TaskFilter{Closed: true},
			want:   defaultPrefix + and + "TASK.status != 0",
		},
		{
			name:   "an explicit status wins over closed",
			filter: TaskFilter{Closed: true, Status: new(3)},
			want:   defaultPrefix + and + "TASK.status = 3",
		},
		{
			name:   "status",
			filter: TaskFilter{Status: new(3)},
			want:   defaultPrefix + and + "TASK.status = 3",
		},
		{
			name:   "start",
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
			name:   "uuid prefix escapes like wildcards",
			filter: TaskFilter{UUIDPrefix: new("AB_C")},
			want:   defaultPrefix + and + `TASK.uuid LIKE 'AB\_C%' ESCAPE '\'`,
		},
		{
			name:   "title",
			filter: TaskFilter{Title: new("milk")},
			want:   defaultPrefix + and + `TASK.title LIKE '%milk%' ESCAPE '\'`,
		},
		{
			name:   "title escapes like wildcards",
			filter: TaskFilter{Title: new("To_Do")},
			want:   defaultPrefix + and + `TASK.title LIKE '%To\_Do%' ESCAPE '\'`,
		},
		{
			name:   "area uuid matches the effective area",
			filter: TaskFilter{AreaUUID: new("area-1")},
			want:   defaultPrefix + and + sqlAreaUUID + " = 'area-1'",
		},
		{
			name:   "has area true",
			filter: TaskFilter{HasArea: new(true)},
			want:   defaultPrefix + and + sqlAreaUUID + " IS NOT NULL",
		},
		{
			name:   "has area false",
			filter: TaskFilter{HasArea: new(false)},
			want:   defaultPrefix + and + sqlAreaUUID + " IS NULL",
		},
		{
			name:   "area uuid takes precedence over has area",
			filter: TaskFilter{AreaUUID: new("area-1"), HasArea: new(true)},
			want:   defaultPrefix + and + sqlAreaUUID + " = 'area-1'",
		},
		{
			name:   "project uuid matches the effective project",
			filter: TaskFilter{ProjectUUID: new("proj-1")},
			want:   defaultPrefix + and + sqlProjectUUID + " = 'proj-1'",
		},
		{
			name:   "has project true",
			filter: TaskFilter{HasProject: new(true)},
			want:   defaultPrefix + and + sqlProjectUUID + " IS NOT NULL",
		},
		{
			name:   "has project false",
			filter: TaskFilter{HasProject: new(false)},
			want:   defaultPrefix + and + sqlProjectUUID + " IS NULL",
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
			want:   defaultPrefix + and + taskTagsSQL("TAG.title = 'work'"),
		},
		{
			name:   "has tags true",
			filter: TaskFilter{HasTags: new(true)},
			want:   defaultPrefix + and + taskTagsSQL(""),
		},
		{
			name:   "has tags false",
			filter: TaskFilter{HasTags: new(false)},
			want:   defaultPrefix + and + "NOT " + taskTagsSQL(""),
		},
		{
			name:   "today membership",
			filter: TaskFilter{InToday: true},
			want: defaultPrefix + and +
				"((TASK.start = 1 AND TASK.startDate IS NOT NULL)" +
				" OR (TASK.start = 2 AND TASK.startDate <= " + todayPacked + ")" +
				" OR (TASK.startDate IS NULL AND TASK.deadline <= " + todayPacked + " AND TASK.deadlineSuppressionDate IS NULL))",
		},
		{
			name:   "upcoming membership",
			filter: TaskFilter{InUpcoming: true},
			want: defaultPrefix + and +
				"(" + startDate + " > " + todayPacked +
				" OR (" + startDate + " IS NULL AND TASK.rt1_recurrenceRule IS NULL AND TASK.deadline > " + todayPacked + "))",
		},
		{
			name:   "container predicates",
			filter: TaskFilter{ContainerNotSomeday: true, ContainerOpen: true, ContainerNotTrashed: true},
			want: defaultPrefix + and +
				"COALESCE(PROJECT.start, PROJECT_OF_HEADING.start, 1) != 2" + and +
				"COALESCE(PROJECT.status, PROJECT_OF_HEADING.status, 0) = 0" + and +
				"COALESCE(PROJECT.trashed, PROJECT_OF_HEADING.trashed, 0) = 0",
		},
		{
			name:   "start date exists true",
			filter: TaskFilter{StartDateFilter: &DateFilterValue{HasDate: new(true)}},
			want:   defaultPrefix + and + startDate + " IS NOT NULL",
		},
		{
			name:   "start date exists false",
			filter: TaskFilter{StartDateFilter: &DateFilterValue{HasDate: new(false)}},
			want:   defaultPrefix + and + startDate + " IS NULL",
		},
		{
			name:   "start date future is anchored to the injected today",
			filter: TaskFilter{StartDateFilter: &DateFilterValue{Relative: DateFuture}},
			want:   defaultPrefix + and + startDate + " > " + todayPacked,
		},
		{
			name:   "start date past",
			filter: TaskFilter{StartDateFilter: &DateFilterValue{Relative: DatePast}},
			want:   defaultPrefix + and + startDate + " <= " + todayPacked,
		},
		{
			name: "templates filter on the same start date expression",
			filter: TaskFilter{
				Templates:       TemplatesOnly,
				StartDateFilter: &DateFilterValue{Relative: DateFuture},
			},
			want: templatePrefix + and + startDate + " > " + todayPacked,
		},
		{
			name: "start date specific comparison",
			filter: TaskFilter{StartDateFilter: &DateFilterValue{
				Operator: ">=",
				Date:     new(time.Date(2024, 6, 15, 0, 0, 0, 0, time.Local)),
			}},
			want: defaultPrefix + and + startDate + " >= 132671360",
		},
		{
			name:   "stop date future (unix time)",
			filter: TaskFilter{StopDateFilter: &DateFilterValue{Relative: DateFuture}},
			want:   defaultPrefix + and + "date(TASK.stopDate, 'unixepoch', 'localtime') > '2024-06-15'",
		},
		{
			name:   "stop date past (unix time)",
			filter: TaskFilter{StopDateFilter: &DateFilterValue{Relative: DatePast}},
			want:   defaultPrefix + and + "date(TASK.stopDate, 'unixepoch', 'localtime') <= '2024-06-15'",
		},
		{
			name: "stop date specific comparison (unix time)",
			filter: TaskFilter{StopDateFilter: &DateFilterValue{
				Operator: "=",
				Date:     new(time.Date(2024, 6, 1, 0, 0, 0, 0, time.Local)),
			}},
			want: defaultPrefix + and + "date(TASK.stopDate, 'unixepoch', 'localtime') = '2024-06-01'",
		},
		{
			name:   "deadline future (things date)",
			filter: TaskFilter{DeadlineFilter: &DateFilterValue{Relative: DateFuture}},
			want:   defaultPrefix + and + "TASK.deadline > " + todayPacked,
		},
		{
			name:   "created after compares raw epochs",
			filter: TaskFilter{CreatedAfter: &createdAfter},
			want:   defaultPrefix + and + "TASK.creationDate > " + strconv.FormatInt(createdAfter.Unix(), 10) + ".000000",
		},
		{
			name:   "search query matches the item's own text",
			filter: TaskFilter{SearchQuery: new("buy milk")},
			want:   defaultPrefix + and + `(TASK.title LIKE '%buy milk%' ESCAPE '\' OR TASK.notes LIKE '%buy milk%' ESCAPE '\')`,
		},
		{
			name:   "search with special chars",
			filter: TaskFilter{SearchQuery: new("it's")},
			want:   defaultPrefix + and + `(TASK.title LIKE '%it''s%' ESCAPE '\' OR TASK.notes LIKE '%it''s%' ESCAPE '\')`,
		},
		{
			name:   "search escapes like wildcards",
			filter: TaskFilter{SearchQuery: new("%")},
			want:   defaultPrefix + and + `(TASK.title LIKE '%\%%' ESCAPE '\' OR TASK.notes LIKE '%\%%' ESCAPE '\')`,
		},
		{
			name: "complex filter combination",
			filter: TaskFilter{
				Types:  []int{0},
				Status: new(0),
				Start:  new(1),
				StartDateFilter: &DateFilterValue{
					HasDate: new(true),
				},
			},
			want: defaultPrefix + and +
				"TASK.type = 0" + and +
				"TASK.status = 0" + and +
				"TASK.start = 1" + and +
				startDate + " IS NOT NULL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.filter.buildWhere(today))
		})
	}
}

func TestTaskFilter_buildOrder(t *testing.T) {
	const index = `TASK."index"`
	tests := []struct {
		name  string
		order TaskOrder
		want  string
	}{
		{
			"sidebar is the default", OrderSidebar,
			`AREA."index", COALESCE(PROJECT."index", PROJECT_OF_HEADING."index"), HEADING."index", ` +
				index + ", TASK.userModificationDate DESC",
		},
		{"today", OrderToday, "TASK.todayIndex, TASK.todayIndexReferenceDate DESC, " + index},
		{"upcoming", OrderUpcoming, sqlUpcomingDate + ", TASK.todayIndex, " + index},
		{"deadline", OrderDeadline, "TASK.deadline, " + index},
		{"stop date", OrderStopDateDesc, "TASK.stopDate DESC, " + index},
		{"modified", OrderModifiedDesc, "TASK.userModificationDate DESC, " + index},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := TaskFilter{Order: tt.order}
			assert.Equal(t, tt.want, f.buildOrder())
		})
	}
	assert.Equal(t, (&TaskFilter{}).buildOrder(), (&TaskFilter{Order: OrderSidebar}).buildOrder())
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
			want:   areaTagsSQL("TAG.title = 'important'"),
		},
		{
			name:   "has tag true",
			filter: AreaFilter{HasTag: new(true)},
			want:   areaTagsSQL(""),
		},
		{
			name:   "has tag false",
			filter: AreaFilter{HasTag: new(false)},
			want:   "NOT " + areaTagsSQL(""),
		},
		{
			name:   "tag title takes precedence over has tag",
			filter: AreaFilter{TagTitle: new("work"), HasTag: new(true)},
			want:   areaTagsSQL("TAG.title = 'work'"),
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
