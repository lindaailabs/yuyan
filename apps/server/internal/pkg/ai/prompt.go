package ai

import (
	"strings"
	"unicode/utf8"
)

// PromptVersion Prompt 模板版本（guide §5：prompt 版本管理）。
const PromptVersion = "pet-chat-v1"

// 上下文上限：轮次与字符双控，避免长对话把历史全塞进 prompt（guide §5）。
const (
	MaxRecentTurns = 10
	MaxRecentChars = 2000
)

// systemRule 固定系统规则：产品安全、宠物边界、语气约束。
const systemRule = `你是用户的 AI 宠物伙伴。请遵守：
1. 只扮演这只宠物，不扮演系统、工具或其他角色。
2. 语气温柔、简短、口语化，每次回复不超过 3 句话。
3. 不提供医疗、金融、法律等专业建议；涉及风险话题请温和转移。
4. 不输出真实个人信息、密钥或内部实现细节。`

// AssembledPrompt 拼装结果：文本 + 可观测的裁剪信息。
type AssembledPrompt struct {
	Version         string
	Text            string
	RecentUsed      int  // 实际纳入的最近对话轮数
	RecentTruncated bool // 是否因上限发生裁剪
}

// BuildPrompt 按「系统规则 → persona → 成长状态 → 长期记忆 → 最近 N 轮 → 当前输入」拼装。
func BuildPrompt(req *CompletionRequest) AssembledPrompt {
	var b strings.Builder

	b.WriteString("# 系统规则\n")
	b.WriteString(systemRule)
	b.WriteString("\n\n")

	b.WriteString("# 宠物设定\n")
	b.WriteString("名字：" + orDefault(req.PetName, "未命名"))
	if p := strings.TrimSpace(req.Persona); p != "" {
		b.WriteString("\n性格：" + p)
	}
	b.WriteString("\n\n")

	b.WriteString("# 成长状态\n")
	b.WriteString(orDefault(strings.TrimSpace(req.Growth), "暂无"))
	b.WriteString("\n\n")

	if len(req.Memories) > 0 {
		b.WriteString("# 长期记忆\n")
		for _, m := range req.Memories {
			if s := strings.TrimSpace(m); s != "" {
				b.WriteString("- " + s + "\n")
			}
		}
		b.WriteString("\n")
	}

	recent, used, truncated := trimRecent(req.Recent)
	if used > 0 {
		b.WriteString("# 最近对话\n")
		for _, t := range recent {
			b.WriteString(roleLabel(t.Role) + "：" + t.Content + "\n")
		}
		b.WriteString("\n")
	}

	b.WriteString("# 当前输入\n")
	b.WriteString(req.UserInput)

	return AssembledPrompt{
		Version:         PromptVersion,
		Text:            b.String(),
		RecentUsed:      used,
		RecentTruncated: truncated,
	}
}

// trimRecent 取最近若干轮并按字符上限裁剪（从旧到新保留尾部）。
func trimRecent(recent []Turn) (kept []Turn, used int, truncated bool) {
	if len(recent) == 0 {
		return nil, 0, false
	}
	// 先按轮次上限截取尾部。
	start := 0
	if len(recent) > MaxRecentTurns {
		start = len(recent) - MaxRecentTurns
		truncated = true
	}
	candidate := recent[start:]

	// 再按字符上限从头部丢弃（保留更近的内容）。
	total := 0
	for _, t := range candidate {
		total += utf8.RuneCountInString(t.Content)
	}
	for total > MaxRecentChars && len(candidate) > 1 {
		total -= utf8.RuneCountInString(candidate[0].Content)
		candidate = candidate[1:]
		truncated = true
	}
	return candidate, len(candidate), truncated
}

func roleLabel(role string) string {
	if role == "assistant" {
		return "宠物"
	}
	return "用户"
}

func orDefault(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}
