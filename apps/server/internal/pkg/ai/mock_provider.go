package ai

import (
	"context"
	"fmt"
	"hash/fnv"
	"math/rand"
	"strings"
	"unicode/utf8"
)

// mockModel mock provider 的模型标识（写入 ai_call_logs.model，便于断言与统计）。
const mockModel = "mock-pet-1"

// mockReplies 确定性回复模板：同一输入恒得同一回复，便于测试断言。
var mockReplies = []string{
	"我在这儿呢，今天想聊点什么？",
	"嗯嗯，我记住了，你继续说。",
	"听你这么说，我也觉得挺有意思的。",
	"要不要一起发会儿呆？我陪你。",
	"我喜欢你跟我说这些，再多讲一点吧。",
	"今天过得怎么样呀？我一直在这儿等你。",
}

// MockProvider 默认 provider：不触网、确定性回复，可注入失败率验证兜底。
type MockProvider struct {
	failRate float64 // 0~1，1 表示必失败
}

// NewMockProvider 构造 mock provider。
func NewMockProvider(failRate float64) *MockProvider {
	if failRate < 0 {
		failRate = 0
	}
	if failRate > 1 {
		failRate = 1
	}
	return &MockProvider{failRate: failRate}
}

func (p *MockProvider) Name() string { return ProviderMock }

func (p *MockProvider) Model() string { return mockModel }

// Complete 生成确定性回复；命中失败率时返回 ErrUpstream。
func (p *MockProvider) Complete(ctx context.Context, prompt string, req *CompletionRequest) (*CompletionResult, error) {
	// 尊重上游超时：调用方已用 context 限时，这里只做轻量检查。
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if p.failRate > 0 && rand.Float64() < p.failRate { //nolint:gosec // 仅测试用随机，非安全场景
		return nil, fmt.Errorf("%w: mock 注入失败", ErrUpstream)
	}

	content := mockReply(req)
	return &CompletionResult{
		Content:      content,
		Provider:     ProviderMock,
		Model:        mockModel,
		InputTokens:  EstimateTokens(prompt),
		OutputTokens: EstimateTokens(content),
		CacheHit:     false, // 耗时由 Gateway 统一测量
	}, nil
}

// mockReply 依据输入与宠物名生成确定性回复。
func mockReply(req *CompletionRequest) string {
	petName := "我"
	if req != nil && strings.TrimSpace(req.PetName) != "" {
		petName = strings.TrimSpace(req.PetName)
	}
	input := ""
	if req != nil {
		input = strings.TrimSpace(req.UserInput)
	}
	idx := 0
	if input != "" {
		h := fnv.New32a()
		_, _ = h.Write([]byte(input))
		idx = int(h.Sum32() % uint32(len(mockReplies)))
	}
	body := mockReplies[idx]
	// 回指用户输入，让回复看起来有上下文（长度受控，避免超长）。
	if n := utf8.RuneCountInString(input); n > 0 && n <= 20 {
		return fmt.Sprintf("%s（你说了「%s」）%s", petName, input, body)
	}
	return fmt.Sprintf("%s：%s", petName, body)
}
