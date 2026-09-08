package api

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/lindaailabs/yuyan/server/internal/model"
	"github.com/lindaailabs/yuyan/server/internal/service"
)

// ContactsHandler 好友关系端点（/api/v1/friends/*，须经 AuthMiddleware）。
type ContactsHandler struct {
	contacts *service.ContactsService
}

// NewContactsHandler 构造。
func NewContactsHandler(contacts *service.ContactsService) *ContactsHandler {
	return &ContactsHandler{contacts: contacts}
}

// SendRequest POST /friends/requests：发起好友申请（防重复矩阵见 service）。
func (h *ContactsHandler) SendRequest(c *gin.Context) {
	var req model.FriendRequestInput
	if err := c.ShouldBindJSON(&req); err != nil {
		FailErr(c, errInvalidParam)
		return
	}
	resp, err := h.contacts.SendRequest(c.Request.Context(), UIDFrom(c), req.UserID)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, resp)
}

// ListRequests GET /friends/requests：我收到的待处理申请。
func (h *ContactsHandler) ListRequests(c *gin.Context) {
	resp, err := h.contacts.ListRequests(c.Request.Context(), UIDFrom(c))
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, resp)
}

// Accept POST /friends/requests/:id/accept：同意申请（事务写双向两行）。
func (h *ContactsHandler) Accept(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	resp, err := h.contacts.Accept(c.Request.Context(), UIDFrom(c), id)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, resp)
}

// Reject POST /friends/requests/:id/reject：拒绝申请（状态标记，不物理删除）。
func (h *ContactsHandler) Reject(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	resp, err := h.contacts.Reject(c.Request.Context(), UIDFrom(c), id)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, resp)
}

// ListFriends GET /friends：我的好友列表。
func (h *ContactsHandler) ListFriends(c *gin.Context) {
	resp, err := h.contacts.ListFriends(c.Request.Context(), UIDFrom(c))
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, resp)
}

// paramID 解析路径参数 id（正整数）；失败已写入响应，调用方检查 ok 即可。
func paramID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		FailErr(c, errInvalidParam)
		return 0, false
	}
	return id, true
}
