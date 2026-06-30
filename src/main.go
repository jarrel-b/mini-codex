package main

import (
	"context"
	"fmt"
	"mini-codex/src/core"
	"mini-codex/src/events"
	"mini-codex/src/model"
	"mini-codex/src/protocol"
	"mini-codex/src/session"
	"mini-codex/src/tools"
)

func main() {
	fmt.Println("mini-codex")
	ctx := context.Background()

	registry := tools.ToolRegistry{}
	registry.Register(protocol.ToolSpec{
		Name:        "read_file",
		Description: "Read contents of a file",
		InputSchema: protocol.InputSchema{
			Type:                 "object",
			Description:          "Read contents of a file",
			Properties:           map[string]protocol.InputSchema{"path": {Type: "string", Description: "Path to file"}},
			Required:             []string{"path"},
			AdditionalProperties: false,
		},
	}, tools.ReadFileTool)

	s := session.Session{
		Model:          &model.DummyProvider{},
		Sink:           &events.StdoutSink{},
		Tools:          registry,
		ContextBuilder: &core.SimpleContextBuilder{Model: "dummy_provider"},
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
