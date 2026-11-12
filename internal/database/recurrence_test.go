package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const plistHeader = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">`

func rulePlist(body string) []byte {
	return []byte(plistHeader + "<dict>" + body + "</dict></plist>")
}

func TestParseRepeatRule(t *testing.T) {
	// Every 2 weeks on Monday and Friday, open-ended, on a fixed schedule.
	rule, err := parseRepeatRule(rulePlist(`
		<key>ed</key><real>64092211200</real>
		<key>fa</key><integer>2</integer>
		<key>fu</key><integer>256</integer>
		<key>ia</key><real>1741910400</real>
		<key>of</key><array>
			<dict><key>wd</key><integer>1</integer></dict>
			<dict><key>wd</key><integer>5</integer></dict>
		</array>
		<key>rc</key><integer>0</integer>
		<key>rrv</key><integer>4</integer>
		<key>sr</key><real>-62135769600</real>
		<key>tp</key><integer>0</integer>
		<key>ts</key><integer>0</integer>`))
	require.NoError(t, err)

	assert.Equal(t, RepeatWeekly, rule.Unit)
	assert.Equal(t, 2, rule.Interval)
	assert.False(t, rule.AfterCompletion)
	assert.Nil(t, rule.Until, "the 4001 end date means the series never ends")
	require.Len(t, rule.Offsets, 2)
	assert.Equal(t, time.Monday, *rule.Offsets[0].Weekday)
	assert.Equal(t, time.Friday, *rule.Offsets[1].Weekday)
	assert.Nil(t, rule.Offsets[0].Day)
}

func TestParseRepeatRuleVariants(t *testing.T) {
	t.Run("after completion with an end date", func(t *testing.T) {
		rule, err := parseRepeatRule(rulePlist(`
			<key>ed</key><real>1712016000</real>
			<key>fa</key><integer>11</integer>
			<key>fu</key><integer>8</integer>
			<key>of</key><array><dict><key>dy</key><integer>-1</integer></dict></array>
			<key>tp</key><integer>1</integer>`))
		require.NoError(t, err)

		assert.Equal(t, RepeatMonthly, rule.Unit)
		assert.Equal(t, 11, rule.Interval)
		assert.True(t, rule.AfterCompletion)
		require.NotNil(t, rule.Until)
		assert.Equal(t, time.Date(2024, time.April, 2, 0, 0, 0, 0, time.Local), *rule.Until)
		assert.Equal(t, -1, *rule.Offsets[0].Day)
	})

	t.Run("yearly with month and day, and a weekday ordinal", func(t *testing.T) {
		rule, err := parseRepeatRule(rulePlist(`
			<key>fa</key><integer>1</integer>
			<key>fu</key><integer>4</integer>
			<key>of</key><array>
				<dict><key>dy</key><integer>13</integer><key>mo</key><integer>2</integer></dict>
				<dict><key>wd</key><integer>0</integer><key>wdo</key><integer>2</integer></dict>
			</array>`))
		require.NoError(t, err)

		assert.Equal(t, RepeatYearly, rule.Unit)
		assert.Equal(t, 13, *rule.Offsets[0].Day)
		assert.Equal(t, 2, *rule.Offsets[0].Month)
		assert.Equal(t, time.Sunday, *rule.Offsets[1].Weekday)
		assert.Equal(t, 2, *rule.Offsets[1].Ordinal)
	})

	t.Run("a missing interval counts as one", func(t *testing.T) {
		rule, err := parseRepeatRule(rulePlist(`<key>fu</key><integer>16</integer>`))
		require.NoError(t, err)
		assert.Equal(t, RepeatDaily, rule.Unit)
		assert.Equal(t, 1, rule.Interval)
		assert.Empty(t, rule.Offsets)
	})
}

func TestParseRepeatRuleRejectsWhatItCannotRead(t *testing.T) {
	for name, blob := range map[string][]byte{
		"empty":            nil,
		"binary":           {0x00, 0xFF, 0x10},
		"truncated":        []byte(plistHeader + "<dict><key>fu</key><integer>16"),
		"not a dictionary": []byte(plistHeader + "<array><integer>1</integer></array></plist>"),
		"non-numeric unit": rulePlist(`<key>fu</key><integer>weekly</integer>`),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := parseRepeatRule(blob)
			assert.Error(t, err)
		})
	}
}

func TestIntegration_RepeatRuleOnTemplateAndInstance(t *testing.T) {
	d := openFixtureDB(t)

	templates, err := d.QueryTasks(t.Context(), &TaskFilter{Templates: TemplatesOnly})
	require.NoError(t, err)
	require.Len(t, templates, 1)
	rule := templates[0].Repeat
	require.NotNil(t, rule)
	assert.Equal(t, RepeatWeekly, rule.Unit)
	assert.True(t, rule.AfterCompletion)

	instance := queryTaskByUUID(t, d, fixtureRepeatingInstance)
	require.NotNil(t, instance.Repeat, "an instance reports the rule of the series it came from")
	assert.Equal(t, rule.Unit, instance.Repeat.Unit)

	assert.Nil(t, queryTaskByUUID(t, d, fixtureTodoInToday).Repeat)
}
