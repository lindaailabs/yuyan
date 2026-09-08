package service

import (
	"context"
	"strings"
	"testing"

	"github.com/lindaailabs/yuyan/server/internal/model"
	"github.com/lindaailabs/yuyan/server/internal/pkg/errcode"
)

func strPtr(s string) *string { return &s }
func i16Ptr(i int16) *int16   { return &i }

func TestProfile(t *testing.T) {
	env := newServiceEnv(t)
	ctx := context.Background()
	uid := env.createUser(t, "13900001111")

	p, err := env.user.Profile(ctx, uid)
	if err != nil {
		t.Fatalf("Profile: %v", err)
	}
	// 自动注册形态：nickname nil、avatar_id 1。
	if p.ID != uid || p.Phone != "13900001111" || p.Nickname != nil || p.AvatarID != 1 {
		t.Errorf("Profile 字段异常: %+v", p)
	}
	if p.CreatedAt <= 0 {
		t.Errorf("CreatedAt 应为 Unix 秒, got %d", p.CreatedAt)
	}
}

func TestProfileNotFound(t *testing.T) {
	env := newServiceEnv(t)
	if _, err := env.user.Profile(context.Background(), 999999); !isErrcode(err, 2013) {
		t.Errorf("不存在用户应 2013, got %v", err)
	}
}

func TestUpdateProfile(t *testing.T) {
	env := newServiceEnv(t)
	ctx := context.Background()
	uid := env.createUser(t, "13900002222")

	p, err := env.user.UpdateProfile(ctx, uid, &model.UpdateProfileInput{Nickname: strPtr("语燕用户"), AvatarID: i16Ptr(3)})
	if err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	if *p.Nickname != "语燕用户" || p.AvatarID != 3 {
		t.Errorf("更新后资料异常: %+v", p)
	}

	// 仅更新昵称：头像保持。
	p, err = env.user.UpdateProfile(ctx, uid, &model.UpdateProfileInput{Nickname: strPtr("新昵称")})
	if err != nil {
		t.Fatalf("UpdateProfile 仅昵称: %v", err)
	}
	if *p.Nickname != "新昵称" || p.AvatarID != 3 {
		t.Errorf("部分更新异常: %+v", p)
	}

	// 空更新（两字段均 nil）：无变化、返回当前资料。
	p, err = env.user.UpdateProfile(ctx, uid, &model.UpdateProfileInput{})
	if err != nil {
		t.Fatalf("UpdateProfile 空更新: %v", err)
	}
	if *p.Nickname != "新昵称" || p.AvatarID != 3 {
		t.Errorf("空更新不应改变资料: %+v", p)
	}

	// nil 入参：1001。
	if _, err := env.user.UpdateProfile(ctx, uid, nil); !isErrcode(err, 1001) {
		t.Errorf("nil 入参应 1001, got %v", err)
	}
}

func TestUpdateProfileValidation(t *testing.T) {
	env := newServiceEnv(t)
	ctx := context.Background()
	uid := env.createUser(t, "13900003333")

	cases := []struct {
		name string
		in   *model.UpdateProfileInput
		want int
	}{
		{"昵称空串", &model.UpdateProfileInput{Nickname: strPtr("")}, 2011},
		{"昵称超长", &model.UpdateProfileInput{Nickname: strPtr(strings.Repeat("燕", 21))}, 2011},
		{"头像 0", &model.UpdateProfileInput{AvatarID: i16Ptr(0)}, 2012},
		{"头像 9", &model.UpdateProfileInput{AvatarID: i16Ptr(9)}, 2012},
	}
	for _, c := range cases {
		if _, err := env.user.UpdateProfile(ctx, uid, c.in); !isErrcode(err, errcode.Code(c.want)) {
			t.Errorf("%s 应 %d, got %v", c.name, c.want, err)
		}
	}

	// 校验失败不落库。
	p, err := env.user.Profile(ctx, uid)
	if err != nil {
		t.Fatalf("Profile: %v", err)
	}
	if p.Nickname != nil || p.AvatarID != 1 {
		t.Errorf("非法更新不应生效: %+v", p)
	}

	// 边界值合法：20 字昵称、头像 1 和 8。
	if _, err := env.user.UpdateProfile(ctx, uid, &model.UpdateProfileInput{Nickname: strPtr(strings.Repeat("燕", 20)), AvatarID: i16Ptr(8)}); err != nil {
		t.Errorf("边界值应合法: %v", err)
	}
	if _, err := env.user.UpdateProfile(ctx, uid, &model.UpdateProfileInput{AvatarID: i16Ptr(1)}); err != nil {
		t.Errorf("头像 1 应合法: %v", err)
	}
}

func TestUpdateProfileNotFound(t *testing.T) {
	env := newServiceEnv(t)
	if _, err := env.user.UpdateProfile(context.Background(), 999999, &model.UpdateProfileInput{Nickname: strPtr("幽灵")}); !isErrcode(err, 2013) {
		t.Errorf("不存在用户应 2013, got %v", err)
	}
}

func TestSearch(t *testing.T) {
	env := newServiceEnv(t)
	ctx := context.Background()
	uid := env.createUser(t, "13911112222")
	if _, err := env.user.UpdateProfile(ctx, uid, &model.UpdateProfileInput{Nickname: strPtr("搜索目标")}); err != nil {
		t.Fatalf("更新昵称: %v", err)
	}

	// 命中：脱敏手机号（前 3 后 4）。
	items, err := env.user.Search(ctx, "13911112222")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("应命中 1 条, got %d", len(items))
	}
	it := items[0]
	if it.ID != uid || it.Phone != "139****2222" || *it.Nickname != "搜索目标" || it.AvatarID != 1 {
		t.Errorf("搜索结果异常: %+v", it)
	}

	// 未注册：空列表、无错误。
	items, err = env.user.Search(ctx, "13999998888")
	if err != nil || len(items) != 0 {
		t.Errorf("未注册应空列表无错误, got items=%d err=%v", len(items), err)
	}

	// 非法 q：1001。
	for _, bad := range []string{"", "12345", "139111122221", "23911112222"} {
		if _, err := env.user.Search(ctx, bad); !isErrcode(err, 1001) {
			t.Errorf("非法 q %q 应 1001, got %v", bad, err)
		}
	}
}
