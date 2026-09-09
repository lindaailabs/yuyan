package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAIProviderParse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("missing auth header")
		}
		if r.URL.Path != "/chat/completions" {
			t.Errorf("path = %s, want /chat/completions", r.URL.Path)
		}
		resp := openaiChatResponse{
			Model: "gpt-3.5-turbo",
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "你好呀"}, FinishReason: "stop"},
			},
		}
		resp.Usage.PromptTokens = 12
		resp.Usage.CompletionTokens = 3
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p, err := NewOpenAIProvider(Config{BaseURL: srv.URL, APIKey: "test-key", Model: "gpt-3.5-turbo"})
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	res, err := p.Complete(context.Background(), "system: 你是宠物\nuser: 嗨", &CompletionRequest{UserInput: "嗨"})
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if res.Content != "你好呀" {
		t.Errorf("content = %q", res.Content)
	}
	if res.Provider != ProviderOpenAI {
		t.Errorf("provider = %q", res.Provider)
	}
	if res.InputTokens != 12 || res.OutputTokens != 3 {
		t.Errorf("usage = %d/%d", res.InputTokens, res.OutputTokens)
	}
}

func TestOpenAIProviderRequiresKey(t *testing.T) {
	if _, err := NewOpenAIProvider(Config{BaseURL: "x", APIKey: ""}); err == nil {
		t.Fatal("缺失 API key 应返回错误")
	}
}
