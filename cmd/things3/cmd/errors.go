package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/moond4rk/things3/cmd/things3/internal/resolve"
)

// NotFoundError reports that a query matched no item. It exits 1.
type NotFoundError struct{ Query string }

func (e *NotFoundError) Error() string { return fmt.Sprintf("no item matches %q", e.Query) }

// ExitCode implements the exit-code contract.
func (e *NotFoundError) ExitCode() int { return 1 }

// Candidate is one row of an ambiguity report.
type Candidate struct {
	UUID  string `json:"uuid" yaml:"uuid"`
	Type  string `json:"type" yaml:"type"`
	Title string `json:"title" yaml:"title"`
}

// AmbiguousError reports that a query matched more than one item. It exits 2
// and carries the candidates so the caller can disambiguate.
type AmbiguousError struct {
	Query      string
	Candidates []Candidate
}

func (e *AmbiguousError) Error() string {
	return fmt.Sprintf("query %q matches %d items", e.Query, len(e.Candidates))
}

// ExitCode implements the exit-code contract.
func (e *AmbiguousError) ExitCode() int { return 2 }

// ExitCode returns the process exit code for an error: 0 for nil, the error's
// own ExitCode() when it implements one, else 1.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var coded interface{ ExitCode() int }
	if errors.As(err, &coded) {
		return coded.ExitCode()
	}
	return 1
}

// fromResolveError maps resolve package errors into the CLI's typed errors so
// exit codes flow through ExitCode. Other errors pass through unchanged.
func fromResolveError(err error) error {
	var notFound *resolve.NotFoundError
	if errors.As(err, &notFound) {
		return &NotFoundError{Query: notFound.Query}
	}
	var ambiguous *resolve.AmbiguousError
	if errors.As(err, &ambiguous) {
		candidates := make([]Candidate, len(ambiguous.Matches))
		for i, m := range ambiguous.Matches {
			candidates[i] = Candidate{UUID: shortUUID(m.UUID()), Type: string(m.Kind), Title: m.Title()}
		}
		return &AmbiguousError{Query: ambiguous.Query, Candidates: candidates}
	}
	return err
}

// errorPayload is the machine-readable error shape written to stderr under
// --json/--yaml.
type errorPayload struct {
	Error      string      `json:"error" yaml:"error"`
	Candidates []Candidate `json:"candidates,omitempty" yaml:"candidates,omitempty"`
}

// maxErrorCandidates caps how many ambiguity rows are printed in text mode.
const maxErrorCandidates = 10

// RenderError writes err to stderr in the format selected by the --text/--json/
// --yaml switches. main owns rendering; commands only return errors. When flag
// parsing itself failed the format falls back to text.
func RenderError(root *cobra.Command, err error) {
	w := root.ErrOrStderr()
	_, format := getOutput(root)

	var ambiguous *AmbiguousError
	isAmbiguous := errors.As(err, &ambiguous)

	switch format {
	case formatJSON:
		payload := errorPayload{Error: err.Error()}
		if isAmbiguous {
			payload.Candidates = ambiguous.Candidates
		}
		data, _ := json.Marshal(payload)
		fmt.Fprintln(w, string(data))
	case formatYAML:
		payload := errorPayload{Error: err.Error()}
		if isAmbiguous {
			payload.Candidates = ambiguous.Candidates
		}
		_ = writeYAML(w, payload)
	default:
		fmt.Fprintf(w, "Error: %s\n", err.Error())
		if isAmbiguous {
			for i, c := range ambiguous.Candidates {
				if i >= maxErrorCandidates {
					fmt.Fprintf(w, "  ... and %d more\n", len(ambiguous.Candidates)-maxErrorCandidates)
					break
				}
				fmt.Fprintf(w, "  %-8s  %-8s  %s\n", c.UUID, c.Type, c.Title)
			}
			fmt.Fprintln(w, "hint: rerun with a UUID from the list as the query (a prefix of 4 or more characters also works)")
		}
	}
}
