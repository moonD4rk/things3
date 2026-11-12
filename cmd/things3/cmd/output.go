package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
)

// outputFormat is the rendering format selected by the --text/--json/--yaml
// switches.
type outputFormat string

const (
	formatText outputFormat = "text"
	formatJSON outputFormat = "json"
	formatYAML outputFormat = "yaml"
)

// rowOptions controls optional segments of a one-line row.
type rowOptions struct {
	// showContainer appends the "@project / heading" or "@area" segment. Views
	// that already group by container (anytime) turn it off to avoid repetition.
	showContainer bool
}

// defaultRow renders every optional segment, including the container.
var defaultRow = rowOptions{showContainer: true}

// noContainerRow suppresses the container segment for container-grouped views.
var noContainerRow = rowOptions{showContainer: false}

// itemContainer returns the "@Project / Heading" or "@Area" segment of a row, or
// an empty string for an item that stands on its own. A project is contained
// by its area; the heading is shown only alongside its project.
func itemContainer(it things3.Item) string {
	if it.Project != nil {
		return prefixed("@", it.Project.AreaTitle)
	}
	t := it.Todo
	if t.ProjectTitle == "" {
		return prefixed("@", t.AreaTitle)
	}
	if t.HeadingTitle == "" {
		return "@" + t.ProjectTitle
	}
	return "@" + t.ProjectTitle + " / " + t.HeadingTitle
}

// prefixed returns prefix+s, or "" when s is empty.
func prefixed(prefix, s string) string {
	if s == "" {
		return ""
	}
	return prefix + s
}

// resolveFormat maps the mutually exclusive --json/--yaml switches to a format,
// defaulting to text. When both are somehow set (an invalid combination surfaced
// only while rendering that very error) it falls back to text so the message
// reads plainly.
func resolveFormat(cmd *cobra.Command) outputFormat {
	jsonSet, _ := cmd.Flags().GetBool(flagJSON)
	yamlSet, _ := cmd.Flags().GetBool(flagYAML)
	switch {
	case jsonSet && yamlSet:
		return formatText
	case jsonSet:
		return formatJSON
	case yamlSet:
		return formatYAML
	default:
		return formatText
	}
}

// outputTodoDetail renders one todo in full detail honoring the format switches.
func outputTodoDetail(cmd *cobra.Command, todo *things3.Todo) error {
	format := resolveFormat(cmd)
	switch format {
	case formatJSON:
		return writeJSON(cmd.OutOrStdout(), todo)
	case formatYAML:
		return writeYAML(cmd.OutOrStdout(), todo)
	default:
		return writeTodoDetail(cmd.OutOrStdout(), todo)
	}
}

// projectDetailOutput is the structured JSON/YAML shape for a project detail.
type projectDetailOutput struct {
	Project *things3.Project `json:"project" yaml:"project"`
	Todos   []things3.Todo   `json:"todos" yaml:"todos"`
}

// outputProjectDetail renders one project with its incomplete todos honoring the
// format switches.
func outputProjectDetail(cmd *cobra.Command, project *things3.Project, todos []things3.Todo) error {
	format := resolveFormat(cmd)
	w := cmd.OutOrStdout()
	switch format {
	case formatJSON:
		return writeJSON(w, projectDetailOutput{Project: project, Todos: todos})
	case formatYAML:
		return writeYAML(w, projectDetailOutput{Project: project, Todos: todos})
	default:
		if err := writeProjectDetail(w, project); err != nil {
			return err
		}
		if len(todos) > 0 {
			fmt.Fprintln(w)
			return writeItems(w, todoItems(todos), defaultRow)
		}
		return nil
	}
}

// todoItems wraps todos as list items.
func todoItems(todos []things3.Todo) []things3.Item {
	items := make([]things3.Item, len(todos))
	for i := range todos {
		items[i] = things3.Item{Todo: &todos[i]}
	}
	return items
}

// projectItems wraps projects as list items.
func projectItems(projects []things3.Project) []things3.Item {
	items := make([]things3.Item, len(projects))
	for i := range projects {
		items[i] = things3.Item{Project: &projects[i]}
	}
	return items
}

