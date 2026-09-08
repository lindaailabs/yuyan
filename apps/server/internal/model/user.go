package model

import "time"

// User 用户实体（users 表映射，guide §4）。
// 昵称可空：NULL 表示未完成首登引导（design D3）。
type User struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Phone     string    `gorm:"column:phone;uniqueIndex;size:20;not null"`
	Nickname  *string   `gorm:"column:nickname;size:20"` // 指针表达 NULL
	AvatarID  int16     `gorm:"column:avatar_id;not null;default:1"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName GORM 表名。
func (User) TableName() string { return "users" }

// UserProfile 用户资料 DTO（GET/PUT /users/me 与搜索结果共用形态）。
type UserProfile struct {
	ID        int64   `json:"id"`
	Phone     string  `json:"phone"`
	Nickname  *string `json:"nickname"`
	AvatarID  int16   `json:"avatar_id"`
	CreatedAt int64   `json:"created_at"` // Unix 秒
}

// UserSearchItem 搜索结果 DTO（手机号脱敏）。
type UserSearchItem struct {
	ID       int64   `json:"id"`
	Nickname *string `json:"nickname"`
	AvatarID int16   `json:"avatar_id"`
	Phone    string  `json:"phone"` // 形如 138****8000
}

// UpdateProfileInput PUT /users/me 请求体（指针表达"仅更新出现的字段"）。
// 越界值须返回 2xxx 业务错误（user_service 校验），故不带 min/max binding——binding 只拦缺失/类型错误。
type UpdateProfileInput struct {
	Nickname *string `json:"nickname"`
	AvatarID *int16  `json:"avatar_id"`
}
