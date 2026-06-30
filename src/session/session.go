package session

import (
	"context"
	"fmt"
	"mini-codex/src/core"
	"mini-codex/src/protocol"
	"mini-codex/src/tools"
	"mini-codex/src/util"
	"os"
)

type Session struct {
	State          core.SessionState
	Model          protocol.ModelProvider
	Tools          tools.ToolRegistry
	ContextBuilder core.ContextBuilder
	Sink           core.EventSink
}

type modelRunResult struct {
	AssistantText string
	ToolCalled    bool
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
				result.ToolCalled = true
				s.emit(ctx, protocol.NewModelToolCallEvent(event.ToolCall))
				s.emit(ctx, protocol.NewToolRequestedEvent(event.ToolCall))
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

					if toolResult.Err != nil {
						return result, toolResult.Err
					}

					toolMessage := protocol.Message{Role: protocol.RoleTool, ToolCallID: event.ToolCall.ID, Content: toolResult.Val.Content}
					if toolResult.Val.Error != nil {
						toolMessage.Content = toolResult.Val.Error.Error()
					}

					s.emit(ctx, protocol.NewModelToolCallCompletedEvent(event.ToolCall.ID))
					s.emit(ctx, protocol.NewToolFinishedEvent(event.ToolCall, toolResult.Val.OK, toolResult.Val.Content, toolResult.Val.Error))
					s.State.History = append(s.State.History, toolMessage)
				}

			case protocol.ModelEventFailed:
				if event.Error == nil {
					event.Error = fmt.Errorf("model failed")
				}
				s.emit(ctx, protocol.NewErrorEvent(event.Error))
				return result, event.Error

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

			if !result.ToolCalled {
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

func (s *Session) toolContext() (tools.ToolContext, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return tools.ToolContext{}, err
	}
	return tools.ToolContext{CurrentDir: cwd}, nil
}
