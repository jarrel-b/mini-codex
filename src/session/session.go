package session

import (
	"context"
	"fmt"
	"mini-codex/src/core"
	"mini-codex/src/protocol"
	"mini-codex/src/util"
	"os"
	"path/filepath"
)

type ToolContext struct {
	CurrentDir string
}

type SessionTools struct {
	Tools map[string]protocol.ToolSpec
}

func (s *SessionTools) Specs() []protocol.ToolSpec {
	specs := make([]protocol.ToolSpec, 0, len(s.Tools))
	for _, spec := range s.Tools {
		specs = append(specs, spec)
	}
	return specs
}

func (s *SessionTools) Execute(ctx context.Context, call protocol.ToolCall, toolCtx ToolContext) <-chan protocol.ToolResult {
	ch := make(chan protocol.ToolResult, 1)

	// Move this to a tool registry later.
	go func() {
		defer close(ch)

		select {
		case <-ctx.Done():
			ch <- protocol.ToolResult{OK: false, Error: ctx.Err()}
			return
		default:
			tool, ok := s.Tools[call.Name]
			if !ok {
				ch <- protocol.ToolResult{OK: false, Error: fmt.Errorf("tool %s not found", call.Name)}
				return
			}

			switch tool.Name {
			case "read_file":
				if len(call.Args) == 0 {
					ch <- protocol.ToolResult{OK: false, Error: fmt.Errorf("expected input file name")}
					return
				}
				content, err := os.ReadFile(filepath.Join(toolCtx.CurrentDir, call.Args[0]))
				if err != nil {
					ch <- protocol.ToolResult{OK: false, Error: err}
					return
				}
				ch <- protocol.ToolResult{OK: true, Content: string(content)}
				return
			default:
				ch <- protocol.ToolResult{OK: false, Error: fmt.Errorf("tool %s not implemented", call.Name)}
				return
			}
		}
	}()

	return ch
}

type Session struct {
	State          core.SessionState
	Model          protocol.ModelProvider
	Tools          SessionTools
	ContextBuilder core.ContextBuilder
	Sink           core.EventSink
}

type modelRunResult struct {
	AssistantText string
	HasToolCall   bool
}

func (s *Session) runModelStream(ctx context.Context, request protocol.ModelRequest) (modelRunResult, error) {
	var result modelRunResult

	stream := s.Model.Stream(ctx, request)

	for {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case event, ok := <-stream:
			if !ok {
				return result, nil
			}

			switch event.Type {
			case protocol.ModelEventTextDelta:
				result.AssistantText += event.TextDelta
				s.emit(ctx, protocol.NewModelTextDeltaEvent(event.TextDelta))
				s.emit(ctx, protocol.NewAssistantDeltaEvent(event.TextDelta))

			case protocol.ModelEventToolCall:
				result.HasToolCall = true
				s.emit(ctx, protocol.NewModelToolCallEvent(event.ToolCall.ID, event.ToolCall.Name, event.ToolCall.Args))
				s.emit(ctx, protocol.NewToolRequestedEvent(event.ToolCall.ID, event.ToolCall.Name, event.ToolCall.Args))
				s.emit(ctx, protocol.NewToolStartedEvent(event.ToolCall.ID, event.ToolCall.Name))

				toolCtx, err := s.toolContext()
				if err != nil {
					return result, err
				}

				select {
				case <-ctx.Done():
					return result, ctx.Err()
				case toolResult, ok := <-s.Tools.Execute(ctx, event.ToolCall, toolCtx):
					if !ok {
						return result, fmt.Errorf("tool execution closed without result")
					}

					s.emit(ctx, protocol.NewModelToolCallCompletedEvent(event.ToolCall.ID))
					s.emit(ctx, protocol.NewToolFinishedEvent(event.ToolCall.ID, event.ToolCall.Name, toolResult.Content, toolResult.OK, toolResult.Error))
					s.State.History = append(s.State.History, protocol.Message{Role: protocol.RoleTool, ToolCallID: event.ToolCall.ID, Content: toolResult.Content})
				}

			case protocol.ModelEventCompleted:
				s.emit(ctx, protocol.NewModelCompletedEvent())
			}
		}
	}
}

const maxLoops = 8

func (s *Session) RunUserTurn(ctx context.Context, text string) <-chan error {
	turnID := util.MustNewID()
	s.emit(ctx, protocol.NewTurnStartedEvent(turnID))
	s.emit(ctx, protocol.NewUserMessageEvent(text))
	s.State.History = append(s.State.History, protocol.Message{Role: protocol.RoleUser, Content: text})

	ch := make(chan error, 1)

	go func() {
		defer close(ch)

		finalAssistantText := ""
		completed := false

		for range maxLoops {
			req := s.ContextBuilder.Build(s.State.History, s.Tools.Specs())

			result, err := s.runModelStream(ctx, req)
			if err != nil {
				ch <- err
				return
			}

			if result.AssistantText != "" {
				s.State.History = append(s.State.History, protocol.Message{
					Role:    protocol.RoleAssistant,
					Content: result.AssistantText,
				})
				finalAssistantText += result.AssistantText
			}

			if !result.HasToolCall {
				completed = true
				break
			}
		}

		if !completed {
			ch <- fmt.Errorf("max tool loop count exceeded")
			return
		}

		if len(finalAssistantText) > 0 {
			s.emit(ctx, protocol.NewAssistantMessageEvent(finalAssistantText))
		}

		s.emit(ctx, protocol.NewTurnFinishedEvent(turnID))
		ch <- nil
	}()

	return ch
}

func (s *Session) emit(ctx context.Context, event protocol.Event) {
	event.ThreadID = s.State.Thread.ID
	select {
	case <-s.Sink.Emit(ctx, event):
	case <-ctx.Done():
	}
}

func (s *Session) toolContext() (ToolContext, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return ToolContext{}, err
	}
	return ToolContext{CurrentDir: cwd}, nil
}
