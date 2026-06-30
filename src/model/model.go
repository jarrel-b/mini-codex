package model

import (
	"context"
	"fmt"
	"mini-codex/src/protocol"
	"mini-codex/src/util"
	"strings"
	"time"
)

type DummyProvider struct{}

func (p *DummyProvider) Stream(ctx context.Context, req protocol.ModelRequest) <-chan protocol.ModelEvent {
	ch := make(chan protocol.ModelEvent)

	var lastMsg protocol.Message

	if len(req.Messages) > 0 {
		lastMsg = req.Messages[len(req.Messages)-1]
	}

	send := func(e protocol.ModelEvent) bool {
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

			event := protocol.ModelEvent{ID: util.MustNewID(), Type: protocol.ModelEventToolCall, ToolCall: toolCall}
			if !send(event) {
				return
			}

			send(protocol.ModelEvent{ID: util.MustNewID(), Type: protocol.ModelEventCompleted})
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
			if !send(protocol.ModelEvent{ID: util.MustNewID(), Type: protocol.ModelEventTextDelta, TextDelta: delta}) {
				return
			}

			select {
			case <-ticker.C:
			case <-ctx.Done():
				return
			}
		}

		send(protocol.ModelEvent{ID: util.MustNewID(), Type: protocol.ModelEventCompleted})
	}()

	return ch
}
