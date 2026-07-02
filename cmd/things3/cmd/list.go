package cmd

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
)

// Shared list-flag names.
const (
	flagPage = "page"
	flagAll  = "all"
	flagSort = "sort"
	flagDesc = "desc"
	flagTag  = "tag"
)

// defaultPageSize is the page size when -n/--limit is not set, applied uniformly
// across text, json, and yaml output.
const defaultPageSize = 10

// sortKey names the field a list is sorted by. The empty value means the view's
// natural order is preserved.
type sortKey string

const (
	sortNone     sortKey = ""
	sortDate     sortKey = "date"
	sortCreated  sortKey = "created"
	sortModified sortKey = "modified"
	sortTitle    sortKey = "title"
)

// sortValue adapts sortKey to pflag.Value so an invalid --sort argument fails at
// parse time rather than silently defaulting.
type sortValue sortKey

// newSortValue returns a sortValue defaulting to the view's natural order.
func newSortValue() *sortValue {
	s := sortValue(sortNone)
	return &s
}

func (s *sortValue) String() string { return string(*s) }

func (s *sortValue) Set(v string) error {
	switch sortKey(v) {
	case sortDate, sortCreated, sortModified, sortTitle:
		*s = sortValue(v)
		return nil
	default:
		return errors.New("must be one of: date, created, modified, title")
	}
}

func (s *sortValue) Type() string { return "sort" }

// pageValue adapts a 1-based page number to pflag.Value so a value below 1 fails
// at parse time.
type pageValue int

// newPageValue returns a pageValue defaulting to the first page.
func newPageValue() *pageValue {
	p := pageValue(1)
	return &p
}

func (p *pageValue) String() string { return strconv.Itoa(int(*p)) }

func (p *pageValue) Set(v string) error {
	n, err := strconv.Atoi(v)
	if err != nil {
		return errors.New("must be an integer")
	}
	if n < 1 {
		return errors.New("must be >= 1")
	}
	*p = pageValue(n)
	return nil
}

func (p *pageValue) Type() string { return "page" }

// listFlags is the resolved shared list state for one command invocation.
type listFlags struct {
	page     int
	all      bool
	sortKey  sortKey
	desc     bool
	tag      string
	limit    int
	limitSet bool
	format   outputFormat
}

// readListFlags resolves the shared list flags plus -o/-n for a command.
func readListFlags(cmd *cobra.Command) listFlags {
	lf := listFlags{page: 1}
	_, lf.format = getOutput(cmd)
	if f := cmd.Flags().Lookup(flagPage); f != nil {
		if p, ok := f.Value.(*pageValue); ok {
			lf.page = int(*p)
		}
	}
	if f := cmd.Flags().Lookup(flagSort); f != nil {
		if s, ok := f.Value.(*sortValue); ok {
			lf.sortKey = sortKey(*s)
		}
	}
	lf.all, _ = cmd.Flags().GetBool(flagAll)
	lf.desc, _ = cmd.Flags().GetBool(flagDesc)
	lf.tag, _ = cmd.Flags().GetString(flagTag)
	lf.limit, _ = cmd.Flags().GetInt(flagLimit)
	lf.limitSet = cmd.Flags().Changed(flagLimit)
	return lf
}

// pageSize returns the effective page size and whether pagination is disabled.
// --all or an explicit -n <= 0 mean unlimited; otherwise -n sets the size and an
// unset -n falls back to defaultPageSize. Pagination is uniform across output
// formats: json/yaml paginate exactly like text and describe the resulting slice
// through the list envelope's total/page/pages fields instead of truncating.
func (lf *listFlags) pageSize() (size int, unlimited bool) {
	switch {
	case lf.all:
		return 0, true
	case lf.limitSet:
		if lf.limit <= 0 {
			return 0, true
		}
		return lf.limit, false
	default:
		return defaultPageSize, false
	}
}

// pageMeta describes one rendered page for the footer.
type pageMeta struct {
	total     int
	start     int // 1-based index of the first row, 0 when the page is empty
	end       int // 1-based index of the last row, 0 when the page is empty
	page      int
	pages     int
	unlimited bool
}

// listAccessors exposes the fields the shared list pipeline needs from an item
// type, so todos, projects, and mixed items share one implementation.
type listAccessors[T any] struct {
	tags     func(T) []string
	date     func(T) *time.Time
	created  func(T) time.Time
	modified func(T) time.Time
	title    func(T) string
}

