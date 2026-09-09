package model

import "time"

// 消息角色（guide §4：pet_messages.role）。
const (
	MessageRoleUser      = "user"
	MessageRoleAssistant = "assistant"
	MessageRoleSystem    = "system"
)

// 消息状态（status SMALLINT，禁止物理删除）。
const (
	MessageStatusOK     int16 = 1 // 正常
	MessageStatusFailed int16 = 2 // AI 生成失败（内容填兜底文案，可原 client_msg_id 重试）
)

// PetConversation 用户与宠物的会话（pet_conversations 表）。
type PetConversation struct {
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID         int64     `gorm:"column:user_id;not null"`
	PetID          int64     `gorm:"column:pet_id;not null"`
	LastMsgID      *int64    `gorm:"column:last_msg_id"`
	LastMsgPreview string    `gorm:"column:last_msg_preview;size:255;not null;default:''"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime"` // MySQL ON UPDATE 维护
}

func (PetConversation) TableName() string { return "pet_conversations" }

// PetMessage 对话消息（pet_messages 表）。id 兼作增量游标。
type PetMessage struct {
	ID           int64   `gorm:"column:id;primaryKey;autoIncrement"`
	ConvID       int64   `gorm:"column:conv_id;not null"`
	UserID       int64   `gorm:"column:user_id;not null"`
	PetID        int64   `gorm:"column:pet_id;not null"`
	Role         string  `gorm:"column:role;size:16;not null"`
	Content      string  `gorm:"column:content;not null"`
	ClientMsgID  *string `gorm:"column:client_msg_id;size:36"`
	Model        string  `gorm:"column:model;size:64;not null;default:''"`
	InputTokens  int     `gorm:"column:input_tokens;not null;default:0"`
	OutputTokens int     `gorm:"column:output_tokens;not null;default:0"`
	LatencyMS    int     `gorm:"column:latency_ms;not null;default:0"`
	Status       int16   `gorm:"column:status;not null;default:1"`
	ErrorCode    int     `gorm:"column:error_code;not null;default:0"`
	CreatedAt    int64   `gorm:"column:created_at;not null"`
}

func (PetMessage) TableName() string { return "pet_messages" }

// AICallLog AI 调用日志（ai_call_logs 表，guide §5 用量统计）。
type AICallLog struct {
	ID           int64  `gorm:"column:id;primaryKey;autoIncrement"`
	UserID       int64  `gorm:"column:user_id;not null"`
	PetID        int64  `gorm:"column:pet_id;not null"`
	ConvID       int64  `gorm:"column:conv_id;not null;default:0"`
	MsgID        int64  `gorm:"column:msg_id;not null;default:0"`
	Provider     string `gorm:"column:provider;size:32;not null;default:''"`
	Model        string `gorm:"column:model;size:64;not null;default:''"`
	PromptVer    string `gorm:"column:prompt_version;size:32;not null;default:''"`
	InputTokens  int    `gorm:"column:input_tokens;not null;default:0"`
	OutputTokens int    `gorm:"column:output_tokens;not null;default:0"`
	LatencyMS    int    `gorm:"column:latency_ms;not null;default:0"`
	ErrCode      int    `gorm:"column:err_code;not null;default:0"`
	CacheHit     bool   `gorm:"column:cache_hit;not null;default:false"`
	CreatedAt    int64  `gorm:"column:created_at;not null"`
}

func (AICallLog) TableName() string { return "ai_call_logs" }

// MessageItem 消息响应 DTO（用户消息与宠物消息同构）。
type MessageItem struct {
	ID           int64   `json:"id"`
	ConvID       int64   `json:"conv_id"`
	Role         string  `json:"role"`
	Content      string  `json:"content"`
	ClientMsgID  *string `json:"client_msg_id"`
	Status       int16   `json:"status"`
	ErrorCode    int     `json:"error_code,omitempty"`
	Model        string  `json:"model,omitempty"`
	InputTokens  int     `json:"input_tokens,omitempty"`
	OutputTokens int     `json:"output_tokens,omitempty"`
	LatencyMS    int     `json:"latency_ms,omitempty"`
	CreatedAt    int64   `json:"created_at"`
}

