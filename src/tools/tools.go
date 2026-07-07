package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"mini-codex/src/core"
	"mini-codex/src/protocol"
	"mini-codex/src/shell"
	"os"
	"path/filepath"
)

type ToolContext struct {
	CurrentDir string
}

type ToolHandlerFn func(context.Context, protocol.ToolCall, ToolContext) protocol.ToolResult

type RegisteredTool struct {
	Spec    protocol.ToolSpec
	Handler ToolHandlerFn
}

type ToolRegistry struct {
	tools map[string]RegisteredTool
}

func (r *ToolRegistry) Register(spec protocol.ToolSpec, handler ToolHandlerFn) {
	if r.tools == nil {
		r.tools = map[string]RegisteredTool{}
	}
	r.tools[spec.Name] = RegisteredTool{spec, handler}
}

func (r *ToolRegistry) Specs() []protocol.ToolSpec {
	specs := make([]protocol.ToolSpec, 0, len(r.tools))
	for _, tool := range r.tools {
		specs = append(specs, tool.Spec)
	}
	return specs
}

func (r *ToolRegistry) Execute(ctx context.Context, call protocol.ToolCall, toolCtx ToolContext) <-chan core.Result[protocol.ToolResult] {
	ch := make(chan core.Result[protocol.ToolResult], 1)

	go func() {
		defer close(ch)

		tool, ok := r.tools[call.Name]
		if !ok {
			ch <- core.Result[protocol.ToolResult]{Err: fmt.Errorf("tool %s not registered", call.Name)}
			return
		}

		resultCh := make(chan core.Result[protocol.ToolResult], 1)
		go func() {
			resultCh <- core.Result[protocol.ToolResult]{Val: tool.Handler(ctx, call, toolCtx)}
		}()

		var result core.Result[protocol.ToolResult]
		var resultOk bool

		select {
		case <-ctx.Done():
			ch <- core.Result[protocol.ToolResult]{Err: ctx.Err()}
			return
		case result, resultOk = <-resultCh:
			if !resultOk {
				result = core.Result[protocol.ToolResult]{Err: fmt.Errorf("tool execution closed without result")}
			}
			ch <- result
			return
		}
	}()

	return ch
}

var ReadFileToolSpec = protocol.ToolSpec{
	Name:        "read_file",
	Description: "Read contents of a file",
	InputSchema: protocol.InputSchema{
		Type:        "object",
		Description: "Read contents of a file",
		Properties: map[string]protocol.InputSchema{
			"path": {
				Type:        "string",
				Description: "Path to file",
			},
		},
		Required:             []string{"path"},
		AdditionalProperties: false,
	},
}

func ReadFileTool(ctx context.Context, call protocol.ToolCall, toolCtx ToolContext) protocol.ToolResult {
	type inputSchema struct {
		Path string `json:"path"`
	}

	var input inputSchema
	if err := json.Unmarshal(call.Args, &input); err != nil {
		return protocol.ToolResult{OK: false, Error: err}
	}

	if input.Path == "" {
		return protocol.ToolResult{OK: false, Error: fmt.Errorf("no path provided")}
	}

	content, err := os.ReadFile(filepath.Join(toolCtx.CurrentDir, input.Path))
	if err != nil {
		return protocol.ToolResult{OK: false, Error: err}
	}

	return protocol.ToolResult{OK: true, Content: string(content)}
}

var ShellToolSpec = protocol.ToolSpec{
	Name:        "shell",
	Description: "Execute a shell command",
	InputSchema: protocol.InputSchema{
		Type:        "object",
		Description: "Execute a shell command",
		Properties: map[string]protocol.InputSchema{
			"command": {
				Type:        "string",
				Description: "Command to execute",
			},
			"args": {
				Type:        "array",
				Description: "Arguments to pass to the command",
				Items:       &protocol.InputSchema{Type: "string"},
			},
			"timeoutMs": {
				Type:        "number",
				Description: "Optional timeout in milliseconds",
			},
		},
		Required:             []string{"command"},
		AdditionalProperties: false,
	},
}

func ShellTool(ctx context.Context, call protocol.ToolCall, toolCtx ToolContext) protocol.ToolResult {
	type inputSchema struct {
		Command   string   `json:"command"`
		Args      []string `json:"args"`
		TimeoutMs int      `json:"timeoutMs"`
	}

	var input inputSchema
	if err := json.Unmarshal(call.Args, &input); err != nil {
		return protocol.ToolResult{OK: false, Error: err}
	}

	if input.Command == "" {
		return protocol.ToolResult{OK: false, Error: fmt.Errorf("no command provided")}
	}

	resultCh := shell.RunExec(ctx, shell.ExecRequest{
		Command:    input.Command,
		Args:       input.Args,
		CurrentDir: toolCtx.CurrentDir,
		TimeoutMs:  input.TimeoutMs,
	})

	select {
	case <-ctx.Done():
		return protocol.ToolResult{OK: false, Error: ctx.Err()}
	case result := <-resultCh:
		if result.Err != nil {
			return protocol.ToolResult{OK: false, Error: result.Err}
		}
		return protocol.ToolResult{OK: true, Content: result.Val.Stdout}
	}
}
