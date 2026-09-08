package service

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/lindaailabs/yuyan/server/internal/model"
	"github.com/lindaailabs/yuyan/server/internal/pkg/errcode"
	"github.com/lindaailabs/yuyan/server/internal/repo"
)

// User 业务错误码（2xxx 段）。
var (
	ErrNicknameInvalid = errcode.New(2011, "昵称须为 1~20 个字符")
	ErrAvatarInvalid   = errcode.New(2012, "头像不存在")
	ErrUserNotFound    = errcode.New(2013, "用户不存在")
)

// 预置头像数量（apps/app assets avatar_1..8）。
const maxAvatarID = 8

// UserService 用户资料业务逻辑（api → service → repo 分层，事务仅发生在本层）。
type UserService struct {
	users *repo.UserRepo
}

// NewUserService 构造。
func NewUserService(users *repo.UserRepo) *UserService {
	return &UserService{users: users}
}

// Profile 查询用户资料；不存在返回 ErrUserNotFound（2xxx）。
func (s *UserService) Profile(ctx context.Context, uid int64) (*model.UserProfile, error) {
	u, err := s.users.FindByID(ctx, uid)
	if repo.IsNotFound(err) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("profile find user: %w", err)
	}
	return toUserProfile(u), nil
}

// UpdateProfile 选择性更新（nil 字段跳过）并返回更新后的资料。
// 业务校验：nickname 1~20 个字符（utf8mb4 按字符计）、avatar_id 1~8。
func (s *UserService) UpdateProfile(ctx context.Context, uid int64, in *model.UpdateProfileInput) (*model.UserProfile, error) {
	if in == nil {
		return nil, errcode.New(errcode.ErrInvalidParam, "请求体不能为空")
	}
	if in.Nickname != nil {
		if n := utf8.RuneCountInString(*in.Nickname); n < 1 || n > 20 {
			return nil, ErrNicknameInvalid
		}
	}
	if in.AvatarID != nil && (*in.AvatarID < 1 || *in.AvatarID > maxAvatarID) {
		return nil, ErrAvatarInvalid
	}

	if err := s.users.UpdateProfile(ctx, uid, in); err != nil {
		if repo.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("update profile: %w", err)
	}
	return s.Profile(ctx, uid)
}

// Search 按手机号精确搜索：命中返回脱敏列表（单元素），未注册返回空列表。
// 脱敏不改变存在性判断——q 本身就是完整手机号，调用方已知晓号码。
func (s *UserService) Search(ctx context.Context, q string) ([]model.UserSearchItem, error) {
	if !phoneRe.MatchString(q) {
		return nil, ErrInvalidPhone
	}
	u, err := s.users.FindByPhone(ctx, q)
	if repo.IsNotFound(err) {
		return []model.UserSearchItem{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("search find user: %w", err)
	}
	return []model.UserSearchItem{{
		ID:       u.ID,
		Nickname: u.Nickname,
		AvatarID: u.AvatarID,
		Phone:    maskPhone(u.Phone),
	}}, nil
}

// maskPhone 手机号脱敏：保留前 3 后 4，中间 4 位以 **** 替换（138****8000）。
func maskPhone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}

// toUserProfile 实体 → 资料 DTO。
func toUserProfile(u *model.User) *model.UserProfile {
	return &model.UserProfile{
		ID:        u.ID,
		Phone:     u.Phone,
		Nickname:  u.Nickname,
		AvatarID:  u.AvatarID,
		CreatedAt: u.CreatedAt.Unix(),
	}
}
