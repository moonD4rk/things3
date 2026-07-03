package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/cmd/things3/internal/resolve"
)

const descSchedule = "Schedule a todo or project (the app's When). when is required: today, tomorrow, evening, " +
	"anytime, someday, or YYYY-MM-DD. target resolves a UUID, prefix, or title. The result reports whether " +
	"the change was verified in the database."

// ScheduleInput is the schedule parameter set.
type ScheduleInput struct {
	Target string `json:"target" jsonschema:"the item to schedule (UUID, prefix, or title)"`
	When   string `json:"when" jsonschema:"today, tomorrow, evening, anytime, someday, or YYYY-MM-DD"`
}

func (s *Server) handleSchedule(ctx context.Context, _ *mcp.CallToolRequest, in ScheduleInput) (*mcp.CallToolResult, WriteResult, error) {
	m, te, err := s.resolveTarget(ctx, in.Target)
	if err != nil {
		return nil, WriteResult{}, err
	}
	if te != nil {
		return nil, writeError(te), nil
	}

	baseline := baselineOf(m)
	builder, serr := s.scheduleBuilder(m, in.When)
	if serr != nil {
		return nil, writeError(serr), nil
	}
	result := s.runWrite(ctx, builder, func(ctx context.Context) WriteResult {
		return s.verifyModified(ctx, m.UUID(), string(m.Kind), baseline)
	})
	return nil, result, nil
}

// scheduleBuilder applies the when grammar to the match's updater.
func (s *Server) scheduleBuilder(m resolve.Match, when string) (URLBuilder, *ToolError) {
	if m.Kind == resolve.KindProject {
		u, err := things3.ParseWhen(s.client.UpdateProject(m.UUID()), when)
		if err != nil {
			return nil, invalidInput(err.Error())
		}
		return u, nil
	}
	u, err := things3.ParseWhen(s.client.UpdateTodo(m.UUID()), when)
	if err != nil {
		return nil, invalidInput(err.Error())
	}
	return u, nil
}
