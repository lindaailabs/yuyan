package repo

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/lindaailabs/yuyan/server/internal/model"
)

// EntitlementRepo 权益数据访问。
type EntitlementRepo struct {
	db *gorm.DB
}

func NewEntitlementRepo(db *gorm.DB) *EntitlementRepo {
	return &EntitlementRepo{db: db}
}

// FindByUser 读取权益（不存在返回 gorm.ErrRecordNotFound）。
func (r *EntitlementRepo) FindByUser(ctx context.Context, userID int64) (*model.Entitlement, error) {
	var e model.Entitlement
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&e).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

// Ensure 无记录时创建（默认免费权益），已存在则原样返回。
func (r *EntitlementRepo) Ensure(ctx context.Context, e *model.Entitlement) (*model.Entitlement, error) {
	res := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(e)
	if res.Error != nil {
		return nil, fmt.Errorf("ensure entitlement: %w", res.Error)
	}
	found, err := r.FindByUser(ctx, e.UserID)
	if err != nil {
		return nil, fmt.Errorf("reload entitlement: %w", err)
	}
	return found, nil
}

// UpdatePlan 更新套餐与续期时间（支付成功/沙盒开通）。renewAt 为绝对时间（DATETIME 列）。
func (r *EntitlementRepo) UpdatePlan(ctx context.Context, userID int64, plan string, status int16, renewAt *time.Time) error {
	updates := map[string]any{"plan": plan, "status": status}
	if renewAt != nil {
		updates["renew_at"] = *renewAt
	}
	res := r.db.WithContext(ctx).Model(&model.Entitlement{}).Where("user_id = ?", userID).Updates(updates)
	if res.Error != nil {
		return fmt.Errorf("update entitlement plan: %w", res.Error)
	}
	return nil
}

// UsageRepo 用量计数器（额度扣减）。
type UsageRepo struct {
	db *gorm.DB
}

func NewUsageRepo(db *gorm.DB) *UsageRepo {
	return &UsageRepo{db: db}
}

// FindCounter 读取计数器（不存在返回 gorm.ErrRecordNotFound）。
func (r *UsageRepo) FindCounter(ctx context.Context, userID int64, action, period string) (*model.UsageCounter, error) {
	var c model.UsageCounter
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND action = ? AND period = ?", userID, action, period).
		First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// EnsureCounter 无记录时插入 0 值（并发下冲突忽略）。
func (r *UsageRepo) EnsureCounter(ctx context.Context, userID int64, action, period string) error {
	c := model.UsageCounter{UserID: userID, Action: action, Period: period, Used: 0}
	res := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&c)
	if res.Error != nil {
		return fmt.Errorf("ensure usage counter: %w", res.Error)
	}
	return nil
}

// Consume 条件扣减：仅当 used+n <= limit 时才累加，避免并发超卖。
// 返回 false 表示已达上限（未扣减）。
func (r *UsageRepo) Consume(ctx context.Context, userID int64, action, period string, n, limit int) (bool, error) {
	res := r.db.WithContext(ctx).
		Model(&model.UsageCounter{}).
		Where("user_id = ? AND action = ? AND period = ? AND used + ? <= ?", userID, action, period, n, limit).
		Update("used", gorm.Expr("used + ?", n))
	if res.Error != nil {
		return false, fmt.Errorf("consume usage: %w", res.Error)
	}
	return res.RowsAffected == 1, nil
}
