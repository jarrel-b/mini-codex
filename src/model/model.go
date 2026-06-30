package model

import (
	"context"
	"fmt"
	"mini-codex/src/protocol"
	"mini-codex/src/util"
	"strings"
	"time"
)

type ModelRequest struct {
	Model    string
	Messages []protocol.Message
	Tools    []protocol.ToolSpec
}

type ModelEvent struct {
	ID        string
	Type      ModelEventType
	TextDelta string
	ToolCall  protocol.ToolCall
}

type ModelEventType string

const (
	ModelEventTextDelta ModelEventType = "text_delta"
	ModelEventToolCall  ModelEventType = "tool_call"
	ModelEventCompleted ModelEventType = "completed"
)

type ModelProvider interface {
	Stream(context.Context, ModelRequest) <-chan ModelEvent
}

type DummyProvider struct{}

func (p *DummyProvider) Stream(ctx context.Context, req ModelRequest) <-chan ModelEvent {
	ch := make(chan ModelEvent)

	var lastMsg protocol.Message

	if len(req.Messages) > 0 {
		lastMsg = req.Messages[len(req.Messages)-1]
	}

	send := func(e ModelEvent) bool {
		select {
		case ch <- e:
			return true
		case <-ctx.Done():
			return false
		}
	}

	go func() {
		defer close(ch)

		if strings.Contains(lastMsg.Content, "read README.md") {
			toolCall := protocol.ToolCall{ID: util.MustNewID(), Name: "read_file", Args: []string{"README.md"}}

			event := ModelEvent{ID: util.MustNewID(), Type: ModelEventToolCall, ToolCall: toolCall}
			if !send(event) {
				return
			}

			send(ModelEvent{ID: util.MustNewID(), Type: ModelEventCompleted})
			return
		}

		var response string

		if lastMsg.ToolCallID != "" {
			response = fmt.Sprintf("Result of tool call: %s", lastMsg.Content)
		} else {
			response = fmt.Sprintf("I received: %s\n\nTry asking: \"read README.md\" to trigger a tool call.", lastMsg.Content)
		}

		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for _, delta := range strings.SplitAfter(response, " ") {
			if !send(ModelEvent{ID: util.MustNewID(), Type: ModelEventTextDelta, TextDelta: delta}) {
				return
			}

			select {
			case <-ticker.C:
			case <-ctx.Done():
				return
			}
		}

		send(ModelEvent{ID: util.MustNewID(), Type: ModelEventCompleted})
	}()

	return ch
}
