package repo

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/lindaailabs/yuyan/server/internal/model"
)

// AICallLogRepo AI 调用日志数据访问（ai_call_logs 表，guide §5 用量统计）。
type AICallLogRepo struct {
	db *gorm.DB
}

func NewAICallLogRepo(db *gorm.DB) *AICallLogRepo {
	return &AICallLogRepo{db: db}
}

// Create 写入一条调用日志（成功与失败都写）。
func (r *AICallLogRepo) Create(ctx context.Context, log *model.AICallLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return fmt.Errorf("insert ai call log: %w", err)
	}
	return nil
}

// FindByMsgID 按消息 id 查调用日志（幂等重放时补齐用量字段）。
func (r *AICallLogRepo) FindByMsgID(ctx context.Context, msgID int64) (*model.AICallLog, error) {
	var log model.AICallLog
	if err := r.db.WithContext(ctx).Where("msg_id = ?", msgID).Order("id DESC").First(&log).Error; err != nil {
		return nil, err
	}
	return &log, nil
}
