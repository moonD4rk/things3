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

// applyLimit applies the limit to a slice of tasks.
func applyLimit[T any](items []T, limit int) []T {
	if limit > 0 && len(items) > limit {
		return items[:limit]
	}
	return items
}

// outputTasks formats and outputs tasks based on the command flags.
func outputTasks(cmd *cobra.Command, tasks []things3.Task) error {
	limit, format := getOutputFlags(cmd)
	tasks = applyLimit(tasks, limit)
	return writeTasks(cmd.OutOrStdout(), tasks, format)
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

// writeTasks writes tasks to the writer in the specified format.
func writeTasks(w io.Writer, tasks []things3.Task, format outputFormat) error {
	switch format {
	case formatJSON:
		return writeJSON(w, tasks)
	case formatYAML:
		return writeYAML(w, tasks)
	default:
		// Print header
		if len(tasks) > 0 {
			if _, err := fmt.Fprintln(w, "STATUS   UUID      TYPE     TITLE"); err != nil {
				return err
			}
		}
		for i := range tasks {
			if _, err := fmt.Fprintln(w, formatTaskLine(&tasks[i])); err != nil {
				return err
			}
		}
		return nil
	}
}

// formatTaskLine formats a single task as a compact one-line string.
// Format: [x] UUID type Title | date | #tag1 #tag2
func formatTaskLine(t *things3.Task) string {
	// Status checkbox
	checkbox := "[ ]"
	switch t.Status {
	case things3.StatusCompleted:
		checkbox = "[x]"
	case things3.StatusCanceled:
		checkbox = "[-]"
	}

	// Task type
	taskType := formatTaskType(t.Type)

	// Build the line: [x] UUID type Title
	line := fmt.Sprintf("%-8s %-9s %-8s %s", checkbox, shortUUID(t.UUID), taskType, t.Title)

	// Add date (prefer stop_date > deadline > start_date)
	if date := getRelevantDate(t); date != "" {
		line += " | " + date
	}

	// Add tags
	if tags := formatTags(t.Tags); tags != "" {
		line += " | " + tags
	}

	return line
}

// shortUUID returns the first 8 characters of a UUID for display.
func shortUUID(uuid string) string {
	if len(uuid) > 8 {
		return uuid[:8]
	}
	return uuid
}

// formatTaskType returns a short string representation of the task type.
func formatTaskType(t things3.TaskType) string {
	switch t {
	case things3.TaskTypeTodo:
		return "todo"
	case things3.TaskTypeProject:
		return "project"
	case things3.TaskTypeHeading:
		return "heading"
	default:
		return "unknown"
	}
}

// getRelevantDate returns the most relevant date for display.
func getRelevantDate(t *things3.Task) string {
	const dateFormat = "2006-01-02"
	if t.StopDate != nil {
		return t.StopDate.Format(dateFormat)
	}
	if t.Deadline != nil {
		return "due:" + t.Deadline.Format(dateFormat)
	}
	if t.StartDate != nil {
		return t.StartDate.Format(dateFormat)
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
