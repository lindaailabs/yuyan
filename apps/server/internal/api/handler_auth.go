package api

import (
	"github.com/gin-gonic/gin"

	"github.com/lindaailabs/yuyan/server/internal/pkg/errcode"
	"github.com/lindaailabs/yuyan/server/internal/service"
)

// AuthHandler 认证端点（/api/v1/auth/*）：只做 binding 校验与 service 调用。
type AuthHandler struct {
	auth *service.AuthService
}

// NewAuthHandler 构造。
func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// errInvalidParam 参数 binding 失败的统一 1001。
var errInvalidParam = errcode.New(errcode.ErrInvalidParam, "参数错误")

// SendSmsCode POST /auth/sms-code：下发图形验证码。
func (h *AuthHandler) SendSmsCode(c *gin.Context) {
	var req struct {
		Phone string `json:"phone" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		FailErr(c, errInvalidParam)
		return
	}
	resp, err := h.auth.SendSmsCode(c.Request.Context(), req.Phone)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, resp)
}

// Login POST /auth/login：验证码登录（不存在则自动注册），签发双 token。
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Phone string `json:"phone" binding:"required"`
		Code  string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		FailErr(c, errInvalidParam)
		return
	}
	resp, err := h.auth.Login(c.Request.Context(), req.Phone, req.Code)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, resp)
}

// Refresh POST /auth/refresh：refresh token 换发新双 token。
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		FailErr(c, errInvalidParam)
		return
	}
	resp, err := h.auth.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, resp)
}
