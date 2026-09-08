package service

import (
	"context"
	"testing"
	"time"

	"github.com/lindaailabs/yuyan/server/internal/pkg/errcode"
	"github.com/lindaailabs/yuyan/server/internal/pkg/jwt"
)

func TestSendSmsCode(t *testing.T) {
	env := newServiceEnv(t)
	ctx := context.Background()

	resp, err := env.svc.SendSmsCode(ctx, "13800138000")
	if err != nil {
		t.Fatalf("SendSmsCode: %v", err)
	}
	if len(resp.CaptchaImage) < 100 {
		t.Errorf("captcha_image 应为有效 base64 PNG, got %d 字节", len(resp.CaptchaImage))
	}

	// 限频：60s 内第二次。
	if _, err := env.svc.SendSmsCode(ctx, "13800138000"); !isErrcode(err, 2001) {
		t.Errorf("第二次应限频 2001, got %v", err)
	}

	// 非法手机号。
	for _, bad := range []string{"", "12345", "12345678901", "23800138000"} {
		if _, err := env.svc.SendSmsCode(ctx, bad); !isErrcode(err, 1001) {
			t.Errorf("非法手机号 %q 应 1001, got %v", bad, err)
		}
	}
}

func TestLoginNewUser(t *testing.T) {
	env := newServiceEnv(t)
	ctx := context.Background()
	phone := "13811112222"

	code := env.sendCode(t, phone)
	resp, err := env.svc.Login(ctx, phone, code)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" || resp.ExpiresIn != 7200 {
		t.Errorf("TokenPair 异常: %+v", resp)
	}

	// access token 可被中间件语义校验（parse 回 uid）。
	uid, err := jwt.NewManager("unit-test-secret").Parse(resp.AccessToken, jwt.TypeAccess)
	if err != nil || uid == 0 {
		t.Errorf("access token 解析异常: uid=%d err=%v", uid, err)
	}

	// 验证码一次性：重放失败。
	if _, err := env.svc.Login(ctx, phone, code); !isErrcode(err, 2002) {
		t.Errorf("重放应 2002, got %v", err)
	}
}

func TestLoginExistingUserSameAccount(t *testing.T) {
	env := newServiceEnv(t)
	ctx := context.Background()
	phone := "13822223333"

	first, err := env.svc.Login(ctx, phone, env.sendCode(t, phone))
	if err != nil {
		t.Fatalf("首次登录: %v", err)
	}
	// 推进 61s 使限频标记过期（代码 TTL 5min 不受影响）。
	env.mr.FastForward(time.Minute + time.Second)
	second, err := env.svc.Login(ctx, phone, env.sendCode(t, phone))
	if err != nil {
		t.Fatalf("二次登录: %v", err)
	}
	// 两次登录 uid 一致（同一账号，自动注册只发生一次）。
	m := jwt.NewManager("unit-test-secret")
	uid1, _ := m.Parse(first.AccessToken, jwt.TypeAccess)
	uid2, _ := m.Parse(second.AccessToken, jwt.TypeAccess)
	if uid1 != uid2 || uid1 == 0 {
		t.Errorf("同手机号两次登录 uid 不一致: %d vs %d", uid1, uid2)
	}
}

func TestLoginWrongCode(t *testing.T) {
	env := newServiceEnv(t)
	ctx := context.Background()
	phone := "13833334444"

	env.sendCode(t, phone)
	if _, err := env.svc.Login(ctx, phone, "000000"); !isErrcode(err, 2002) {
		t.Errorf("错误验证码应 2002, got %v", err)
	}
	// 错误尝试不消耗原验证码。
	if _, err := env.svc.Login(ctx, phone, env.storedCode(t, phone)); err != nil {
		t.Fatalf("错误尝试后正确码登录: %v", err)
	}
}

func TestLoginBadPhone(t *testing.T) {
	env := newServiceEnv(t)
	ctx := context.Background()
	if _, err := env.svc.Login(ctx, "12345", "123456"); !isErrcode(err, 1001) {
		t.Errorf("非法手机号登录应 1001, got %v", err)
	}
	if _, err := env.svc.Login(ctx, "13800138000", ""); !isErrcode(err, 1001) {
		t.Errorf("空验证码应 1001, got %v", err)
	}
}

func TestRefresh(t *testing.T) {
	env := newServiceEnv(t)
	ctx := context.Background()
	phone := "13844445555"

	login, err := env.svc.Login(ctx, phone, env.sendCode(t, phone))
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	refreshed, err := env.svc.Refresh(ctx, login.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if refreshed.AccessToken == "" || refreshed.RefreshToken == "" {
		t.Error("刷新后 token 为空")
	}

	// 用 access 调 refresh：拒绝。
	if _, err := env.svc.Refresh(ctx, login.AccessToken); !isErrcode(err, 1002) {
		t.Errorf("access 刷新应 1002, got %v", err)
	}
	// 垃圾串 / 空串。
	if _, err := env.svc.Refresh(ctx, "garbage"); !isErrcode(err, 1002) {
		t.Errorf("垃圾串应 1002, got %v", err)
	}
	if _, err := env.svc.Refresh(ctx, ""); !isErrcode(err, 1001) {
		t.Errorf("空串应 1001, got %v", err)
	}
}

// isErrcode 断言 err 为指定错误码的 errcode.Error。
func isErrcode(err error, want errcode.Code) bool {
	if err == nil {
		return false
	}
	var ec *errcode.Error
	if e, ok := err.(*errcode.Error); ok {
		ec = e
	} else {
		return false
	}
	return ec.Code == want
}
