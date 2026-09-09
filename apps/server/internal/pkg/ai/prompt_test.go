package ai

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestBuildPromptContainsSectionsInOrder(t *testing.T) {
	p := BuildPrompt(&CompletionRequest{
		PetName:   "小燕",
		Persona:   "活泼好奇",
		Growth:    "等级 2，心情好奇",
		Memories:  []string{"用户喜欢咖啡"},
		Recent:    []Turn{{Role: "user", Content: "早上好"}, {Role: "assistant", Content: "早呀"}},
		UserInput: "今天也很累",
	})

	if p.Version != PromptVersion {
		t.Errorf("version = %s, want %s", p.Version, PromptVersion)
	}
	sysIdx := strings.Index(p.Text, "# 系统规则")
	personaIdx := strings.Index(p.Text, "# 宠物设定")
	growthIdx := strings.Index(p.Text, "# 成长状态")
	memoryIdx := strings.Index(p.Text, "# 长期记忆")
	recentIdx := strings.Index(p.Text, "# 最近对话")
	inputIdx := strings.Index(p.Text, "# 当前输入")
	if !(sysIdx < personaIdx && personaIdx < growthIdx && growthIdx < memoryIdx && memoryIdx < recentIdx && recentIdx < inputIdx) {
		t.Fatalf("段落顺序不符合 guide §5: %s", p.Text)
	}
	for _, want := range []string{"小燕", "活泼好奇", "等级 2", "用户喜欢咖啡", "宠物：早呀", "今天也很累"} {
		if !strings.Contains(p.Text, want) {
			t.Errorf("prompt 缺少 %q", want)
		}
	}
	if p.RecentUsed != 2 || p.RecentTruncated {
		t.Errorf("recent used=%d truncated=%v, want 2/false", p.RecentUsed, p.RecentTruncated)
	}
}

func TestBuildPromptOmitsEmptySections(t *testing.T) {
	p := BuildPrompt(&CompletionRequest{PetName: "小燕", UserInput: "你好"})
	if strings.Contains(p.Text, "# 长期记忆") {
		t.Error("无记忆时不应出现长期记忆段落")
	}
	if strings.Contains(p.Text, "# 最近对话") {
		t.Error("无历史时不应出现最近对话段落")
	}
	if !strings.Contains(p.Text, "暂无") {
		t.Error("成长状态缺失时应填默认值")
	}
}

func TestBuildPromptTruncatesByTurns(t *testing.T) {
	var recent []Turn
	for i := 0; i < MaxRecentTurns+5; i++ {
		recent = append(recent, Turn{Role: "user", Content: "短消息"})
	}
	p := BuildPrompt(&CompletionRequest{PetName: "小燕", UserInput: "继续", Recent: recent})
	if p.RecentUsed != MaxRecentTurns {
		t.Errorf("recent used = %d, want %d", p.RecentUsed, MaxRecentTurns)
	}
	if !p.RecentTruncated {
		t.Error("超出轮次上限应标记 truncated")
	}
}

func TestBuildPromptTruncatesByChars(t *testing.T) {
	long := strings.Repeat("长", MaxRecentChars)
	recent := []Turn{
		{Role: "user", Content: long},
		{Role: "assistant", Content: "好的"},
	}
	p := BuildPrompt(&CompletionRequest{PetName: "小燕", UserInput: "继续", Recent: recent})
	if p.RecentUsed != 1 {
		t.Errorf("字符超限应丢弃较旧轮次, used = %d", p.RecentUsed)
	}
	if !p.RecentTruncated {
		t.Error("字符超限应标记 truncated")
	}
	if utf8.RuneCountInString(p.Text) > MaxRecentChars*2 {
		t.Error("prompt 长度失控，裁剪未生效")
	}
}

func TestEstimateTokens(t *testing.T) {
	if got := EstimateTokens(""); got != 0 {
		t.Errorf("空串 token = %d, want 0", got)
	}
	if got := EstimateTokens("你好世界"); got < 1 {
		t.Errorf("非空 token 应 >= 1, got %d", got)
	}
}
