package api

import (
	"github.com/gin-gonic/gin"

	"github.com/lindaailabs/yuyan/server/internal/model"
	"github.com/lindaailabs/yuyan/server/internal/service"
)

// UserHandler 用户资料端点（/api/v1/users/*，须经 AuthMiddleware）。
type UserHandler struct {
	user *service.UserService
}

// NewUserHandler 构造。
func NewUserHandler(user *service.UserService) *UserHandler {
	return &UserHandler{user: user}
}

// Me GET /users/me：当前用户资料（首登判定：nickname null → 客户端跳引导页）。
func (h *UserHandler) Me(c *gin.Context) {
	resp, err := h.user.Profile(c.Request.Context(), UIDFrom(c))
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, resp)
}

// UpdateMe PUT /users/me：选择性更新资料（nickname/avatar_id；越界为 2xxx 业务错误）。
func (h *UserHandler) UpdateMe(c *gin.Context) {
	var req model.UpdateProfileInput
	if err := c.ShouldBindJSON(&req); err != nil {
		FailErr(c, errInvalidParam)
		return
	}
	resp, err := h.user.UpdateProfile(c.Request.Context(), UIDFrom(c), &req)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, resp)
}

// Search GET /users/search?q=：按手机号精确搜索（脱敏）。
func (h *UserHandler) Search(c *gin.Context) {
	var req struct {
		Q string `form:"q" binding:"required"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		FailErr(c, errInvalidParam)
		return
	}
	resp, err := h.user.Search(c.Request.Context(), req.Q)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, resp)
}
