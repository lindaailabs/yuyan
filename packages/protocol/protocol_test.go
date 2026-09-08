package protocol

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// 编译 schema 并校验 doc（map 形式），返回校验错误。
func validate(t *testing.T, schemaPath string, doc string) error {
	t.Helper()
	raw, err := SchemaFS().ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}
	var schemaDoc any
	if err := json.Unmarshal(raw, &schemaDoc); err != nil {
		return fmt.Errorf("unmarshal schema: %w", err)
	}
	c := jsonschema.NewCompiler()
	if err := c.AddResource(schemaPath, schemaDoc); err != nil {
		return fmt.Errorf("add resource: %w", err)
	}
	sch, err := c.Compile(schemaPath)
	if err != nil {
		return fmt.Errorf("compile schema: %w", err)
	}
	var v any
	if err := json.Unmarshal([]byte(doc), &v); err != nil {
		return fmt.Errorf("unmarshal doc: %w", err)
	}
	return sch.Validate(v)
}

func TestFrameSchemas(t *testing.T) {
	cases := []struct {
		schema  string
		name    string
		valid   string
		invalid string
	}{
		{
			schema:  "schema/ws/conn.auth.json",
			name:    "conn.auth",
			valid:   `{"cmd":"conn.auth","seq":1,"data":{"token":"jwt-token"}}`,
			invalid: `{"cmd":"conn.auth","seq":1}`, // 缺 data.token
		},
		{
			schema:  "schema/ws/conn.heartbeat.json",
			name:    "conn.heartbeat",
			valid:   `{"cmd":"conn.heartbeat","seq":42}`,
			invalid: `{"cmd":"conn.heartbeat"}`, // 缺 seq
		},
		{
			schema:  "schema/ws/msg.send.json",
			name:    "msg.send",
			valid:   `{"cmd":"msg.send","seq":12345,"data":{"conv_id":9,"content":"hello"},"client_msg_id":"550e8400-e29b-41d4-a716-446655440000"}`,
			invalid: `{"cmd":"msg.send","seq":1,"data":{"conv_id":9,"content":"hi"}}`, // 缺 client_msg_id
		},
		{
			schema:  "schema/ws/msg.push.json",
			name:    "msg.push",
			valid:   `{"cmd":"msg.push","seq":2,"data":{"msg_id":101,"conv_id":9,"sender_id":1,"msg_type":1,"content":"hello","created_at":1757300000000}}`,
			invalid: `{"cmd":"msg.push","seq":2,"data":{"msg_id":101}}`, // data 缺必填字段
		},
		{
			schema:  "schema/ws/msg.ack.json",
			name:    "msg.ack",
			valid:   `{"cmd":"msg.ack","seq":3,"data":{"client_msg_id":"550e8400-e29b-41d4-a716-446655440000","msg_id":101,"created_at":1757300000000}}`,
			invalid: `{"cmd":"msg.ack","seq":3,"data":{"msg_id":101,"created_at":1}}`, // data 缺 client_msg_id
		},
		{
			schema:  "schema/ws/msg.pull.json",
			name:    "msg.pull",
			valid:   `{"cmd":"msg.pull","seq":4,"data":{"cursor":0,"limit":20}}`,
			invalid: `{"cmd":"msg.pull","seq":4,"data":{"cursor":0,"limit":1000}}`, // limit 超上限
		},
		{
			schema:  "schema/ws/conv.unread.json",
			name:    "conv.unread",
			valid:   `{"cmd":"conv.unread","seq":5,"data":{"conv_id":9,"unread_count":3}}`,
			invalid: `{"cmd":"conv.unread","seq":5,"data":{"conv_id":9,"unread_count":-1}}`, // 负未读数
		},
		{
			schema:  "schema/ws/response.json",
			name:    "ws response",
			valid:   `{"cmd":"msg.send","seq":12345,"code":0,"msg":"ok","data":{"msg_id":1}}`,
			invalid: `{"cmd":"msg.send","code":0,"msg":"ok"}`, // 缺 seq
		},
		{
			schema:  "schema/rest/response.json",
			name:    "rest wrapper",
			valid:   `{"code":0,"msg":"ok","data":null}`,
			invalid: `{"msg":"ok"}`, // 缺 code
		},
	}

	for _, tc := range cases {
		t.Run(tc.name+" valid", func(t *testing.T) {
			if err := validate(t, tc.schema, tc.valid); err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
		})
		t.Run(tc.name+" invalid", func(t *testing.T) {
			err := validate(t, tc.schema, tc.invalid)
			if err == nil {
				t.Errorf("expected invalid, got pass")
			} else if !strings.Contains(err.Error(), "jsonschema") {
				t.Logf("rejected with: %v", err)
			}
		})
	}
}

// TestSchemaCount 确保命令字全集（7 个 + WS 响应 + REST 包裹）齐全。
func TestSchemaCount(t *testing.T) {
	want := []string{
		"schema/ws/conn.auth.json",
		"schema/ws/conn.heartbeat.json",
		"schema/ws/msg.send.json",
		"schema/ws/msg.push.json",
		"schema/ws/msg.ack.json",
		"schema/ws/msg.pull.json",
		"schema/ws/conv.unread.json",
		"schema/ws/response.json",
		"schema/rest/response.json",
	}
	for _, p := range want {
		if _, err := SchemaFS().ReadFile(p); err != nil {
			t.Errorf("missing schema %s: %v", p, err)
		}
	}
}
