# RFC 007: Test Fixture Database

Status: Draft
Author: @moond4rk
Date: 2026-02-22

## Summary

This document provides a complete inventory of all data in the test fixture database (`testdata/main.sqlite`), maps each record to its Things 3 concept, and identifies coverage gaps for future test improvements.

## Database Metadata

| Key | Value |
|-----|-------|
| databaseVersion | 24 |
| didCreateDefaultTags | true |
| didRemoveOrphanHeadings | true |

## Schema: Enum Reference

### TaskType (TMTask.type)

| Value | Name | Description |
|-------|------|-------------|
| 0 | Todo | A to-do item (actionable task) |
| 1 | Project | A project (container for todos and headings) |
| 2 | Heading | A heading (organizational grouping within a project) |

### Status (TMTask.status)

| Value | Name | Description |
|-------|------|-------------|
| 0 | Incomplete | Not yet done |
| 2 | Canceled | Explicitly canceled (note: value 1 is unused) |
| 3 | Completed | Done |

### StartBucket (TMTask.start)

| Value | Name | Description |
|-------|------|-------------|
| 0 | Inbox | Unprocessed, in the inbox |
| 1 | Anytime | Scheduled for "anytime" (active) |
| 2 | Someday | Deferred to "someday" |

### Date Encoding

Things uses a custom binary date format stored as integers:

- **Date** (startDate, deadline): `YYYYYYYYYYYMMMMDDDDD0000000` (27-bit integer)
- **Time** (reminderTime): `hhhhhmmmmmm00000000000000000000`
- **Timestamp** (stopDate, creationDate, userModificationDate): Core Data epoch (seconds since 2001-01-01)

## Complete Data Inventory

### TMTask: Todos (type=0)

#### Incomplete Todos (status=0)

##### Start = Inbox (start=0)

| UUID | Title | Trashed | Notes | Area | Project | Heading | StartDate | Deadline | Reminder | Repeating | Checklist |
|------|-------|---------|-------|------|---------|---------|-----------|----------|----------|-----------|-----------|
| `DfYoiXcNLQssk9DkSoJV3Y` | To-Do in Inbox | no | "With\nNotes" | - | - | - | - | - | - | - | 0 items |
| `3Eva4XFof6zWb9iSfYy4ej` | To-Do in Inbox with Checklist Items | no | - | - | - | - | - | - | - | - | 3 items |
| `Tmt3omxtu61wAoTaSLweQy` | Another Deleted Todo | **yes** | - | - | - | - | - | - | - | - | 0 items |

##### Start = Anytime (start=1)

| UUID | Title | Trashed | Notes | Area | Project | Heading | StartDate | Deadline | Reminder | Repeating | Checklist | Tags |
|------|-------|---------|-------|------|---------|---------|-----------|----------|----------|-----------|-----------|------|
| `5pUx6PESj3ctFYbgth1PXY` | To-Do in Today | no | "With\nNotes" | - | - | - | 132464128 | - | - | - | 0 | - |
| `QqhVksfbsAVaNnwB1x3CuD` | To-Do in Anytime | no | "With\nNotes" | - | - | - | - | - | - | - | 0 | - |
| `E18tg5qepzrQk9J6jQtb5C` | To-Do in Project | no | "With\nNotes" | - | Project without Area | - | - | - | - | - | 0 | - |
| `Q7uN9y3jp5ChZAGjZJhMfY` | To-Do in Area 1 | no | "With\nNotes" | Area 1 | - | - | - | - | - | - | 0 | - |
| `HbKGAeZKFDkWH5osSBNHvz` | To-Do in Heading | no | "With\nNotes" | - | - | Heading | - | 133739008 | - | - | 0 | - |
| `A2oPvtt4dXoypeoLc8uYzY` | Deleted Todo | **yes** | "With\nNotes" | - | Cancelled Project in Area | - | - | - | - | - | 0 | - |
| `K9bx7h1xCJdevvyWardZDq` | Repeating To-Do | no | - | - | - | - | 132434304 | 132464128 | - | **yes** (template: `N1PJHsbjct4mb1bhcs7aHa`) | 0 | - |
| `EJxkdyCLyjJx6wucDPUcvu` | Todo in Area 3 | no | - | Area 3 | - | - | - | - | - | - | 0 | - |
| `W5JYfjY2xtLdmedQKU6caM` | Todo in Area 1 | no | - | - | Project in Area 1 | - | - | - | - | - | 0 | Errand, Home |
| `NoQLFamrMMooAELuBznao8` | Task in Deleted Project | no | - | - | Deleted Project | - | - | - | - | - | 0 | - |
| `KisAmSsnzCcRRumjY4TkVV` | Overdue Todo automatically shown in Today | no | - | - | Project in Area 1 | - | - | 132471424 | - | - | 0 | - |
| `Cc73oaq1C2mDMpZZUJaBxe` | Overdue Todo not shown in Today | no | - | - | Project in Area 1 | - | - | 132471424 | - | - | 0 | - |
| `LeMGbV4SZyMgSRTCWq72Yg` | Deleted Task in Deleted Project | **yes** | - | - | Deleted Project | - | - | - | - | - | 0 | - |

