package service

import (
	"context"
	"testing"
	"time"

	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"

	"github.com/lindaailabs/yuyan/server/internal/model"
	"github.com/lindaailabs/yuyan/server/internal/repo"
)

type growthEnv struct {
	svc  *GrowthService
	pets *PetService
	ur   *repo.UserRepo
	evts *repo.GrowthRepo
	loc  *time.Location
}

// newGrowthEnv 构建成长测试环境；clock 为 nil 时用真实时间。
func newGrowthEnv(t *testing.T, clock func() time.Time) *growthEnv {
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
	loc := time.FixedZone("CST", 8*60*60)
	petRepo := repo.NewPetRepo(gdb)
	return &growthEnv{
		svc:  NewGrowthService(repo.NewGrowthRepo(gdb), petRepo, clock),
		pets: NewPetService(petRepo),
		ur:   repo.NewUserRepo(gdb),
		evts: repo.NewGrowthRepo(gdb),
		loc:  loc,
	}
}

func (e *growthEnv) user(t *testing.T, phone string) int64 {
	t.Helper()
	u, err := e.ur.CreateUser(context.Background(), phone)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return u.ID
}

func (e *growthEnv) pet(t *testing.T, uid int64) int64 {
	t.Helper()
	pet, err := e.pets.Create(context.Background(), uid, &model.CreatePetInput{Name: "小燕"})
	if err != nil {
		t.Fatalf("create pet: %v", err)
	}
	return pet.ID
}

func TestLevelOfAndMoodOf(t *testing.T) {
	if got := LevelOf(0); got != 1 {
		t.Errorf("LevelOf(0) = %d, want 1", got)
	}
	if got := LevelOf(20); got != 2 {
		t.Errorf("LevelOf(20) = %d, want 2", got)
	}
	if got := LevelOf(45); got != 3 {
		t.Errorf("LevelOf(45) = %d, want 3", got)
	}

	base := time.Date(2026, 9, 12, 20, 0, 0, 0, time.UTC)
	cases := []struct {
		gap  time.Duration
		want string
	}{
		{30 * time.Minute, model.MoodHappy},
		{10 * time.Hour, model.MoodCurious},
		{48 * time.Hour, model.MoodSleepy},
		{96 * time.Hour, model.MoodLonely},
	}
	for _, c := range cases {
		if got := MoodOf(base, base.Add(-c.gap)); got != c.want {
			t.Errorf("MoodOf(gap=%v) = %s, want %s", c.gap, got, c.want)
		}
	}
}

