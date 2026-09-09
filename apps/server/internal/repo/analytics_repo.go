package repo

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/lindaailabs/yuyan/server/internal/model"
)

// EventRepo 事件日志与支付订单的数据访问。
type EventRepo struct {
	db *gorm.DB
}

func NewEventRepo(db *gorm.DB) *EventRepo {
	return &EventRepo{db: db}
}

// InsertEvents 批量写入事件（Ignore 冲突，保证写入不阻断业务）。
func (r *EventRepo) InsertEvents(ctx context.Context, rows []model.EventLog) error {
	if len(rows) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&rows)
	if res.Error != nil {
		return fmt.Errorf("insert events: %w", res.Error)
	}
	return nil
}

// ListEvents 按事件名倒序查询（内部观察出口）。
func (r *EventRepo) ListEvents(ctx context.Context, name string, limit int) ([]model.EventLog, error) {
	q := r.db.WithContext(ctx).Order("id DESC")
	if name != "" {
		q = q.Where("name = ?", name)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	var rows []model.EventLog
	if err := q.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	return rows, nil
}

// PaymentOrderRepo 支付订单（订单号唯一键保证回调幂等）。
type PaymentOrderRepo struct {
	db *gorm.DB
}

func NewPaymentOrderRepo(db *gorm.DB) *PaymentOrderRepo {
	return &PaymentOrderRepo{db: db}
}

// Create 写入订单；order_no 冲突表示重复回调，返回已有订单。
func (r *PaymentOrderRepo) Create(ctx context.Context, order *model.PaymentOrder) (bool, error) {
	res := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(order)
	if res.Error != nil {
		return false, fmt.Errorf("create payment order: %w", res.Error)
	}
	return res.RowsAffected == 1, nil
}

// FindByOrderNo 按订单号查询（幂等重放时回读）。
func (r *PaymentOrderRepo) FindByOrderNo(ctx context.Context, orderNo string) (*model.PaymentOrder, error) {
	var order model.PaymentOrder
	if err := r.db.WithContext(ctx).Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}
