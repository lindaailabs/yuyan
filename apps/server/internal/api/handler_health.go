package api

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

// handleHealthz 健康检查：无条件返回 200 与统一包裹。
func handleHealthz(c *gin.Context) {
	slog.InfoContext(c.Request.Context(), "healthz ok", "trace_id", TraceIDFrom(c.Request.Context()))
	OK(c, nil)
}
