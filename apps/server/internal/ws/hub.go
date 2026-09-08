// Package ws 是 WebSocket 网关层：连接管理、帧编解码、心跳（W4 实现）。
// 帧格式见 LLM_DEV_GUIDE.md §5.1；trace_id 需随帧透传。
package ws

// Hub 维护在线用户的连接路由表（一期单机内存版）。
// W1 仅占位；W4 实现连接注册/注销与按 user_id 推送。
type Hub struct{}

// NewHub 构造空 Hub。
func NewHub() *Hub { return &Hub{} }