// writeItems writes a flat list in text mode. Machine formats render the
// self-describing envelope through writeListEnvelope instead.
func writeItems(w io.Writer, items []things3.Item, opts rowOptions) error {
	if len(items) > 0 {
		if _, err := fmt.Fprintln(w, "STATUS   UUID      TITLE"); err != nil {
			return err
		}
	}
	for _, it := range items {
		if _, err := fmt.Fprintln(w, formatItemLine(it, opts)); err != nil {
			return err
		}
	}
	return nil
}

// itemGroup is a labeled run of items for sectioned text output.
type itemGroup struct {
	Header string
	Items  []things3.Item
}

// writeGroupedItems writes items grouped under headers in text mode. Machine
// formats render the flat envelope instead, so grouping never leaks into them.
func writeGroupedItems(w io.Writer, groups []itemGroup, opts rowOptions) error {
	for i := range groups {
		if i > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w, groups[i].Header); err != nil {
			return err
		}
		for _, it := range groups[i].Items {
			if _, err := fmt.Fprintln(w, formatItemLine(it, opts)); err != nil {
				return err
			}
		}
	}
	return nil
}

// formatItemLine formats one row:
// STATUS UUID TITLE [| date] [| @container] [| #tags] [| repeats].
func formatItemLine(it things3.Item, opts rowOptions) string {
	segments := []string{
		fmt.Sprintf("%-8s %-9s %s", statusMark(it), shortUUID(it.UUID()), it.Title()),
		itemRelevantDate(it),
	}
	if opts.showContainer {
		segments = append(segments, itemContainer(it))
	}
	segments = append(segments, formatTags(it.Tags()), repeatMarker(it))
	return strings.Join(slices.DeleteFunc(segments, func(s string) bool { return s == "" }), " | ")
}

// repeatMarker flags a repeating item, and says why a series is not producing
// occurrences when it is paused or has ended.
func repeatMarker(it things3.Item) string {
	repeating, repeat := it.Todo != nil && it.Todo.Repeating, (*things3.Repeat)(nil)
	if it.Project != nil {
		repeating, repeat = it.Project.Repeating, it.Project.Repeat
	} else {
		repeat = it.Todo.Repeat
	}
	switch {
	case !repeating:
		return ""
	case repeat != nil && repeat.Paused:
		return "repeats (paused)"
	case repeat != nil && repeat.Ended:
		return "repeats (ended)"
	default:
		return "repeats"
	}
}

// statusMark draws the row's status the way the app does: a square checkbox for
// a todo, a round one for a project, so a mixed list needs no type column.
func statusMark(it things3.Item) string {
	if it.IsProject() {
		return "(" + statusFill(it.Status()) + ")"
	}
	return statusCheckbox(it.Status())
}

// statusCheckbox returns a todo-style checkbox for the given status.
func statusCheckbox(s things3.Status) string {
	return "[" + statusFill(s) + "]"
}

func statusFill(s things3.Status) string {
	switch s {
	case things3.StatusCompleted:
		return "x"
	case things3.StatusCanceled:
		return "-"
	default:
		return " "
	}
}

// shortUUID returns the first 8 characters of a UUID for display.
func shortUUID(uuid string) string {
	if len(uuid) > 8 {
		return uuid[:8]
	}
	return uuid
}

// itemRelevantDate returns the date a row shows: when the item was closed, else
// its deadline (marked "due:"), else its start date.
func itemRelevantDate(it things3.Item) string {
	switch {
	case it.StoppedAt() != nil:
		return it.StoppedAt().Format(time.DateOnly)
	case it.Deadline() != nil:
		return "due:" + it.Deadline().Format(time.DateOnly)
	case it.StartDate() != nil:
		return it.StartDate().Format(time.DateOnly)
	default:
		return ""
	}
}

// formatTags formats tags as #tag1 #tag2.
func formatTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	return "#" + strings.Join(tags, " #")
}