##### Start = Someday (start=2)

| UUID | Title | Trashed | Notes | Area | Project | Heading | StartDate | Deadline | Reminder | Repeating |
|------|-------|---------|-------|------|---------|---------|-----------|----------|----------|-----------|
| `7F4vqUNiTvGKaCUfv5pqYG` | To-Do in Upcoming | no | "With\nNotes" | - | - | - | 132814976 | - | 840957952 | - |
| `JLYSEPFkLfBC5rhGJRa5S1` | To-Do in Someday | no | "With\nNotes" | - | - | - | - | - | - | - |
| `N1PJHsbjct4mb1bhcs7aHa` | Repeating To-Do | no | - | - | - | - | - | 262213760 | 805306368 | - |
| `6Hf2qWBjWhq7B1xszwdo34` | Upcoming To-Do in Today (yellow) | no | - | - | - | - | 132469248 | - | - | - |

#### Canceled Todos (status=2)

| UUID | Title | Start | Trashed | StopDate | Area | Project | Heading |
|------|-------|-------|---------|----------|------|---------|---------|
| `9DyzgLkZf1cBDbJ2dYFGBR` | Cancelled To-Do in Inbox | 0 (inbox) | no | 1616958852.56 | - | - | - |
| `SzgXfYgNV4kWp5anvjsdJT` | Cancelled To-Do in Today | 1 (anytime) | no | 1616958854.87 | - | - | - |
| `Ak7cN3VDSnpW6MQt7tf4cd` | Cancelled To-Do in Anytime | 1 | no | 1616958869.56 | - | - | - |
| `5HLnvorXMbqcbjUuPN6ywi` | Cancelled To-Do in Project | 1 | no | 1616958929.65 | - | Project without Area | - |
| `BWzcy7ZSQ6T48AX8vsaPC8` | Cancelled To-Do in Area 1 | 1 | no | 1616958929.65 | Area 1 | - | - |
| `RqRi38gMxTFyhPh2X1vH1i` | Cancelled To-Do in Heading | 1 | no | 1616958852.56 | - | - | Heading |
| `8xADVLww2JjHu3BWV4iK2U` | Cancelled Deleted Todo | 1 | **yes** | 1616960010.34 | Area 1 | - | - |
| `NsEyVWNres9441aCBtz9bF` | Cancelled To-Do in cancelled Project in Area | 1 | no | 1616958929.65 | - | Cancelled Project in Area | - |
| `ADLex1EmJzLpu2GHxFvLvc` | Cancelled To-Do in Upcoming | 2 (someday) | no | 1616958864.48 | - | - | - |
| `SuSafUtGHGKatpo3rqUdsh` | Cancelled To-Do in Logbook | 2 | no | 1616958799.96 | - | - | - |
| `DkVUPkCVM9mNq8yQuLrDo` | Cancelled To-Do in Someday | 2 | no | 1616958880.27 | - | - | - |

