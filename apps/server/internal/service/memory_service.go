package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/lindaailabs/yuyan/server/internal/model"
	"github.com/lindaailabs/yuyan/server/internal/pkg/errcode"
	"github.com/lindaailabs/yuyan/server/internal/repo"
)

// 记忆域业务错误码（2xxx 段，24xx 子段）。
var (
	ErrMemoryNotFound = errcode.New(2401, "记忆不存在或无权限")
)

const (
	recallCandidateLimit = 200 // 参与打分的 active 记忆上限
	DefaultRecallLimit   = 3   // 进入上下文的记忆条数上限（guide §5：少量）
	maxMemoryContentLen  = 100
)

// 抽取规则：一期用确定性正则（guide §13：一期 MySQL + 规则筛选）。
// 真实 provider 接入后可叠加模型抽取，表结构与 DTO 已预留 type/confidence。
type extractionRule struct {
	memoryType string
	re         *regexp.Regexp
	confidence float64
	template   string
	negative   bool // 否定式（喜欢/讨厌）需要排除前导「不」
}

var extractionRules = []extractionRule{
	{model.MemoryTypePreference, regexp.MustCompile(`我(?:不喜欢|讨厌|不爱)([^，。！？!?,.]{1,20})`), 0.8, "用户不喜欢$1", true},
	{model.MemoryTypePreference, regexp.MustCompile(`我(?:最喜欢|很喜欢|喜欢|爱)([^，。！？!?,.]{1,20})`), 0.8, "用户喜欢$1", false},
	{model.MemoryTypeProfile, regexp.MustCompile(`我(?:叫|的名字是)([^，。！？!?,.]{1,12})`), 0.7, "用户称呼$1", false},
	{model.MemoryTypeProfile, regexp.MustCompile(`我住在([^，。！？!?,.]{1,12})`), 0.6, "用户住在$1", false},
	{model.MemoryTypeProfile, regexp.MustCompile(`我在([^，。！？!?,.]{1,12})工作`), 0.6, "用户在$1工作", false},
	{model.MemoryTypeEvent, regexp.MustCompile(`我(?:今天|昨天|明天)([^，。！？!?,.]{1,20})`), 0.6, "用户$1（时间相关）", false},
}

// MemoryService 长期记忆：规则抽取、相关度召回、用户查看与软删除。
type MemoryService struct {
	mem      *repo.MemoryRepo
	pets     *repo.PetRepo
	analytics *AnalyticsService
}

func NewMemoryService(mem *repo.MemoryRepo, pets *repo.PetRepo, analytics *AnalyticsService) *MemoryService {
	return &MemoryService{mem: mem, pets: pets, analytics: analytics}
}

// Extract 从一条用户消息中抽取并持久化长期事实，返回本次新增的记忆。
// 抽取失败不阻断对话：调用方以告警处理。
func (s *MemoryService) Extract(ctx context.Context, userID, petID, sourceMsgID int64, text string) ([]model.MemoryItem, error) {
	candidates := extractCandidates(text)
	if len(candidates) == 0 {
		return nil, nil
	}
	items := make([]model.MemoryItem, 0, len(candidates))
	src := sourceMsgID
	for _, c := range candidates {
		m := &model.PetMemory{
			UserID:      userID,
			PetID:       petID,
			MemoryType:  c.memoryType,
			Content:     c.content,
			ContentHash: ContentHash(petID, c.memoryType, c.content),
			SourceMsgID: &src,
			Confidence:  c.confidence,
			Status:      model.MemoryStatusActive,
		}
		inserted, err := s.mem.Upsert(ctx, m)
		if err != nil {
			return items, fmt.Errorf("upsert memory: %w", err)
		}
		if inserted {
			items = append(items, model.ToMemoryItem(m))
		}
	}
	if len(items) > 0 {
		slog.Info("memory extracted",
			"user_id", userID, "pet_id", petID, "count", len(items))
	}
	return items, nil
}

// Recall 按当前输入相关度召回少量 active 记忆（guide §5：只召回少量相关事实）。
func (s *MemoryService) Recall(ctx context.Context, petID int64, query string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = DefaultRecallLimit
	}
	rows, err := s.mem.ListActive(ctx, petID, recallCandidateLimit)
	if err != nil {
		return nil, err
	}
	queryGrams := bigrams(query)
	if len(queryGrams) == 0 {
		return nil, nil
	}

	type scored struct {
		m     model.PetMemory
		score float64
	}
	var hits []scored
	for _, m := range rows {
		score := overlapScore(queryGrams, bigrams(m.Content))
		if score <= 0 {
			continue
		}
		hits = append(hits, scored{m: m, score: score})
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].score != hits[j].score {
			return hits[i].score > hits[j].score
		}
		if hits[i].m.Confidence != hits[j].m.Confidence {
			return hits[i].m.Confidence > hits[j].m.Confidence
		}
		return hits[i].m.UpdatedAt.After(hits[j].m.UpdatedAt)
	})

	if len(hits) > limit {
		hits = hits[:limit]
	}
	recalled := make([]string, 0, len(hits))
	ids := make([]int64, 0, len(hits))
	for _, h := range hits {
		recalled = append(recalled, h.m.Content)
		ids = append(ids, h.m.ID)
	}
	if err := s.mem.TouchUsed(ctx, ids); err != nil {
		slog.Warn("touch memory used failed", "err", err, "pet_id", petID)
	}
	if len(recalled) > 0 {
		slog.Info("memory recalled", "pet_id", petID, "count", len(recalled))
	}
	return recalled, nil
}

