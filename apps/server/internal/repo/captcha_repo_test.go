package repo

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/lindaailabs/yuyan/server/internal/pkg/errcode"
)

func newTestCaptchaRepo(t *testing.T) (*CaptchaRepo, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return NewCaptchaRepo(rdb), mr
}

func TestCaptchaSetAndVerify(t *testing.T) {
	repo, _ := newTestCaptchaRepo(t)
	ctx := context.Background()

	if err := repo.SetCode(ctx, "13800138000", "123456"); err != nil {
		t.Fatalf("SetCode: %v", err)
	}
	if err := repo.VerifyCode(ctx, "13800138000", "123456"); err != nil {
		t.Fatalf("VerifyCode 正确码: %v", err)
	}
}

func TestCaptchaRateLimit(t *testing.T) {
	repo, _ := newTestCaptchaRepo(t)
	ctx := context.Background()

	if err := repo.SetCode(ctx, "13800138000", "111111"); err != nil {
		t.Fatalf("首次 SetCode: %v", err)
	}
	err := repo.SetCode(ctx, "13800138000", "222222")
	if err == nil {
		t.Fatal("60s 内第二次 SetCode 应被限频")
	}
	if !isErrcode(err, 2001) {
		t.Errorf("限频错误码 = %v, want 2001", err)
	}
}

func TestCaptchaVerifyOneShot(t *testing.T) {
	repo, _ := newTestCaptchaRepo(t)
	ctx := context.Background()

	_ = repo.SetCode(ctx, "13800138000", "123456")
	if err := repo.VerifyCode(ctx, "13800138000", "123456"); err != nil {
		t.Fatalf("首次校验: %v", err)
	}
	// 重放：码已被删除。
	err := repo.VerifyCode(ctx, "13800138000", "123456")
	if err == nil {
		t.Fatal("重放应被拒绝")
	}
	if !isErrcode(err, 2002) {
		t.Errorf("重放错误码 = %v, want 2002", err)
	}
}

func TestCaptchaVerifyMismatch(t *testing.T) {
	repo, _ := newTestCaptchaRepo(t)
	ctx := context.Background()

	_ = repo.SetCode(ctx, "13800138000", "123456")
	err := repo.VerifyCode(ctx, "13800138000", "654321")
	if err == nil {
		t.Fatal("错误码应被拒绝")
	}
	if !isErrcode(err, 2002) {
		t.Errorf("错误码错误 = %v, want 2002", err)
	}
	// 错误码不消耗原验证码：正确码仍可用。
	if err := repo.VerifyCode(ctx, "13800138000", "123456"); err != nil {
		t.Fatalf("错误尝试后正确码应仍可用: %v", err)
	}
}

func TestCaptchaCodeExpiry(t *testing.T) {
	repo, mr := newTestCaptchaRepo(t)
	ctx := context.Background()

	_ = repo.SetCode(ctx, "13800138000", "123456")
	mr.FastForward(6 * 60 * 1e9) // 超过 5min TTL
	err := repo.VerifyCode(ctx, "13800138000", "123456")
	if err == nil {
		t.Fatal("过期验证码应被拒绝")
	}
}

func TestCaptchaFreqExpiry(t *testing.T) {
	repo, mr := newTestCaptchaRepo(t)
	ctx := context.Background()

	_ = repo.SetCode(ctx, "13800138000", "111111")
	mr.FastForward(61 * 1e9) // 限频窗口 60s 过去
	if err := repo.SetCode(ctx, "13800138000", "222222"); err != nil {
		t.Fatalf("限频窗口过后 SetCode 应放行: %v", err)
	}
}

func isErrcode(err error, want errcode.Code) bool {
	var ec *errcode.Error
	if ok := asErrcode(err, &ec); !ok {
		return false
	}
	return ec.Code == want
}

func asErrcode(err error, target **errcode.Error) bool {
	if e, ok := err.(*errcode.Error); ok {
		*target = e
		return true
	}
	return false
}