// applyListPipeline runs the uniform tag-filter -> sort -> paginate steps and
// returns the page slice together with its footer metadata.
func applyListPipeline[T any](items []T, acc listAccessors[T], lf *listFlags) ([]T, pageMeta) {
	if lf.tag != "" {
		items = filterByTag(items, acc.tags, lf.tag)
	}
	if lf.sortKey != sortNone {
		slices.SortStableFunc(items, func(a, b T) int {
			return compareListItems(a, b, acc, lf.sortKey, lf.desc)
		})
	}
	total := len(items)

	size, unlimited := lf.pageSize()
	if unlimited {
		return items, pageMeta{total: total, page: 1, pages: 1, unlimited: true}
	}

	pages := max((total+size-1)/size, 1)
	meta := pageMeta{total: total, page: lf.page, pages: pages}
	offset := (lf.page - 1) * size
	if offset >= total {
		return []T{}, meta
	}
	end := min(offset+size, total)
	meta.start = offset + 1
	meta.end = end
	return items[offset:end], meta
}

// filterByTag keeps items carrying want (compared case-insensitively) in any of
// their tags.
func filterByTag[T any](items []T, tagsOf func(T) []string, want string) []T {
	out := make([]T, 0, len(items))
	for _, it := range items {
		if slices.ContainsFunc(tagsOf(it), func(tag string) bool {
			return strings.EqualFold(tag, want)
		}) {
			out = append(out, it)
		}
	}
	return out
}

// compareListItems orders two items by the selected sort key. Date sorts nil
// last in both directions; other keys simply reverse under --desc.
func compareListItems[T any](a, b T, acc listAccessors[T], key sortKey, desc bool) int {
	switch key {
	case sortDate:
		return compareDatePtr(acc.date(a), acc.date(b), desc)
	case sortCreated:
		return signed(acc.created(a).Compare(acc.created(b)), desc)
	case sortModified:
		return signed(acc.modified(a).Compare(acc.modified(b)), desc)
	case sortTitle:
		return signed(strings.Compare(strings.ToLower(acc.title(a)), strings.ToLower(acc.title(b))), desc)
	default:
		return 0
	}
}

// signed negates a comparison result when descending order is requested.
func signed(c int, desc bool) int {
	if desc {
		return -c
	}
	return c
}

// compareDatePtr orders two dates, always ranking a nil date last regardless of
// direction.
func compareDatePtr(a, b *time.Time, desc bool) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return 1
	case b == nil:
		return -1
	case desc:
		return b.Compare(*a)
	default:
		return a.Compare(*b)
	}
}

// todoSortDate returns the display-relevant date for a todo with the same
// precedence as todoRelevantDate: completed, else canceled, else deadline, else
// start.
func todoSortDate(t *things3.Todo) *time.Time {
	switch {
	case t.CompletedAt != nil:
		return t.CompletedAt
	case t.CanceledAt != nil:
		return t.CanceledAt
	case t.Deadline != nil:
		return t.Deadline
	case t.StartDate != nil:
		return t.StartDate
	default:
		return nil
	}
}

// projectSortDate mirrors todoSortDate for projects.
func projectSortDate(p *things3.Project) *time.Time {
	switch {
	case p.CompletedAt != nil:
		return p.CompletedAt
	case p.CanceledAt != nil:
		return p.CanceledAt
	case p.Deadline != nil:
		return p.Deadline
	case p.StartDate != nil:
		return p.StartDate
	default:
		return nil
	}
}

// todoAccessors adapts things3.Todo to the list pipeline.
//
//nolint:dupl // type-specific adapter, mirrors projectAccessors
var todoAccessors = listAccessors[things3.Todo]{
	tags:     func(t things3.Todo) []string { return t.Tags },
	date:     func(t things3.Todo) *time.Time { return todoSortDate(&t) },
	created:  func(t things3.Todo) time.Time { return t.CreatedAt },
	modified: func(t things3.Todo) time.Time { return t.ModifiedAt },
	title:    func(t things3.Todo) string { return t.Title },
}

// projectAccessors adapts things3.Project to the list pipeline.
//
//nolint:dupl // type-specific adapter, mirrors todoAccessors
var projectAccessors = listAccessors[things3.Project]{
	tags:     func(p things3.Project) []string { return p.Tags },
	date:     func(p things3.Project) *time.Time { return projectSortDate(&p) },
	created:  func(p things3.Project) time.Time { return p.CreatedAt },
	modified: func(p things3.Project) time.Time { return p.ModifiedAt },
	title:    func(p things3.Project) string { return p.Title },
}

// mixedAccessors adapts a mixedItem to the list pipeline by delegating to its
// underlying todo or project.
var mixedAccessors = listAccessors[mixedItem]{
	tags: func(m mixedItem) []string {
		if m.Project != nil {
			return m.Project.Tags
		}
		return m.Todo.Tags
	},
	date: func(m mixedItem) *time.Time {
		if m.Project != nil {
			return projectSortDate(m.Project)
		}
		return todoSortDate(m.Todo)
	},
	created: func(m mixedItem) time.Time {
		if m.Project != nil {
			return m.Project.CreatedAt
		}
		return m.Todo.CreatedAt
	},
	modified: func(m mixedItem) time.Time {
		if m.Project != nil {
			return m.Project.ModifiedAt
		}
		return m.Todo.ModifiedAt
	},
	title: func(m mixedItem) string {
		if m.Project != nil {
			return m.Project.Title
		}
		return m.Todo.Title
	},
}

