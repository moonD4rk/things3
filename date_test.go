package things3

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_thingsDateToTime(t *testing.T) {
	tests := []struct {
		name       string
		thingsDate int64
		wantYear   int
		wantMonth  time.Month
		wantDay    int
	}{
		{
			name:       "2021-03-28",
			thingsDate: 132464128, // From Python test
			wantYear:   2021,
			wantMonth:  time.March,
			wantDay:    28,
		},
		{
			name:       "zero",
			thingsDate: 0,
			wantYear:   1,
			wantMonth:  time.January,
			wantDay:    1,
		},
		{
			name:       "negative",
			thingsDate: -1,
			wantYear:   1,
			wantMonth:  time.January,
			wantDay:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := thingsDateToTime(tt.thingsDate)
			if tt.thingsDate <= 0 {
				assert.True(t, got.IsZero(), "thingsDateToTime(%d) should return zero time", tt.thingsDate)
				return
			}
			assert.Equal(t, tt.wantYear, got.Year())
			assert.Equal(t, tt.wantMonth, got.Month())
			assert.Equal(t, tt.wantDay, got.Day())
		})
	}
}

func Test_timeToThingsDate(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
		want int64
	}{
		{
			name: "2021-03-28",
			time: time.Date(2021, time.March, 28, 0, 0, 0, 0, time.Local),
			want: 132464128,
		},
		{
			name: "zero",
			time: time.Time{},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := timeToThingsDate(tt.time)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_thingsDateToString(t *testing.T) {
	tests := []struct {
		name       string
		thingsDate int64
		want       string
	}{
		{
			name:       "2021-03-28",
			thingsDate: 132464128,
			want:       "2021-03-28",
		},
		{
			name:       "zero",
			thingsDate: 0,
			want:       "",
		},
		{
			name:       "negative",
			thingsDate: -1,
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := thingsDateToString(tt.thingsDate)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_stringToThingsDate(t *testing.T) {
	tests := []struct {
		name    string
		isoDate string
		want    int64
		wantErr bool
	}{
		{
			name:    "2021-03-28",
			isoDate: "2021-03-28",
			want:    132464128,
			wantErr: false,
		},
		{
			name:    "empty",
			isoDate: "",
			want:    0,
			wantErr: false,
		},
		{
			name:    "invalid",
			isoDate: "invalid",
			want:    0,
			wantErr: true,
		},
		{
			name:    "wrong format",
			isoDate: "03-28-2021",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := stringToThingsDate(tt.isoDate)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_thingsTimeToString(t *testing.T) {
	tests := []struct {
		name       string
		thingsTime int64
		want       string
	}{
		{
			name:       "12:34",
			thingsTime: (12 << 26) | (34 << 20), // Correct calculation
			want:       "12:34",
		},
		{
			name:       "00:00",
			thingsTime: 0,
			want:       "",
		},
		{
			name:       "negative",
			thingsTime: -1,
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := thingsTimeToString(tt.thingsTime)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_nowThingsDate(t *testing.T) {
	now := time.Now()
	thingsDate := nowThingsDate()

	// Convert back and check
	converted := thingsDateToTime(thingsDate)

	assert.Equal(t, now.Year(), converted.Year())
	assert.Equal(t, now.Month(), converted.Month())
	assert.Equal(t, now.Day(), converted.Day())
}

func Test_todayThingsDateSQL(t *testing.T) {
	sql := todayThingsDateSQL()
	assert.NotEmpty(t, sql)
	assert.Greater(t, len(sql), 50, "todayThingsDateSQL() seems too short")
}

func Test_unixToTime(t *testing.T) {
	tests := []struct {
		name     string
		unixTime float64
		wantZero bool
	}{
		{
			name:     "zero",
			unixTime: 0,
			wantZero: true,
		},
		{
			name:     "valid",
			unixTime: 1616946000, // 2021-03-28 15:00:00 UTC
			wantZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := unixToTime(tt.unixTime)
			if tt.wantZero {
				assert.True(t, got.IsZero())
			} else {
				assert.False(t, got.IsZero())
			}
		})
	}
}
