package model

import "time"

// 记忆类型（pet_memories.memory_type）。
const (
	MemoryTypePreference = "preference" // 偏好（喜欢/讨厌）
	MemoryTypeProfile    = "profile"    // 用户画像（称呼、年龄、城市、职业）
	MemoryTypeEvent      = "event"      // 重要事件
	MemoryTypeRelation   = "relation"   // 宠物关系进展
)

// 记忆状态（status SMALLINT，禁止物理删除）。
const (
	MemoryStatusActive   int16 = 1
	MemoryStatusArchived int16 = 2
	MemoryStatusDeleted  int16 = 3
)

// PetMemory 长期记忆（pet_memories 表）。
type PetMemory struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement"`
	UserID      int64      `gorm:"column:user_id;not null"`
	PetID       int64      `gorm:"column:pet_id;not null"`
	MemoryType  string     `gorm:"column:memory_type;size:32;not null"`
	Content     string     `gorm:"column:content;not null"`
	ContentHash string     `gorm:"column:content_hash;size:64;not null"`
	SourceMsgID *int64     `gorm:"column:source_msg_id"`
	Confidence  float64    `gorm:"column:confidence;not null;default:0.5"`
	LastUsedAt  *time.Time `gorm:"column:last_used_at"`
	Status      int16      `gorm:"column:status;not null;default:1"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (PetMemory) TableName() string { return "pet_memories" }

// MemoryItem 记忆响应 DTO（列表与新增提示同构）。
type MemoryItem struct {
	ID          int64   `json:"id"`
	PetID       int64   `json:"pet_id"`
	Type        string  `json:"type"`
	Content     string  `json:"content"`
	Confidence  float64 `json:"confidence"`
	CreatedAt   int64   `json:"created_at"`
	LastUsedAt  *int64  `json:"last_used_at,omitempty"`
	SourceMsgID *int64  `json:"source_msg_id,omitempty"`
}

// ToMemoryItem 实体 → 响应 DTO。
func ToMemoryItem(m *PetMemory) MemoryItem {
	item := MemoryItem{
		ID:         m.ID,
		PetID:      m.PetID,
		Type:       m.MemoryType,
		Content:    m.Content,
		Confidence: m.Confidence,
		CreatedAt:  m.CreatedAt.Unix(),
	}
	if m.LastUsedAt != nil {
		unix := m.LastUsedAt.Unix()
		item.LastUsedAt = &unix
	}
	if m.SourceMsgID != nil {
		item.SourceMsgID = m.SourceMsgID
	}
	return item
}

// MemoryItems 批量转换（保持顺序）。
func MemoryItems(rows []PetMemory) []MemoryItem {
	items := make([]MemoryItem, 0, len(rows))
	for i := range rows {
		items = append(items, ToMemoryItem(&rows[i]))
	}
	return items
}
