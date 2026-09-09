package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/lindaailabs/yuyan/server/internal/model"
	"github.com/lindaailabs/yuyan/server/internal/pkg/errcode"
	"github.com/lindaailabs/yuyan/server/internal/repo"
)

// 权益域业务错误码（2xxx 段，25xx 子段）。
var (
	ErrQuotaExceeded    = errcode.New(2501, "今日额度已用完，明天再来或升级订阅")
	ErrEntitlementMissing = errcode.New(2502, "权益不存在或无权限")
	ErrPaymentOrderInvalid = errcode.New(2503, "支付订单参数非法")
	ErrSandboxDisabled  = errcode.New(2504, "沙盒支付未启用（生产环境不可用）")
)

// 额度常量（一期：免费每日 50 条 AI 对话、记忆 200 条；订阅 500 条、记忆 1000 条）。
const (
	FreeDailyMessages = 50
	ProDailyMessages  = 500
	FreeMemoryLimit   = 200
	ProMemoryLimit    = 1000
)

// EntitlementService 权益与额度：服务端唯一事实源（guide §7）。
type EntitlementService struct {
	ents   *repo.EntitlementRepo
	usage  *repo.UsageRepo
	orders *repo.PaymentOrderRepo
	now    func() time.Time
	loc    *time.Location

	// sandboxEnabled 沙盒支付开关：生产环境必须关闭。
	sandboxEnabled bool
}

// NewEntitlementService 构造（now 为 nil 用真实时间）。
func NewEntitlementService(
	ents *repo.EntitlementRepo,
	usage *repo.UsageRepo,
	orders *repo.PaymentOrderRepo,
	now func() time.Time,
	sandboxEnabled bool,
) *EntitlementService {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*60*60)
	}
	return &EntitlementService{
		ents:           ents,
		usage:          usage,
		orders:         orders,
		now:            now,
		loc:            loc,
		sandboxEnabled: sandboxEnabled,
	}
}

func (s *EntitlementService) timeNow() time.Time {
	if s.now != nil {
		return s.now().In(s.loc)
	}
	return time.Now().In(s.loc)
}

