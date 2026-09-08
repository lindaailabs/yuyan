// Package api 是 HTTP handler 层（Gin）：只做参数校验与调用 service，禁止直接访问 repo。
package api

import (
	"github.com/gin-gonic/gin"
)

// NewRouter 装配 HTTP 路由与中间件。
func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(TraceMiddleware())

	r.GET("/healthz", handleHealthz)
	return r
}