func TestGrowthApplyAddsIntimacyAndEvent(t *testing.T) {
	fixed := time.Date(2026, 9, 12, 20, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	env := newGrowthEnv(t, func() time.Time { return fixed })
	ctx := context.Background()
	uid := env.user(t, "13805000001")
	petID := env.pet(t, uid)

	state, err := env.svc.ApplyInteraction(ctx, uid, petID, 101, false)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if state.Intimacy != intimacyPerMessage {
		t.Errorf("intimacy = %d, want %d", state.Intimacy, intimacyPerMessage)
	}
	if state.Level != 1 {
		t.Errorf("level = %d, want 1", state.Level)
	}
	if state.StreakDays != 1 {
		t.Errorf("streak = %d, want 1", state.StreakDays)
	}

	items, err := env.svc.List(ctx, uid, petID, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 1 || items[0].EventType != model.GrowthEventMessage {
		t.Fatalf("事件异常: %+v", items)
	}
	if items[0].SourceMsgID == nil || *items[0].SourceMsgID != 101 {
		t.Error("事件缺少来源消息 id（不可追溯）")
	}
}

func TestGrowthReplayDoesNotAddIntimacy(t *testing.T) {
	fixed := time.Date(2026, 9, 12, 20, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	env := newGrowthEnv(t, func() time.Time { return fixed })
	ctx := context.Background()
	uid := env.user(t, "13805000002")
	petID := env.pet(t, uid)

	if _, err := env.svc.ApplyInteraction(ctx, uid, petID, 201, false); err != nil {
		t.Fatalf("first apply: %v", err)
	}
	again, err := env.svc.ApplyInteraction(ctx, uid, petID, 201, false)
	if err != nil {
		t.Fatalf("replay apply: %v", err)
	}
	if again.Intimacy != intimacyPerMessage {
		t.Errorf("重放后 intimacy = %d, want %d（重复请求不应重复加经验）", again.Intimacy, intimacyPerMessage)
	}

	items, err := env.svc.List(ctx, uid, petID, 50)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("重放后事件数 = %d, want 1", len(items))
	}
}

func TestGrowthLevelUpEvent(t *testing.T) {
	fixed := time.Date(2026, 9, 12, 20, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	env := newGrowthEnv(t, func() time.Time { return fixed })
	ctx := context.Background()
	uid := env.user(t, "13805000003")
	petID := env.pet(t, uid)

	// 每轮 +2，10 轮达到 20 点 → 2 级。
	for i := 1; i <= 10; i++ {
		if _, err := env.svc.ApplyInteraction(ctx, uid, petID, int64(300+i), false); err != nil {
			t.Fatalf("apply %d: %v", i, err)
		}
	}
	pet, err := env.pets.Detail(ctx, uid, petID)
	if err != nil {
		t.Fatalf("detail: %v", err)
	}
	if pet.Level != 2 || pet.Intimacy != 20 {
		t.Errorf("升级异常: level=%d intimacy=%d, want 2/20", pet.Level, pet.Intimacy)
	}

	items, err := env.svc.List(ctx, uid, petID, 50)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	var levelUp bool
	for _, it := range items {
		if it.EventType == model.GrowthEventLevelUp {
			levelUp = true
		}
	}
	if !levelUp {
		t.Error("缺少 level_up 成长事件")
	}
}

func TestGrowthStreakAcrossDays(t *testing.T) {
	base := time.Date(2026, 9, 12, 20, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	current := base
	env := newGrowthEnv(t, func() time.Time { return current })
	ctx := context.Background()
	uid := env.user(t, "13805000004")
	petID := env.pet(t, uid)

	// 连续三天各互动一次。
	for i := 0; i < 3; i++ {
		current = base.AddDate(0, 0, i)
		state, err := env.svc.ApplyInteraction(ctx, uid, petID, int64(400+i), false)
		if err != nil {
			t.Fatalf("apply day %d: %v", i, err)
		}
		wantStreak := i + 1
		if state.StreakDays != wantStreak {
			t.Errorf("第 %d 天 streak = %d, want %d", i+1, state.StreakDays, wantStreak)
		}
	}

	items, err := env.svc.List(ctx, uid, petID, 50)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	var milestone bool
	for _, it := range items {
		if it.EventType == model.GrowthEventStreak {
			milestone = true
		}
	}
	if !milestone {
		t.Error("连续 3 天应生成 streak_milestone 事件")
	}

	// 断天后重置为 1。
	current = base.AddDate(0, 0, 10)
	state, err := env.svc.ApplyInteraction(ctx, uid, petID, 500, false)
	if err != nil {
		t.Fatalf("apply after gap: %v", err)
	}
	if state.StreakDays != 1 {
		t.Errorf("断天后 streak = %d, want 1", state.StreakDays)
	}
}

func TestGrowthMoodChangesAfterLongGap(t *testing.T) {
	base := time.Date(2026, 9, 12, 20, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	current := base
	env := newGrowthEnv(t, func() time.Time { return current })
	ctx := context.Background()
	uid := env.user(t, "13805000005")
	petID := env.pet(t, uid)

	if _, err := env.svc.ApplyInteraction(ctx, uid, petID, 601, false); err != nil {
		t.Fatalf("first: %v", err)
	}

	// 96 小时后再互动：心情应变为 lonely 并生成 mood_change 事件。
	current = base.Add(96 * time.Hour)
	state, err := env.svc.ApplyInteraction(ctx, uid, petID, 602, false)
	if err != nil {
		t.Fatalf("after gap: %v", err)
	}
	if state.Mood != model.MoodLonely {
		t.Errorf("mood = %s, want %s", state.Mood, model.MoodLonely)
	}

	items, err := env.svc.List(ctx, uid, petID, 50)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	var moodEvent bool
	for _, it := range items {
		if it.EventType == model.GrowthEventMood {
			moodEvent = true
		}
	}
	if !moodEvent {
		t.Error("心情变化应生成 mood_change 事件")
	}
}

func TestGrowthAccessControl(t *testing.T) {
	fixed := time.Date(2026, 9, 12, 20, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	env := newGrowthEnv(t, func() time.Time { return fixed })
	ctx := context.Background()
	owner := env.user(t, "13805000006")
	other := env.user(t, "13805000007")
	petID := env.pet(t, owner)

	if _, err := env.svc.ApplyInteraction(ctx, owner, petID, 701, false); err != nil {
		t.Fatalf("apply: %v", err)
	}

	if _, err := env.svc.List(ctx, other, petID, 10); err == nil {
		t.Error("他人查看成长事件应报错")
	} else {
		wantCode(t, err, ErrPetNotAccessible.Code)
	}

	if _, err := env.svc.ApplyInteraction(ctx, other, petID, 702, false); err == nil {
		t.Error("他人触发成长结算应报错")
	} else {
		wantCode(t, err, ErrPetNotAccessible.Code)
	}
}
