package scheme

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tokenRecorder records how often the token function ran and with which context.
type tokenRecorder struct {
	calls int
	ctx   context.Context
	token string
	err   error
}

func (r *tokenRecorder) fn(ctx context.Context) (string, error) {
	r.calls++
	r.ctx = ctx
	return r.token, r.err
}

type ctxKey struct{}

// An empty token from the token function must fail immediately on the first
// fetch: no second fetch, and the caller's context must be used.
func TestExecuteEmptyTokenFailsOnFirstFetch(t *testing.T) {
	s := New()
	ctx := context.WithValue(t.Context(), ctxKey{}, "marker")

	t.Run("todo updater", func(t *testing.T) {
		rec := &tokenRecorder{}
		err := NewTodoUpdater(s, rec.fn, "uuid").Completed(true).Execute(ctx)
		require.ErrorIs(t, err, ErrEmptyToken)
		assert.Equal(t, 1, rec.calls, "token function must run exactly once")
		assert.Equal(t, "marker", rec.ctx.Value(ctxKey{}), "caller context must be passed to the token function")
	})

	t.Run("project updater", func(t *testing.T) {
		rec := &tokenRecorder{}
		err := NewProjectUpdater(s, rec.fn, "uuid").Completed(true).Execute(ctx)
		require.ErrorIs(t, err, ErrEmptyToken)
		assert.Equal(t, 1, rec.calls)
		assert.Equal(t, "marker", rec.ctx.Value(ctxKey{}))
	})

	t.Run("auth batch with update", func(t *testing.T) {
		rec := &tokenRecorder{}
		err := NewAuthBatch(s, rec.fn).
			UpdateTodo("uuid", func(todo BatchTodoConfigurator) { todo.Completed(true) }).
			Execute(ctx)
		require.ErrorIs(t, err, ErrEmptyToken)
		assert.Equal(t, 1, rec.calls)
		assert.Equal(t, "marker", rec.ctx.Value(ctxKey{}))
	})
}

// The corrected error text must not reference the nonexistent WithToken API.
func TestEmptyTokenErrorText(t *testing.T) {
	assert.NotContains(t, ErrEmptyToken.Error(), "WithToken")
}

// Token function errors must surface unchanged from Build.
func TestBuildPropagatesTokenFuncError(t *testing.T) {
	s := New()
	tokenErr := errors.New("token store unavailable")
	failing := func(context.Context) (string, error) { return "", tokenErr }

	_, err := NewTodoUpdater(s, failing, "uuid").Completed(true).Build()
	require.ErrorIs(t, err, tokenErr)

	_, err = NewProjectUpdater(s, failing, "uuid").Completed(true).Build()
	assert.ErrorIs(t, err, tokenErr)
}

// A create-only auth batch never emits auth-token, so it must build even when
// the token function fails; a batch containing an update must still fail.
func TestAuthBatchTokenOnlyRequiredForUpdates(t *testing.T) {
	s := New()
	tokenErr := errors.New("token store unavailable")

	t.Run("create-only builds despite failing token func", func(t *testing.T) {
		rec := &tokenRecorder{err: tokenErr}
		thingsURL, err := NewAuthBatch(s, rec.fn).
			AddTodo(func(todo BatchTodoConfigurator) { todo.Title("Test") }).
			Build()
		require.NoError(t, err)
		assert.Zero(t, rec.calls, "create-only batches must not fetch a token")
		assert.False(t, parseQuery(t, thingsURL).Has(KeyAuthToken))
	})

	t.Run("update batch fails with failing token func", func(t *testing.T) {
		rec := &tokenRecorder{err: tokenErr}
		_, err := NewAuthBatch(s, rec.fn).
			UpdateTodo("uuid", func(todo BatchTodoConfigurator) { todo.Completed(true) }).
			Build()
		require.ErrorIs(t, err, tokenErr)
		assert.Equal(t, 1, rec.calls)
	})

	t.Run("update batch fails with empty token", func(t *testing.T) {
		_, err := NewAuthBatch(s, staticTokenFunc("")).
			UpdateTodo("uuid", func(todo BatchTodoConfigurator) { todo.Completed(true) }).
			Build()
		assert.ErrorIs(t, err, ErrEmptyToken)
	})

	t.Run("mixed batch includes token", func(t *testing.T) {
		thingsURL, err := NewAuthBatch(s, staticTokenFunc("secret")).
			AddTodo(func(todo BatchTodoConfigurator) { todo.Title("New") }).
			UpdateTodo("uuid", func(todo BatchTodoConfigurator) { todo.Completed(true) }).
			Build()
		require.NoError(t, err)
		assert.Equal(t, "secret", parseQuery(t, thingsURL).Get(KeyAuthToken))
	})
}

// The token is resolved once and reused by subsequent Build calls.
func TestTokenResolvedOnce(t *testing.T) {
	s := New()
	rec := &tokenRecorder{token: "secret"}
	builder := NewTodoUpdater(s, rec.fn, "uuid").Completed(true)

	first, err := builder.Build()
	require.NoError(t, err)
	second, err := builder.Build()
	require.NoError(t, err)

	assert.Equal(t, first, second)
	assert.Equal(t, 1, rec.calls, "token must be cached after the first fetch")
}
