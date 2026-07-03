package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/moond4rk/things3"
	"github.com/moond4rk/things3/cmd/things3/internal/resolve"
)

const descComplete = "Set a todo or project's status. status is required: completed, canceled, or incomplete " +
	"(which reopens a completed or canceled item). target resolves a UUID, prefix, or title. The result " +
	"reports whether the status change was verified in the database."

// CompleteInput is the complete parameter set.
type CompleteInput struct {
	Target string         `json:"target" jsonschema:"the item to update (UUID, prefix, or title)"`
	Status CompleteStatus `json:"status" jsonschema:"completed, canceled, or incomplete (reopen)"`
}

func (s *Server) handleComplete(ctx context.Context, _ *mcp.CallToolRequest, in CompleteInput) (*mcp.CallToolResult, WriteResult, error) {
	m, te, err := s.resolveTarget(ctx, in.Target)
	if err != nil {
		return nil, WriteResult{}, err
	}
	if te != nil {
		return nil, writeError(te), nil
	}

	want, builder, cerr := s.completeBuilder(m, string(in.Status))
	if cerr != nil {
		return nil, writeError(cerr), nil
	}
	result := s.runWrite(ctx, builder, func(ctx context.Context) WriteResult {
		return s.verifyStatus(ctx, m.UUID(), string(m.Kind), want)
	})
	return nil, result, nil
}

// completeBuilder builds the status-flip updater for the match and the status
// the verifier should confirm. incomplete clears both completed and canceled.
func (s *Server) completeBuilder(m resolve.Match, status string) (things3.Status, URLBuilder, *ToolError) {
	project := m.Kind == resolve.KindProject
	switch status {
	case statusCompleted:
		if project {
			return things3.StatusCompleted, s.client.UpdateProject(m.UUID()).Completed(true), nil
		}
		return things3.StatusCompleted, s.client.UpdateTodo(m.UUID()).Completed(true), nil
	case statusCanceled:
		if project {
			return things3.StatusCanceled, s.client.UpdateProject(m.UUID()).Canceled(true), nil
		}
		return things3.StatusCanceled, s.client.UpdateTodo(m.UUID()).Canceled(true), nil
	case statusIncomplete:
		if project {
			return things3.StatusIncomplete, s.client.UpdateProject(m.UUID()).Completed(false).Canceled(false), nil
		}
		return things3.StatusIncomplete, s.client.UpdateTodo(m.UUID()).Completed(false).Canceled(false), nil
	default:
		return 0, nil, invalidInput("status must be completed, canceled, or incomplete")
	}
}