#### Completed Todos (status=3)

| UUID | Title | Start | Trashed | StopDate | Area | Project | Heading |
|------|-------|-------|---------|----------|------|---------|---------|
| `LgqUAQAdNsS3CGHok4EjLa` | Completed To-Do in Inbox | 0 (inbox) | no | 1616958852.01 | - | - | - |
| `56dtXSk3A373M6n4eqGyr3` | Completed To-Do in Today | 1 (anytime) | no | 1616958854.38 | - | - | - |
| `NSzDo18ibpJ1H8xStXLvto` | Completed To-Do in Anytime | 1 | no | 1616958868.97 | - | - | - |
| `Eb85bTZUJWjJe8ppmaRzjN` | Completed Deleted Todo | 1 | **yes** | 1616960019.32 | Area 1 | - | - |
| `5u2yGhP4rMQUmPQYEpGYDd` | Completed To-Do in Project | 1 | no | 1616958932.53 | - | Project without Area | - |
| `UwNEL2WdQTd92ZLa2HkHnc` | Completed To-Do in Area 1 | 1 | no | 1616958932.53 | Area 1 | - | - |
| `2qBNNhNuDUBEGcB2tVRH9W` | Completed To-Do in Heading | 1 | no | 1616958852.01 | - | - | Heading |
| `S8QU6gEvQec7XRMkN5Vjwg` | Completed To-Do in cancelled Project in Area | 1 | no | 1616958932.53 | - | Cancelled Project in Area | - |
| `LnGwkFDZw78ydwp98jqo3z` | Completed To-Do before midnight UTC+0 | 1 | no | 1718668800.00 | - | Project without Area | - |
| `JM91cry5BMFP7R3vXDns9z` | Completed To-Do after midnight UTC+0 | 1 | no | 1718668800.00 | - | Project without Area | - |
| `LE2WEGxANmtHWD3c9g5iWA` | Completed To-Do in Upcoming | 2 (someday) | no | 1616958861.56 | - | - | - |
| `6gM3LexGhMGawEjGmKm3Z4` | Completed To-Do in Logbook | 2 | no | 1616958799.11 | - | - | - |
| `WQ8p2mhuHWd7g9tMJfed2W` | Completed To-Do in Someday | 2 | no | 1616958879.77 | - | - | - |

### TMTask: Projects (type=1)

| UUID | Title | Status | Start | Trashed | Area | StartDate | StopDate | Actions (total/open) |
|------|-------|--------|-------|---------|------|-----------|----------|----------------------|
| `TCozQqXVbB2TJkXXXQj2H9` | Project without Area | incomplete | 1 (anytime) | no | - | - | - | 5 / 1 |
| `3x1QqJqfvZyhtw8NSdnZqG` | Project in Area 1 | incomplete | 1 (anytime) | no | Area 1 | - | - | 6 / 4 |
| `PgsWnDkzXRz6zvofTqtHqn` | Project in Today | incomplete | 1 (anytime) | no | - | 132466816 | - | 0 / 0 |
| `Tc7DABDNNMZvV4ZGB8tLDh` | Deleted Project | incomplete | 1 (anytime) | **yes** | - | - | - | 1 / 1 |
| `SkLdfSe1MXR5vMV1gMYkHE` | Cancelled Project in Area | canceled | 1 (anytime) | no | Area 1 | - | 1616959972.93 | 2 / 0 |
| `CmpltdProjTestFixture01` | Completed Project | completed | 1 (anytime) | no | Area 1 | - | 1616959500.0 | 0 / 0 |
| `SmdyProjTestFixture001` | Project in Someday | incomplete | 2 (someday) | no | - | - | - | 0 / 0 |

### TMTask: Headings (type=2)

