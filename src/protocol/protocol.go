package protocol

import (
	"context"
	"encoding/json"
)

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type Message struct {
	Role       Role
	Content    string
	ToolCallID string
}

type Op string

const (
	OpNewThread    Op = "new_thread"
	OpResumeThread Op = "resume_thread"
	OpUserInput    Op = "user_input"
	OpInterrupt    Op = "interupt"
	OpApprove      Op = "approve"
	OpDeny         Op = "deny"
)

type Event struct {
	Type            EventType       `json:"type"`
	ThreadID        string          `json:"threadId,omitempty"`
	TurnID          string          `json:"turnId,omitempty"`
	Text            string          `json:"text,omitempty"`
	ToolID          string          `json:"toolId,omitempty"`
	ToolName        string          `json:"toolName,omitempty"`
	ToolArgs        json.RawMessage `json:"toolArgs,omitempty"`
	ToolOutput      string          `json:"toolOutput,omitempty"`
	ToolCallOK      bool            `json:"toolCallOk"`
	ApprovalID      string          `json:"approvalId,omitempty"`
	ApprovalSummary string          `json:"approvalSummary,omitempty"`
	ApprovalDetails string          `json:"approvalDetails,omitempty"`
	Approved        bool            `json:"approved"`
	FileChanged     string          `json:"fileChanged,omitempty"`
	DiffUpdated     string          `json:"diffUpdated,omitempty"`
	Error           string          `json:"error,omitempty"`
}

type EventType string

const (
	EventThreadStarted          EventType = "thread_started"
	EventThreadResumed          EventType = "thread_resumed"
	EventTurnStarted            EventType = "turn_started"
	EventUserMessage            EventType = "user_message"
	EventAssistantDelta         EventType = "assistant_delta"
	EventAssistantMessage       EventType = "assistant_message"
	EventModelTextDelta         EventType = "model_text_delta"
	EventModelToolCall          EventType = "model_tool_call"
	EventModelToolCallCompleted EventType = "model_tool_call_completed"
	EventModelCompleted         EventType = "model_completed"
	EventToolRequested          EventType = "tool_requested"
	EventToolStarted            EventType = "tool_started"
	EventToolOutputDelta        EventType = "tool_output_delta"
	EventToolFinished           EventType = "tool_finished"
	EventApprovalRequested      EventType = "approval_requested"
	EventApprovalResolved       EventType = "approval_resolved"
	EventFileChanged            EventType = "file_changed"
	EventDiffUpdated            EventType = "diff_updated"
	EventTurnFinished           EventType = "turn_finished"
	EventError                  EventType = "error"
)

func NewThreadStartedEvent(threadID string) Event {
	return Event{Type: EventThreadStarted, ThreadID: threadID}
}

func NewThreadResumedEvent(threadID string) Event {
	return Event{Type: EventThreadResumed, ThreadID: threadID}
}

func NewTurnStartedEvent(turnID string) Event {
	return Event{Type: EventTurnStarted, TurnID: turnID}
}

func NewUserMessageEvent(text string) Event {
	return Event{Type: EventUserMessage, Text: text}
}

func NewAssistantDeltaEvent(text string) Event {
	return Event{Type: EventAssistantDelta, Text: text}
}

func NewAssistantMessageEvent(text string) Event {
	return Event{Type: EventAssistantMessage, Text: text}
}

func NewModelTextDeltaEvent(text string) Event {
	return Event{Type: EventModelTextDelta, Text: text}
}

func NewModelToolCallEvent(call ToolCall) Event {
	return Event{Type: EventModelToolCall, ToolID: call.ID, ToolName: call.Name, ToolArgs: call.Args}
}

func NewModelToolCallCompletedEvent(id string) Event {
	return Event{Type: EventModelToolCallCompleted, ToolID: id}
}

func NewModelCompletedEvent() Event {
	return Event{Type: EventModelCompleted}
}

func NewToolRequestedEvent(call ToolCall) Event {
	return Event{Type: EventToolRequested, ToolID: call.ID, ToolName: call.Name, ToolArgs: call.Args}
}

func NewToolStartedEvent(id, name string) Event {
	return Event{Type: EventToolStarted, ToolID: id, ToolName: name}
}

func NewToolOutputDeltaEvent(id, text string) Event {
	return Event{Type: EventToolOutputDelta, ToolID: id, Text: text}
}

func NewToolFinishedEvent(call ToolCall, ok bool, output string, err error) Event {
	event := Event{Type: EventToolFinished, ToolID: call.ID, ToolName: call.Name, ToolOutput: output, ToolCallOK: ok}
	if err != nil {
		event.Error = err.Error()
	}
	return event
}

func NewApprovalRequestedEvent(id, summary, details string) Event {
	return Event{Type: EventApprovalRequested, ApprovalID: id, ApprovalSummary: summary, ApprovalDetails: details}
}

func NewApprovalResolvedEvent(id string, approved bool) Event {
	return Event{Type: EventApprovalResolved, ApprovalID: id, Approved: approved}
}

func NewFileChangedEvent(path string) Event {
	return Event{Type: EventFileChanged, FileChanged: path}
}

func NewDiffUpdatedEvent(diff string) Event {
	return Event{Type: EventDiffUpdated, DiffUpdated: diff}
}

func NewTurnFinishedEvent(turnID string) Event {
	return Event{Type: EventTurnFinished, TurnID: turnID}
}

func NewErrorEvent(err error) Event {
	if err == nil {
		return Event{Type: EventError}
	}
	return Event{Type: EventError, Error: err.Error()}
}

type ModelRequest struct {
	Model    string
	Messages []Message
	Tools    []ToolSpec
}

type ModelEvent struct {
	ID        string
	Type      ModelEventType
	TextDelta string
	ToolCall  ToolCall
	Error     error
}

type ModelEventType string

const (
	ModelEventTextDelta ModelEventType = "text_delta"
	ModelEventToolCall  ModelEventType = "tool_call"
	ModelEventFailed    ModelEventType = "failed"
	ModelEventCompleted ModelEventType = "completed"
)

type ModelProvider interface {
	Stream(context.Context, ModelRequest) <-chan ModelEvent
}

type ToolSpec struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type                 string                 `json:"type"`
	Description          string                 `json:"description"`
	Properties           map[string]InputSchema `json:"properties"`
	Items                *InputSchema           `json:"items,omitempty"`
	Required             []string               `json:"required"`
	AdditionalProperties bool                   `json:"additionalProperties"`
}

type ToolCall struct {
	ID   string          `json:"id"`
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}

type ToolResult struct {
	OK       bool           `json:"ok"`
	Content  string         `json:"content"`
	Metadata map[string]any `json:"metadata"`
	Error    error          `json:"error"`
}
