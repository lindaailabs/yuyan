package api

import (
	"github.com/gin-gonic/gin"

	"github.com/lindaailabs/yuyan/server/internal/service"
)

// GrowthHandler 成长端点（/api/v1/pets/{id}/growth-events，须经 AuthMiddleware）。
type GrowthHandler struct {
	growth *service.GrowthService
}

// NewGrowthHandler 构造。
func NewGrowthHandler(growth *service.GrowthService) *GrowthHandler {
	return &GrowthHandler{growth: growth}
}

// List GET /pets/:id/growth-events：成长事件时间线（倒序）。
func (h *GrowthHandler) List(c *gin.Context) {
	petID, ok := paramID(c)
	if !ok {
		return
	}
	var q struct {
		Limit int `form:"limit"`
	}
	// 查询参数可选，绑定失败按默认处理（不打断读取）。
	_ = c.ShouldBindQuery(&q)

	items, err := h.growth.List(c.Request.Context(), UIDFrom(c), petID, q.Limit)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, items)
}
