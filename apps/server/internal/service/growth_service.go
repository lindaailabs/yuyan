package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/lindaailabs/yuyan/server/internal/model"
	"github.com/lindaailabs/yuyan/server/internal/repo"
)

// 成长规则常量：集中一处，纯数值、可测试、可回放（guide §4 约定 4：模型不得决定核心数值）。
const (
	intimacyPerMessage = 2  // 每条宠物消息增加的亲密度
	intimacyPerLevel   = 20 // 每级所需亲密度
	maxGrowthEvents    = 50 // 时间线单次返回上限
)

// 连续互动里程碑（达到即生成事件）。
var streakMilestones = []int{3, 7, 30}

// 心情阈值：距上次互动的时长。
const (
	moodHappyWindow   = time.Hour
	moodCuriousWindow = 24 * time.Hour
	moodSleepyWindow  = 72 * time.Hour
)

// GrowthService 成长规则与成长事件（确定性，不依赖模型）。
type GrowthService struct {
	events *repo.GrowthRepo
	pets   *repo.PetRepo
	now    func() time.Time // 注入时钟，便于测试跨天/长时间未互动
	loc    *time.Location
}

// NewGrowthService 构造（now 为 nil 时使用真实时间）。
func NewGrowthService(events *repo.GrowthRepo, pets *repo.PetRepo, now func() time.Time) *GrowthService {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*60*60)
	}
	return &GrowthService{events: events, pets: pets, now: now, loc: loc}
}

func (s *GrowthService) timeNow() time.Time {
	if s.now != nil {
		return s.now().In(s.loc)
	}
	return time.Now().In(s.loc)
}

// ApplyInteraction 结算一次互动的成长：写入事件、更新宠物状态、生成可解释事件。
// 幂等：同一 sourceMsgID 重复调用不重复加经验，直接返回当前状态。
func (s *GrowthService) ApplyInteraction(ctx context.Context, uid, petID, sourceMsgID int64, aiFailed bool) (*model.GrowthState, error) {
	if err := s.insertMessageEvent(ctx, uid, petID, sourceMsgID, aiFailed); err != nil {
		if errors.Is(err, repo.ErrGrowthEventDuplicated) {
			// 已结算过：直接返回当前状态（重复请求不加经验）。
			return s.currentState(ctx, uid, petID)
		}
		return nil, err
	}

	pet, err := s.pets.FindByUserAndID(ctx, uid, petID)
	if repo.IsNotFound(err) {
		return nil, ErrPetNotAccessible
	}
	if err != nil {
		return nil, fmt.Errorf("find pet for growth: %w", err)
	}

	now := s.timeNow()
	intimacy := pet.Intimacy + intimacyPerMessage
	level := LevelOf(intimacy)

	// 心情反映「距上一次互动」的时长，因此必须先读旧统计再累加今日。
	prevStat, err := s.events.FindStat(ctx, petID)
	if err != nil && !repo.IsNotFound(err) {
		slog.Warn("growth find stat failed", "err", err, "pet_id", petID)
	}
	lastAt := lastInteractionOf(prevStat, now)

	stat, err := s.applyDailyStat(ctx, petID, now)
	if err != nil {
		slog.Warn("growth daily stat failed", "err", err, "pet_id", petID)
		stat = &model.PetDailyStat{StreakDays: 1}
	}

	mood := MoodOf(now, lastAt)
	updates := map[string]any{"intimacy": intimacy, "mood": mood}
	if level != pet.Level {
		updates["level"] = level
	}
	if err := s.pets.Update(ctx, uid, petID, updates); err != nil {
		return nil, fmt.Errorf("update pet growth: %w", err)
	}

	if level != pet.Level {
		_ = s.insertPlainEvent(ctx, uid, petID, model.GrowthEventLevelUp,
			deltaOf(map[string]int{"level": level - pet.Level}),
			fmt.Sprintf("亲密度达到 %d，从 %d 级升到 %d 级", intimacy, pet.Level, level))
	}
	if mood != pet.Mood {
		_ = s.insertPlainEvent(ctx, uid, petID, model.GrowthEventMood,
			deltaOf(map[string]string{"mood": mood}),
			fmt.Sprintf("心情从 %s 变为 %s", pet.Mood, mood))
	}
	for _, m := range streakMilestones {
		if stat.StreakDays == m {
			_ = s.insertPlainEvent(ctx, uid, petID, model.GrowthEventStreak,
				deltaOf(map[string]int{"streak_days": m}),
				fmt.Sprintf("连续互动 %d 天", m))
		}
	}

	return &model.GrowthState{
		PetID:      petID,
		Level:      level,
		Intimacy:   intimacy,
		Mood:       mood,
		StreakDays: stat.StreakDays,
	}, nil
}

