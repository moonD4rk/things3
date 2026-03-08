package things3

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatus_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		status  Status
		jsonStr string
	}{
		{StatusIncomplete, `"incomplete"`},
		{StatusCanceled, `"canceled"`},
		{StatusCompleted, `"completed"`},
	}

	for _, tt := range tests {
		t.Run(tt.jsonStr, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.status)
			require.NoError(t, err)
			assert.Equal(t, tt.jsonStr, string(data))

			// Unmarshal
			var got Status
			err = json.Unmarshal(data, &got)
			require.NoError(t, err)
			assert.Equal(t, tt.status, got)
		})
	}
}

func TestStatus_UnmarshalJSON_Unknown(t *testing.T) {
	var got Status
	err := json.Unmarshal([]byte(`"invalid"`), &got)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown status")
}

func TestStatus_StructRoundTrip(t *testing.T) {
	type wrapper struct {
		Status Status `json:"status"`
	}

	original := wrapper{Status: StatusCompleted}
	data, err := json.Marshal(original)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"completed"`)

	var decoded wrapper
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, original, decoded)
}
