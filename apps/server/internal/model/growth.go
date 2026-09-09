package model

import "time"

// 成长事件类型（pet_growth_events.event_type）。
const (
	GrowthEventMessage = "message"          // 一次互动
	GrowthEventLevelUp = "level_up"         // 升级
	GrowthEventMood    = "mood_change"      // 心情变化
	GrowthEventStreak  = "streak_milestone" // 连续互动里程碑
)

// 宠物心情（pets.mood，由确定性规则推导，模型不得决定）。
const (
	MoodHappy   = "happy"
	MoodCurious = "curious"
	MoodSleepy  = "sleepy"
	MoodLonely  = "lonely"
)

// PetGrowthEvent 成长事件（pet_growth_events 表）。
type PetGrowthEvent struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID      int64     `gorm:"column:user_id;not null"`
	PetID       int64     `gorm:"column:pet_id;not null"`
	EventType   string    `gorm:"column:event_type;size:32;not null"`
	Delta       *string   `gorm:"column:delta"`
	Reason      string    `gorm:"column:reason;size:255;not null;default:''"`
	SourceMsgID *int64    `gorm:"column:source_msg_id"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (PetGrowthEvent) TableName() string { return "pet_growth_events" }

// PetDailyStat 宠物每日统计（pet_daily_stats 表）。
type PetDailyStat struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement"`
	PetID         int64     `gorm:"column:pet_id;not null"`
	LastDate      string    `gorm:"column:last_date;size:10;not null"`
	StreakDays    int       `gorm:"column:streak_days;not null;default:0"`
	TotalMessages int       `gorm:"column:total_messages;not null;default:0"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (PetDailyStat) TableName() string { return "pet_daily_stats" }

// GrowthEventItem 成长事件响应 DTO（时间线展示）。
type GrowthEventItem struct {
	ID          int64   `json:"id"`
	PetID       int64   `json:"pet_id"`
	EventType   string  `json:"event_type"`
	Delta       *string `json:"delta,omitempty"`
	Reason      string  `json:"reason"`
	SourceMsgID *int64  `json:"source_msg_id,omitempty"`
	CreatedAt   int64   `json:"created_at"`
}

// GrowthState 结算后的宠物成长状态（供对话响应与状态查询复用）。
type GrowthState struct {
	PetID      int64  `json:"pet_id"`
	Level      int    `json:"level"`
	Intimacy   int    `json:"intimacy"`
	Mood       string `json:"mood"`
	StreakDays int    `json:"streak_days"`
}

// ToGrowthEventItem 实体 → 响应 DTO。
func ToGrowthEventItem(e *PetGrowthEvent) GrowthEventItem {
	return GrowthEventItem{
		ID:          e.ID,
		PetID:       e.PetID,
		EventType:   e.EventType,
		Delta:       e.Delta,
		Reason:      e.Reason,
		SourceMsgID: e.SourceMsgID,
		CreatedAt:   e.CreatedAt.Unix(),
	}
}

// GrowthEventItems 批量转换（保持顺序）。
func GrowthEventItems(rows []PetGrowthEvent) []GrowthEventItem {
	items := make([]GrowthEventItem, 0, len(rows))
	for i := range rows {
		items = append(items, ToGrowthEventItem(&rows[i]))
	}
	return items
}
