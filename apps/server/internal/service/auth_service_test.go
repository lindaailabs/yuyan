package service

import (
	"context"
	"testing"

	"github.com/lindaailabs/yuyan/server/internal/pkg/errcode"
	"github.com/lindaailabs/yuyan/server/internal/pkg/jwt"
)

func TestRegister(t *testing.T) {
	env := newServiceEnv(t)
	ctx := context.Background()
	phone, password := "13800138000", "secret123"

	resp, err := env.svc.Register(ctx, phone, password)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" || resp.ExpiresIn != 7200 {
		t.Errorf("TokenPair 异常: %+v", resp)
	}

	// 重复注册应被拒（引导走登录）。
	if _, err := env.svc.Register(ctx, phone, password); !isErrcode(err, 2004) {
		t.Errorf("重复注册应 2004, got %v", err)
	}

	// 非法手机号。
	if _, err := env.svc.Register(ctx, "12345", password); !isErrcode(err, 1001) {
		t.Errorf("非法手机号应 1001, got %v", err)
	}
	// 密码过短。
	if _, err := env.svc.Register(ctx, "13800138099", "123"); !isErrcode(err, 1001) {
		t.Errorf("短密码应 1001, got %v", err)
	}
}

func TestLoginNewUser(t *testing.T) {
	env := newServiceEnv(t)
	ctx := context.Background()
	phone, password := "13811112222", "secret123"

	// 未注册直接登录：凭证错误。
	if _, err := env.svc.Login(ctx, phone, password); !isErrcode(err, 2005) {
		t.Errorf("未注册登录应 2005, got %v", err)
	}

	// 注册后再登录成功。
	env.register(t, phone, password)
	resp, err := env.svc.Login(ctx, phone, password)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Errorf("TokenPair 异常: %+v", resp)
	}

	// access token 可被中间件语义校验（parse 回 uid）。
	uid, err := jwt.NewManager("unit-test-secret").Parse(resp.AccessToken, jwt.TypeAccess)
	if err != nil || uid == 0 {
		t.Errorf("access token 解析异常: uid=%d err=%v", uid, err)
	}

	// 错误密码：凭证错误（且不泄露账号是否存在）。
	if _, err := env.svc.Login(ctx, phone, "wrongpass"); !isErrcode(err, 2005) {
		t.Errorf("错误密码应 2005, got %v", err)
	}
}

func TestLoginExistingUserSameAccount(t *testing.T) {
	env := newServiceEnv(t)
	ctx := context.Background()
	phone, password := "13822223333", "secret123"

	first, err := env.svc.Login(ctx, phone, password)
	// 首次未注册：2005，随后注册。
	if !isErrcode(err, 2005) {
		t.Fatalf("首次未注册应 2005, got %v", err)
	}
	env.register(t, phone, password)
	first, err = env.svc.Login(ctx, phone, password)
	if err != nil {
		t.Fatalf("注册后登录: %v", err)
	}
	second, err := env.svc.Login(ctx, phone, password)
	if err != nil {
		t.Fatalf("二次登录: %v", err)
	}
	// 两次登录 uid 一致（同一账号）。
	m := jwt.NewManager("unit-test-secret")
	uid1, _ := m.Parse(first.AccessToken, jwt.TypeAccess)
	uid2, _ := m.Parse(second.AccessToken, jwt.TypeAccess)
	if uid1 != uid2 || uid1 == 0 {
		t.Errorf("同手机号两次登录 uid 不一致: %d vs %d", uid1, uid2)
	}
}

func TestLoginBadPhone(t *testing.T) {
	env := newServiceEnv(t)
	ctx := context.Background()
	if _, err := env.svc.Login(ctx, "12345", "secret123"); !isErrcode(err, 1001) {
		t.Errorf("非法手机号登录应 1001, got %v", err)
	}
	if _, err := env.svc.Login(ctx, "13800138000", ""); !isErrcode(err, 1001) {
		t.Errorf("空密码应 1001, got %v", err)
	}
}

func TestRefresh(t *testing.T) {
	env := newServiceEnv(t)
	ctx := context.Background()
	phone, password := "13844445555", "secret123"

	env.register(t, phone, password)
	login, err := env.svc.Login(ctx, phone, password)
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
