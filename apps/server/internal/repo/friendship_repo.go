package repo

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/lindaailabs/yuyan/server/internal/model"
)

// FriendshipRepo 好友关系数据访问（GORM）。
// 事务由 service 层编排：Transaction 内拿到的 txRepo 绑定事务连接。
type FriendshipRepo struct {
	db *gorm.DB
}

// NewFriendshipRepo 构造。
func NewFriendshipRepo(db *gorm.DB) *FriendshipRepo {
	return &FriendshipRepo{db: db}
}

// Transaction 在单事务内执行 fn（fn 内使用绑定事务的 txRepo）。
// fn 返回错误（含 errcode 业务错误）时整体回滚——service 的原子性依赖此语义。
func (r *FriendshipRepo) Transaction(ctx context.Context, fn func(txRepo *FriendshipRepo) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&FriendshipRepo{db: tx})
	})
}

// FindPairForUpdate 行锁读取双向两行（me→target 与 target→me），缺行为 nil 而非错误。
// OR 条件经 uk_user_friend 索引扫描，锁顺序由索引序决定，两方向并发申请不会交叉死锁。
func (r *FriendshipRepo) FindPairForUpdate(ctx context.Context, me, target int64) (forward, reverse *model.Friendship, err error) {
	var rows []model.Friendship
	q := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)", me, target, target, me)
	if err := q.Find(&rows).Error; err != nil {
		return nil, nil, fmt.Errorf("find pair for update: %w", err)
	}
	for i := range rows {
		if rows[i].UserID == me {
			forward = &rows[i]
		} else {
			reverse = &rows[i]
		}
	}
	return forward, reverse, nil
}

// FindByIDForUpdate 行锁读取申请行（accept/reject 用）；不存在返回 gorm.ErrRecordNotFound。
func (r *FriendshipRepo) FindByIDForUpdate(ctx context.Context, id int64) (*model.Friendship, error) {
	var f model.Friendship
	q := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id)
	if err := q.First(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

// InsertPending 插入 pending 行（发起申请/重新申请均可能调用）。
func (r *FriendshipRepo) InsertPending(ctx context.Context, me, target int64) (*model.Friendship, error) {
	f := model.Friendship{
		UserID:   me,
		FriendID: target,
		Status:   model.FriendshipPending,
	}
	if err := r.db.WithContext(ctx).Create(&f).Error; err != nil {
		return nil, fmt.Errorf("insert friendship: %w", err)
	}
	return &f, nil
}

// UpdateStatus 条件状态流转（from 不匹配则 0 行）：0 行视为"记录不存在或状态已变"。
func (r *FriendshipRepo) UpdateStatus(ctx context.Context, id int64, from, to int16) error {
	res := r.db.WithContext(ctx).Model(&model.Friendship{}).
		Where("id = ? AND status = ?", id, from).
		Update("status", to)
	if res.Error != nil {
		return fmt.Errorf("update friendship status: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ReactivatePending 被拒后重新申请：rejected → pending 并刷新 created_at（申请时间语义）。
func (r *FriendshipRepo) ReactivatePending(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Model(&model.Friendship{}).
		Where("id = ? AND status = ?", id, model.FriendshipRejected).
		Updates(map[string]any{"status": model.FriendshipPending, "created_at": time.Now()})
	if res.Error != nil {
		return fmt.Errorf("reactivate friendship: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpsertAccepted 写 accepted 行：不存在则插入，已存在（历史 rejected/pending）则覆盖为 accepted
// 并刷新 created_at（结交时间语义）。accept 的反向行复用它保证最终一致。
func (r *FriendshipRepo) UpsertAccepted(ctx context.Context, userID, friendID int64) error {
	f := model.Friendship{
		UserID:    userID,
		FriendID:  friendID,
		Status:    model.FriendshipAccepted,
		CreatedAt: time.Now(),
	}
	res := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "friend_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "created_at"}),
	}).Create(&f)
	if res.Error != nil {
		return fmt.Errorf("upsert accepted friendship: %w", res.Error)
	}
	return nil
}

// ListPendingRequests 我收到的 pending 申请（id 倒序，新的在前）。
func (r *FriendshipRepo) ListPendingRequests(ctx context.Context, friendID int64) ([]model.Friendship, error) {
	var rows []model.Friendship
	q := r.db.WithContext(ctx).
		Where("friend_id = ? AND status = ?", friendID, model.FriendshipPending).
		Order("id DESC")
	if err := q.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list pending requests: %w", err)
	}
	return rows, nil
}

// ListFriends 我的 accepted 好友行（id 升序 = 结交时间序）。
func (r *FriendshipRepo) ListFriends(ctx context.Context, userID int64) ([]model.Friendship, error) {
	var rows []model.Friendship
	q := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, model.FriendshipAccepted).
		Order("id ASC")
	if err := q.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list friends: %w", err)
	}
	return rows, nil
}
