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

		if strings.Contains(lastMsg.Content, "read go.mod") {
			toolCall := protocol.ToolCall{ID: util.MustNewID(), Name: "read_file", Args: []string{"go.mod"}}

			event := ModelEvent{Type: ModelEventToolCall, ToolCall: toolCall}
			if !send(event) {
				return
			}

			send(ModelEvent{Type: ModelEventCompleted})
			return
		}

		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		response := fmt.Sprintf("I received: %s\n\nTry asking: \"read go.mod\" to trigger a tool call.", lastMsg.Content)

		for _, delta := range strings.SplitAfter(response, " ") {
			if !send(ModelEvent{Type: ModelEventTextDelta, TextDelta: delta}) {
				return
			}

			select {
			case <-ticker.C:
			case <-ctx.Done():
				return
			}
		}

		send(ModelEvent{Type: ModelEventCompleted})
	}()

	return ch
}
