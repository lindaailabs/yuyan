package model

import "time"

// 权益套餐。
const (
	PlanFree = "free"
	PlanPro  = "pro"
)

// 权益状态。
const (
	EntitlementActive  int16 = 1
	EntitlementExpired int16 = 2
	EntitlementRevoked int16 = 3
)

// 支付订单状态。
const (
	PaymentStatusSuccess int16 = 1
	PaymentStatusFailed  int16 = 2
)

// 消耗型动作（usage_counters.action）。
const (
	ActionAIMessage = "ai_message"
)

// Entitlement 订阅权益（entitlements 表）。
type Entitlement struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement"`
	UserID    int64      `gorm:"column:user_id;not null"`
	Plan      string     `gorm:"column:plan;size:32;not null;default:free"`
	Status    int16      `gorm:"column:status;not null;default:1"`
	QuotaJSON *string    `gorm:"column:quota_json"`
	RenewAt   *time.Time `gorm:"column:renew_at"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (Entitlement) TableName() string { return "entitlements" }

// UsageCounter 用量计数器（usage_counters 表）。
type UsageCounter struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID    int64     `gorm:"column:user_id;not null"`
	Action    string    `gorm:"column:action;size:32;not null"`
	Period    string    `gorm:"column:period;size:10;not null"`
	Used      int       `gorm:"column:used;not null;default:0"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (UsageCounter) TableName() string { return "usage_counters" }

// EventLog 行为事件（event_logs 表）。
type EventLog struct {
	ID        int64   `gorm:"column:id;primaryKey;autoIncrement"`
	UserID    int64   `gorm:"column:user_id;not null;default:0"`
	Name      string  `gorm:"column:name;size:64;not null"`
	Props     *string `gorm:"column:props"`
	CreatedAt int64   `gorm:"column:created_at;not null"`
}

func (EventLog) TableName() string { return "event_logs" }

// PaymentOrder 支付订单（payment_orders 表）。
type PaymentOrder struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID    int64     `gorm:"column:user_id;not null"`
	OrderNo   string    `gorm:"column:order_no;size:64;not null"`
	Platform  string    `gorm:"column:platform;size:32;not null;default:sandbox"`
	Plan      string    `gorm:"column:plan;size:32;not null"`
	Status    int16     `gorm:"column:status;not null;default:1"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (PaymentOrder) TableName() string { return "payment_orders" }

// QuotaView 额度视图（响应用）。
type QuotaView struct {
	DailyMessages int  `json:"daily_messages"`
	DailyUsed     int  `json:"daily_used"`
	DailyRemain   int  `json:"daily_remain"`
	MemoryLimit   int  `json:"memory_limit"`
	AdvancedModel bool `json:"advanced_model"`
}

// EntitlementView 权益响应 DTO（客户端只展示，不作判断源）。
type EntitlementView struct {
	UserID    int64     `json:"user_id"`
	Plan      string    `json:"plan"`
	Status    int16     `json:"status"`
	Quota     QuotaView `json:"quota"`
	RenewAt   *int64    `json:"renew_at,omitempty"`
	UpdatedAt int64     `json:"updated_at"`
}

// SandboxPurchaseInput POST /entitlements/sandbox-purchase 请求体。
type SandboxPurchaseInput struct {
	Plan string `json:"plan" binding:"required,oneof=free pro"`
}

// PaymentCallbackInput POST /entitlements/payments/callback 请求体。
type PaymentCallbackInput struct {
	OrderNo  string `json:"order_no" binding:"required"`
	Platform string `json:"platform"`
	Plan     string `json:"plan" binding:"required,oneof=free pro"`
	Success  *bool  `json:"success"`
}

// EventInput 单条上报事件。
type EventInput struct {
	Name  string         `json:"name" binding:"required"`
	Props map[string]any `json:"props"`
	Ts    *int64         `json:"ts"`
}

// EventReportInput POST /events 请求体。
type EventReportInput struct {
	Events []EventInput `json:"events" binding:"required,min=1,max=50"`
}

// EventItem 事件响应 DTO（内部查询）。
type EventItem struct {
	ID        int64          `json:"id"`
	UserID    int64          `json:"user_id"`
	Name      string         `json:"name"`
	Props     map[string]any `json:"props,omitempty"`
	CreatedAt int64          `json:"created_at"`
}
