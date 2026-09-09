package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/lindaailabs/yuyan/server/internal/model"
	"github.com/lindaailabs/yuyan/server/internal/pkg/ai"
	"github.com/lindaailabs/yuyan/server/internal/pkg/errcode"
	"github.com/lindaailabs/yuyan/server/internal/repo"
)

// 宠物对话域业务错误码（2xxx 段，23xx 子段）。
var (
	ErrConversationNotFound = errcode.New(2301, "会话不存在或无权限")
	ErrMessageInvalid       = errcode.New(2302, "消息内容不合法")
	ErrPetNotAccessible     = errcode.New(2303, "宠物不存在或无权限")
	ErrAIProviderMissing    = errcode.New(2304, "AI 服务未配置")
)

const (
	maxMessageLen       = 2000 // 单条消息字符上限
	defaultHistoryLimit = 20   // 历史消息默认每页条数
	maxHistoryLimit     = 50   // 历史消息每页上限
	promptRecentTurns   = 10   // 进入 prompt 的最近消息条数（Gateway 内再做字符上限）
	fallbackReply       = "（刚才走神了，你再说一次好不好？）"
)

// ConversationService 宠物对话业务逻辑：会话、消息、AI Gateway 编排与用量落库。
// AI 调用不进事务：单行写入本身原子，模型 RT 不占用 DB 连接与行锁。
type ConversationService struct {
	conv     *repo.ConversationRepo
	msgs     *repo.MessageRepo
	logs     *repo.AICallLogRepo
	pets     *repo.PetRepo
	mem      *MemoryService
	growth   *GrowthService
	ent      *EntitlementService
	analytics *AnalyticsService
	gw       ai.Gateway
}

// NewConversationService 构造。mem 为 nil 时跳过记忆抽取与召回（便于单测裁剪）。
func NewConversationService(
	conv *repo.ConversationRepo,
	msgs *repo.MessageRepo,
	logs *repo.AICallLogRepo,
	pets *repo.PetRepo,
	mem *MemoryService,
	growth *GrowthService,
	gw ai.Gateway,
	ent *EntitlementService,
	analytics *AnalyticsService,
) *ConversationService {
	return &ConversationService{
		conv:     conv,
		msgs:     msgs,
		logs:     logs,
		pets:     pets,
		mem:      mem,
		growth:   growth,
		ent:      ent,
		analytics: analytics,
		gw:       gw,
	}
}

// GetOrCreateConversation 创建或获取「当前用户 + 宠物」的会话。
func (s *ConversationService) GetOrCreateConversation(ctx context.Context, uid int64, in *model.CreateConversationInput) (*model.ConversationItem, error) {
	if in == nil {
		return nil, errcode.New(errcode.ErrInvalidParam, "请求体不能为空")
	}
	pet, err := s.ownPet(ctx, uid, in.PetID)
	if err != nil {
		return nil, err
	}
	conv, err := s.conv.GetOrCreate(ctx, uid, pet.ID)
	if err != nil {
		return nil, fmt.Errorf("get or create conversation: %w", err)
	}
	item := model.ToConversationItem(conv)
	return &item, nil
}

