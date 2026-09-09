package api

import (
	"github.com/gin-gonic/gin"

	"github.com/lindaailabs/yuyan/server/internal/model"
	"github.com/lindaailabs/yuyan/server/internal/service"
)

// AnalyticsHandler 数据看板端点：客户端批量上报 + 内部查询（非生产）。
// 埋点范围与脱敏见 guide §8 与 §12 红线。
type AnalyticsHandler struct {
	svc *service.AnalyticsService
}

func NewAnalyticsHandler(svc *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc}
}

// Report POST /events：客户端批量上报（服务端再做脱敏与长度上限校验）。
func (h *AnalyticsHandler) Report(c *gin.Context) {
	var in model.EventReportInput
	if err := c.ShouldBindJSON(&in); err != nil {
		FailErr(c, errInvalidParam)
		return
	}
	n, err := h.svc.Report(c.Request.Context(), UIDFrom(c), &in)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, gin.H{"accepted": n})
}

// List GET /admin/events：内部查询出口（仅非生产环境由 router 注册）。
func (h *AnalyticsHandler) List(c *gin.Context) {
	var q struct {
		Name  string `form:"name"`
		Limit int    `form:"limit"`
	}
	_ = c.ShouldBindQuery(&q)
	items, err := h.svc.List(c.Request.Context(), q.Name, q.Limit)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, items)
}
