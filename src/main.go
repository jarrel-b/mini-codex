package main

import (
	"context"
	"fmt"
	"mini-codex/src/model"
	"mini-codex/src/protocol"
)

func main() {
	fmt.Println("mini-codex")

	ctx := context.Background()
	provider := model.DummyProvider{}

	fmt.Println("user message:")
	req := model.ModelRequest{Messages: []protocol.Message{{Content: "Hello!"}}}
	response := provider.Stream(ctx, req)
	for event := range response {
		fmt.Printf("%s", event.TextDelta)
	}

	fmt.Println()
	fmt.Println("tool call:")
	req = model.ModelRequest{Messages: []protocol.Message{{Content: "read go.mod"}}}
	response = provider.Stream(ctx, req)
	for event := range response {
		fmt.Printf("%v", event.ToolCall)
	}
}
