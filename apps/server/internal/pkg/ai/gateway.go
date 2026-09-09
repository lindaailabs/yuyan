// Package ai 是 AI Gateway：模型调用的唯一入口（LLM_DEV_GUIDE.md §5）。
//
// 业务 service 只依赖 Gateway 接口，禁止直接调用具体模型供应商 SDK；
// 本包负责 prompt 拼装、provider 适配、超时与重试、用量统计字段与日志脱敏。
package ai

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"
)

// Gateway 层错误：由 service 翻译为业务错误码（23xx）或写入调用日志 err_code。
var (
	// ErrProviderNotConfigured provider 未实现或配置缺失（不静默降级为 mock）。
	ErrProviderNotConfigured = errors.New("ai: provider 未配置或未实现")
	// ErrEmptyInput 输入为空（调用方缺陷，直接失败不消耗额度）。
	ErrEmptyInput = errors.New("ai: 用户输入为空")
	// ErrUpstream 上游模型调用失败（可重试/兜底）。
	ErrUpstream = errors.New("ai: 上游模型调用失败")
	// ErrTimeout 上游超时（不重试，避免拉长用户等待）。
	ErrTimeout = errors.New("ai: 上游模型调用超时")
)

// 调用日志错误码（与 errcode 段位一致：5xxx 服务端）。
const (
	ErrCodeOK       = 0
	ErrCodeUpstream = 5001
	ErrCodeTimeout  = 5002
)

// Provider 名称常量。
const (
	ProviderMock = "mock"
	// ProviderOpenAI 为 OpenAI 兼容端点（base_url + api_key + model），支持主流兼容服务。
	ProviderOpenAI = "openai"
)

// defaultAttempts 最多尝试次数（1 次正常 + 1 次重试）。
const defaultAttempts = 2

// Turn 一轮历史对话（role: user/assistant）。
type Turn struct {
	Role    string
	Content string
}

// CompletionRequest Gateway 输入：业务语义字段，不含拼装后的 prompt。
type CompletionRequest struct {
	UserID    int64
	PetID     int64
	PetName   string
	Persona   string
	Growth    string
	Memories  []string
	Recent    []Turn
	UserInput string
}

// CompletionResult Gateway 输出：内容与用量统计（guide §5）。
// 调用失败时返回非 nil result（含 ErrCode）与非 nil error，便于调用方既兜底又留痕。
type CompletionResult struct {
	Content      string
	Provider     string
	Model        string
	PromptVer    string
	InputTokens  int
	OutputTokens int
	LatencyMS    int
	CacheHit     bool
	ErrCode      int
}

// Gateway 模型调用唯一入口。
type Gateway interface {
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResult, error)
}

// Provider 具体模型实现（mock / 未来的真实供应商）。
type Provider interface {
	// Name provider 名称（写入 ai_call_logs.provider）。
	Name() string
	// Model 模型标识（写入 ai_call_logs.model）。
	Model() string
	// Complete 执行调用；入参为已拼装的 prompt 文本。
	Complete(ctx context.Context, prompt string, req *CompletionRequest) (*CompletionResult, error)
}

// Config Gateway 配置（由 config 包读取环境变量后传入）。
type Config struct {
	Provider     string  // mock（默认）；openai（OpenAI 兼容）；未实现的取值启动即失败
	TimeoutMS    int     // 单次调用超时
	MockFailRate float64 // 仅 mock 生效：注入失败率，用于验证兜底路径

	// OpenAI 兼容配置（provider=openai 时使用）。
	BaseURL string // 兼容端点，如 https://api.openai.com/v1
	APIKey  string // 模型服务 API Key
	Model   string // 模型名
}

type gateway struct {
	provider Provider
	timeout  time.Duration
	attempts int
}

