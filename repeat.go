package things3

import (
	"fmt"
	"strings"
	"time"

	"github.com/moond4rk/things3/internal/database"
)

// RepeatUnit is the calendar unit a repeating series counts in.
type RepeatUnit string

// Units of a repeating series.
const (
	RepeatDaily   RepeatUnit = "day"
	RepeatWeekly  RepeatUnit = "week"
	RepeatMonthly RepeatUnit = "month"
	RepeatYearly  RepeatUnit = "year"
)

// Repeat describes how a repeating series recurs. A todo or project carries it
// when it belongs to a series: the series itself (Template) and every copy
// Things created from it.
type Repeat struct {
	// Unit and Every give the interval: every Every Units.
	Unit  RepeatUnit `json:"unit"`
	Every int        `json:"every"`
	// AfterCompletion reports a series whose next occurrence is counted from the
	// day the previous one was completed, rather than from a fixed schedule.
	AfterCompletion bool `json:"after_completion,omitempty"`
	// Until is the last day of the series, nil when it never ends.
	Until *time.Time `json:"until,omitempty"`
	// Paused reports a series the user paused. It creates no occurrences.
	Paused bool `json:"paused,omitempty"`
	// Ended reports a series that has run out: it is not paused, yet has no next
	// occurrence. It is only known on the series itself, not on its copies.
	Ended bool `json:"ended,omitempty"`
	// Summary is the rule in words, such as "every 2 weeks on Mon, Fri".
	Summary string `json:"summary"`
}

// String returns the summary followed by the state that keeps the series from
// producing occurrences, if any.
func (r *Repeat) String() string {
	switch {
	case r.Paused:
		return r.Summary + " (paused)"
	case r.Ended:
		return r.Summary + " (ended)"
	default:
		return r.Summary
	}
}

// convertRepeat builds the public rule of the series a row belongs to.
func convertRepeat(r *database.TaskRow) *Repeat {
	rule := r.Repeat
	if rule == nil {
		return nil
	}
	return &Repeat{
		Unit:            repeatUnits[rule.Unit],
		Every:           rule.Interval,
		AfterCompletion: rule.AfterCompletion,
		Until:           rule.Until,
		Paused:          rule.Paused,
		Ended:           r.Template && r.StartDate == nil && !rule.Paused,
		Summary:         summarizeRepeat(rule),
	}
}

var repeatUnits = map[database.RepeatUnit]RepeatUnit{
	database.RepeatDaily:   RepeatDaily,
	database.RepeatWeekly:  RepeatWeekly,
	database.RepeatMonthly: RepeatMonthly,
	database.RepeatYearly:  RepeatYearly,
}

// summarizeRepeat renders a rule in words. An after-completion rule has no
// anchor within its period, so it stops at the interval.
func summarizeRepeat(rule *database.RepeatRule) string {
	unit := string(repeatUnits[rule.Unit])
	if unit == "" {
		return "repeats"
	}
	summary := "every " + unit
	if rule.Interval > 1 {
		summary = fmt.Sprintf("every %d %ss", rule.Interval, unit)
	}
	if rule.AfterCompletion {
		summary += " after completion"
	} else if on := repeatOffsets(rule); on != "" {
		summary += " on " + on
	}
	if rule.Until != nil {
		summary += ", until " + rule.Until.Format(time.DateOnly)
	}
	return summary
}

func repeatOffsets(rule *database.RepeatRule) string {
	parts := make([]string, 0, len(rule.Offsets))
	for _, o := range rule.Offsets {
		if part := repeatOffset(rule.Unit, o); part != "" {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, ", ")
}

func repeatOffset(unit database.RepeatUnit, o database.RepeatOffset) string {
	var day string
	switch {
	case o.Weekday != nil && o.Ordinal != nil:
		day = "the " + ordinal(*o.Ordinal) + " " + shortWeekday(*o.Weekday)
	case o.Weekday != nil:
		day = shortWeekday(*o.Weekday)
	case o.Day != nil && *o.Day < 0:
		day = "the last day"
	case o.Day != nil:
		// Stored zero-based.
		day = "the " + ordinal(*o.Day+1)
	}

	if unit != database.RepeatYearly || o.Month == nil {
		return day
	}
	month := time.Month(*o.Month + 1).String()[:3]
	switch {
	case o.Weekday != nil || (o.Day != nil && *o.Day < 0):
		return day + " of " + month
	case o.Day != nil:
		return fmt.Sprintf("%s %d", month, *o.Day+1)
	default:
		return month
	}
}

func shortWeekday(d time.Weekday) string {
	return d.String()[:3]
}

// ordinal renders 1 as "1st", 22 as "22nd", and -1 as "last".
func ordinal(n int) string {
	if n < 0 {
		return "last"
	}
	suffix := "th"
	if n%100 < 11 || n%100 > 13 {
		switch n % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		case 3:
			suffix = "rd"
		}
	}
	return fmt.Sprintf("%d%s", n, suffix)
}
