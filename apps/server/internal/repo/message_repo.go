package repo

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/lindaailabs/yuyan/server/internal/model"
)

// ErrDuplicateClientMsgID 同一 client_msg_id 重复提交（幂等重放）。
var ErrDuplicateClientMsgID = errors.New("repo: client_msg_id 重复")

// MessageRepo 消息数据访问（pet_messages 表；id 兼作增量游标）。
type MessageRepo struct {
	db *gorm.DB
}

func NewMessageRepo(db *gorm.DB) *MessageRepo {
	return &MessageRepo{db: db}
}

// Create 插入消息；uk_user_client_msg_id 冲突时返回 ErrDuplicateClientMsgID（不产生第二条）。
// 客户端未带 client_msg_id 时不参与幂等（MySQL 唯一键允许多个 NULL）。
func (r *MessageRepo) Create(ctx context.Context, m *model.PetMessage) error {
	if m.ClientMsgID == nil {
		if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
			return fmt.Errorf("insert message: %w", err)
		}
		return nil
	}
	res := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}, {Name: "client_msg_id"}}, DoNothing: true}).
		Create(m)
	if res.Error != nil {
		return fmt.Errorf("insert message: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrDuplicateClientMsgID
	}
	return nil
}

// FindByClientMsgID 按当前用户与会话查找 client_msg_id（幂等重放时取首次结果）。
func (r *MessageRepo) FindByClientMsgID(ctx context.Context, uid, convID int64, clientMsgID string) (*model.PetMessage, error) {
	var m model.PetMessage
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND conv_id = ? AND client_msg_id = ? AND role = ?", uid, convID, clientMsgID, model.MessageRoleUser).
		Order("id ASC").First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// FindReplyAfter 取指定消息之后的第一条 assistant 消息（幂等重放时补齐回复）。
func (r *MessageRepo) FindReplyAfter(ctx context.Context, convID, afterID int64) (*model.PetMessage, error) {
	var m model.PetMessage
	if err := r.db.WithContext(ctx).
		Where("conv_id = ? AND id > ? AND role = ?", convID, afterID, model.MessageRoleAssistant).
		Order("id ASC").First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// ListAfter 游标分页：id > cursor 升序取 limit+1 条（多取一条用于判断 has_more）。
// 禁止 offset 与时间戳排序（guide §12 红线）。
func (r *MessageRepo) ListAfter(ctx context.Context, convID, cursor int64, limit int) ([]model.PetMessage, error) {
	var rows []model.PetMessage
	if err := r.db.WithContext(ctx).
		Where("conv_id = ? AND id > ?", convID, cursor).
		Order("id ASC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	return rows, nil
}

// RecentTurns 取会话最近 limit 条消息（按 id 升序，供 prompt 拼装）。
func (r *MessageRepo) RecentTurns(ctx context.Context, convID int64, limit int) ([]model.PetMessage, error) {
	var rows []model.PetMessage
	if err := r.db.WithContext(ctx).
		Where("conv_id = ?", convID).
		Order("id DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("recent messages: %w", err)
	}
	// 反转为时间正序（旧 → 新）。
	for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
		rows[i], rows[j] = rows[j], rows[i]
	}
	return rows, nil
}
