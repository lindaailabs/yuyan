// Package protocol 是语燕 v1 通信协议的单一事实源：
// WS 帧（schema/ws/）与 REST 统一响应包裹（schema/rest/）的 JSON Schema 定义，
// 双端通过本包引用；字段语义见 LLM_DEV_GUIDE.md §5，修改需走协议版本化评审。
package protocol

import "embed"

//go:embed schema
var schemaFS embed.FS

// SchemaFS 返回协议 Schema 文件系统（schema/ws/*.json、schema/rest/*.json）。
func SchemaFS() embed.FS { return schemaFS }
