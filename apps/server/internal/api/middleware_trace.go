package api

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TraceIDHeader trace_id 透传用的请求/响应头。
const TraceIDHeader = "X-Trace-Id"

// traceIDKey context 中 trace_id 的键类型（私有，防止跨包冲突）。
type traceIDKey struct{}

// TraceMiddleware 为每个请求生成或透传 trace_id：
// 请求头存在 X-Trace-Id 则复用，否则生成 uuid；
// 写入响应头并注入 request context，供 slog.InfoContext 输出。
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(TraceIDHeader)
		if traceID == "" {
			traceID = uuid.NewString()
		}
		c.Header(TraceIDHeader, traceID)
		ctx := context.WithValue(c.Request.Context(), traceIDKey{}, traceID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// TraceIDFrom 从 context 取 trace_id；不存在时返回空串。
func TraceIDFrom(ctx context.Context) string {
	if v, ok := ctx.Value(traceIDKey{}).(string); ok {
		return v
	}
	return ""
}
