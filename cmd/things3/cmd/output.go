package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"

	"github.com/moond4rk/things3"
)

// outputFormat represents the output format type.
type outputFormat int

const (
	formatText outputFormat = iota
	formatJSON
	formatYAML
)

// getOutputFlags extracts the common output flags from the command.
func getOutputFlags(cmd *cobra.Command) (limit int, format outputFormat) {
	limit, _ = cmd.Flags().GetInt(flagLimit)
	if asYAML, _ := cmd.Flags().GetBool(flagYAML); asYAML {
		return limit, formatYAML
	}
	if asJSON, _ := cmd.Flags().GetBool(flagJSON); asJSON {
		return limit, formatJSON
	}
	return limit, formatText
}

// applyLimit applies the limit to a slice.
func applyLimit[T any](items []T, limit int) []T {
	if limit > 0 && len(items) > limit {
		return items[:limit]
	}
	return items
}

// outputTodos formats and outputs todos based on the command flags.
func outputTodos(cmd *cobra.Command, todos []things3.Todo) error {
	limit, format := getOutputFlags(cmd)
	todos = applyLimit(todos, limit)
	return writeTodos(cmd.OutOrStdout(), todos, format)
}

// outputProjects formats and outputs projects based on the command flags.
func outputProjects(cmd *cobra.Command, projects []things3.Project) error {
	limit, format := getOutputFlags(cmd)
	projects = applyLimit(projects, limit)
	return writeProjects(cmd.OutOrStdout(), projects, format)
}

// outputAreas formats and outputs areas based on the command flags.
func outputAreas(cmd *cobra.Command, areas []things3.Area) error {
	limit, format := getOutputFlags(cmd)
	areas = applyLimit(areas, limit)
	return writeAreas(cmd.OutOrStdout(), areas, format)
}

// outputTags formats and outputs tags based on the command flags.
func outputTags(cmd *cobra.Command, tags []things3.Tag) error {
	limit, format := getOutputFlags(cmd)
	tags = applyLimit(tags, limit)
	return writeTags(cmd.OutOrStdout(), tags, format)
}

// writeTodos writes todos to the writer in the specified format.
func writeTodos(w io.Writer, todos []things3.Todo, format outputFormat) error {
	switch format {
	case formatJSON:
		return writeJSON(w, todos)
	case formatYAML:
		return writeYAML(w, todos)
	default:
		if len(todos) > 0 {
			if _, err := fmt.Fprintln(w, "STATUS   UUID      TITLE"); err != nil {
				return err
			}
		}
		for i := range todos {
			if _, err := fmt.Fprintln(w, formatTodoLine(&todos[i])); err != nil {
				return err
			}
		}
		return nil
	}
}

// writeProjects writes projects to the writer in the specified format.
func writeProjects(w io.Writer, projects []things3.Project, format outputFormat) error {
	switch format {
	case formatJSON:
		return writeJSON(w, projects)
	case formatYAML:
		return writeYAML(w, projects)
	default:
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
}

// formatTodoLine formats a single todo as a compact one-line string.
func formatTodoLine(t *things3.Todo) string {
	checkbox := statusCheckbox(t.Status)
	line := fmt.Sprintf("%-8s %-9s %s", checkbox, shortUUID(t.UUID), t.Title)

	if date := todoRelevantDate(t); date != "" {
		line += " | " + date
	}
	if tags := formatTags(t.Tags); tags != "" {
		line += " | " + tags
	}
	return line
}

// formatProjectLine formats a single project as a compact one-line string.
func formatProjectLine(p *things3.Project) string {
	checkbox := statusCheckbox(p.Status)
	line := fmt.Sprintf("%-8s %-9s %s", checkbox, shortUUID(p.UUID), p.Title)

	if date := projectRelevantDate(p); date != "" {
		line += " | " + date
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

// todoRelevantDate returns the most relevant date for display.
func todoRelevantDate(t *things3.Todo) string {
	const dateFormat = "2006-01-02"
	if t.CompletedAt != nil {
		return t.CompletedAt.Format(dateFormat)
	}
	if t.CanceledAt != nil {
		return t.CanceledAt.Format(dateFormat)
	}
	if t.Deadline != nil {
		return "due:" + t.Deadline.Format(dateFormat)
	}
	if t.StartDate != nil {
		return t.StartDate.Format(dateFormat)
	}
	return ""
}

// projectRelevantDate returns the most relevant date for display.
func projectRelevantDate(p *things3.Project) string {
	const dateFormat = "2006-01-02"
	if p.CompletedAt != nil {
		return p.CompletedAt.Format(dateFormat)
	}
	if p.CanceledAt != nil {
		return p.CanceledAt.Format(dateFormat)
	}
	if p.Deadline != nil {
		return "due:" + p.Deadline.Format(dateFormat)
	}
	if p.StartDate != nil {
		return p.StartDate.Format(dateFormat)
	}
	return ""
}

// formatTags formats tags as #tag1 #tag2.
func formatTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	result := ""
	for i, tag := range tags {
		if i > 0 {
			result += " "
		}
		result += "#" + tag
	}
	return result
}

// writeAreas writes areas to the writer in the specified format.
func writeAreas(w io.Writer, areas []things3.Area, format outputFormat) error {
	switch format {
	case formatJSON:
		return writeJSON(w, areas)
	case formatYAML:
		return writeYAML(w, areas)
	default:
		for i := range areas {
			if _, err := fmt.Fprintf(w, "- %s\n", areas[i].Title); err != nil {
				return err
			}
		}
		return nil
	}
}

// writeTags writes tags to the writer in the specified format.
func writeTags(w io.Writer, tags []things3.Tag, format outputFormat) error {
	switch format {
	case formatJSON:
		return writeJSON(w, tags)
	case formatYAML:
		return writeYAML(w, tags)
	default:
		for i := range tags {
			if _, err := fmt.Fprintf(w, "- %s\n", tags[i].Title); err != nil {
				return err
			}
		}
		return nil
	}
}

// outputTodoDetail formats and outputs a single todo with full details.
func outputTodoDetail(cmd *cobra.Command, todo *things3.Todo) error {
	_, format := getOutputFlags(cmd)
	if format == formatJSON {
		return writeJSON(cmd.OutOrStdout(), todo)
	}
	if format == formatYAML {
		return writeYAML(cmd.OutOrStdout(), todo)
	}
	return writeTodoDetail(cmd.OutOrStdout(), todo)
}

// projectDetailOutput is the structured representation for JSON/YAML output.
type projectDetailOutput struct {
	Project *things3.Project `json:"project" yaml:"project"`
	Todos   []things3.Todo   `json:"todos,omitempty" yaml:"todos,omitempty"`
}

// outputProjectDetail formats and outputs a single project with full details and its todos.
func outputProjectDetail(cmd *cobra.Command, project *things3.Project, todos []things3.Todo) error {
	_, format := getOutputFlags(cmd)
	w := cmd.OutOrStdout()
	if format == formatJSON {
		return writeJSON(w, projectDetailOutput{Project: project, Todos: todos})
	}
	if format == formatYAML {
		return writeYAML(w, projectDetailOutput{Project: project, Todos: todos})
	}
	if err := writeProjectDetail(w, project); err != nil {
		return err
	}
	if len(todos) > 0 {
		fmt.Fprintln(w)
		return writeTodos(w, todos, formatText)
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

// writeJSON writes any value as JSON to the writer.
func writeJSON(w io.Writer, v any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

// writeYAML writes any value as YAML to the writer.
func writeYAML(w io.Writer, v any) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}
