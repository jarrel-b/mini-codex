package core

import "mini-codex/src/protocol"

type ContextBuilder interface {
	Build([]protocol.Message, []protocol.ToolSpec) protocol.ModelRequest
}

type SimpleContextBuilder struct {
	Model string
}

func (b *SimpleContextBuilder) Build(messages []protocol.Message, tools []protocol.ToolSpec) protocol.ModelRequest {
	return protocol.ModelRequest{
		Model: b.Model,
		Messages: append([]protocol.Message{{
			Role:    protocol.RoleSystem,
			Content: "You are mini-codex, a local coding agent. Use tools when you need to inspect or modify files.",
		}}, messages...),
		Tools: tools,
	}
}
