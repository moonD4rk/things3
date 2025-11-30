package things3

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThingsDateToTime(t *testing.T) {
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
			got := ThingsDateToTime(tt.thingsDate)
			if tt.thingsDate <= 0 {
				assert.True(t, got.IsZero(), "ThingsDateToTime(%d) should return zero time", tt.thingsDate)
				return
			}
			assert.Equal(t, tt.wantYear, got.Year())
			assert.Equal(t, tt.wantMonth, got.Month())
			assert.Equal(t, tt.wantDay, got.Day())
		})
	}
}

func TestTimeToThingsDate(t *testing.T) {
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
			got := TimeToThingsDate(tt.time)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestThingsDateToString(t *testing.T) {
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
			got := ThingsDateToString(tt.thingsDate)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStringToThingsDate(t *testing.T) {
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
			got, err := StringToThingsDate(tt.isoDate)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestThingsTimeToString(t *testing.T) {
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
			got := ThingsTimeToString(tt.thingsTime)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNowThingsDate(t *testing.T) {
	now := time.Now()
	thingsDate := NowThingsDate()

	// Convert back and check
	converted := ThingsDateToTime(thingsDate)

	assert.Equal(t, now.Year(), converted.Year())
	assert.Equal(t, now.Month(), converted.Month())
	assert.Equal(t, now.Day(), converted.Day())
}

func TestTodayThingsDateSQL(t *testing.T) {
	sql := TodayThingsDateSQL()
	assert.NotEmpty(t, sql)
	assert.Greater(t, len(sql), 50, "TodayThingsDateSQL() seems too short")
}

func TestUnixToTime(t *testing.T) {
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
			got := UnixToTime(tt.unixTime)
			if tt.wantZero {
				assert.True(t, got.IsZero())
			} else {
				assert.False(t, got.IsZero())
			}
		})
	}
}