| UUID | Title | Status | Start | Trashed | Project | Actions (total/open) |
|------|-------|--------|-------|---------|---------|----------------------|
| `6QpDLSHZMRAUSAeZ9mNvgt` | Heading | incomplete | 1 (anytime) | no | Project in Area 1 | 3 / 1 |
| `CmpltdHdngTestFixture1` | Completed Heading | completed | 1 (anytime) | no | Project without Area | 0 / 0 |
| `AddtnlHdngTestFixture1` | Empty Heading | incomplete | 1 (anytime) | no | Project without Area | 0 / 0 |

Tasks under "Heading" (in Project in Area 1):

| UUID | Title | Status |
|------|-------|--------|
| `HbKGAeZKFDkWH5osSBNHvz` | To-Do in Heading | incomplete |
| `RqRi38gMxTFyhPh2X1vH1i` | Cancelled To-Do in Heading | canceled |
| `2qBNNhNuDUBEGcB2tVRH9W` | Completed To-Do in Heading | completed |

### TMArea

| UUID | Title | Visible | Index | Tags | Todo Count | Project Count |
|------|-------|---------|-------|------|------------|---------------|
| `DciSFacytdrNG1nRaMJPgY` | Area 1 | NULL | 0 | Errand, Important | 3 | 2 |
| `3UXZmXt9qNMTWL5iZNyrxj` | Area 2 | NULL | -343 | - | 0 | 0 |
| `Y3JC4XeyGWxzDocQL4aobo` | Area 3 | NULL | -723 | - | 1 | 0 |

### TMTag

| UUID | Title | Shortcut | Parent | Index | Task Count | Area Count |
|------|-------|----------|--------|-------|------------|------------|
| `H96sVJwE7VJveAnv7itmux` | Errand | - | - | 0 | 1 | 1 |
| `CK9dARrf2ezbFvrVUUxkHE` | Home | - | - | 592 | 1 | 0 |
| `Qt2AY87x2QDdowSn9HKTt1` | Office | - | - | 1000 | 1 | 0 |
| `XdDBCjmEXEhjZy9A2wFFKP` | Important | - | - | 1595 | 1 | 1 |
| `BULfa35PCAn1LtsmBA6A2u` | Pending | - | - | 2020 | 1 | 0 |

### TMChecklistItem

All belong to task `3Eva4XFof6zWb9iSfYy4ej` (To-Do in Inbox with Checklist Items).
Parent task has checklistItemsCount=3, openChecklistItemsCount=2.

| UUID | Title | Status | StopDate |
|------|-------|--------|----------|
| `Ka8uwUstDgQWkugYyVHB1a` | Item 1 | 0 (incomplete) | - |
| `UR9qjvuykBsv2dp8yPzWGT` | Item 2 | 0 (incomplete) | - |
| `XufyKEcAa9vAUxiJuwChK` | Item 3 | 3 (completed) | 1616959000.0 |

### Relationship Tables

#### TMTaskTag (Task-Tag associations)

| Task | Tag |
|------|-----|
| Todo in Area 1 (`W5JYfjY2xtLdmedQKU6caM`) | Errand (`H96sVJwE7VJveAnv7itmux`) |
| Todo in Area 1 (`W5JYfjY2xtLdmedQKU6caM`) | Home (`CK9dARrf2ezbFvrVUUxkHE`) |
| To-Do in Today (`5pUx6PESj3ctFYbgth1PXY`) | Office (`Qt2AY87x2QDdowSn9HKTt1`) |
| To-Do in Anytime (`QqhVksfbsAVaNnwB1x3CuD`) | Pending (`BULfa35PCAn1LtsmBA6A2u`) |
| To-Do in Project (`E18tg5qepzrQk9J6jQtb5C`) | Important (`XdDBCjmEXEhjZy9A2wFFKP`) |

#### TMAreaTag (Area-Tag associations)

| Area | Tag |
|------|-----|
| Area 1 (`DciSFacytdrNG1nRaMJPgY`) | Errand (`H96sVJwE7VJveAnv7itmux`) |
| Area 1 (`DciSFacytdrNG1nRaMJPgY`) | Important (`XdDBCjmEXEhjZy9A2wFFKP`) |

