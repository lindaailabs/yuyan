package service

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"

	"github.com/mojocn/base64Captcha"

	"github.com/lindaailabs/yuyan/server/internal/pkg/errcode"
	"github.com/lindaailabs/yuyan/server/internal/pkg/jwt"
	"github.com/lindaailabs/yuyan/server/internal/repo"
)

// randomCode 生成 6 位数字验证码（图形验证码明文）。
func randomCode() string {
	return strconv.Itoa(100000 + rand.Intn(900000)) //nolint:gosec // 非安全用途（图形验证码）
}

// phoneRe 中国大陆手机号：11 位、1 开头、第二位 3-9。
var phoneRe = regexp.MustCompile(`^1[3-9]\d{9}$`)

// ErrInvalidPhone 手机号格式错误。
var ErrInvalidPhone = errcode.New(1001, "手机号格式不正确")

// Auth 业务错误码（2xxx 段）。
var (
	ErrCodeInvalid  = errcode.New(2003, "验证码错误或已过期")
	ErrTokenInvalid = errcode.New(1002, "登录状态无效，请重新登录")
)

// AuthService 账号/认证业务逻辑（api → service → repo 分层，事务仅发生在本层）。
type AuthService struct {
	captcha *repo.CaptchaRepo
	users   *repo.UserRepo
	jwt     *jwt.Manager
	// 验证码图片驱动（依赖注入便于测试替换）。
	captchaDriver base64Captcha.Driver
}

// NewAuthService 构造。
func NewAuthService(captcha *repo.CaptchaRepo, users *repo.UserRepo, jwtMgr *jwt.Manager) *AuthService {
	// 6 位数字图形验证码；宽 120 高 40，干扰适中（一期不对接短信时的登录凭证）。
	driver := base64Captcha.NewDriverDigit(40, 120, 6, 0.7, 20)
	return &AuthService{captcha: captcha, users: users, jwt: jwtMgr, captchaDriver: driver}
}

// SendSmsCodeResponse 下发验证码响应：一期携带图形验证码（base64 PNG）。
// 生产接入短信后 captcha_image 缺省（客户端两态兼容，见 MILESTONES W2）。
type SendSmsCodeResponse struct {
	CaptchaImage string `json:"captcha_image,omitempty"`
}

// SendSmsCode 生成 6 位数字验证码：存 Redis（5min TTL、限频 1/min）并渲染为图形验证码。
func (s *AuthService) SendSmsCode(ctx context.Context, phone string) (*SendSmsCodeResponse, error) {
	if !phoneRe.MatchString(phone) {
		return nil, ErrInvalidPhone
	}

	code := randomCode()

	if err := s.captcha.SetCode(ctx, phone, code); err != nil {
		return nil, err // 限频错误（2001）或 Redis 故障，直接上抛
	}

	// 将明文验证码渲染为图形验证码（base64 PNG）。
	item, err := s.captchaDriver.DrawCaptcha(code)
	if err != nil {
		return nil, fmt.Errorf("draw captcha: %w", err)
	}
	return &SendSmsCodeResponse{CaptchaImage: item.EncodeB64string()}, nil
}

// LoginResponse 登录/刷新响应。
type LoginResponse struct {
	jwt.TokenPair
}

// Login 验证码登录：校验（一次性）→ 不存在则自动注册 → 签发双 token。
func (s *AuthService) Login(ctx context.Context, phone, code string) (*LoginResponse, error) {
	if !phoneRe.MatchString(phone) {
		return nil, ErrInvalidPhone
	}
	if code == "" {
		return nil, errcode.New(errcode.ErrInvalidParam, "验证码不能为空")
	}

	if err := s.captcha.VerifyCode(ctx, phone, code); err != nil {
		return nil, err
	}

	// 自动注册：不存在则建（先查后插；唯一索引兜底竞态）。
	u, err := s.users.FindByPhone(ctx, phone)
	if repo.IsNotFound(err) {
		u, err = s.users.CreateUser(ctx, phone)
	}
	if err != nil {
		return nil, fmt.Errorf("login find/create user: %w", err)
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
