package main

import (
	"context"
	"fmt"
	"mini-codex/src/events"
	"mini-codex/src/model"
	"mini-codex/src/protocol"
	"mini-codex/src/session"
)

func main() {
	fmt.Println("mini-codex")
	ctx := context.Background()

	tools := session.SessionTools{
		Tools: map[string]protocol.ToolSpec{
			"read_file": protocol.ToolSpec{Name: "read_file"},
		},
	}

	s := session.Session{
		Model: &model.DummyProvider{},
		Sink:  &events.StdoutSink{},
		Tools: tools,
	}

	err := <-s.RunUserTurn(ctx, "Hello World!")
	if err != nil {
		panic(err)
	}

	err = <-s.RunUserTurn(ctx, "read README.md")
	if err != nil {
		panic(err)
	}
}
