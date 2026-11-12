package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/moond4rk/things3/cmd/things3/internal/change"
	"github.com/moond4rk/things3/cmd/things3/internal/resolve"
)

const descSchedule = "Schedule a todo or project (the app's When). when is required: today, tomorrow, evening, " +
	"anytime, someday, or YYYY-MM-DD. reminder (HH:MM) optionally sets a reminder on that day; it needs a day, " +
	"so it is rejected with anytime or someday. target resolves a UUID, prefix, or title. The result reports " +
	"whether the change was verified in the database."

// ScheduleInput is the schedule parameter set.
type ScheduleInput struct {
	Target   string `json:"target" jsonschema:"the item to schedule (UUID, prefix, or title)"`
	When     string `json:"when" jsonschema:"today, tomorrow, evening, anytime, someday, or YYYY-MM-DD"`
	Reminder string `json:"reminder,omitempty" jsonschema:"reminder time on that day, HH:MM"`
}

func (s *Server) handleSchedule(ctx context.Context, _ *mcp.CallToolRequest, in ScheduleInput) (*mcp.CallToolResult, WriteResult, error) {
	schedule := change.Schedule{When: in.When}
	if in.Reminder != "" {
		clock, err := change.ParseClock(in.Reminder)
		if err != nil {
			return nil, writeError(changeError(err)), nil
		}
		schedule.Reminder = &clock
	}

	m, te, err := s.resolveTarget(ctx, in.Target, resolve.PreferOpen)
	if err != nil {
		return nil, WriteResult{}, err
	}
	if te != nil {
		return nil, writeError(te), nil
	}

	builder, err := schedule.Apply(s.client, m)
	if err != nil {
		return nil, writeError(changeError(err)), nil
	}
	baseline := baselineOf(m)
	result := s.runWrite(ctx, builder, func(ctx context.Context) WriteResult {
		return s.verifyModified(ctx, m.UUID(), string(m.Kind), baseline)
	})
	return nil, result, nil
}