## Test Constants Mapping

| Constant | Value | Derivation |
|----------|-------|------------|
| `Headings` | 3 | Tasks under the heading (incomplete + canceled + completed) |
| `Inbox` | 2 | type=0, status=0, start=0, trashed=0 |
| `TrashedTodos` | 3 | type=0, trashed=1 (Another Deleted Todo + Deleted Todo + Deleted Task in Deleted Project) |
| `TrashedProjects` | 1 | type=1, trashed=1 (Deleted Project) |
| `TrashedCanceled` | 1 | trashed=1, status=2 (Cancelled Deleted Todo) |
| `TrashedCompleted` | 1 | trashed=1, status=3 (Completed Deleted Todo) |
| `TrashedProjectTodos` | 1 | type=0, trashed=0, project.trashed=1 but task.trashed=0 (Task in Deleted Project; excludes LeMGbV4 which is self-trashed) |
| `Trashed` | 6 | All records with trashed=1 (3 todos + 1 deleted todo in project + 1 project + 1 canceled) |
| `Projects` | 5 | type=1, status=0 (excludes canceled and completed projects) |
| `ProjectsNotTrashed` | 4 | type=1, status=0, trashed=0 |
| `Upcoming` | 1 | start=2, startDate exists, startDate is future, status=0 (To-Do in Upcoming) |
| `DeadlinePast` | 3 | deadline exists, deadline is past, status=0 |
| `DeadlineFuture` | 1 | deadline exists, deadline is future, status=0 |
| `Deadlines` | 4 | DeadlinePast + DeadlineFuture |
| `TodayProjects` | 1 | type=1, startDate exists, start=1 (Project in Today) |
| `TodayTasks` | 4 | Tasks shown in Today view (startDate=today + overdue + upcoming-today) |
| `Today` | 5 | TodayProjects + TodayTasks |
| `Anytime` | 15 | start=1, status=0, contextTrashed=false (excludes task in deleted project) |
| `Logbook` | 25 | status=2 or status=3, contextTrashed=false (canceled + completed) |
| `Canceled` | 11 | status=2, contextTrashed=false |
| `Completed` | 14 | status=3, contextTrashed=false |
| `Someday` | 2 | start=2, startDate not exists, status=0 (To-Do in Someday + Project in Someday) |
| `Tags` | 5 | Total tag count |
| `Areas` | 3 | Total area count |
| `TodosIncomplete` | 15 | type=0, status=0, contextTrashed=false |
| `TodosAnytime` | 11 | type=0, start=1, status=0 (includes contextTrashed) |
| `TodosAnytimeComplete` | 8 | type=0, start=1, status=3 |
| `TodosComplete` | 12 | type=0, status=3 |
| `TasksIncomplete` | 22 | status=0 (includes contextTrashed, all types) |
| `TasksIncompleteFiltered` | 21 | status=0, contextTrashed=false (TasksIncomplete minus 1 context-trashed) |
| `TasksInProjectAll` | 7 | Tasks in "Project without Area" (all statuses including heading) |
| `TasksInProject` | 5 | Tasks in "Project without Area" (status=0 only) |
| `ProjectItems` | 7 | Items in "Project in Area 1" (todos + headings, nested) |
| `DatabaseVersion` | 24 | Meta table databaseVersion value |

Note: `Headings = 3` counts the tasks belonging to the single heading, not the number of heading records.

## Coverage Analysis

### Well Covered

- **Todo status combinations**: All 3 statuses (incomplete/canceled/completed) x all 3 start buckets (inbox/anytime/someday) are represented
- **Trashed variants**: Trashed todos, trashed projects, trashed tasks in trashed projects, context-trashed tasks
- **Date fields**: startDate, deadline, stopDate all have multiple samples
- **ContextTrashed logic**: Task in Deleted Project tests the parent-trashed-but-task-not-trashed case
- **Repeating tasks**: Both template and instance exist
- **UTC boundary**: Two completed todos testing before/after midnight UTC+0

