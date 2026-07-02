package things3

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTodayHelper(t *testing.T) {
	today := Today()
	now := time.Now()

	assert.Equal(t, now.Year(), today.Year())
	assert.Equal(t, now.Month(), today.Month())
	assert.Equal(t, now.Day(), today.Day())
	assert.Equal(t, 0, today.Hour())
	assert.Equal(t, 0, today.Minute())
	assert.Equal(t, 0, today.Second())
}

func TestTomorrowHelper(t *testing.T) {
	tomorrow := Tomorrow()
	today := Today()

	assert.Equal(t, today.AddDate(0, 0, 1), tomorrow)
}

func TestDaysAgo(t *testing.T) {
	sevenDaysAgo := DaysAgo(7)
	expected := time.Now().AddDate(0, 0, -7)

	// Compare within 1 second tolerance
	assert.WithinDuration(t, expected, sevenDaysAgo, time.Second)
}

func TestWeeksAgo(t *testing.T) {
	twoWeeksAgo := WeeksAgo(2)
	expected := time.Now().AddDate(0, 0, -14)

	assert.WithinDuration(t, expected, twoWeeksAgo, time.Second)
}

func TestMonthsAgo(t *testing.T) {
	oneMonthAgo := MonthsAgo(1)
	expected := time.Now().AddDate(0, -1, 0)

	assert.WithinDuration(t, expected, oneMonthAgo, time.Second)
}

func TestYearsAgo(t *testing.T) {
	oneYearAgo := YearsAgo(1)
	expected := time.Now().AddDate(-1, 0, 0)

	assert.WithinDuration(t, expected, oneYearAgo, time.Second)
}

// testWhenDate is a fixed calendar date used by the when-parsing tests.
const testWhenDate = "2024-12-25"

func TestParseWhen(t *testing.T) {
	scheme := newScheme()

	// Get expected dates for today/tomorrow
	todayStr := Today().Format(time.DateOnly)
	tomorrowStr := Tomorrow().Format(time.DateOnly)

	tests := []struct {
		name     string
		when     string
		wantWhen string
		wantErr  bool
	}{
		{"today keyword", whenKeywordToday, todayStr, false},
		{"tomorrow keyword", whenKeywordTomorrow, tomorrowStr, false},
		{"evening keyword", whenKeywordEvening, whenKeywordEvening, false},
		{"anytime keyword", whenKeywordAnytime, whenKeywordAnytime, false},
		{"someday keyword", whenKeywordSomeday, whenKeywordSomeday, false},
		{"specific date", testWhenDate, testWhenDate, false},
		{"unrecognized word", "invalid", "", true},
		{"malformed date", "2024-13-45", "", true},
		{"empty string", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todo, err := ParseWhen(scheme.AddTodo().Title("Test"), tt.when)

			if tt.wantErr {
				require.ErrorContains(t, err, "unrecognized when value")
				if tt.when != "" {
					require.ErrorContains(t, err, tt.when)
				}
			} else {
				require.NoError(t, err)
			}

			thingsURL, err := todo.Build()
			require.NoError(t, err)

			_, params := parseThingsURL(t, thingsURL)
			whenValue := params.Get("when")

			if tt.wantErr {
				assert.Empty(t, whenValue, "builder should be returned unchanged for invalid input")
			} else {
				assert.Equal(t, tt.wantWhen, whenValue)
			}
		})
	}
}

func TestApplyWhen(t *testing.T) {
	scheme := newScheme()

	// Get expected dates for today/tomorrow
	todayStr := Today().Format(time.DateOnly)
	tomorrowStr := Tomorrow().Format(time.DateOnly)

	tests := []struct {
		name      string
		when      string
		wantWhen  string
		wantEmpty bool
	}{
		{"today keyword", whenKeywordToday, todayStr, false},
		{"tomorrow keyword", whenKeywordTomorrow, tomorrowStr, false},
		{"evening keyword", whenKeywordEvening, whenKeywordEvening, false},
		{"anytime keyword", whenKeywordAnytime, whenKeywordAnytime, false},
		{"someday keyword", whenKeywordSomeday, whenKeywordSomeday, false},
		{"specific date", testWhenDate, testWhenDate, false},
		{"invalid format silently ignored", "invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todo := scheme.AddTodo().Title("Test")
			todo = ApplyWhen(todo, tt.when)
			thingsURL, err := todo.Build()
			require.NoError(t, err)

			_, params := parseThingsURL(t, thingsURL)
			whenValue := params.Get("when")

			if tt.wantEmpty {
				assert.Empty(t, whenValue, "when parameter should be empty for invalid input")
			} else {
				assert.Equal(t, tt.wantWhen, whenValue)
			}
		})
	}
}

func TestApplyWhenWithDifferentBuilders(t *testing.T) {
	scheme := newScheme()
	todayStr := Today().Format(time.DateOnly)
	tomorrowStr := Tomorrow().Format(time.DateOnly)

	// Test AddTodoBuilder
	todo := scheme.AddTodo().Title("Todo")
	todo = ApplyWhen(todo, whenKeywordToday)
	todoURL, err := todo.Build()
	require.NoError(t, err)
	_, todoParams := parseThingsURL(t, todoURL)
	assert.Equal(t, todayStr, todoParams.Get("when"))

	// Test AddProjectBuilder
	project := scheme.AddProject().Title("Project")
	project = ApplyWhen(project, whenKeywordTomorrow)
	projectURL, err := project.Build()
	require.NoError(t, err)
	_, projectParams := parseThingsURL(t, projectURL)
	assert.Equal(t, tomorrowStr, projectParams.Get("when"))

	// Test UpdateTodoBuilder (requires auth token)
	authScheme := scheme.WithToken("test-token")
	updateTodo := authScheme.UpdateTodo("test-uuid")
	updateTodo = ApplyWhen(updateTodo, whenKeywordEvening)
	updateURL, err := updateTodo.Build()
	require.NoError(t, err)
	_, updateParams := parseThingsURL(t, updateURL)
	assert.Equal(t, whenKeywordEvening, updateParams.Get("when"))

	// Test UpdateProjectBuilder
	updateProject := authScheme.UpdateProject("test-uuid")
	updateProject = ApplyWhen(updateProject, whenKeywordSomeday)
	updateProjectURL, err := updateProject.Build()
	require.NoError(t, err)
	_, updateProjectParams := parseThingsURL(t, updateProjectURL)
	assert.Equal(t, whenKeywordSomeday, updateProjectParams.Get("when"))
}