// ConversationItem 会话响应 DTO。
type ConversationItem struct {
	ID             int64  `json:"id"`
	UserID         int64  `json:"user_id"`
	PetID          int64  `json:"pet_id"`
	LastMsgID      *int64 `json:"last_msg_id"`
	LastMsgPreview string `json:"last_msg_preview"`
	UpdatedAt      int64  `json:"updated_at"`
}

// CreateConversationInput POST /pet-conversations 请求体。
type CreateConversationInput struct {
	PetID int64 `json:"pet_id" binding:"required,gt=0"`
}

// SendMessageInput POST /pet-messages 请求体。
// client_msg_id 由客户端生成（UUID），服务端以此幂等；缺省时服务端不保证重放安全。
type SendMessageInput struct {
	PetID       int64   `json:"pet_id" binding:"required,gt=0"`
	ClientMsgID *string `json:"client_msg_id" binding:"omitempty,len=36"`
	Content     string  `json:"content" binding:"required"`
}

// SendMessageResult POST /pet-messages 响应 data。
// Streaming 为协议预留位：本变更恒为 false（流式留待 protocol-v2）。
type SendMessageResult struct {
	ConversationID int64       `json:"conversation_id"`
	UserMessage    MessageItem `json:"user_message"`
	// AssistantMessage 为 nil 表示回复尚未生成（幂等重放时可能出现），客户端按待回复处理。
	AssistantMessage *MessageItem `json:"assistant_message,omitempty"`
	// NewMemories 本轮新形成的长期记忆（聊天页轻量提示用）。
	NewMemories []MemoryItem `json:"new_memories,omitempty"`
	Streaming   bool         `json:"streaming"`
	Usage       UsageSummary `json:"usage"`
}

// UsageSummary 单次调用的用量摘要（guide §5：token/模型/耗时/缓存命中）。
type UsageSummary struct {
	Provider     string `json:"provider"`
	Model        string `json:"model"`
	PromptVer    string `json:"prompt_version"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	LatencyMS    int    `json:"latency_ms"`
	CacheHit     bool   `json:"cache_hit"`
	ErrCode      int    `json:"err_code"`
}

// MessagePage GET /pet-messages 响应 data（游标分页，禁止 offset）。
type MessagePage struct {
	Items      []MessageItem `json:"items"`
	NextCursor int64         `json:"next_cursor"`
	HasMore    bool          `json:"has_more"`
}

// ToMessageItem 实体 → 响应 DTO。
func ToMessageItem(m *PetMessage) MessageItem {
	return MessageItem{
		ID:           m.ID,
		ConvID:       m.ConvID,
		Role:         m.Role,
		Content:      m.Content,
		ClientMsgID:  m.ClientMsgID,
		Status:       m.Status,
		ErrorCode:    m.ErrorCode,
		Model:        m.Model,
		InputTokens:  m.InputTokens,
		OutputTokens: m.OutputTokens,
		LatencyMS:    m.LatencyMS,
		CreatedAt:    m.CreatedAt,
	}
}

// ToConversationItem 会话实体 → 响应 DTO。
func ToConversationItem(c *PetConversation) ConversationItem {
	return ConversationItem{
		ID:             c.ID,
		UserID:         c.UserID,
		PetID:          c.PetID,
		LastMsgID:      c.LastMsgID,
		LastMsgPreview: c.LastMsgPreview,
		UpdatedAt:      c.UpdatedAt.Unix(),
	}
}

// MessageItems 批量转换（保持顺序）。
func MessageItems(rows []PetMessage) []MessageItem {
	items := make([]MessageItem, 0, len(rows))
	for i := range rows {
		items = append(items, ToMessageItem(&rows[i]))
	}
	return items
}