// writeAreas writes areas as "- Title" lines in text mode.
func writeAreas(w io.Writer, areas []things3.Area) error {
	for i := range areas {
		if _, err := fmt.Fprintf(w, "- %s\n", areas[i].Title); err != nil {
			return err
		}
	}
	return nil
}

// writeTags writes tags as "- Title" lines in text mode.
func writeTags(w io.Writer, tags []things3.Tag) error {
	for i := range tags {
		if _, err := fmt.Fprintf(w, "- %s\n", tags[i].Title); err != nil {
			return err
		}
	}
	return nil
}

// writeTodoDetail writes a single todo with full details in text format.
func writeTodoDetail(w io.Writer, t *things3.Todo) error {
	const dateFormat = "2006-01-02"
	fmt.Fprintf(w, "Title:    %s\n", t.Title)
	fmt.Fprintf(w, "UUID:     %s\n", t.UUID)
	fmt.Fprintf(w, "Status:   %s\n", t.Status)
	fmt.Fprintf(w, "Start:    %s\n", t.Start)
	writeStateLines(w, t.Trashed, t.Repeating, t.Template, t.Repeat)
	if t.ProjectTitle != "" {
		fmt.Fprintf(w, "Project:  %s\n", t.ProjectTitle)
	}
	if t.AreaTitle != "" {
		fmt.Fprintf(w, "Area:     %s\n", t.AreaTitle)
	}
	if t.HeadingTitle != "" {
		fmt.Fprintf(w, "Heading:  %s\n", t.HeadingTitle)
	}
	if len(t.Tags) > 0 {
		fmt.Fprintf(w, "Tags:     %s\n", formatTags(t.Tags))
	}
	if t.StartDate != nil {
		fmt.Fprintf(w, "When:     %s\n", t.StartDate.Format(dateFormat))
	}
	if t.Deadline != nil {
		fmt.Fprintf(w, "Deadline: %s\n", t.Deadline.Format(dateFormat))
	}
	if t.Reminder != nil {
		fmt.Fprintf(w, "Reminder: %s\n", t.Reminder.Format("15:04"))
	}
	if t.CompletedAt != nil {
		fmt.Fprintf(w, "Done:     %s\n", t.CompletedAt.Format(dateFormat))
	}
	if t.CanceledAt != nil {
		fmt.Fprintf(w, "Canceled: %s\n", t.CanceledAt.Format(dateFormat))
	}
	if t.Notes != "" {
		fmt.Fprintf(w, "\nNotes:\n%s\n", t.Notes)
	}
	if len(t.Checklist) > 0 {
		fmt.Fprintln(w, "\nChecklist:")
		for i := range t.Checklist {
			fmt.Fprintf(w, "  %s %s\n", statusCheckbox(t.Checklist[i].Status), t.Checklist[i].Title)
		}
	}
	return nil
}

// writeStateLines writes the detail lines for the states that set an item apart
// from an ordinary live one. For a repeating series, When is its next
// occurrence.
func writeStateLines(w io.Writer, trashed, repeating, template bool, repeat *things3.Repeat) {
	if trashed {
		fmt.Fprintf(w, "Trashed:  yes\n")
	}
	if !repeating {
		return
	}
	rule := "yes"
	if repeat != nil {
		rule = repeat.String()
	}
	fmt.Fprintf(w, "Repeats:  %s\n", rule)
	if template {
		fmt.Fprintf(w, "Series:   the repeating series itself, read-only\n")
	}
}

