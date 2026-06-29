package storage

import (
	"encoding/json"
	"fmt"
	"mini-codex/src/core"
	"mini-codex/src/protocol"
	"mini-codex/src/util"
	"os"
	"path/filepath"
	"time"
)

type JsonLThreadStore struct {
	Root string
}

func (s *JsonLThreadStore) CreateThread() <-chan core.Result[core.Thread] {
	now := time.Now().UnixMilli()
	thread := core.Thread{
		ID:        util.MustNewID(),
		CreatedAt: now,
		UpdatedAt: now,
	}

	ch := make(chan core.Result[core.Thread], 1)

	err := os.MkdirAll(threadDirName(s.Root, thread.ID), 0755)
	if err != nil {
		ch <- core.Result[core.Thread]{Err: err}
		return ch
	}

	err = writeThreadFile(metaFileName(s.Root, thread.ID), thread)
	if err != nil {
		ch <- core.Result[core.Thread]{Err: err}
		return ch
	}

	event := protocol.Event{Type: protocol.EventThreadStarted, ThreadID: thread.ID}
	err = appendEvent(eventsFileName(s.Root, thread.ID), event)
	if err != nil {
		ch <- core.Result[core.Thread]{Err: err}
		return ch
	}

	ch <- core.Result[core.Thread]{Val: thread}
	return ch
}

func (s *JsonLThreadStore) AppendMessage(threadID string, msg protocol.Message) <-chan error {
	ch := make(chan error, 1)
	switch msg.Role {
	case protocol.RoleUser:
		err := appendEvent(eventsFileName(s.Root, threadID), protocol.Event{Type: protocol.EventUserMessage, ThreadID: threadID, Text: msg.Content})
		ch <- err
		return ch
	case protocol.RoleAssistant:
		err := appendEvent(eventsFileName(s.Root, threadID), protocol.Event{Type: protocol.EventAssistantMessage, ThreadID: threadID, Text: msg.Content})
		ch <- err
		return ch
	default:
		ch <- fmt.Errorf("unknown role: %s", msg.Role)
		return ch
	}
}

func writeThreadFile(fileName string, thread core.Thread) error {
	content, err := json.Marshal(thread)
	if err != nil {
		return err
	}
	return writeFile(fileName, content)
}

func writeFile(name string, content []byte) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	_, err = f.Write(content)
	return err
}

func eventsFileName(root, threadID string) string {
	return filepath.Join(threadDirName(root, threadID), "events.jsonl")
}

func metaFileName(root, threadID string) string {
	return filepath.Join(threadDirName(root, threadID), "meta.json")
}

func threadDirName(root, threadID string) string {
	return filepath.Join(root, "threads", threadID)
}

func appendEvent(fileName string, event protocol.Event) error {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	content, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = file.Write(content)
	return err
}
