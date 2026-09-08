package server

import (
	"testing"

	"github.com/lindaailabs/yuyan/protocol"
)

// TestProtocolLinked 验证 apps/server 可通过 workspace/replace 引用 protocol 包并读取 Schema
// （openspec 变更 add-repo-scaffolding 任务 2.5）。
func TestProtocolLinked(t *testing.T) {
	b, err := protocol.SchemaFS().ReadFile("schema/ws/msg.send.json")
	if err != nil {
		t.Fatalf("read msg.send schema: %v", err)
	}
	if len(b) == 0 {
		t.Fatal("msg.send schema is empty")
	}
}
