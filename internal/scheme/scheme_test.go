package scheme

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrapExecError(t *testing.T) {
	execErr := errors.New("exit status 1")

	t.Run("nil error returns nil", func(t *testing.T) {
		assert.NoError(t, wrapExecError(nil, []byte("ignored output")))
	})

	tests := []struct {
		name       string
		stderr     string
		wantStderr string
	}{
		{
			name:       "stderr included in error message",
			stderr:     "execution error: Things3 got an error: AppleEvent handler failed. (-10000)\n",
			wantStderr: "AppleEvent handler failed",
		},
		{
			name:   "empty stderr still wraps error",
			stderr: "",
		},
		{
			name:   "whitespace-only stderr treated as empty",
			stderr: "  \n\t",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapExecError(execErr, []byte(tt.stderr))
			require.Error(t, got)
			require.ErrorIs(t, got, execErr, "original error must stay matchable")
			assert.Contains(t, got.Error(), execErr.Error())
			if tt.wantStderr != "" {
				assert.Contains(t, got.Error(), tt.wantStderr)
			}
		})
	}
}
