package events

import (
	"context"
	"encoding/json"
	"fmt"
	"mini-codex/src/protocol"
)

type StdoutSink struct{}

func (s *StdoutSink) Emit(ctx context.Context, event protocol.Event) <-chan struct{} {
	ch := make(chan struct{}, 1)

	go func() {
		defer close(ch)

		select {
		case <-ctx.Done():
			return
		default:
			b, err := json.Marshal(event)
			if err != nil {
				fmt.Printf("event marshal error: %v\n", err)
				return
			}
			fmt.Println(string(b))
		}
	}()

	return ch
}
