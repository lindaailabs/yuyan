package service

import (
	"context"
	"strings"
	"testing"

	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"

	"github.com/lindaailabs/yuyan/server/internal/model"
	"github.com/lindaailabs/yuyan/server/internal/repo"
)

type memoryEnv struct {
	svc  *MemoryService
	pets *PetService
	ur   *repo.UserRepo
	db   *repo.MemoryRepo
}

func newMemoryEnv(t *testing.T) *memoryEnv {
	t.Helper()
	ctx := context.Background()

	container, err := tcmysql.Run(ctx, "mysql:8.0",
		tcmysql.WithDatabase("yuyan"),
		tcmysql.WithUsername("yuyan"),
		tcmysql.WithPassword("yuyan123"),
	)
	if err != nil {
		t.Fatalf("start mysql: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	dsn := container.MustConnectionString(ctx) + "?charset=utf8mb4&parseTime=True&loc=Local"
	gdb, err := gormOpen(dsn)
	if err != nil {
		t.Fatalf("gorm open: %v", err)
	}
	petRepo := repo.NewPetRepo(gdb)
	return &memoryEnv{
		svc:  NewMemoryService(repo.NewMemoryRepo(gdb), petRepo),
		pets: NewPetService(petRepo),
		ur:   repo.NewUserRepo(gdb),
		db:   repo.NewMemoryRepo(gdb),
	}
}

func (e *memoryEnv) user(t *testing.T, phone string) int64 {
	t.Helper()
	u, err := e.ur.CreateUser(context.Background(), phone)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return u.ID
}

func (e *memoryEnv) pet(t *testing.T, uid int64) int64 {
	t.Helper()
	pet, err := e.pets.Create(context.Background(), uid, &model.CreatePetInput{Name: "小燕"})
	if err != nil {
		t.Fatalf("create pet: %v", err)
	}
	return pet.ID
}

func TestExtractCandidatesRules(t *testing.T) {
	cases := []struct {
		in       string
		wantType string
		wantIn   string
	}{
		{"我喜欢蓝色", model.MemoryTypePreference, "蓝色"},
		{"我不喜欢咖啡", model.MemoryTypePreference, "咖啡"},
		{"我叫小明", model.MemoryTypeProfile, "小明"},
		{"我住在杭州", model.MemoryTypeProfile, "杭州"},
	}
	for _, c := range cases {
		got := extractCandidates(c.in)
		if len(got) == 0 {
			t.Fatalf("%q 未抽取到记忆", c.in)
		}
		found := false
		for _, g := range got {
			if g.memoryType == c.wantType && strings.Contains(g.content, c.wantIn) {
				found = true
			}
			// 否定表达不得被抽取成肯定偏好。
			if c.wantIn == "咖啡" && strings.Contains(g.content, "用户喜欢") {
				t.Errorf("%q 不应抽取为肯定偏好: %s", c.in, g.content)
			}
		}
		if !found {
			t.Errorf("%q 抽取结果 %+v 不含 %s/%s", c.in, got, c.wantType, c.wantIn)
		}
	}

	if got := extractCandidates("今天天气不错"); len(got) != 0 {
		t.Errorf("无事实的句子不应抽取记忆: %+v", got)
	}
	// 疑问句不是事实：不得形成「用户喜欢什么颜色吗」这类记忆。
	for _, q := range []string{"你记得我喜欢什么颜色吗", "我喜欢什么颜色呢", "你知道我喜欢谁吗"} {
		if got := extractCandidates(q); len(got) != 0 {
			t.Errorf("疑问句 %q 不应抽取记忆: %+v", q, got)
		}
	}
}

func TestMemoryExtractAndDedup(t *testing.T) {
	env := newMemoryEnv(t)
	ctx := context.Background()
	uid := env.user(t, "13804000001")
	petID := env.pet(t, uid)

	first, err := env.svc.Extract(ctx, uid, petID, 1, "我喜欢蓝色")
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if len(first) != 1 {
		t.Fatalf("首次抽取 = %d 条, want 1", len(first))
	}
	if first[0].Type != model.MemoryTypePreference || !strings.Contains(first[0].Content, "蓝色") {
		t.Errorf("记忆内容异常: %+v", first[0])
	}
	if first[0].SourceMsgID == nil || *first[0].SourceMsgID != 1 {
		t.Error("缺少来源消息 id（guide §4：可追溯来源）")
	}

	// 同一事实重复表达：不新增，仅更新。
	second, err := env.svc.Extract(ctx, uid, petID, 2, "我喜欢蓝色")
	if err != nil {
		t.Fatalf("extract again: %v", err)
	}
	if len(second) != 0 {
		t.Errorf("重复表达不应产生新记忆: %+v", second)
	}
	list, err := env.svc.List(ctx, uid, petID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("记忆总数 = %d, want 1", len(list))
	}
}

func TestMemoryRecallAndLimit(t *testing.T) {
	env := newMemoryEnv(t)
	ctx := context.Background()
	uid := env.user(t, "13804000002")
	petID := env.pet(t, uid)

	for _, text := range []string{"我喜欢蓝色", "我不喜欢咖啡", "我住在杭州"} {
		if _, err := env.svc.Extract(ctx, uid, petID, 1, text); err != nil {
			t.Fatalf("extract %s: %v", text, err)
		}
	}

	recalled, err := env.svc.Recall(ctx, petID, "你记得我喜欢什么颜色吗", 3)
	if err != nil {
		t.Fatalf("recall: %v", err)
	}
	if len(recalled) == 0 {
		t.Fatal("相关记忆未被召回")
	}
	if !strings.Contains(recalled[0], "蓝色") {
		t.Errorf("最相关记忆应为蓝色, got %q", recalled[0])
	}

	limited, err := env.svc.Recall(ctx, petID, "你记得我喜欢什么颜色吗", 1)
	if err != nil {
		t.Fatalf("recall limit: %v", err)
	}
	if len(limited) != 1 {
		t.Errorf("limit=1 召回 %d 条", len(limited))
	}

	none, err := env.svc.Recall(ctx, petID, "今天天气如何", 3)
	if err != nil {
		t.Fatalf("recall none: %v", err)
	}
	if len(none) != 0 {
		t.Errorf("无关输入不应召回记忆: %v", none)
	}
}

func TestMemoryDeleteStopsRecall(t *testing.T) {
	env := newMemoryEnv(t)
	ctx := context.Background()
	uid := env.user(t, "13804000003")
	petID := env.pet(t, uid)

	items, err := env.svc.Extract(ctx, uid, petID, 1, "我喜欢蓝色")
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("抽取 = %d", len(items))
	}

	if err := env.svc.Delete(ctx, uid, items[0].ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	recalled, err := env.svc.Recall(ctx, petID, "我喜欢什么颜色", 3)
	if err != nil {
		t.Fatalf("recall after delete: %v", err)
	}
	for _, m := range recalled {
		if strings.Contains(m, "蓝色") {
			t.Errorf("删除后仍被召回: %q", m)
		}
	}

	list, err := env.svc.List(ctx, uid, petID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("删除后列表仍有 %d 条", len(list))
	}

	// 再次删除 → 2401（已删除或不存在）。
	if err := env.svc.Delete(ctx, uid, items[0].ID); err == nil {
		t.Error("重复删除应返回错误")
	}
}

func TestMemoryAccessControl(t *testing.T) {
	env := newMemoryEnv(t)
	ctx := context.Background()
	owner := env.user(t, "13804000004")
	other := env.user(t, "13804000005")
	petID := env.pet(t, owner)

	items, err := env.svc.Extract(ctx, owner, petID, 1, "我喜欢蓝色")
	if err != nil {
		t.Fatalf("extract: %v", err)
	}

	if _, err := env.svc.List(ctx, other, petID); err == nil {
		t.Error("他人查看宠物记忆应报错")
	} else {
		wantCode(t, err, ErrPetNotAccessible.Code)
	}

	if err := env.svc.Delete(ctx, other, items[0].ID); err == nil {
		t.Error("他人删除记忆应报错")
	} else {
		wantCode(t, err, ErrMemoryNotFound.Code)
	}

	if err := env.svc.Delete(ctx, owner, 999999); err == nil {
		t.Error("删除不存在的记忆应报错")
	} else {
		wantCode(t, err, ErrMemoryNotFound.Code)
	}
}

func TestMemoryRecallTouchesLastUsed(t *testing.T) {
	env := newMemoryEnv(t)
	ctx := context.Background()
	uid := env.user(t, "13804000006")
	petID := env.pet(t, uid)

	if _, err := env.svc.Extract(ctx, uid, petID, 1, "我喜欢蓝色"); err != nil {
		t.Fatalf("extract: %v", err)
	}
	if _, err := env.svc.Recall(ctx, petID, "我喜欢什么颜色", 3); err != nil {
		t.Fatalf("recall: %v", err)
	}

	rows, err := env.db.ListActive(ctx, petID, 10)
	if err != nil {
		t.Fatalf("list active: %v", err)
	}
	if len(rows) != 1 || rows[0].LastUsedAt == nil {
		t.Fatalf("召回后应更新 last_used_at: %+v", rows)
	}
}