### Coverage Gaps

#### Heading (Partially Addressed)

| Gap | Description | Impact | Status |
|-----|-------------|--------|--------|
| Only 1 heading exists | Cannot test multiple headings in a project | `InHeading()` filter only tested with 1 UUID | **Fixed**: Added 2 headings in "Project without Area" |
| No heading in other projects | All heading data is in "Project in Area 1" | Cross-project heading queries not tested | **Fixed**: New headings are in "Project without Area" |
| No trashed heading | No type=2 with trashed=1 | `Type().Heading().Trashed(true)` never tested | Open |
| No empty heading | The only heading has 3 tasks | Edge case of heading with 0 items not tested | **Fixed**: "Empty Heading" has 0 tasks |
| testHeadings=3 is misleading | Actually counts tasks under heading, not heading records | May confuse downstream users | Open (naming issue) |

#### Tag (Partially Addressed)

| Gap | Description | Impact | Status |
|-----|-------------|--------|--------|
| Only 1 task has tags | "Todo in Area 1" has Errand + Home | `InTag()` filter tested with minimal data | **Fixed**: 3 more tasks have tag associations |
| No nested tags | TMTag.parent is NULL for all tags | `WithParent()` query never returns data | Open |
| No tag shortcuts | TMTag.shortcut is empty for all tags | Tag.Shortcut field always empty in tests | Open |
| No project with tags | Only tasks and areas have tag associations | Tag-project relationship not tested | Open |

#### Project (Partially Addressed)

| Gap | Description | Impact | Status |
|-----|-------------|--------|--------|
| No completed project | Only incomplete(3) and canceled(1) | `Type().Project().Status().Completed()` returns 0 | **Fixed**: "Completed Project" added |
| No someday project | All projects have start=1(anytime) | `Type().Project().Start().Someday()` returns 0 | **Fixed**: "Project in Someday" added |
| No project with deadline | No project has deadline set | Project deadline queries untested | Open |
| No project with notes | All projects have empty notes | Project.Notes always empty in tests | Open |
| No project with tags | TMTaskTag has no project entries | `HasTag()` on projects untested | Open |

#### Area (Minor)

| Gap | Description | Impact |
|-----|-------------|--------|
| Area 2 is completely empty | No tasks, no projects, no tags | Only tests "area exists with no items" |
| visible field always NULL | No area with explicit visible=true/false | `Visible()` filter not meaningfully tested |

#### ChecklistItem (Partially Addressed)

| Gap | Description | Impact | Status |
|-----|-------------|--------|--------|
| All items are incomplete | No status=3(completed) or status=2(canceled) | ChecklistItem.IsCompleted()/IsCanceled() return false on all data | **Fixed**: Item 3 is now completed with stopDate |
| Only 1 task has checklist | 3 items all on same task | No variety in checklist ownership | Open |
| No checklist stopDate | All stopDate are NULL | ChecklistItem.StopDate always nil | **Fixed**: Item 3 has stopDate |

#### Task Fields (Minor)

| Gap | Description | Impact |
|-----|-------------|--------|
| contact always NULL | No task has a contact | Task model likely does not expose this, low priority |
| deadlineSuppressionDate always NULL | No suppressed deadlines | `WithDeadlineSuppressed()` tested but with no matching data |
| Only 1 task has reminderTime | To-Do in Upcoming | Reminder parsing tested minimally |
| startBucket always default | Field exists but unclear if used | May be legacy field |

## Remaining Recommendations

Priority order for further enriching test data:

1. **Add nested tags** with parent-child relationship
2. **Add tag shortcuts** to at least one tag
3. **Add a trashed heading** (type=2, trashed=1)
4. **Add project with deadline** and notes
5. **Add project-tag associations** (TMTaskTag with project UUID)
6. **Set Area.visible** explicitly on at least one area
