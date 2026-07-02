package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

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

// repeatsSuffix marks a repeating item in one-line output.
const repeatsSuffix = " | repeats"

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

// todoContainer returns the "@project / heading" or "@area" segment for a todo,
// or an empty string when it belongs to neither. The heading is shown only
// alongside its project.
func todoContainer(t *things3.Todo) string {
	switch {
	case t.ProjectTitle != "":
		seg := "@" + t.ProjectTitle
		if t.HeadingTitle != "" {
			seg += " / " + t.HeadingTitle
		}
		return seg
	case t.AreaTitle != "":
		return "@" + t.AreaTitle
	default:
		return ""
	}
}

// projectContainer returns the "@area" segment for a project, or an empty
// string when it belongs to no area.
func projectContainer(p *things3.Project) string {
	if p.AreaTitle != "" {
		return "@" + p.AreaTitle
	}
	return ""
}

// getOutput reads the -n/--limit flag and resolves the output format for a
// command from the inherited --text/--json/--yaml switches.
func getOutput(cmd *cobra.Command) (limit int, format outputFormat) {
	limit, _ = cmd.Flags().GetInt(flagLimit)
	return limit, resolveFormat(cmd)
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

// applyLimit truncates a slice to at most limit items (0 = unlimited).
func applyLimit[T any](items []T, limit int) []T {
	if limit > 0 && len(items) > limit {
		return items[:limit]
	}
	return items
}

// outputMixed renders a cross-type list honoring the format switches and -n. In
// json/yaml the list is wrapped in the self-describing envelope with total set to
// the pre-limit count.
func outputMixed(cmd *cobra.Command, items []mixedItem) error {
	limit, format := getOutput(cmd)
	w := cmd.OutOrStdout()
	page := applyLimit(items, limit)
	switch format {
	case formatJSON, formatYAML:
		return writeListEnvelope(w, page, pageMeta{total: len(items), page: 1, pages: 1}, format)
	default:
		return writeMixed(w, page)
	}
}

// outputTodoDetail renders one todo in full detail honoring the format switches.
func outputTodoDetail(cmd *cobra.Command, todo *things3.Todo) error {
	_, format := getOutput(cmd)
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
	_, format := getOutput(cmd)
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
			return writeTodos(w, todos, defaultRow)
		}
		return nil
	}
}

// writeTodos writes a flat todo list in text mode. Machine formats render the
// self-describing envelope through writeListEnvelope instead.
func writeTodos(w io.Writer, todos []things3.Todo, opts rowOptions) error {
	if len(todos) > 0 {
		if _, err := fmt.Fprintln(w, "STATUS   UUID      TITLE"); err != nil {
			return err
		}
	}
	for i := range todos {
		if _, err := fmt.Fprintln(w, formatTodoLine(&todos[i], opts)); err != nil {
			return err
		}
	}
	return nil
}

// writeProjects writes a flat project list in text mode.
func writeProjects(w io.Writer, projects []things3.Project) error {
	if len(projects) > 0 {
		if _, err := fmt.Fprintln(w, "STATUS   UUID      TITLE"); err != nil {
			return err
		}
	}
	for i := range projects {
		if _, err := fmt.Fprintln(w, formatProjectLine(&projects[i])); err != nil {
			return err
		}
	}
	return nil
}

// todoGroup is a labeled run of todos for sectioned/grouped text output.
type todoGroup struct {
	Header string
	Todos  []things3.Todo
}

// writeGroupedTodos writes todos grouped under headers in text mode. Machine
// formats render the flat envelope instead, so grouping never leaks into them.
func writeGroupedTodos(w io.Writer, groups []todoGroup, opts rowOptions) error {
	for i := range groups {
		if i > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w, groups[i].Header); err != nil {
			return err
		}
		for j := range groups[i].Todos {
			if _, err := fmt.Fprintln(w, formatTodoLine(&groups[i].Todos[j], opts)); err != nil {
				return err
			}
		}
	}
	return nil
}

// mixedItem wraps a todo or project for cross-type lists (trash, show, search).
// Exactly one of Todo/Project is set.
type mixedItem struct {
	Type    string
	Todo    *things3.Todo
	Project *things3.Project
}

// todoMixed wraps a todo as a mixed item.
func todoMixed(t *things3.Todo) mixedItem { return mixedItem{Type: typeTodo, Todo: t} }

// projectMixed wraps a project as a mixed item.
func projectMixed(p *things3.Project) mixedItem { return mixedItem{Type: typeProject, Project: p} }

// asMap marshals the embedded model and adds the type discriminator so both
// JSON and YAML carry the fields inline (encoding/json has no ",inline").
func (m mixedItem) asMap() (map[string]any, error) {
	var raw []byte
	var err error
	if m.Project != nil {
		raw, err = json.Marshal(m.Project)
	} else {
		raw, err = json.Marshal(m.Todo)
	}
	if err != nil {
		return nil, err
	}
	obj := map[string]any{}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, err
	}
	obj["type"] = m.Type
	return obj, nil
}

// MarshalJSON renders the item with an inline "type" discriminator.
func (m mixedItem) MarshalJSON() ([]byte, error) {
	obj, err := m.asMap()
	if err != nil {
		return nil, err
	}
	return json.Marshal(obj)
}

// MarshalYAML mirrors MarshalJSON so YAML carries the same inline shape.
func (m mixedItem) MarshalYAML() (any, error) {
	return m.asMap()
}

