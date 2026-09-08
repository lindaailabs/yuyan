package api

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/lindaailabs/yuyan/server/internal/pkg/errcode"
	"github.com/lindaailabs/yuyan/server/internal/pkg/jwt"
)

// bearerPrefix Authorization 头凭证前缀。
const bearerPrefix = "Bearer "

// ctxUIDKey gin context 中 uid 的键。
const ctxUIDKey = "uid"

// AuthMiddleware Bearer access token 校验（签名/有效期/typ=access），uid 注入 gin context。
// 缺失/无效/typ 不符 → 1002（HTTP 401，客户端拦截器据此触发 refresh），不进入 handler。
func AuthMiddleware(jwtMgr *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, bearerPrefix) {
			abortUnauthorized(c, "缺少 Bearer 凭证")
			return
		}
		uid, err := jwtMgr.Parse(strings.TrimPrefix(h, bearerPrefix), jwt.TypeAccess)
		if err != nil {
			abortUnauthorized(c, "凭证无效或已过期")
			return
		}
		c.Set(ctxUIDKey, uid)
		c.Next()
	}
}

// abortUnauthorized 以 1002/401 中断请求链。
func abortUnauthorized(c *gin.Context, msg string) {
	FailErr(c, errcode.New(errcode.ErrUnauthorized, msg))
	c.Abort()
}

// UIDFrom 从 gin context 取认证中间件注入的 uid；不存在返回 0。
func UIDFrom(c *gin.Context) int64 {
	v, ok := c.Get(ctxUIDKey)
	if !ok {
		return 0
	}
	uid, ok := v.(int64)
	if !ok {
		return 0
	}
	return uid
}
