package repo

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/lindaailabs/yuyan/server/internal/model"
)

// MemoryRepo 长期记忆数据访问（pet_memories 表）。
type MemoryRepo struct {
	db *gorm.DB
}

func NewMemoryRepo(db *gorm.DB) *MemoryRepo {
	return &MemoryRepo{db: db}
}

// Upsert 写入记忆：同一 (pet_id, content_hash) 已存在时更新置信度与来源并复活为 active，
// 返回 inserted=true 表示本次是新增（用于「新形成的记忆」提示）。
func (r *MemoryRepo) Upsert(ctx context.Context, m *model.PetMemory) (bool, error) {
	var existing model.PetMemory
	err := r.db.WithContext(ctx).
		Where("pet_id = ? AND content_hash = ?", m.PetID, m.ContentHash).
		First(&existing).Error
	switch {
	case err == nil:
		updates := map[string]any{
			"confidence":    m.Confidence,
			"source_msg_id": m.SourceMsgID,
			"status":        model.MemoryStatusActive, // 再次表达视为重新授权
			"updated_at":    time.Now(),
		}
		if res := r.db.WithContext(ctx).Model(&model.PetMemory{}).
			Where("id = ?", existing.ID).Updates(updates); res.Error != nil {
			return false, fmt.Errorf("update memory: %w", res.Error)
		}
		m.ID = existing.ID
		return false, nil
	case !IsNotFound(err):
		return false, fmt.Errorf("find memory by hash: %w", err)
	}

	res := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(m)
	if res.Error != nil {
		return false, fmt.Errorf("insert memory: %w", res.Error)
	}
	return res.RowsAffected > 0, nil
}

// ListActive 宠物的 active 记忆（最近更新在前）。
func (r *MemoryRepo) ListActive(ctx context.Context, petID int64, limit int) ([]model.PetMemory, error) {
	var rows []model.PetMemory
	q := r.db.WithContext(ctx).
		Where("pet_id = ? AND status = ?", petID, model.MemoryStatusActive).
		Order("updated_at DESC, id DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list memories: %w", err)
	}
	return rows, nil
}

// FindByUserAndID 按归属读取（越权与不存在同返回 gorm.ErrRecordNotFound）。
func (r *MemoryRepo) FindByUserAndID(ctx context.Context, userID, memoryID int64) (*model.PetMemory, error) {
	var m model.PetMemory
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND id = ?", userID, memoryID).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// MarkDeleted 软删除（status=3），返回影响行数。
func (r *MemoryRepo) MarkDeleted(ctx context.Context, userID, memoryID int64) (int64, error) {
	res := r.db.WithContext(ctx).Model(&model.PetMemory{}).
		Where("user_id = ? AND id = ? AND status <> ?", userID, memoryID, model.MemoryStatusDeleted).
		Update("status", model.MemoryStatusDeleted)
	if res.Error != nil {
		return 0, fmt.Errorf("delete memory: %w", res.Error)
	}
	return res.RowsAffected, nil
}

// TouchUsed 更新最近使用时间（召回后置位，供后续排序）。
func (r *MemoryRepo) TouchUsed(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).Model(&model.PetMemory{}).
		Where("id IN ?", ids).Update("last_used_at", time.Now())
	if res.Error != nil {
		return fmt.Errorf("touch memory used: %w", res.Error)
	}
	return nil
}