// Get 查询权益与当日剩余额度（无记录时自动建立免费权益）。
func (s *EntitlementService) Get(ctx context.Context, userID int64) (*model.EntitlementView, error) {
	ent, err := s.ensure(ctx, userID)
	if err != nil {
		return nil, err
	}
	used, err := s.usedToday(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.viewOf(ent, used), nil
}

// ConsumeAIMessage 扣减一次 AI 对话额度；超额返回 ErrQuotaExceeded（2501）。
func (s *EntitlementService) ConsumeAIMessage(ctx context.Context, userID int64) error {
	ent, err := s.ensure(ctx, userID)
	if err != nil {
		return err
	}
	limit := DailyMessageLimit(ent.Plan)
	period := s.timeNow().Format("2006-01-02")

	// 计数器不存在时先建 0 值，再由条件 UPDATE 保证不超卖。
	if _, err := s.usage.FindCounter(ctx, userID, model.ActionAIMessage, period); err != nil {
		if !repo.IsNotFound(err) {
			return fmt.Errorf("find usage counter: %w", err)
		}
		if err := s.usage.EnsureCounter(ctx, userID, model.ActionAIMessage, period); err != nil {
			return err
		}
	}
	ok, err := s.usage.Consume(ctx, userID, model.ActionAIMessage, period, 1, limit)
	if err != nil {
		return err
	}
	if !ok {
		slog.Info("quota exceeded", "user_id", userID, "plan", ent.Plan, "limit", limit)
		return ErrQuotaExceeded
	}
	return nil
}

// SandboxPurchase 沙盒开通/变更套餐（仅非生产环境）。
func (s *EntitlementService) SandboxPurchase(ctx context.Context, userID int64, plan string) (*model.EntitlementView, error) {
	if !s.sandboxEnabled {
		return nil, ErrSandboxDisabled
	}
	if plan != model.PlanFree && plan != model.PlanPro {
		return nil, ErrPaymentOrderInvalid
	}
	orderNo := fmt.Sprintf("sandbox-%d-%d", userID, s.timeNow().UnixNano())
	return s.grant(ctx, userID, orderNo, "sandbox", plan)
}

// HandlePaymentCallback 支付回调：按订单号幂等，重复回调不重复发放权益。
func (s *EntitlementService) HandlePaymentCallback(ctx context.Context, userID int64, in *model.PaymentCallbackInput) (*model.EntitlementView, error) {
	if in == nil || in.OrderNo == "" {
		return nil, ErrPaymentOrderInvalid
	}
	if in.Plan != model.PlanFree && in.Plan != model.PlanPro {
		return nil, ErrPaymentOrderInvalid
	}
	platform := in.Platform
	if platform == "" {
		platform = "unknown"
	}
	failed := in.Success != nil && !*in.Success

	order := &model.PaymentOrder{
		UserID:   userID,
		OrderNo:  in.OrderNo,
		Platform: platform,
		Plan:     in.Plan,
		Status:   model.PaymentStatusSuccess,
	}
	if failed {
		order.Status = model.PaymentStatusFailed
	}
	created, err := s.orders.Create(ctx, order)
	if err != nil {
		return nil, err
	}
	if !created {
		// 重复回调：直接返回当前权益（幂等成功，不重复发放）。
		slog.Info("payment callback duplicated", "order_no", in.OrderNo, "user_id", userID)
		return s.Get(ctx, userID)
	}
	if failed {
		return s.Get(ctx, userID)
	}
	return s.grant(ctx, userID, in.OrderNo, platform, in.Plan)
}

// grant 发放权益（订单已写入后调用，更新套餐与续期时间）。
func (s *EntitlementService) grant(ctx context.Context, userID int64, orderNo, platform, plan string) (*model.EntitlementView, error) {
	if _, err := s.ensure(ctx, userID); err != nil {
		return nil, err
	}
	renewAt := s.timeNow().AddDate(0, 1, 0)
	if err := s.ents.UpdatePlan(ctx, userID, plan, model.EntitlementActive, &renewAt); err != nil {
		return nil, err
	}
	slog.Info("entitlement granted",
		"user_id", userID, "plan", plan, "platform", platform, "order_no", orderNo)
	return s.Get(ctx, userID)
}

// ensure 保证权益记录存在（默认免费）。
func (s *EntitlementService) ensure(ctx context.Context, userID int64) (*model.Entitlement, error) {
	ent, err := s.ents.FindByUser(ctx, userID)
	if err == nil {
		return ent, nil
	}
	if !repo.IsNotFound(err) {
		return nil, fmt.Errorf("find entitlement: %w", err)
	}
	quota := s.defaultQuota(model.PlanFree)
	created, err := s.ents.Ensure(ctx, &model.Entitlement{
		UserID:    userID,
		Plan:      model.PlanFree,
		Status:    model.EntitlementActive,
		QuotaJSON: &quota,
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// usedToday 当日 AI 对话已用量（无记录为 0）。
func (s *EntitlementService) usedToday(ctx context.Context, userID int64) (int, error) {
	period := s.timeNow().Format("2006-01-02")
	c, err := s.usage.FindCounter(ctx, userID, model.ActionAIMessage, period)
	if err != nil {
		if repo.IsNotFound(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("find usage counter: %w", err)
	}
	return c.Used, nil
}

// viewOf 组装权益视图（含剩余额度）。
func (s *EntitlementService) viewOf(ent *model.Entitlement, used int) *model.EntitlementView {
	limit := DailyMessageLimit(ent.Plan)
	remain := limit - used
	if remain < 0 {
		remain = 0
	}
	view := &model.EntitlementView{
		UserID:    ent.UserID,
		Plan:      ent.Plan,
		Status:    ent.Status,
		UpdatedAt: ent.UpdatedAt.Unix(),
		Quota: model.QuotaView{
			DailyMessages: limit,
			DailyUsed:     used,
			DailyRemain:   remain,
			MemoryLimit:   MemoryLimit(ent.Plan),
			AdvancedModel: ent.Plan == model.PlanPro,
		},
	}
	if ent.RenewAt != nil {
		unix := ent.RenewAt.Unix()
		view.RenewAt = &unix
	}
	return view
}

// defaultQuota 默认额度 JSON（按套餐）。
func (s *EntitlementService) defaultQuota(plan string) string {
	b, _ := json.Marshal(map[string]any{
		"daily_messages": DailyMessageLimit(plan),
		"memory_limit":   MemoryLimit(plan),
		"advanced_model": plan == model.PlanPro,
	})
	return string(b)
}

// DailyMessageLimit 套餐每日 AI 对话额度（纯函数）。
func DailyMessageLimit(plan string) int {
	if plan == model.PlanPro {
		return ProDailyMessages
	}
	return FreeDailyMessages
}

// MemoryLimit 套餐记忆容量上限（纯函数）。
func MemoryLimit(plan string) int {
	if plan == model.PlanPro {
		return ProMemoryLimit
	}
	return FreeMemoryLimit
}
