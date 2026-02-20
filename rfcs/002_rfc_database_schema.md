# RFC 002: Database Schema

Status: Implemented
Author: @moond4rk

## Summary

This document describes the Things 3 SQLite database schema that the things3 library reads from. It provides complete table definitions, column types, and the relationships between tables required for implementing the read-only API.

## Database Location

Default paths (in order of precedence):
1. `$THINGSDB` environment variable
2. `~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/ThingsData-*/Things Database.thingsdatabase/main.sqlite` (v3.15.16+)
3. `~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/Things Database.thingsdatabase/main.sqlite` (older versions)

## Core Tables

### TMTask

Primary table for tasks, projects, and headings.

| Column | Type | Description |
|--------|------|-------------|
| uuid | TEXT | Primary key |
| leavesTombstone | INTEGER | Tombstone flag for sync |
| creationDate | REAL | Unix timestamp (UTC) |
| userModificationDate | REAL | Unix timestamp (UTC) |
| type | INTEGER | 0=to-do, 1=project, 2=heading |
| status | INTEGER | 0=incomplete, 2=canceled, 3=completed |
| stopDate | REAL | Unix timestamp when completed/canceled |
| trashed | INTEGER | 0=not trashed, 1=trashed |
| title | TEXT | Task title |
| notes | TEXT | Task notes/description |
| notesSync | INTEGER | Notes sync flag |
| cachedTags | BLOB | Cached tag data |
| start | INTEGER | 0=Inbox, 1=Anytime, 2=Someday |
| startDate | INTEGER | Things date format (see Date Formats) |
| startBucket | INTEGER | Start bucket |
| reminderTime | INTEGER | Things time format (see Date Formats) |
| lastReminderInteractionDate | REAL | Unix timestamp |
| deadline | INTEGER | Things date format |
| deadlineSuppressionDate | INTEGER | Things date format |
| t2_deadlineOffset | INTEGER | Deadline offset |
| index | INTEGER | Sort order in default view |
| todayIndex | INTEGER | Sort order in Today view |
| todayIndexReferenceDate | INTEGER | Reference date for Today index |
| area | TEXT | Foreign key to TMArea.uuid |
| project | TEXT | Foreign key to TMTask.uuid (parent project) |
| heading | TEXT | Foreign key to TMTask.uuid (parent heading) |
| contact | TEXT | Foreign key to TMContact.uuid |
| untrashedLeafActionsCount | INTEGER | Count of untrashed subtasks |
| openUntrashedLeafActionsCount | INTEGER | Count of open untrashed subtasks |
| checklistItemsCount | INTEGER | Total checklist items |
| openChecklistItemsCount | INTEGER | Open checklist items |
| rt1_repeatingTemplate | TEXT | Repeating template reference |
| rt1_recurrenceRule | BLOB | Recurrence rule data |
| rt1_instanceCreationStartDate | INTEGER | Things date format |
| rt1_instanceCreationPaused | INTEGER | Pause flag |
| rt1_instanceCreationCount | INTEGER | Instance count |
| rt1_afterCompletionReferenceDate | INTEGER | Things date format |
| rt1_nextInstanceStartDate | INTEGER | Things date format |
| experimental | BLOB | Experimental features data |
| repeater | BLOB | New repeater data |
| repeaterMigrationDate | REAL | Unix timestamp |

**Indexes**: stopDate, project, heading, area, rt1_repeatingTemplate

### TMArea

Areas for organizing tasks and projects.

| Column | Type | Description |
|--------|------|-------------|
| uuid | TEXT | Primary key |
| title | TEXT | Area name |
| visible | INTEGER | Visibility flag |
| index | INTEGER | Sort order |
| cachedTags | BLOB | Cached tag data |
| experimental | BLOB | Experimental features data |

### TMTag

Tags for categorizing tasks and areas.

| Column | Type | Description |
|--------|------|-------------|
| uuid | TEXT | Primary key |
| title | TEXT | Tag name |
| shortcut | TEXT | Keyboard shortcut |
| usedDate | REAL | Last used timestamp |
| parent | TEXT | Parent tag UUID (for nested tags) |
| index | INTEGER | Sort order |
| experimental | BLOB | Experimental features data |

### TMChecklistItem

Checklist items within tasks.

| Column | Type | Description |
|--------|------|-------------|
| uuid | TEXT | Primary key |
| userModificationDate | REAL | Unix timestamp |
| creationDate | REAL | Unix timestamp |
| title | TEXT | Item text |
| status | INTEGER | 0=incomplete, 2=canceled, 3=completed |
| stopDate | REAL | Unix timestamp when completed |
| index | INTEGER | Sort order within task |
| task | TEXT | Foreign key to TMTask.uuid |
| leavesTombstone | INTEGER | Tombstone flag |
| experimental | BLOB | Experimental features data |

