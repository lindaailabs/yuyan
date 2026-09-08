package model

import "time"

// 好友关系状态常量（guide §4：SMALLINT 而非 ENUM，避免 ALTER TABLE 问题）。
const (
	FriendshipPending  int16 = 1 // 待处理
	FriendshipAccepted int16 = 2 // 已是好友
	FriendshipBlocked  int16 = 3 // 拉黑（一期无入口，保留常量）
	FriendshipRejected int16 = 4 // 已拒绝
)

// Friendship 好友关系实体（friendships 表映射，guide §4）。
// 关系双向各存一行：A↔B 好友 = (A,B) 与 (B,A) 两行均为 accepted。
type Friendship struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID    int64     `gorm:"column:user_id;not null"`
	FriendID  int64     `gorm:"column:friend_id;not null"`
	Status    int16     `gorm:"column:status;not null;default:1"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`

	// 申请人/好友资料（JOIN users 查询时填充，非表字段）。
	Friend *User `gorm:"-"`
}

// TableName GORM 表名。
func (Friendship) TableName() string { return "friendships" }

// FriendshipRequestItem 申请列表项 DTO（GET /friends/requests）。
type FriendshipRequestItem struct {
	ID        int64          `json:"id"`         // 申请行 id（accept/reject 路径参数）
	FromUser  UserSearchItem `json:"from_user"`  // 申请人资料（手机号脱敏）
	CreatedAt int64          `json:"created_at"` // Unix 秒
}

// FriendshipItem 好友列表项 DTO（GET /friends）。
type FriendshipItem struct {
	User      UserSearchItem `json:"user"`       // 好友资料（手机号脱敏）
	CreatedAt int64          `json:"created_at"` // 结交时间，Unix 秒
}

// FriendRequestResult 发起申请响应 DTO（POST /friends/requests）。
type FriendRequestResult struct {
	ID        int64 `json:"id"`
	UserID    int64 `json:"user_id"`   // 申请人
	FriendID  int64 `json:"friend_id"` // 被申请人
	Status    int16 `json:"status"`
	CreatedAt int64 `json:"created_at"` // Unix 秒
}

// FriendRequestInput POST /friends/requests 请求体。
// 目标用户不存在/为本人等由 service 层校验（2xxx 业务错误），故只拦缺失/类型错误。
type FriendRequestInput struct {
	UserID int64 `json:"user_id" binding:"required,gt=0"`
}

// FriendOpResult accept/reject 响应 DTO（POST /friends/requests/{id}/accept|reject）。
type FriendOpResult struct {
	ID     int64 `json:"id"`
	Status int16 `json:"status"`
}
