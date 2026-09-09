package ai

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// stubProvider 测试用 provider：可编排结果、错误与耗时；failTimes 控制前 N 次失败。
type stubProvider struct {
	name      string
	model     string
	content   string
	err       error
	failTimes int
	delay     time.Duration
	calls     int
}

func (p *stubProvider) Name() string  { return p.name }
func (p *stubProvider) Model() string { return p.model }

func (p *stubProvider) Complete(ctx context.Context, prompt string, req *CompletionRequest) (*CompletionResult, error) {
	p.calls++
	if p.delay > 0 {
		select {
		case <-time.After(p.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if p.calls <= p.failTimes {
		if p.err != nil {
			return nil, p.err
		}
		return nil, ErrUpstream
	}
	if p.err != nil {
		return nil, p.err
	}
	return &CompletionResult{
		Content:     p.content,
		Provider:    p.name,
		Model:       p.model,
		CacheHit:    true,
		InputTokens: 11,
	}, nil
}

func newTestGateway(t *testing.T, p Provider, timeout time.Duration, attempts int) *gateway {
	t.Helper()
	if attempts <= 0 {
		attempts = defaultAttempts
	}
	return &gateway{provider: p, timeout: timeout, attempts: attempts}
}

func TestNewDefaultsToMock(t *testing.T) {
	g, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res, err := g.Complete(context.Background(), &CompletionRequest{
		UserID: 1, PetID: 2, PetName: "小燕", UserInput: "你好",
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if res.Provider != ProviderMock || res.Model != mockModel {
		t.Errorf("provider/model = %s/%s, want mock/%s", res.Provider, res.Model, mockModel)
	}
	if res.PromptVer != PromptVersion {
		t.Errorf("prompt version = %s, want %s", res.PromptVer, PromptVersion)
	}
	if res.Content == "" {
		t.Error("mock 回复不应为空")
	}
	if !strings.Contains(res.Content, "小燕") {
		t.Errorf("mock 回复应包含宠物名: %q", res.Content)
	}
	if res.InputTokens <= 0 || res.OutputTokens <= 0 {
		t.Errorf("token 统计缺失: input=%d output=%d", res.InputTokens, res.OutputTokens)
	}
	if res.ErrCode != ErrCodeOK {
		t.Errorf("err_code = %d, want 0", res.ErrCode)
	}
}

func TestNewUnknownProviderFails(t *testing.T) {
	if _, err := New(Config{Provider: "openai"}); !errors.Is(err, ErrProviderNotConfigured) {
		t.Fatalf("未实现 provider 应返回 ErrProviderNotConfigured, got %v", err)
	}
}

func TestCompleteEmptyInput(t *testing.T) {
	g, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := g.Complete(context.Background(), &CompletionRequest{UserInput: "   "}); !errors.Is(err, ErrEmptyInput) {
		t.Fatalf("空输入应返回 ErrEmptyInput, got %v", err)
	}
	if _, err := g.Complete(context.Background(), nil); !errors.Is(err, ErrEmptyInput) {
		t.Fatalf("nil 请求应返回 ErrEmptyInput, got %v", err)
	}
}

func TestCompleteRetriesOnceThenSucceeds(t *testing.T) {
	// 前 1 次失败、第 2 次成功：验证确实发生了重试。
	p := &stubProvider{name: "stub", model: "stub-1", content: "好", failTimes: 1}
	g := newTestGateway(t, p, time.Second, 2)

	res, err := g.Complete(context.Background(), &CompletionRequest{UserInput: "你好"})
	if err != nil {
		t.Fatalf("重试后应成功: %v", err)
	}
	if p.calls < 2 {
		t.Errorf("期望重试一次，实际调用 %d 次", p.calls)
	}
	if res.Content != "好" {
		t.Errorf("content = %q, want 好", res.Content)
	}
}

func TestCompleteFailureReturnsErrCode(t *testing.T) {
	p := &stubProvider{name: "stub", model: "stub-1", failTimes: 99}
	g := newTestGateway(t, p, time.Second, 2)

	res, err := g.Complete(context.Background(), &CompletionRequest{UserInput: "你好"})
	if !errors.Is(err, ErrUpstream) {
		t.Fatalf("期望 ErrUpstream, got %v", err)
	}
	if res == nil {
		t.Fatal("失败时也应返回 result 以便落调用日志")
	}
	if res.ErrCode != ErrCodeUpstream {
		t.Errorf("err_code = %d, want %d", res.ErrCode, ErrCodeUpstream)
	}
	if res.LatencyMS < 0 {
		t.Error("latency 不应为负")
	}
}

func TestCompleteTimeoutNoRetry(t *testing.T) {
	p := &stubProvider{name: "stub", model: "stub-1", delay: 500 * time.Millisecond}
	g := newTestGateway(t, p, 50*time.Millisecond, 2)

	res, err := g.Complete(context.Background(), &CompletionRequest{UserInput: "你好"})
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("期望 ErrTimeout, got %v", err)
	}
	if res.ErrCode != ErrCodeTimeout {
		t.Errorf("err_code = %d, want %d", res.ErrCode, ErrCodeTimeout)
	}
	if p.calls != 1 {
		t.Errorf("超时不应重试，实际调用 %d 次", p.calls)
	}
}

func TestMockProviderDeterministic(t *testing.T) {
	p := NewMockProvider(0)
	req := &CompletionRequest{PetName: "小燕", UserInput: "今天很累"}
	first, err := p.Complete(context.Background(), "prompt", req)
	if err != nil {
		t.Fatalf("mock complete: %v", err)
	}
	second, err := p.Complete(context.Background(), "prompt", req)
	if err != nil {
		t.Fatalf("mock complete: %v", err)
	}
	if first.Content != second.Content {
		t.Errorf("mock 回复应确定性: %q vs %q", first.Content, second.Content)
	}
}

func TestMockProviderFailRate(t *testing.T) {
	p := NewMockProvider(1)
	if _, err := p.Complete(context.Background(), "prompt", &CompletionRequest{UserInput: "你好"}); !errors.Is(err, ErrUpstream) {
		t.Fatalf("failRate=1 应返回 ErrUpstream, got %v", err)
	}
}
