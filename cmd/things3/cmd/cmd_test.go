package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/moond4rk/things3/thingstest"
)

// setupFixtureDB points the CLI at a writable copy of the shared test fixture.
func setupFixtureDB(t *testing.T) {
	t.Helper()
	t.Setenv("THINGSDB", thingstest.DatabasePath(t))
}

// executeCommand runs the root command with the given args and returns the
// captured stdout, stderr, and execution error.
func executeCommand(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	root := NewRootCmd()
	var outBuf, errBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs(args)
	err = root.Execute()
	return outBuf.String(), errBuf.String(), err
}

func TestListInbox(t *testing.T) {
	setupFixtureDB(t)

	stdout, stderr, err := executeCommand(t, "list", "inbox")
	if err != nil {
		t.Fatalf("list inbox failed: %v (stderr: %s)", err, stderr)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if !strings.HasPrefix(lines[0], "STATUS") {
		t.Errorf("expected header line, got %q", lines[0])
	}
	if got, want := len(lines)-1, thingstest.Inbox; got != want {
		t.Errorf("expected %d inbox rows, got %d\noutput:\n%s", want, got, stdout)
	}
}

func TestListInboxJSONArray(t *testing.T) {
	setupFixtureDB(t)

	stdout, stderr, err := executeCommand(t, "list", "inbox", "--json")
	if err != nil {
		t.Fatalf("list inbox --json failed: %v (stderr: %s)", err, stderr)
	}

	trimmed := strings.TrimSpace(stdout)
	if !strings.HasPrefix(trimmed, "[") {
		t.Errorf("expected JSON array output starting with [, got %q", trimmed)
	}
	var todos []map[string]any
	if jsonErr := json.Unmarshal([]byte(trimmed), &todos); jsonErr != nil {
		t.Fatalf("output is not a valid JSON array: %v", jsonErr)
	}
	if len(todos) != thingstest.Inbox {
		t.Errorf("expected %d todos in JSON array, got %d", thingstest.Inbox, len(todos))
	}
}

func TestProjectShortUUIDRoundTrip(t *testing.T) {
	setupFixtureDB(t)

	listOut, stderr, err := executeCommand(t, "list", "projects")
	if err != nil {
		t.Fatalf("list projects failed: %v (stderr: %s)", err, stderr)
	}

	lines := strings.Split(strings.TrimSpace(listOut), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected header plus at least one project row, got:\n%s", listOut)
	}

	// Row format is "%-8s %-9s %s" (status, short UUID, title): the displayed
	// short UUID column starts at index 9.
	firstRow := lines[1]
	if len(firstRow) < 10 {
		t.Fatalf("project row too short to contain a UUID column: %q", firstRow)
	}
	shortID := strings.Fields(firstRow[9:])[0]
	if len(shortID) != 8 {
		t.Fatalf("expected 8-char short UUID in list output, got %q from row %q", shortID, firstRow)
	}

	detailOut, stderr, err := executeCommand(t, "project", shortID)
	if err != nil {
		t.Fatalf("project %s failed: %v (stderr: %s)", shortID, err, stderr)
	}
	if !strings.Contains(detailOut, "Title:") {
		t.Errorf("expected detail view with Title field, got:\n%s", detailOut)
	}

	var fullUUID string
	for line := range strings.Lines(detailOut) {
		if rest, ok := strings.CutPrefix(line, "UUID:"); ok {
			fullUUID = strings.TrimSpace(rest)
			break
		}
	}
	if fullUUID == "" {
		t.Fatalf("expected detail view with UUID field, got:\n%s", detailOut)
	}
	if !strings.HasPrefix(fullUUID, shortID) {
		t.Errorf("detail UUID %q does not start with displayed short UUID %q", fullUUID, shortID)
	}
	if len(fullUUID) <= len(shortID) {
		t.Errorf("expected full UUID longer than short form, got %q", fullUUID)
	}
}

func TestCommandErrors(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		wantErrContains []string
	}{
		{
			name:            "todo not found",
			args:            []string{"todo", "zzznosuchtodo"},
			wantErrContains: []string{"zzznosuchtodo"},
		},
		{
			name:            "project not found",
			args:            []string{"project", "zzznosuchproject"},
			wantErrContains: []string{"zzznosuchproject"},
		},
		{
			name:            "json and yaml are mutually exclusive",
			args:            []string{"list", "inbox", "--json", "--yaml"},
			wantErrContains: []string{"json", "yaml"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupFixtureDB(t)

			stdout, stderr, err := executeCommand(t, tt.args...)
			if err == nil {
				t.Fatalf("expected error for args %v, got nil", tt.args)
			}
			for _, want := range tt.wantErrContains {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q does not contain %q", err.Error(), want)
				}
				if !strings.Contains(stderr, want) {
					t.Errorf("stderr %q does not contain %q", stderr, want)
				}
			}
			if strings.Contains(stdout, "Usage:") || strings.Contains(stderr, "Usage:") {
				t.Errorf("runtime error must not dump usage help\nstdout:\n%s\nstderr:\n%s", stdout, stderr)
			}
		})
	}
}

// logbookRow captures the fields needed to verify logbook ordering.
type logbookRow struct {
	UUID        string     `json:"uuid"`
	CompletedAt *time.Time `json:"completed_at"`
	CanceledAt  *time.Time `json:"canceled_at"`
}

func (r *logbookRow) stopTime() *time.Time {
	if r.CompletedAt != nil {
		return r.CompletedAt
	}
	return r.CanceledAt
}

func runLogbookJSON(t *testing.T, args ...string) []logbookRow {
	t.Helper()
	stdout, stderr, err := executeCommand(t, append([]string{"list", "logbook", "--days", "0", "--json"}, args...)...)
	if err != nil {
		t.Fatalf("list logbook failed: %v (stderr: %s)", err, stderr)
	}
	var rows []logbookRow
	if jsonErr := json.Unmarshal([]byte(stdout), &rows); jsonErr != nil {
		t.Fatalf("cannot decode logbook JSON: %v", jsonErr)
	}
	return rows
}

func TestLogbookOrderedByStopTimeDesc(t *testing.T) {
	setupFixtureDB(t)

	rows := runLogbookJSON(t)
	if len(rows) < 2 {
		t.Fatalf("expected at least 2 logbook rows, got %d", len(rows))
	}

	for i := 1; i < len(rows); i++ {
		prev, curr := rows[i-1].stopTime(), rows[i].stopTime()
		if prev == nil {
			if curr != nil {
				t.Fatalf("row %d (%s) has a stop time but follows row %d (%s) without one",
					i, rows[i].UUID, i-1, rows[i-1].UUID)
			}
			continue
		}
		if curr == nil {
			continue
		}
		if curr.After(*prev) {
			t.Fatalf("logbook not sorted by stop time descending: row %d (%s, %s) is more recent than row %d (%s, %s)",
				i, rows[i].UUID, curr, i-1, rows[i-1].UUID, prev)
		}
	}

	limited := runLogbookJSON(t, "--limit", "2")
	if len(limited) != 2 {
		t.Fatalf("expected 2 rows with --limit 2, got %d", len(limited))
	}
	for i := range limited {
		if limited[i].UUID != rows[i].UUID {
			t.Errorf("--limit 2 row %d = %s, want most recent entry %s", i, limited[i].UUID, rows[i].UUID)
		}
	}
}