**Indexes**: task

## Junction Tables

### TMTaskTag

Many-to-many relationship between tasks and tags.

| Column | Type | Description |
|--------|------|-------------|
| tasks | TEXT | Foreign key to TMTask.uuid |
| tags | TEXT | Foreign key to TMTag.uuid |

**Indexes**: tasks

### TMAreaTag

Many-to-many relationship between areas and tags.

| Column | Type | Description |
|--------|------|-------------|
| areas | TEXT | Foreign key to TMArea.uuid |
| tags | TEXT | Foreign key to TMTag.uuid |

**Indexes**: areas

## Metadata Tables

### TMSettings

Application settings.

| Column | Type | Description |
|--------|------|-------------|
| uuid | TEXT | Primary key (always 'RhAzEf6qDxCD5PmnZVtBZR') |
| logInterval | INTEGER | Log interval |
| manualLogDate | REAL | Manual log date |
| groupTodayByParent | INTEGER | Group Today by parent flag |
| uriSchemeAuthenticationToken | TEXT | Auth token for URL scheme |
| experimental | BLOB | Experimental features data |

### Meta

Database metadata.

| Column | Type | Description |
|--------|------|-------------|
| key | TEXT | Primary key (e.g., 'databaseVersion') |
| value | TEXT | Plist-encoded value |

## Date Formats

### Things Date (Integer)

Binary format: `YYYYYYYYYYYMMMMDDDDD0000000`

- Bits 16-26: Year (11 bits)
- Bits 12-15: Month (4 bits)
- Bits 7-11: Day (5 bits)
- Bits 0-6: Unused (7 bits)

**Masks**:
- Year: `0x7FF0000` (134152192)
- Month: `0xF000` (61440)
- Day: `0xF80` (3968)

**SQL Conversion to ISO Date**:
```sql
printf('%d-%02d-%02d',
    (startDate & 134152192) >> 16,
    (startDate & 61440) >> 12,
    (startDate & 3968) >> 7)
```

### Things Time (Integer)

Binary format: `hhhhhmmmmmm00000000000000000000`

- Bits 26-30: Hours (5 bits)
- Bits 20-25: Minutes (6 bits)
- Bits 0-19: Unused (20 bits)

**Masks**:
- Hours: `0x7C000000` (2080374784)
- Minutes: `0x3F00000` (66060288)

**SQL Conversion to HH:MM**:
```sql
printf('%02d:%02d',
    (reminderTime & 2080374784) >> 26,
    (reminderTime & 66060288) >> 20)
```

### Unix Timestamp (REAL)

Standard Unix timestamp stored as floating-point. Used for:
- `creationDate`
- `userModificationDate`
- `stopDate`

**SQL Conversion**:
```sql
datetime(creationDate, 'unixepoch', 'localtime')
```

## SQL Query Patterns

### Task Query (Base)

