package protocol

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
	Type     EventType
	ThreadID string
	Text     string
}

type EventType string

const (
	EventThreadStarted     EventType = "thread_started"
	EventThreadResumed     EventType = "thread_resumed"
	EventTurnStarted       EventType = "turn_started"
	EventUserMessage       EventType = "user_message"
	EventAssistantDelta    EventType = "assistant_delta"
	EventAssistantMessage  EventType = "assistant_message"
	EventToolRequested     EventType = "tool_requested"
	EventToolStarted       EventType = "tool_started"
	EventToolOutput_delta  EventType = "tool_output_delta"
	EventToolFinished      EventType = "tool_finished"
	EventApprovalRequested EventType = "approval_requested"
	EventApprovalResolved  EventType = "approval_resolved"
	EventFileChanged       EventType = "file_changed"
	EventDiffUpdated       EventType = "diff_updated"
	EventTurnFinished      EventType = "turn_finished"
	EventError             EventType = "error"
)