// writeProjectDetail writes a single project with full details in text format.
func writeProjectDetail(w io.Writer, p *things3.Project) error {
	const dateFormat = "2006-01-02"
	fmt.Fprintf(w, "Title:    %s\n", p.Title)
	fmt.Fprintf(w, "UUID:     %s\n", p.UUID)
	fmt.Fprintf(w, "Status:   %s\n", p.Status)
	fmt.Fprintf(w, "Start:    %s\n", p.Start)
	writeStateLines(w, p.Trashed, p.Repeating, p.Template, p.Repeat)
	if p.AreaTitle != "" {
		fmt.Fprintf(w, "Area:     %s\n", p.AreaTitle)
	}
	if len(p.Tags) > 0 {
		fmt.Fprintf(w, "Tags:     %s\n", formatTags(p.Tags))
	}
	if p.StartDate != nil {
		fmt.Fprintf(w, "When:     %s\n", p.StartDate.Format(dateFormat))
	}
	if p.Deadline != nil {
		fmt.Fprintf(w, "Deadline: %s\n", p.Deadline.Format(dateFormat))
	}
	if p.CompletedAt != nil {
		fmt.Fprintf(w, "Done:     %s\n", p.CompletedAt.Format(dateFormat))
	}
	if p.CanceledAt != nil {
		fmt.Fprintf(w, "Canceled: %s\n", p.CanceledAt.Format(dateFormat))
	}
	if p.Notes != "" {
		fmt.Fprintf(w, "\nNotes:\n%s\n", p.Notes)
	}
	return nil
}

// writeResult is the output shape of every action command.
type writeResult struct {
	Action   string           `json:"action"`
	DryRun   bool             `json:"dry_run,omitempty"`
	Verified bool             `json:"verified"`
	URL      string           `json:"url,omitempty"`
	Type     string           `json:"type,omitempty"`
	Todo     *things3.Todo    `json:"todo,omitempty"`
	Project  *things3.Project `json:"project,omitempty"`
	UUID     string           `json:"uuid,omitempty"`
	Message  string           `json:"message,omitempty"`
}

// displayVerb maps an action to its past-tense display form where they differ.
func displayVerb(action string) string {
	if action == actionOpen {
		return "opened"
	}
	return action
}

// writeWriteResult renders a writeResult. Text: dry-run prints only the URL;
// a confirmed write prints "<verb>: <item line>"; an unverified send prints
// "<verb>: sent to Things (not yet confirmed)".
func writeWriteResult(w io.Writer, r *writeResult, format outputFormat) error {
	switch format {
	case formatJSON:
		return writeJSON(w, r)
	case formatYAML:
		return writeYAML(w, r)
	}
	verb := displayVerb(r.Action)
	switch {
	case r.DryRun:
		_, err := fmt.Fprintln(w, r.URL)
		return err
	case !r.Verified:
		_, err := fmt.Fprintf(w, "%s: sent to Things (not yet confirmed)\n", verb)
		return err
	case r.Todo != nil:
		_, err := fmt.Fprintf(w, "%s: %s\n", verb, formatItemLine(things3.Item{Todo: r.Todo}, defaultRow))
		return err
	case r.Project != nil:
		_, err := fmt.Fprintf(w, "%s: %s\n", verb, formatItemLine(things3.Item{Project: r.Project}, defaultRow))
		return err
	default:
		_, err := fmt.Fprintf(w, "%s: %s\n", verb, r.Message)
		return err
	}
}

// listPage is the self-describing envelope wrapping a paginated list in json and
// yaml. Items always encodes as [] (never null) so consumers can iterate safely,
// while total/page/pages describe the slice within the full result set.
type listPage[T any] struct {
	Items []T `json:"items" yaml:"items"`
	Total int `json:"total" yaml:"total"`
	Page  int `json:"page" yaml:"page"`
	Pages int `json:"pages" yaml:"pages"`
}

// writeListEnvelope renders a page slice and its pagination metadata as a
// listPage in json or yaml, guaranteeing a non-null items array.
func writeListEnvelope[T any](w io.Writer, page []T, meta pageMeta, format outputFormat) error {
	env := listPage[T]{Items: page, Total: meta.total, Page: meta.page, Pages: meta.pages}
	if env.Items == nil {
		env.Items = []T{}
	}
	if format == formatYAML {
		return writeYAML(w, env)
	}
	return writeJSON(w, env)
}

// writeJSON writes any value as indented JSON.
func writeJSON(w io.Writer, v any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

// writeYAML writes any value as YAML.
func writeYAML(w io.Writer, v any) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}
