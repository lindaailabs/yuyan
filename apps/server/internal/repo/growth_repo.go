package repo

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/lindaailabs/yuyan/server/internal/model"
)

// ErrGrowthEventDuplicated 同一来源消息的同类事件已存在（幂等重放）。
var ErrGrowthEventDuplicated = errors.New("repo: 成长事件已存在")

// GrowthRepo 成长事件与每日统计的数据访问。
type GrowthRepo struct {
	db *gorm.DB
}

func NewGrowthRepo(db *gorm.DB) *GrowthRepo {
	return &GrowthRepo{db: db}
}

// InsertEvent 写入成长事件；唯一键冲突返回 ErrGrowthEventDuplicated（幂等闸门）。
func (r *GrowthRepo) InsertEvent(ctx context.Context, e *model.PetGrowthEvent) error {
	res := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(e)
	if res.Error != nil {
		return fmt.Errorf("insert growth event: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrGrowthEventDuplicated
	}
	return nil
}

// ListEvents 按宠物倒序查询成长事件（时间线）。
func (r *GrowthRepo) ListEvents(ctx context.Context, petID int64, limit int) ([]model.PetGrowthEvent, error) {
	var rows []model.PetGrowthEvent
	q := r.db.WithContext(ctx).Where("pet_id = ?", petID).Order("id DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list growth events: %w", err)
	}
	return rows, nil
}

// FindStat 读取每日统计（不存在返回 gorm.ErrRecordNotFound）。
func (r *GrowthRepo) FindStat(ctx context.Context, petID int64) (*model.PetDailyStat, error) {
	var stat model.PetDailyStat
	if err := r.db.WithContext(ctx).Where("pet_id = ?", petID).First(&stat).Error; err != nil {
		return nil, err
	}
	return &stat, nil
}

// UpsertStat 写入或更新每日统计。
func (r *GrowthRepo) UpsertStat(ctx context.Context, stat *model.PetDailyStat) error {
	res := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "pet_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"last_date", "streak_days", "total_messages", "updated_at"}),
	}).Create(stat)
	if res.Error != nil {
		return fmt.Errorf("upsert daily stat: %w", res.Error)
	}
	return nil
}
