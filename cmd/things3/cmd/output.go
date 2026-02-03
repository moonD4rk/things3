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
		for i := range tasks {
			checkbox := "[ ]"
			switch tasks[i].Status {
			case things3.StatusCompleted:
				checkbox = "[x]"
			case things3.StatusCanceled:
				checkbox = "[-]"
			}
			if _, err := fmt.Fprintf(w, "- %s %s\n", checkbox, tasks[i].Title); err != nil {
				return err
			}
		}
		return nil
	}
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
