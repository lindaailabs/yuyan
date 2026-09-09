package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// openaiProvider OpenAI 兼容端点实现（guide §5）：模型调用的唯一真实出口。
// 兼容 OpenAI / DeepSeek / Moonshot / 本地 vLLM 等满足 /chat/completions 协议的服务。
type openaiProvider struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

// NewOpenAIProvider 构造 OpenAI 兼容 provider；缺失 api_key 时启动即失败，不静默降级。
func NewOpenAIProvider(cfg Config) (Provider, error) {
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, fmt.Errorf("%w: AI_API_KEY 缺失（provider=openai）", ErrProviderNotConfigured)
	}
	base := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if base == "" {
		base = "https://api.openai.com/v1"
	}
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = "gpt-3.5-turbo"
	}
	return &openaiProvider{
		baseURL: base,
		apiKey:  cfg.APIKey,
		model:   model,
		client:  &http.Client{Timeout: 30 * time.Second}, // 单次总超时；业务层另有 ctx 限时
	}, nil
}

func (p *openaiProvider) Name() string { return ProviderOpenAI }

func (p *openaiProvider) Model() string { return p.model }

// openaiChatRequest / openaiChatResponse 仅覆盖 chat/completions 所需字段。
type openaiChatRequest struct {
	Model       string          `json:"model"`
	Messages    []openaiMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Stream      bool            `json:"stream"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiChatResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// Complete 调用 OpenAI 兼容端点；入参 prompt 为已拼装文本，整体作为一条 user 消息发送。
func (p *openaiProvider) Complete(ctx context.Context, prompt string, req *CompletionRequest) (*CompletionResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	body := openaiChatRequest{
		Model:       p.model,
		Messages:    []openaiMessage{{Role: "user", Content: prompt}},
		Temperature: 0.8,
		Stream:      false,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("%w: marshal request: %v", ErrUpstream, err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("%w: build request: %v", ErrUpstream, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		// 网络错误：由 Gateway 按是否为超时映射到 5002/5001。
		return nil, fmt.Errorf("%w: %v", ErrUpstream, err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: read response: %v", ErrUpstream, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: http %d: %s", ErrUpstream, resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var parsed openaiChatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("%w: parse response: %v", ErrUpstream, err)
	}
	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("%w: empty choices", ErrUpstream)
	}
	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if content == "" {
		return nil, fmt.Errorf("%w: empty content", ErrUpstream)
	}

	modelName := parsed.Model
	if modelName == "" {
		modelName = p.model
	}
	return &CompletionResult{
		Content:      content,
		Provider:     ProviderOpenAI,
		Model:        modelName,
		InputTokens:  parsed.Usage.PromptTokens,
		OutputTokens: parsed.Usage.CompletionTokens,
		CacheHit:     false,
	}, nil
}
