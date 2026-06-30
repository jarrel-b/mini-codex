package protocol

type ToolSpec struct {
	Name        string
	Description string
	InputSchema map[string]any
}

type ToolCall struct {
	ID   string
	Name string
	Args []string
}

type ToolResult struct {
	OK       bool
	Content  string
	Metadata map[string]any
	Error    error
}
