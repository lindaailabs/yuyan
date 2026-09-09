package service

import (
	"context"
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"

	"github.com/lindaailabs/yuyan/server/internal/pkg/errcode"
	"github.com/lindaailabs/yuyan/server/internal/pkg/jwt"
	"github.com/lindaailabs/yuyan/server/internal/repo"
)

// phoneRe 中国大陆手机号：11 位、1 开头、第二位 3-9。
var phoneRe = regexp.MustCompile(`^1[3-9]\d{9}$`)

const (
	minPasswordLen = 6
	maxPasswordLen = 64
)

// Auth 业务错误码（2xxx 段）。
var (
	ErrInvalidPhone    = errcode.New(1001, "手机号格式不正确")
	ErrInvalidPassword = errcode.New(1001, "密码需为 6~64 位字符")
	ErrPhoneRegistered = errcode.New(2004, "该手机号已注册，请直接登录")
	ErrCredential      = errcode.New(2005, "手机号或密码错误")
	ErrTokenInvalid    = errcode.New(1002, "登录状态无效，请重新登录")
)

// AuthService 账号/认证业务逻辑（api → service → repo 分层，事务仅发生在本层）。
type AuthService struct {
	users *repo.UserRepo
	jwt   *jwt.Manager
}

// NewAuthService 构造。
func NewAuthService(users *repo.UserRepo, jwtMgr *jwt.Manager) *AuthService {
	return &AuthService{users: users, jwt: jwtMgr}
}

// LoginResponse 注册/登录/刷新响应（同一结构）。
type LoginResponse struct {
	jwt.TokenPair
}

// Register 手机号+密码注册（不存在才可注册），成功后直接签发双 token。
func (s *AuthService) Register(ctx context.Context, phone, password string) (*LoginResponse, error) {
	if !phoneRe.MatchString(phone) {
		return nil, ErrInvalidPhone
	}
	if l := len(password); l < minPasswordLen || l > maxPasswordLen {
		return nil, ErrInvalidPassword
	}
	// 已注册：引导走登录，避免重复创建。
	if _, err := s.users.FindByPhone(ctx, phone); err == nil {
		return nil, ErrPhoneRegistered
	} else if !repo.IsNotFound(err) {
		return nil, fmt.Errorf("register find user: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	u, err := s.users.CreateUserWithPassword(ctx, phone, string(hash))
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	pair, err := s.jwt.Issue(u.ID)
	if err != nil {
		return nil, fmt.Errorf("issue token: %w", err)
	}
	return &LoginResponse{TokenPair: pair}, nil
}

// Login 手机号+密码登录：校验密码哈希 → 签发双 token。
func (s *AuthService) Login(ctx context.Context, phone, password string) (*LoginResponse, error) {
	if !phoneRe.MatchString(phone) {
		return nil, ErrInvalidPhone
	}
	if password == "" {
		return nil, errcode.New(errcode.ErrInvalidParam, "密码不能为空")
	}

	u, err := s.users.FindByPhone(ctx, phone)
	if repo.IsNotFound(err) {
		return nil, ErrCredential
	}
	if err != nil {
		return nil, fmt.Errorf("login find user: %w", err)
	}
	// 旧验证码注册用户无密码哈希：统一返回凭证错误（一期数据，生产可走重置流程）。
	if u.PasswordHash == "" {
		return nil, ErrCredential
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, ErrCredential
	}

	pair, err := s.jwt.Issue(u.ID)
	if err != nil {
		return nil, fmt.Errorf("issue token: %w", err)
	}
	return &LoginResponse{TokenPair: pair}, nil
}

// Refresh 用 refresh token 换发新双 token。
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	if refreshToken == "" {
		return nil, errcode.New(errcode.ErrInvalidParam, "refresh_token 不能为空")
	}
	uid, err := s.jwt.Parse(refreshToken, jwt.TypeRefresh)
	if err != nil {
		return nil, ErrTokenInvalid
	}
	// 确认用户仍存在（防止已注销用户续签）。
	if _, err := s.users.FindByID(ctx, uid); err != nil {
		if repo.IsNotFound(err) {
			return nil, ErrTokenInvalid
		}
		return nil, fmt.Errorf("refresh find user: %w", err)
	}
	pair, err := s.jwt.Issue(uid)
	if err != nil {
		return nil, fmt.Errorf("issue token: %w", err)
	}
	return &LoginResponse{TokenPair: pair}, nil
}
