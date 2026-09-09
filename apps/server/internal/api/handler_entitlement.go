package api

import (
	"github.com/gin-gonic/gin"

	"github.com/lindaailabs/yuyan/server/internal/model"
	"github.com/lindaailabs/yuyan/server/internal/service"
)

// EntitlementHandler 权益与额度端点（/api/v1/entitlements/*，须经 AuthMiddleware）。
// 服务端是权益唯一事实源（guide §7）：额度校验、扣减、支付幂等均在此完成。
type EntitlementHandler struct {
	ents *service.EntitlementService
}

func NewEntitlementHandler(ents *service.EntitlementService) *EntitlementHandler {
	return &EntitlementHandler{ents: ents}
}

// Me GET /entitlements/me：当前权益与当日剩余额度（无记录自动建免费权益）。
func (h *EntitlementHandler) Me(c *gin.Context) {
	view, err := h.ents.Get(c.Request.Context(), UIDFrom(c))
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, view)
}

// SandboxPurchase POST /entitlements/sandbox-purchase：沙盒开通/变更套餐（仅非生产）。
// 生产环境由 service 层返回 2504，本处仅做参数校验与转发。
func (h *EntitlementHandler) SandboxPurchase(c *gin.Context) {
	var in model.SandboxPurchaseInput
	if err := c.ShouldBindJSON(&in); err != nil {
		FailErr(c, errInvalidParam)
		return
	}
	view, err := h.ents.SandboxPurchase(c.Request.Context(), UIDFrom(c), in.Plan)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, view)
}

// PaymentCallback POST /entitlements/payments/callback：支付回调（按订单号幂等）。
func (h *EntitlementHandler) PaymentCallback(c *gin.Context) {
	var in model.PaymentCallbackInput
	if err := c.ShouldBindJSON(&in); err != nil {
		FailErr(c, errInvalidParam)
		return
	}
	view, err := h.ents.HandlePaymentCallback(c.Request.Context(), UIDFrom(c), &in)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, view)
}
