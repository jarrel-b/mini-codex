package protocol

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type MessageSchema struct {
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

type Event string

const (
	OpThreadStarted     Event = "thread_started"
	OpThreadResumed     Event = "thread_resumed"
	OpTurnStarted       Event = "turn_started"
	OpUserMessage       Event = "user_message"
	OpAssistantDelta    Event = "assistant_delta"
	OpAssistantMessage  Event = "assistant_message"
	OpToolRequested     Event = "tool_requested"
	OpToolStarted       Event = "tool_started"
	OpToolOutput_delta  Event = "tool_output_delta"
	OpToolFinished      Event = "tool_finished"
	OpApprovalRequested Event = "approval_requested"
	OpApprovalResolved  Event = "approval_resolved"
	OpFileChanged       Event = "file_changed"
	OpDiffUpdated       Event = "diff_updated"
	OpTurnFinished      Event = "turn_finished"
	OpError             Event = "error"
)
