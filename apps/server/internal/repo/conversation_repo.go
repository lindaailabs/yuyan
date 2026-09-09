package repo

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/lindaailabs/yuyan/server/internal/model"
)

// ConversationRepo 会话数据访问（pet_conversations 表）。
type ConversationRepo struct {
	db *gorm.DB
}

func NewConversationRepo(db *gorm.DB) *ConversationRepo {
	return &ConversationRepo{db: db}
}

// GetOrCreate 获取或创建「用户 + 宠物」会话（uk_user_pet 保证唯一）。
func (r *ConversationRepo) GetOrCreate(ctx context.Context, userID, petID int64) (*model.PetConversation, error) {
	var conv model.PetConversation
	err := r.db.WithContext(ctx).Where("user_id = ? AND pet_id = ?", userID, petID).First(&conv).Error
	if err == nil {
		return &conv, nil
	}
	if !IsNotFound(err) {
		return nil, fmt.Errorf("find conversation: %w", err)
	}

	created := model.PetConversation{UserID: userID, PetID: petID}
	// 并发首条消息：唯一键冲突时什么都不做，再查一次即可拿到同一会话。
	res := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&created)
	if res.Error != nil {
		return nil, fmt.Errorf("create conversation: %w", res.Error)
	}
	if res.RowsAffected > 0 {
		if err := r.db.WithContext(ctx).Where("id = ?", created.ID).First(&conv).Error; err != nil {
			return nil, fmt.Errorf("reload conversation: %w", err)
		}
		return &conv, nil
	}
	if err := r.db.WithContext(ctx).Where("user_id = ? AND pet_id = ?", userID, petID).First(&conv).Error; err != nil {
		return nil, fmt.Errorf("reload conversation after conflict: %w", err)
	}
	return &conv, nil
}

// FindByUserAndID 按会话 id 读取，并校验归属（越权与不存在同返回 gorm.ErrRecordNotFound）。
func (r *ConversationRepo) FindByUserAndID(ctx context.Context, userID, convID int64) (*model.PetConversation, error) {
	var conv model.PetConversation
	if err := r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, convID).First(&conv).Error; err != nil {
		return nil, err
	}
	return &conv, nil
}

// UpdateLastMessage 更新会话的最后一条消息与预览。
func (r *ConversationRepo) UpdateLastMessage(ctx context.Context, convID, msgID int64, preview string) error {
	if len([]rune(preview)) > 255 {
		preview = string([]rune(preview)[:255])
	}
	res := r.db.WithContext(ctx).Model(&model.PetConversation{}).
		Where("id = ?", convID).
		Updates(map[string]any{"last_msg_id": msgID, "last_msg_preview": preview})
	if res.Error != nil {
		return fmt.Errorf("update conversation last message: %w", res.Error)
	}
	return nil
}