// areaAccessors adapts things3.Area to the list pipeline. Areas carry no dates or
// timestamps, so those accessors yield nil and the zero time; --tag filters on
// the area's own tags.
var areaAccessors = listAccessors[things3.Area]{
	tags:     func(a things3.Area) []string { return a.Tags },
	date:     func(things3.Area) *time.Time { return nil },
	created:  func(things3.Area) time.Time { return time.Time{} },
	modified: func(things3.Area) time.Time { return time.Time{} },
	title:    func(a things3.Area) string { return a.Title },
}

// tagAccessors adapts things3.Tag to the list pipeline. A tag has no nested tags,
// dates, or timestamps, so --tag filters by the tag's own title (equality) and
// the date/timestamp accessors yield nil and the zero time.
var tagAccessors = listAccessors[things3.Tag]{
	tags:     func(t things3.Tag) []string { return []string{t.Title} },
	date:     func(things3.Tag) *time.Time { return nil },
	created:  func(things3.Tag) time.Time { return time.Time{} },
	modified: func(things3.Tag) time.Time { return time.Time{} },
	title:    func(t things3.Tag) string { return t.Title },
}

// outputTodoList applies the shared pipeline to a todo view and renders the
// page. group, when non-nil and returning a non-empty result, sections the page
// in text mode; opts controls the container segment on each row.
func outputTodoList(cmd *cobra.Command, todos []things3.Todo, group func([]things3.Todo) []todoGroup, opts rowOptions) error {
	lf := readListFlags(cmd)
	page, meta := applyListPipeline(todos, todoAccessors, &lf)
	w := cmd.OutOrStdout()
	switch lf.format {
	case formatJSON, formatYAML:
		return writeListEnvelope(w, page, meta, lf.format)
	default:
		var groups []todoGroup
		if group != nil {
			groups = group(page)
		}
		if len(groups) > 0 {
			if err := writeGroupedTodos(w, groups, opts); err != nil {
				return err
			}
		} else if err := writeTodos(w, page, opts); err != nil {
			return err
		}
		return writeListFooter(w, meta)
	}
}

// outputList runs the shared pipeline over a flat (ungrouped) list and renders
// the page: the self-describing envelope for json/yaml, or writeText plus the
// pagination footer for text.
func outputList[T any](cmd *cobra.Command, items []T, acc listAccessors[T], writeText func(io.Writer, []T) error) error {
	lf := readListFlags(cmd)
	page, meta := applyListPipeline(items, acc, &lf)
	w := cmd.OutOrStdout()
	switch lf.format {
	case formatJSON, formatYAML:
		return writeListEnvelope(w, page, meta, lf.format)
	default:
		if err := writeText(w, page); err != nil {
			return err
		}
		return writeListFooter(w, meta)
	}
}

// outputProjectList applies the shared pipeline to a project list and renders
// the page.
func outputProjectList(cmd *cobra.Command, projects []things3.Project) error {
	return outputList(cmd, projects, projectAccessors, writeProjects)
}

// outputMixedList applies the shared pipeline to a cross-type list and renders
// the page.
func outputMixedList(cmd *cobra.Command, items []mixedItem) error {
	return outputList(cmd, items, mixedAccessors, writeMixed)
}

// outputAreas applies the shared pipeline to an area list and renders the page.
func outputAreas(cmd *cobra.Command, areas []things3.Area) error {
	return outputList(cmd, areas, areaAccessors, writeAreas)
}

// outputTags applies the shared pipeline to a tag list and renders the page.
func outputTags(cmd *cobra.Command, tags []things3.Tag) error {
	return outputList(cmd, tags, tagAccessors, writeTags)
}

// writeListFooter prints the pagination footer in text mode. It appears only
// when the list spans more than one page or a non-first page is shown, and never
// for json/yaml or unlimited output.
func writeListFooter(w io.Writer, m pageMeta) error {
	if m.unlimited {
		return nil
	}
	shown := 0
	if m.start > 0 {
		shown = m.end - m.start + 1
	}
	if m.total <= shown && m.page <= 1 {
		return nil
	}
	line := fmt.Sprintf("-- %d-%d of %d (page %d/%d)", m.start, m.end, m.total, m.page, m.pages)
	if m.page < m.pages {
		line += fmt.Sprintf(" | next: --page %d", m.page+1)
	}
	line += " | all: --all"
	_, err := fmt.Fprintln(w, line)
	return err
}
