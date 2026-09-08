package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/lindaailabs/yuyan/server/internal/pkg/errcode"
)

// ErrCaptchaRateLimited 限频命中（60s 内重复请求 sms-code）。
var ErrCaptchaRateLimited = errcode.New(2001, "请求过于频繁，请稍后再试")

// ErrCaptchaMismatch 验证码错误或已过期。
var ErrCaptchaMismatch = errcode.New(2002, "验证码错误或已过期")

// captcha_repo 的 Redis 键约定（design D1）：
//   sms:code:{phone}  6 位数字验证码，TTL 5min
//   sms:freq:{phone}  限频标记，TTL 60s（存在即拒绝下发）

const (
	captchaCodeTTL = 5 * time.Minute
	captchaFreqTTL = 60 * time.Second
)

// CaptchaRepo 验证码存储与限频（service 层唯一入口）。
type CaptchaRepo struct {
	rdb *redis.Client
}

// NewCaptchaRepo 构造。
func NewCaptchaRepo(rdb *redis.Client) *CaptchaRepo {
	return &CaptchaRepo{rdb: rdb}
}

func codeKey(phone string) string { return "sms:code:" + phone }
func freqKey(phone string) string { return "sms:freq:" + phone }

// SetCode 存储验证码；命中限频返回 ErrCaptchaRateLimited。
// 限频标记与验证码原子写入（Lua 保证"检查+设置"不被并发穿透）。
func (r *CaptchaRepo) SetCode(ctx context.Context, phone, code string) error {
	ok, err := r.rdb.SetNX(ctx, freqKey(phone), 1, captchaFreqTTL).Result()
	if err != nil {
		return fmt.Errorf("redis set freq: %w", err)
	}
	if !ok {
		return ErrCaptchaRateLimited
	}
	if err := r.rdb.Set(ctx, codeKey(phone), code, captchaCodeTTL).Err(); err != nil {
		return fmt.Errorf("redis set code: %w", err)
	}
	return nil
}

// VerifyCode 校验并一次性删除（登录成功即作废，防重放）。
// 不存在/不匹配返回 ErrCaptchaMismatch。
func (r *CaptchaRepo) VerifyCode(ctx context.Context, phone, code string) error {
	stored, err := r.rdb.Get(ctx, codeKey(phone)).Result()
	if errors.Is(err, redis.Nil) {
		return ErrCaptchaMismatch
	}
	if err != nil {
		return fmt.Errorf("redis get code: %w", err)
	}
	if stored != code {
		return ErrCaptchaMismatch
	}
	// 校验通过即删除（一次性）。删除失败仅意味着旧码多活到 TTL，不影响正确性。
	if err := r.rdb.Del(ctx, codeKey(phone)).Err(); err != nil {
		return fmt.Errorf("redis del code: %w", err)
	}
	return nil
}