// List 用户查看宠物的 active 记忆。
func (s *MemoryService) List(ctx context.Context, userID, petID int64) ([]model.MemoryItem, error) {
	if _, err := s.pets.FindByUserAndID(ctx, userID, petID); repo.IsNotFound(err) {
		return nil, ErrPetNotAccessible
	} else if err != nil {
		return nil, fmt.Errorf("find pet: %w", err)
	}
	rows, err := s.mem.ListActive(ctx, petID, 0)
	if err != nil {
		return nil, err
	}
	return model.MemoryItems(rows), nil
}

// Delete 软删除记忆（status=3，不物理删除）；越权与不存在同返回 ErrMemoryNotFound。
func (s *MemoryService) Delete(ctx context.Context, userID, memoryID int64) error {
	if _, err := s.mem.FindByUserAndID(ctx, userID, memoryID); repo.IsNotFound(err) {
		return ErrMemoryNotFound
	} else if err != nil {
		return fmt.Errorf("find memory: %w", err)
	}
	affected, err := s.mem.MarkDeleted(ctx, userID, memoryID)
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrMemoryNotFound
	}
	slog.Info("memory deleted", "user_id", userID, "memory_id", memoryID)
	if s.analytics != nil {
		s.analytics.Track(ctx, userID, "memory_deleted", map[string]any{"memory_id": memoryID})
	}
	return nil
}

// ContentHash 记忆去重键：宠物 + 类型 + 内容。
func ContentHash(petID int64, memoryType, content string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%d|%s|%s", petID, memoryType, content)))
	return hex.EncodeToString(sum[:])
}

type memoryCandidate struct {
	memoryType string
	content    string
	confidence float64
}

// extractCandidates 按规则表抽取（不含落库），命中多条时各自独立入库。
func extractCandidates(text string) []memoryCandidate {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil
	}
	var out []memoryCandidate
	seen := map[string]bool{}
	for _, rule := range extractionRules {
		locs := rule.re.FindStringSubmatchIndex(trimmed)
		if locs == nil {
			continue
		}
		// 否定式规则跳过前导「不」的肯定匹配（RE2 不支持后向断言，手工判断）。
		if !rule.negative && locs[0] > 0 {
			prev := trimmed[locs[0]-1 : locs[0]]
			if strings.ContainsAny(prev, "不没") {
				continue
			}
		}
		value := strings.TrimSpace(trimmed[locs[2]:locs[3]])
		value = strings.TrimRight(value, "的了啊呀呢吧")
		if value == "" || utf8.RuneCountInString(value) > maxMemoryContentLen {
			continue
		}
		// 疑问句不是事实（"我喜欢什么颜色吗"），必须排除，否则会污染记忆库。
		if isInterrogative(value) {
			continue
		}
		content := strings.ReplaceAll(rule.template, "$1", value)
		if seen[content] {
			continue
		}
		seen[content] = true
		out = append(out, memoryCandidate{
			memoryType: rule.memoryType,
			content:    content,
			confidence: rule.confidence,
		})
	}
	return out
}

// isInterrogative 判断捕获值是否为疑问句/未知指代（不构成长期事实）。
func isInterrogative(value string) bool {
	if strings.ContainsAny(value, "吗呢吧?？") {
		return true
	}
	for _, w := range []string{"什么", "怎么", "为什么", "谁", "哪", "多少", "几个"} {
		if strings.Contains(value, w) {
			return true
		}
	}
	return false
}

// bigrams 生成字符二元组集合（中文无空格，二元组是低成本的相关性信号）。
func bigrams(s string) map[string]bool {
	runes := []rune(s)
	grams := make(map[string]bool, len(runes))
	for i := 0; i+1 < len(runes); i++ {
		grams[string(runes[i:i+2])] = true
	}
	return grams
}

// overlapScore 查询与记忆内容的二元组重合度（Jaccard）：
// 命中数 / 并集大小，兼顾「长内容不该天然占优」与「短内容不该被淹没」。
func overlapScore(queryGrams, contentGrams map[string]bool) float64 {
	if len(queryGrams) == 0 || len(contentGrams) == 0 {
		return 0
	}
	hit := 0
	for g := range queryGrams {
		if contentGrams[g] {
			hit++
		}
	}
	if hit == 0 {
		return 0
	}
	return float64(hit) / float64(len(queryGrams)+len(contentGrams)-hit)
}