```sql
SELECT DISTINCT
    TASK.uuid,
    CASE
        WHEN TASK.type = 0 THEN 'to-do'
        WHEN TASK.type = 1 THEN 'project'
        WHEN TASK.type = 2 THEN 'heading'
    END AS type,
    CASE
        WHEN TASK.trashed = 1 THEN 1
    END AS trashed,
    TASK.title,
    CASE
        WHEN TASK.status = 0 THEN 'incomplete'
        WHEN TASK.status = 2 THEN 'canceled'
        WHEN TASK.status = 3 THEN 'completed'
    END AS status,
    CASE
        WHEN AREA.uuid IS NOT NULL THEN AREA.uuid
    END AS area,
    CASE
        WHEN AREA.uuid IS NOT NULL THEN AREA.title
    END AS area_title,
    CASE
        WHEN PROJECT.uuid IS NOT NULL THEN PROJECT.uuid
    END AS project,
    CASE
        WHEN PROJECT.uuid IS NOT NULL THEN PROJECT.title
    END AS project_title,
    CASE
        WHEN HEADING.uuid IS NOT NULL THEN HEADING.uuid
    END AS heading,
    CASE
        WHEN HEADING.uuid IS NOT NULL THEN HEADING.title
    END AS heading_title,
    TASK.notes,
    CASE
        WHEN TAG.uuid IS NOT NULL THEN 1
    END AS tags,
    CASE
        WHEN TASK.start = 0 THEN 'Inbox'
        WHEN TASK.start = 1 THEN 'Anytime'
        WHEN TASK.start = 2 THEN 'Someday'
    END AS start,
    CASE
        WHEN CHECKLIST_ITEM.uuid IS NOT NULL THEN 1
    END AS checklist,
    -- Things date conversion for start_date
    CASE WHEN TASK.startDate THEN
        printf('%d-%02d-%02d',
            (TASK.startDate & 134152192) >> 16,
            (TASK.startDate & 61440) >> 12,
            (TASK.startDate & 3968) >> 7)
    ELSE TASK.startDate END AS start_date,
    -- Things date conversion for deadline
    CASE WHEN TASK.deadline THEN
        printf('%d-%02d-%02d',
            (TASK.deadline & 134152192) >> 16,
            (TASK.deadline & 61440) >> 12,
            (TASK.deadline & 3968) >> 7)
    ELSE TASK.deadline END AS deadline,
    -- Things time conversion for reminder_time
    CASE WHEN TASK.reminderTime THEN
        printf('%02d:%02d',
            (TASK.reminderTime & 2080374784) >> 26,
            (TASK.reminderTime & 66060288) >> 20)
    ELSE TASK.reminderTime END AS reminder_time,
    datetime(TASK.stopDate, 'unixepoch', 'localtime') AS stop_date,
    datetime(TASK.creationDate, 'unixepoch', 'localtime') AS created,
    datetime(TASK.userModificationDate, 'unixepoch', 'localtime') AS modified,
    TASK.'index',
    TASK.todayIndex AS today_index
FROM
    TMTask AS TASK
LEFT OUTER JOIN
    TMTask PROJECT ON TASK.project = PROJECT.uuid
LEFT OUTER JOIN
    TMArea AREA ON TASK.area = AREA.uuid
LEFT OUTER JOIN
    TMTask HEADING ON TASK.heading = HEADING.uuid
LEFT OUTER JOIN
    TMTask PROJECT_OF_HEADING ON HEADING.project = PROJECT_OF_HEADING.uuid
LEFT OUTER JOIN
    TMTaskTag TAGS ON TASK.uuid = TAGS.tasks
LEFT OUTER JOIN
    TMTag TAG ON TAGS.tags = TAG.uuid
LEFT OUTER JOIN
    TMChecklistItem CHECKLIST_ITEM ON TASK.uuid = CHECKLIST_ITEM.task
WHERE
    {where_predicate}
ORDER BY
    {order_predicate}
```

### Common Filter Expressions

| Filter | SQL Expression |
|--------|----------------|
| Is to-do | `type = 0` |
| Is project | `type = 1` |
| Is heading | `type = 2` |
| Is incomplete | `status = 0` |
| Is canceled | `status = 2` |
| Is completed | `status = 3` |
| Is in Inbox | `start = 0` |
| Is Anytime | `start = 1` |
| Is Someday | `start = 2` |
| Is trashed | `trashed = 1` |
| Not trashed | `trashed = 0` |
| Not recurring | `rt1_recurrenceRule IS NULL` |

## Implementation Notes

### Go Constants Mapping

The Go implementation in `constants.go` maps directly to this schema:

```go
// Table names (unexported - internal implementation)
const (
    tableTask          = "TMTask"
    tableArea          = "TMArea"
    tableTag           = "TMTag"
    tableChecklistItem = "TMChecklistItem"
    tableTaskTag       = "TMTaskTag"
    tableAreaTag       = "TMAreaTag"
    tableSettings      = "TMSettings"
    tableMeta          = "Meta"
)

// Filter expressions (unexported - internal implementation)
const (
    filterIsTodo       = "type = 0"
    filterIsProject    = "type = 1"
    filterIsHeading    = "type = 2"
    filterIsIncomplete = "status = 0"
    filterIsCanceled   = "status = 2"
    filterIsCompleted  = "status = 3"
    filterIsInbox      = "start = 0"
    filterIsAnytime    = "start = 1"
    filterIsSomeday    = "start = 2"
)
```

### Python things.py Parity

The SQL generated by this Go library matches the Python things.py implementation:

| Feature | Python | Go | Match |
|---------|--------|-----|-------|
| Task query structure | `make_tasks_sql_query()` | `buildTasksSQL()` | Yes |
| Date conversion masks | `y_mask`, `m_mask`, `d_mask` | `yearMask`, `monthMask`, `dayMask` | Yes |
| Time conversion masks | `h_mask`, `m_mask` | `hourMask`, `minuteMask` | Yes |
| Filter expressions | String constants | String constants | Yes |
| Table names | `TABLE_*` constants | `table*` constants | Yes |
| JOIN structure | LEFT OUTER JOIN | LEFT OUTER JOIN | Yes |

## References

- Python source: https://github.com/thingsapi/things.py
- Test database: `testdata/main.sqlite`
