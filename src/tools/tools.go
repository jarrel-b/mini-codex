package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"mini-codex/src/core"
	"mini-codex/src/protocol"
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
