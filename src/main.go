package main

import (
	"context"
	"fmt"
	"mini-codex/src/core"
	"mini-codex/src/events"
	"mini-codex/src/model"
	"mini-codex/src/session"
	"mini-codex/src/tools"
)

func main() {
	fmt.Println("mini-codex")
	ctx := context.Background()

	registry := tools.ToolRegistry{}
	registry.Register(tools.ReadFileToolSpec, tools.ReadFileTool)
	registry.Register(tools.ShellToolSpec, tools.ShellTool)

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

	err = <-s.RunUserTurn(ctx, "show me the contents of this directory")
	if err != nil {
		panic(err)
	}
}