// List 成长事件时间线（倒序）。
func (s *GrowthService) List(ctx context.Context, uid, petID int64, limit int) ([]model.GrowthEventItem, error) {
	if _, err := s.pets.FindByUserAndID(ctx, uid, petID); repo.IsNotFound(err) {
		return nil, ErrPetNotAccessible
	} else if err != nil {
		return nil, fmt.Errorf("find pet: %w", err)
	}
	if limit <= 0 || limit > maxGrowthEvents {
		limit = maxGrowthEvents
	}
	rows, err := s.events.ListEvents(ctx, petID, limit)
	if err != nil {
		return nil, err
	}
	return model.GrowthEventItems(rows), nil
}

// currentState 读取宠物当前成长状态（幂等重放时返回）。
func (s *GrowthService) currentState(ctx context.Context, uid, petID int64) (*model.GrowthState, error) {
	pet, err := s.pets.FindByUserAndID(ctx, uid, petID)
	if repo.IsNotFound(err) {
		return nil, ErrPetNotAccessible
	}
	if err != nil {
		return nil, fmt.Errorf("find pet: %w", err)
	}
	streak := 0
	if stat, err := s.events.FindStat(ctx, petID); err == nil {
		streak = stat.StreakDays
	}
	return &model.GrowthState{
		PetID:      petID,
		Level:      pet.Level,
		Intimacy:   pet.Intimacy,
		Mood:       pet.Mood,
		StreakDays: streak,
	}, nil
}

// insertMessageEvent 写入「一次互动」事件（幂等闸门）。
func (s *GrowthService) insertMessageEvent(ctx context.Context, uid, petID, sourceMsgID int64, aiFailed bool) error {
	reason := "完成一次对话，亲密度 +2"
	if aiFailed {
		reason = "陪它说了一句话（这次没听清），亲密度 +2"
	}
	src := sourceMsgID
	return s.events.InsertEvent(ctx, &model.PetGrowthEvent{
		UserID:      uid,
		PetID:       petID,
		EventType:   model.GrowthEventMessage,
		Delta:       deltaOf(map[string]int{"intimacy": intimacyPerMessage}),
		Reason:      reason,
		SourceMsgID: &src,
	})
}

// insertPlainEvent 写入无来源消息的事件（升级/心情/里程碑）。
func (s *GrowthService) insertPlainEvent(ctx context.Context, uid, petID int64, typ string, delta *string, reason string) error {
	return s.events.InsertEvent(ctx, &model.PetGrowthEvent{
		UserID:    uid,
		PetID:     petID,
		EventType: typ,
		Delta:     delta,
		Reason:    reason,
	})
}

// applyDailyStat 累加连续互动天数与消息数（按自然日）。
func (s *GrowthService) applyDailyStat(ctx context.Context, petID int64, now time.Time) (*model.PetDailyStat, error) {
	today := now.Format("2006-01-02")
	stat, err := s.events.FindStat(ctx, petID)
	if err != nil && !repo.IsNotFound(err) {
		return nil, fmt.Errorf("find daily stat: %w", err)
	}
	if err != nil { // 首次互动
		stat = &model.PetDailyStat{PetID: petID, LastDate: today, StreakDays: 1, TotalMessages: 0}
	} else if stat.LastDate != today {
		yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
		if stat.LastDate == yesterday {
			stat.StreakDays++
		} else {
			stat.StreakDays = 1 // 断天重置
		}
		stat.LastDate = today
	}
	stat.TotalMessages++
	if err := s.events.UpsertStat(ctx, stat); err != nil {
		return nil, err
	}
	return stat, nil
}

// LevelOf 亲密度 → 等级（纯函数）。
func LevelOf(intimacy int) int {
	if intimacy < 0 {
		intimacy = 0
	}
	return 1 + intimacy/intimacyPerLevel
}

// MoodOf 依据距上次互动时长推导心情（纯函数）。
func MoodOf(now, last time.Time) string {
	if last.IsZero() {
		return model.MoodCurious
	}
	gap := now.Sub(last)
	switch {
	case gap <= moodHappyWindow:
		return model.MoodHappy
	case gap <= moodCuriousWindow:
		return model.MoodCurious
	case gap <= moodSleepyWindow:
		return model.MoodSleepy
	default:
		return model.MoodLonely
	}
}

// lastInteractionOf 由每日统计推导上次互动时间（统计日期当天视为刚互动过）。
func lastInteractionOf(stat *model.PetDailyStat, now time.Time) time.Time {
	if stat == nil || stat.LastDate == "" {
		return time.Time{}
	}
	date, err := time.ParseInLocation("2006-01-02", stat.LastDate, now.Location())
	if err != nil {
		return time.Time{}
	}
	// 统计日期当天：以上一日的 24 点前视为刚互动（保守按当天开始时间计）。
	return date.Add(12 * time.Hour)
}

func deltaOf(v any) *string {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	s := string(b)
	return &s
}