// History 按游标分页返回会话历史（禁止 offset/时间戳排序）。
func (s *ConversationService) History(ctx context.Context, uid, convID, cursor int64, limit int) (*model.MessagePage, error) {
	if convID <= 0 {
		return nil, errcode.New(errcode.ErrInvalidParam, "conv_id 不合法")
	}
	if _, err := s.conv.FindByUserAndID(ctx, uid, convID); repo.IsNotFound(err) {
		return nil, ErrConversationNotFound
	} else if err != nil {
		return nil, fmt.Errorf("find conversation: %w", err)
	}

	if limit <= 0 {
		limit = defaultHistoryLimit
	}
	if limit > maxHistoryLimit {
		limit = maxHistoryLimit
	}
	if cursor < 0 {
		cursor = 0
	}

	// 多取一条用于判断 has_more（不额外 COUNT，省一次查询）。
	rows, err := s.msgs.ListAfter(ctx, convID, cursor, limit+1)
	if err != nil {
		return nil, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	page := &model.MessagePage{Items: model.MessageItems(rows), HasMore: hasMore}
	if len(rows) > 0 {
		page.NextCursor = rows[len(rows)-1].ID
	}
	return page, nil
}

// SendMessage 发送一条用户消息并生成宠物回复。
// 幂等：同一 client_msg_id 重复提交返回首次结果，不产生重复消息。
func (s *ConversationService) SendMessage(ctx context.Context, uid int64, in *model.SendMessageInput) (*model.SendMessageResult, error) {
	if in == nil {
		return nil, errcode.New(errcode.ErrInvalidParam, "请求体不能为空")
	}
	content := strings.TrimSpace(in.Content)
	if content == "" || utf8.RuneCountInString(content) > maxMessageLen {
		return nil, ErrMessageInvalid
	}
	pet, err := s.ownPet(ctx, uid, in.PetID)
	if err != nil {
		return nil, err
	}
	// 额度校验：超额直接拒绝（25xx），不写消息、不调用 AI（guide §7：客户端仅展示，服务端判定）。
	if s.ent != nil {
		if err := s.ent.ConsumeAIMessage(ctx, uid); err != nil {
			s.track(ctx, uid, "ai_quota_rejected", map[string]any{"pet_id": pet.ID})
			return nil, err
		}
	}

	conv, err := s.conv.GetOrCreate(ctx, uid, pet.ID)
	if err != nil {
		return nil, fmt.Errorf("get or create conversation: %w", err)
	}

	now := time.Now().Unix()
	userMsg := &model.PetMessage{
		ConvID:      conv.ID,
		UserID:      uid,
		PetID:       pet.ID,
		Role:        model.MessageRoleUser,
		Content:     content,
		ClientMsgID: in.ClientMsgID,
		Status:      model.MessageStatusOK,
		CreatedAt:   now,
	}
	if err := s.msgs.Create(ctx, userMsg); err != nil {
		if errors.Is(err, repo.ErrDuplicateClientMsgID) && in.ClientMsgID != nil {
			return s.replayResult(ctx, *in.ClientMsgID, conv.ID)
		}
		return nil, fmt.Errorf("insert user message: %w", err)
	}

	// 记忆抽取：失败只告警，不阻断本轮对话（记忆是增值能力）。
	newMemories, err := s.extractMemories(ctx, uid, pet.ID, userMsg.ID, content)
	if err != nil {
		slog.Warn("memory extract failed", "err", err, "user_id", uid)
	} else if len(newMemories) > 0 {
		s.track(ctx, uid, "memory_created", map[string]any{"pet_id": pet.ID, "count": len(newMemories)})
	}
	// 记忆召回：只取少量相关事实进入上下文（guide §5）。
	memories, err := s.recallMemories(ctx, pet.ID, content)
	if err != nil {
		slog.Warn("memory recall failed", "err", err, "user_id", uid)
	}

	turns, err := s.recentTurns(ctx, conv.ID, userMsg.ID)
	if err != nil {
		return nil, err
	}

	res, aiErr := s.gw.Complete(ctx, &ai.CompletionRequest{
		UserID:    uid,
		PetID:     pet.ID,
		PetName:   pet.Name,
		Persona:   deref(pet.Persona),
		Growth:    fmt.Sprintf("等级 %d，亲密度 %d，心情 %s", pet.Level, pet.Intimacy, pet.Mood),
		Memories:  memories,
		Recent:    turns,
		UserInput: content,
	})
	if aiErr != nil && errors.Is(aiErr, ai.ErrProviderNotConfigured) {
		return nil, ErrAIProviderMissing
	}

	assistantMsg := &model.PetMessage{
		ConvID:    conv.ID,
		UserID:    uid,
		PetID:     pet.ID,
		Role:      model.MessageRoleAssistant,
		Status:    model.MessageStatusOK,
		CreatedAt: now,
	}
	switch {
	case aiErr != nil:
		assistantMsg.Status = model.MessageStatusFailed
		assistantMsg.Content = fallbackReply
		assistantMsg.ErrorCode = aiErrCode(res, aiErr)
	case res != nil:
		assistantMsg.Content = res.Content
		assistantMsg.Model = res.Model
		assistantMsg.InputTokens = res.InputTokens
		assistantMsg.OutputTokens = res.OutputTokens
		assistantMsg.LatencyMS = res.LatencyMS
	default:
		assistantMsg.Status = model.MessageStatusFailed
		assistantMsg.Content = fallbackReply
		assistantMsg.ErrorCode = ai.ErrCodeUpstream
	}
	if err := s.msgs.Create(ctx, assistantMsg); err != nil {
		return nil, fmt.Errorf("insert assistant message: %w", err)
	}
	// 成长结算：确定性规则，以宠物回复消息 id 为幂等键（重复结算不加经验）。
	s.applyGrowth(ctx, uid, pet.ID, assistantMsg.ID, aiErr != nil, pet.Level, pet.Intimacy)

	// 会话预览与调用日志失败不影响本次对话结果，仅告警。
	if err := s.conv.UpdateLastMessage(ctx, conv.ID, assistantMsg.ID, assistantMsg.Content); err != nil {
		slog.Warn("update conversation last message failed", "err", err, "conv_id", conv.ID)
	}
	s.writeCallLog(ctx, uid, pet.ID, conv.ID, assistantMsg.ID, res, aiErr)

	item := model.ToMessageItem(assistantMsg)
	return &model.SendMessageResult{
		ConversationID:   conv.ID,
		UserMessage:      model.ToMessageItem(userMsg),
		AssistantMessage: &item,
		NewMemories:      newMemories,
		Streaming:        false, // 流式留待 protocol-v2
		Usage:            usageOf(res, aiErr),
	}, nil
}

// applyGrowth 结算成长（未配置成长服务时跳过；失败仅告警，不影响对话）。
// prevLevel/prevIntimacy 为结算前数值，用于埋点「升级」「首轮对话」。
func (s *ConversationService) applyGrowth(ctx context.Context, uid, petID, msgID int64, aiFailed bool, prevLevel, prevIntimacy int) {
	if s.growth == nil {
		return
	}
	state, err := s.growth.ApplyInteraction(ctx, uid, petID, msgID, aiFailed)
	if err != nil {
		slog.Warn("growth apply failed", "err", err, "pet_id", petID)
		return
	}
	if prevIntimacy == 0 {
		s.track(ctx, uid, "first_message", map[string]any{"pet_id": petID})
	}
	if state.Level > prevLevel {
		s.track(ctx, uid, "level_up", map[string]any{"pet_id": petID, "from": prevLevel, "to": state.Level})
	}
}

// track 关键路径埋点（未配置分析服务时跳过；失败仅告警，不阻断业务）。
func (s *ConversationService) track(ctx context.Context, uid int64, name string, props map[string]any) {
	if s.analytics == nil {
		return
	}
	s.analytics.Track(ctx, uid, name, props)
}

// extractMemories 抽取本轮新形成的记忆（未配置记忆服务时返回 nil）。
func (s *ConversationService) extractMemories(ctx context.Context, uid, petID, msgID int64, content string) ([]model.MemoryItem, error) {
	if s.mem == nil {
		return nil, nil
	}
	return s.mem.Extract(ctx, uid, petID, msgID, content)
}

// recallMemories 召回与当前输入相关的少量记忆。
func (s *ConversationService) recallMemories(ctx context.Context, petID int64, query string) ([]string, error) {
	if s.mem == nil {
		return nil, nil
	}
	return s.mem.Recall(ctx, petID, query, DefaultRecallLimit)
}

// replayResult 幂等重放：返回首次的用户消息与其后的宠物回复。
func (s *ConversationService) replayResult(ctx context.Context, clientMsgID string, convID int64) (*model.SendMessageResult, error) {
	userMsg, err := s.msgs.FindByClientMsgID(ctx, clientMsgID)
	if repo.IsNotFound(err) {
		return nil, ErrConversationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find duplicated message: %w", err)
	}
	result := &model.SendMessageResult{
		ConversationID: convID,
		UserMessage:    model.ToMessageItem(userMsg),
	}
	if reply, err := s.msgs.FindReplyAfter(ctx, convID, userMsg.ID); err == nil {
		item := model.ToMessageItem(reply)
		result.AssistantMessage = &item
		usage := model.UsageSummary{
			Model:        reply.Model,
			InputTokens:  reply.InputTokens,
			OutputTokens: reply.OutputTokens,
			LatencyMS:    reply.LatencyMS,
			ErrCode:      reply.ErrorCode,
		}
		// 用量细节（provider / prompt 版本）只在调用日志里，重放时补齐以保持响应一致。
		if log, err := s.logs.FindByMsgID(ctx, reply.ID); err == nil {
			usage.Provider = log.Provider
			usage.PromptVer = log.PromptVer
			usage.CacheHit = log.CacheHit
		}
		result.Usage = usage
	}
	return result, nil
}

// recentTurns 取进入 prompt 的最近对话（排除本次输入，跳过失败回复）。
func (s *ConversationService) recentTurns(ctx context.Context, convID, excludeID int64) ([]ai.Turn, error) {
	rows, err := s.msgs.RecentTurns(ctx, convID, promptRecentTurns+1)
	if err != nil {
		return nil, err
	}
	turns := make([]ai.Turn, 0, len(rows))
	for i := range rows {
		m := &rows[i]
		if m.ID == excludeID || m.Status == model.MessageStatusFailed {
			continue
		}
		turns = append(turns, ai.Turn{Role: m.Role, Content: m.Content})
	}
	return turns, nil
}

// writeCallLog 落 AI 调用日志（成功与失败都写；写入失败仅告警，不阻断对话）。
func (s *ConversationService) writeCallLog(ctx context.Context, uid, petID, convID, msgID int64, res *ai.CompletionResult, aiErr error) {
	log := &model.AICallLog{
		UserID:    uid,
		PetID:     petID,
		ConvID:    convID,
		MsgID:     msgID,
		CreatedAt: time.Now().Unix(),
	}
	if res != nil {
		log.Provider = res.Provider
		log.Model = res.Model
		log.PromptVer = res.PromptVer
		log.InputTokens = res.InputTokens
		log.OutputTokens = res.OutputTokens
		log.LatencyMS = res.LatencyMS
		log.CacheHit = res.CacheHit
		log.ErrCode = res.ErrCode
	}
	if aiErr != nil && log.ErrCode == 0 {
		log.ErrCode = ai.ErrCodeUpstream
	}
	if err := s.logs.Create(ctx, log); err != nil {
		slog.Warn("ai call log insert failed", "err", err, "user_id", uid, "conv_id", convID)
	}
}

// ownPet 校验宠物归属（越权与不存在同返回 ErrPetNotAccessible）。
func (s *ConversationService) ownPet(ctx context.Context, uid, petID int64) (*model.Pet, error) {
	pet, err := s.pets.FindByUserAndID(ctx, uid, petID)
	if repo.IsNotFound(err) {
		return nil, ErrPetNotAccessible
	}
	if err != nil {
		return nil, fmt.Errorf("find pet: %w", err)
	}
	return pet, nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func aiErrCode(res *ai.CompletionResult, err error) int {
	if err == nil && (res == nil || res.ErrCode == 0) {
		return ai.ErrCodeOK
	}
	if res != nil && res.ErrCode != 0 {
		return res.ErrCode
	}
	if errors.Is(err, ai.ErrTimeout) {
		return ai.ErrCodeTimeout
	}
	return ai.ErrCodeUpstream
}

func usageOf(res *ai.CompletionResult, err error) model.UsageSummary {
	if res == nil {
		return model.UsageSummary{ErrCode: aiErrCode(nil, err)}
	}
	return model.UsageSummary{
		Provider:     res.Provider,
		Model:        res.Model,
		PromptVer:    res.PromptVer,
		InputTokens:  res.InputTokens,
		OutputTokens: res.OutputTokens,
		LatencyMS:    res.LatencyMS,
		CacheHit:     res.CacheHit,
		ErrCode:      aiErrCode(res, err),
	}
}
