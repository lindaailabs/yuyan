package repo

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/lindaailabs/yuyan/server/internal/model"
)

// UserRepo 用户数据访问（GORM，仅 CRUD；复杂查询才写原生 SQL，本域无）。
type UserRepo struct {
	db *gorm.DB
}

// NewUserRepo 构造。
func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

// FindByPhone 按手机号精确查找；不存在返回 gorm.ErrRecordNotFound。
func (r *UserRepo) FindByPhone(ctx context.Context, phone string) (*model.User, error) {
	var u model.User
	if err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// FindByID 按 id 查找；不存在返回 gorm.ErrRecordNotFound。
func (r *UserRepo) FindByID(ctx context.Context, id int64) (*model.User, error) {
	var u model.User
	if err := r.db.WithContext(ctx).First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// FindByIDs 批量查找（好友/申请列表组装资料用）；空入参返回空切片。
// 缺失的 id 静默跳过（ friendships 行存在而用户行缺失属脏数据，不阻断列表）。
func (r *UserRepo) FindByIDs(ctx context.Context, ids []int64) ([]model.User, error) {
	if len(ids) == 0 {
		return []model.User{}, nil
	}
	var users []model.User
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("find users by ids: %w", err)
	}
	return users, nil
}

// CreateUser 创建用户（登录自动注册路径，nickname 为 NULL、avatar_id 默认 1）。
// phone 唯一冲突由 DB 唯一索引兜底，调用方（service）先查后插，竞态时返回错误。
func (r *UserRepo) CreateUser(ctx context.Context, phone string) (*model.User, error) {
	u := model.User{Phone: phone, AvatarID: 1, Nickname: nil}
	if err := r.db.WithContext(ctx).Create(&u).Error; err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &u, nil
}

// UpdateProfile 选择性更新（nil 字段跳过）。
func (r *UserRepo) UpdateProfile(ctx context.Context, id int64, in *model.UpdateProfileInput) error {
	updates := map[string]any{}
	if in.Nickname != nil {
		updates["nickname"] = *in.Nickname
	}
	if in.AvatarID != nil {
		updates["avatar_id"] = *in.AvatarID
	}
	if len(updates) == 0 {
		return nil
	}
	res := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return fmt.Errorf("update profile: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ErrNotFound 统一的"未找到"哨兵（service 层判断用）。
var ErrNotFound = gorm.ErrRecordNotFound

// IsNotFound 判断 err 是否为"记录不存在"。
func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