// writeMixed writes a cross-type list with a TYPE column in text mode.
func writeMixed(w io.Writer, items []mixedItem) error {
	if len(items) > 0 {
		if _, err := fmt.Fprintf(w, "%-8s %-9s %-8s %s\n", "STATUS", "UUID", "TYPE", "TITLE"); err != nil {
			return err
		}
	}
	for i := range items {
		if _, err := fmt.Fprintln(w, formatMixedLine(items[i])); err != nil {
			return err
		}
	}
	return nil
}

// formatMixedLine formats one cross-type row:
// STATUS UUID TYPE TITLE [| date] [| @container] [| #tags] [| repeats].
func formatMixedLine(m mixedItem) string {
	var (
		status    things3.Status
		uuid      string
		title     string
		date      string
		container string
		tags      string
		repeating bool
	)
	if m.Project != nil {
		status, uuid, title = m.Project.Status, m.Project.UUID, m.Project.Title
		date, container, repeating = projectRelevantDate(m.Project), projectContainer(m.Project), m.Project.Repeating
	} else {
		status, uuid, title = m.Todo.Status, m.Todo.UUID, m.Todo.Title
		date, container, tags, repeating = todoRelevantDate(m.Todo), todoContainer(m.Todo), formatTags(m.Todo.Tags), m.Todo.Repeating
	}
	line := fmt.Sprintf("%-8s %-9s %-8s %s", statusCheckbox(status), shortUUID(uuid), m.Type, title)
	if date != "" {
		line += " | " + date
	}
	if container != "" {
		line += " | " + container
	}
	if tags != "" {
		line += " | " + tags
	}
	if repeating {
		line += repeatsSuffix
	}
	return line
}

// formatTodoLine formats a single todo as a compact one-line string:
// STATUS UUID TITLE [| date] [| @container] [| #tags] [| repeats].
func formatTodoLine(t *things3.Todo, opts rowOptions) string {
	line := fmt.Sprintf("%-8s %-9s %s", statusCheckbox(t.Status), shortUUID(t.UUID), t.Title)
	if date := todoRelevantDate(t); date != "" {
		line += " | " + date
	}
	if opts.showContainer {
		if container := todoContainer(t); container != "" {
			line += " | " + container
		}
	}
	if tags := formatTags(t.Tags); tags != "" {
		line += " | " + tags
	}
	if t.Repeating {
		line += repeatsSuffix
	}
	return line
}

// formatProjectLine formats a single project as a compact one-line string.
func formatProjectLine(p *things3.Project) string {
	line := fmt.Sprintf("%-8s %-9s %s", statusCheckbox(p.Status), shortUUID(p.UUID), p.Title)
	if date := projectRelevantDate(p); date != "" {
		line += " | " + date
	}
	if p.Repeating {
		line += repeatsSuffix
	}
	return line
}

// statusCheckbox returns a checkbox string for the given status.
func statusCheckbox(s things3.Status) string {
	switch s {
	case things3.StatusCompleted:
		return "[x]"
	case things3.StatusCanceled:
		return "[-]"
	default:
		return "[ ]"
	}
}

// shortUUID returns the first 8 characters of a UUID for display.
func shortUUID(uuid string) string {
	if len(uuid) > 8 {
		return uuid[:8]
	}
	return uuid
}

// todoRelevantDate returns the most relevant date for a todo row.
func todoRelevantDate(t *things3.Todo) string {
	const dateFormat = "2006-01-02"
	switch {
	case t.CompletedAt != nil:
		return t.CompletedAt.Format(dateFormat)
	case t.CanceledAt != nil:
		return t.CanceledAt.Format(dateFormat)
	case t.Deadline != nil:
		return "due:" + t.Deadline.Format(dateFormat)
	case t.StartDate != nil:
		return t.StartDate.Format(dateFormat)
	default:
		return ""
	}
}

// projectRelevantDate returns the most relevant date for a project row.
func projectRelevantDate(p *things3.Project) string {
	const dateFormat = "2006-01-02"
	switch {
	case p.CompletedAt != nil:
		return p.CompletedAt.Format(dateFormat)
	case p.CanceledAt != nil:
		return p.CanceledAt.Format(dateFormat)
	case p.Deadline != nil:
		return "due:" + p.Deadline.Format(dateFormat)
	case p.StartDate != nil:
		return p.StartDate.Format(dateFormat)
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
	if t.Repeating {
		fmt.Fprintf(w, "Repeats:  yes\n")
	}
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

// writeProjectDetail writes a single project with full details in text format.
func writeProjectDetail(w io.Writer, p *things3.Project) error {
	const dateFormat = "2006-01-02"
	fmt.Fprintf(w, "Title:    %s\n", p.Title)
	fmt.Fprintf(w, "UUID:     %s\n", p.UUID)
	fmt.Fprintf(w, "Status:   %s\n", p.Status)
	fmt.Fprintf(w, "Start:    %s\n", p.Start)
	if p.Repeating {
		fmt.Fprintf(w, "Repeats:  yes\n")
	}
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
		_, err := fmt.Fprintf(w, "%s: %s\n", verb, formatTodoLine(r.Todo, defaultRow))
		return err
	case r.Project != nil:
		_, err := fmt.Fprintf(w, "%s: %s\n", verb, formatProjectLine(r.Project))
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
