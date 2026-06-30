package core

import (
	"context"
	"fmt"
	"mini-codex/src/protocol"
)

type Thread struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

type SessionState struct {
	Thread  Thread
	History []protocol.Message
}

type Turn struct {
	ID        string
	ThreadID  string
	StartedAt int64
	UserText  string
}

type EventSink interface {
	Emit(context.Context, protocol.Event) <-chan struct{}
}

type Result[T any] struct {
	Val T
	Err error
}

type ThreadStore interface {
	CreateThread() <-chan Result[Thread]
	GetThread(threadID string) <-chan Result[Thread]
	ListThreads() []<-chan Result[Thread]
	AppendMessage(threadID string, message protocol.Message) <-chan error
	LoadMessages(threadID string) Result[[]protocol.Message]
}

type ThreadManager struct {
	store ThreadStore
}

func (t *ThreadManager) createSession() (SessionState, error) {
	thread := <-t.store.CreateThread()
	if thread.Err != nil {
		return SessionState{}, thread.Err
	}
	return SessionState{Thread: thread.Val, History: []protocol.Message{}}, nil
}

func (t *ThreadManager) resumeSession(threadID string) (SessionState, error) {
	thread := <-t.store.GetThread(threadID)
	if thread.Err != nil {
		return SessionState{}, thread.Err
	}

	if thread.Val.ID == "" {
		return SessionState{}, fmt.Errorf("threadID=%s not found", threadID)
	}

	history := t.store.LoadMessages(threadID)
	if history.Err != nil {
		return SessionState{}, history.Err
	}

	return SessionState{thread.Val, history.Val}, nil
}