// New 构造 Gateway；provider 未实现时返回 ErrProviderNotConfigured（明确失败，不静默降级）。
func New(cfg Config) (Gateway, error) {
	provider, err := newProvider(cfg)
	if err != nil {
		return nil, err
	}
	timeout := time.Duration(cfg.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 8 * time.Second // guide §1.3：文本回复完整延迟 P95 < 8s
	}
	return &gateway{provider: provider, timeout: timeout, attempts: defaultAttempts}, nil
}

func newProvider(cfg Config) (Provider, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case "", ProviderMock:
		return NewMockProvider(cfg.MockFailRate), nil
	case ProviderOpenAI:
		return NewOpenAIProvider(cfg)
	default:
		return nil, fmt.Errorf("%w: %s", ErrProviderNotConfigured, cfg.Provider)
	}
}

// Complete 执行一次带超时与重试的模型调用。
// 超时不重试（避免用户等待翻倍）；其余错误重试一次。
func (g *gateway) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResult, error) {
	if req == nil || strings.TrimSpace(req.UserInput) == "" {
		return nil, ErrEmptyInput
	}

	prompt := BuildPrompt(req)
	start := time.Now()

	var lastErr error
	for attempt := 1; attempt <= g.attempts; attempt++ {
		callCtx, cancel := context.WithTimeout(ctx, g.timeout)
		res, err := g.provider.Complete(callCtx, prompt.Text, req)
		cancel()

		if err == nil {
			res.PromptVer = prompt.Version
			res.LatencyMS = int(time.Since(start).Milliseconds())
			if res.InputTokens == 0 {
				res.InputTokens = EstimateTokens(prompt.Text)
			}
			if res.OutputTokens == 0 {
				res.OutputTokens = EstimateTokens(res.Content)
			}
			logCall(req, res, prompt)
			return res, nil
		}

		lastErr = err
		if errors.Is(err, context.DeadlineExceeded) {
			lastErr = fmt.Errorf("%w: %dms", ErrTimeout, g.timeout.Milliseconds())
			break
		}
	}

	res := &CompletionResult{
		Provider:    g.provider.Name(),
		Model:       g.provider.Model(),
		PromptVer:   prompt.Version,
		InputTokens: EstimateTokens(prompt.Text),
		LatencyMS:   int(time.Since(start).Milliseconds()),
		ErrCode:     errCodeOf(lastErr),
	}
	logCall(req, res, prompt)
	return res, lastErr
}

// errCodeOf 把错误映射为调用日志错误码。
func errCodeOf(err error) int {
	switch {
	case err == nil:
		return ErrCodeOK
	case errors.Is(err, ErrTimeout), errors.Is(err, context.DeadlineExceeded):
		return ErrCodeTimeout
	default:
		return ErrCodeUpstream
	}
}

// logCall 脱敏日志：只记结构与统计，禁止输出 prompt 正文、手机号或 token（guide §12）。
func logCall(req *CompletionRequest, res *CompletionResult, prompt AssembledPrompt) {
	slog.Info("ai gateway call",
		"user_id", req.UserID,
		"pet_id", req.PetID,
		"provider", res.Provider,
		"model", res.Model,
		"prompt_version", res.PromptVer,
		"prompt_chars", utf8.RuneCountInString(prompt.Text),
		"recent_turns", prompt.RecentUsed,
		"recent_truncated", prompt.RecentTruncated,
		"input_tokens", res.InputTokens,
		"output_tokens", res.OutputTokens,
		"latency_ms", res.LatencyMS,
		"err_code", res.ErrCode,
	)
}

// EstimateTokens 粗估 token 数：中文按 2 字符≈1 token，其余按 4 字符≈1 token。
// 仅用于 mock 与兜底统计；真实 provider 应返回上游给出的精确值。
func EstimateTokens(s string) int {
	if s == "" {
		return 0
	}
	runes := utf8.RuneCountInString(s)
	// 粗略折中：整体按 2 字符 1 token 估算，最小 1。
	n := runes / 2
	if n < 1 {
		n = 1
	}
	return n
}
