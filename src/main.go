package main

import (
	"fmt"
	"mini-codex/src/protocol"
	"mini-codex/src/storage"
)

func main() {
	fmt.Println("mini-codex")
	store := storage.JsonLThreadStore{Root: "state"}

	result := <-store.CreateThread()
	if result.Err != nil {
		panic(result.Err)
	}

	err := <-store.AppendMessage(result.Val.ID, protocol.Message{Role: protocol.RoleUser, Content: "Hello World!"})
	if err != nil {
		panic(err)
	}
}
