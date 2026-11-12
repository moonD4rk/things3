package database

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// RepeatUnit is the calendar unit a repeating rule counts in, as stored in the
// rule's "fu" field.
type RepeatUnit int

// Frequency units of a repeating rule.
const (
	RepeatYearly  RepeatUnit = 4
	RepeatMonthly RepeatUnit = 8
	RepeatDaily   RepeatUnit = 16
	RepeatWeekly  RepeatUnit = 256
)

// RepeatOffset pins an occurrence within its period. Fields are nil when the
// rule does not use them. Day and Month are zero-based as stored, and Day -1
// means the last day of the month. Ordinal qualifies Weekday ("the 2nd
// Friday"), with -1 for the last one.
type RepeatOffset struct {
	Day     *int
	Weekday *time.Weekday
	Month   *int
	Ordinal *int
}

// RepeatRule is a decoded TMTask.rt1_recurrenceRule: a property list whose keys
// Things abbreviates to two letters.
type RepeatRule struct {
	Unit     RepeatUnit
	Interval int
	// AfterCompletion counts the interval from the day the previous occurrence
	// was completed instead of following a fixed schedule.
	AfterCompletion bool
	Offsets         []RepeatOffset
	// Until is the last day of the series, nil when it never ends.
	Until *time.Time
	// Paused reports a series the user paused; it creates no occurrences.
	Paused bool
}

// neverEnds is the "ed" value of an open-ended rule: 4001-01-01, the same
// sentinel Things uses in packed dates.
var neverEnds = time.Date(4001, time.January, 1, 0, 0, 0, 0, time.UTC)

// parseRepeatRule decodes a recurrence rule blob. A blob that is not the
// expected property list yields an error rather than a guessed rule.
func parseRepeatRule(blob []byte) (*RepeatRule, error) {
	root, err := parsePlist(blob)
	if err != nil {
		return nil, fmt.Errorf("recurrence rule: %w", err)
	}
	dict, ok := root.(map[string]any)
	if !ok {
		return nil, errors.New("recurrence rule: not a dictionary")
	}

	rule := &RepeatRule{
		Unit:            RepeatUnit(plistInt(dict["fu"])),
		Interval:        max(plistInt(dict["fa"]), 1),
		AfterCompletion: plistInt(dict["tp"]) == 1,
	}
	if ends, ok := dict["ed"].(float64); ok {
		// The end date is midnight UTC of a calendar day.
		if day := time.Unix(int64(ends), 0).UTC(); day.Before(neverEnds) {
			local := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.Local)
			rule.Until = &local
		}
	}
	offsets, _ := dict["of"].([]any)
	for _, raw := range offsets {
		entry, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		var offset RepeatOffset
		if v, ok := entry["dy"]; ok {
			offset.Day = new(plistInt(v))
		}
		if v, ok := entry["mo"]; ok {
			offset.Month = new(plistInt(v))
		}
		if v, ok := entry["wd"]; ok {
			offset.Weekday = new(time.Weekday(plistInt(v)))
		}
		if v, ok := entry["wdo"]; ok {
			offset.Ordinal = new(plistInt(v))
		}
		rule.Offsets = append(rule.Offsets, offset)
	}
	return rule, nil
}

func plistInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	default:
		return 0
	}
}

// parsePlist decodes the subset of XML property lists Things writes: dict,
// array, integer, real, string, true, and false.
func parsePlist(data []byte) (any, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		if start, ok := tok.(xml.StartElement); ok && start.Name.Local != "plist" {
			return plistValue(dec, start)
		}
	}
}

func plistValue(dec *xml.Decoder, start xml.StartElement) (any, error) {
	switch start.Name.Local {
	case "dict":
		return plistDict(dec)
	case "array":
		return plistArray(dec)
	case "true", "false":
		return start.Name.Local == "true", dec.Skip()
	}

	var text string
	if err := dec.DecodeElement(&text, &start); err != nil {
		return nil, err
	}
	switch start.Name.Local {
	case "integer":
		return strconv.Atoi(text)
	case "real":
		return strconv.ParseFloat(text, 64)
	default:
		return text, nil
	}
}

func plistDict(dec *xml.Decoder) (map[string]any, error) {
	dict := map[string]any{}
	key := ""
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch el := tok.(type) {
		case xml.EndElement:
			return dict, nil
		case xml.StartElement:
			if el.Name.Local == "key" {
				if err := dec.DecodeElement(&key, &el); err != nil {
					return nil, err
				}
				continue
			}
			value, err := plistValue(dec, el)
			if err != nil {
				return nil, err
			}
			dict[key] = value
		}
	}
}

func plistArray(dec *xml.Decoder) ([]any, error) {
	var items []any
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch el := tok.(type) {
		case xml.EndElement:
			return items, nil
		case xml.StartElement:
			value, err := plistValue(dec, el)
			if err != nil {
				return nil, err
			}
			items = append(items, value)
		}
	}
}
