package api

import (
	"github.com/gin-gonic/gin"

	"github.com/lindaailabs/yuyan/server/internal/model"
	"github.com/lindaailabs/yuyan/server/internal/service"
)

// ConversationHandler 宠物对话端点（/api/v1/pet-conversations、/pet-messages，须经 AuthMiddleware）。
type ConversationHandler struct {
	conv *service.ConversationService
}

// NewConversationHandler 构造。
func NewConversationHandler(conv *service.ConversationService) *ConversationHandler {
	return &ConversationHandler{conv: conv}
}

// CreateConversation POST /pet-conversations：创建或获取会话。
func (h *ConversationHandler) CreateConversation(c *gin.Context) {
	var req model.CreateConversationInput
	if err := c.ShouldBindJSON(&req); err != nil {
		FailErr(c, errInvalidParam)
		return
	}
	resp, err := h.conv.GetOrCreateConversation(c.Request.Context(), UIDFrom(c), &req)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, resp)
}

// SendMessage POST /pet-messages：发送用户消息并返回宠物回复（非流式）。
func (h *ConversationHandler) SendMessage(c *gin.Context) {
	var req model.SendMessageInput
	if err := c.ShouldBindJSON(&req); err != nil {
		FailErr(c, errInvalidParam)
		return
	}
	resp, err := h.conv.SendMessage(c.Request.Context(), UIDFrom(c), &req)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, resp)
}

// History GET /pet-messages?conv_id=&cursor=&limit=：按游标分页拉取历史消息。
func (h *ConversationHandler) History(c *gin.Context) {
	var q struct {
		ConvID int64 `form:"conv_id" binding:"required,gt=0"`
		Cursor int64 `form:"cursor"`
		Limit  int   `form:"limit"`
	}
	if err := c.ShouldBindQuery(&q); err != nil {
		FailErr(c, errInvalidParam)
		return
	}
	resp, err := h.conv.History(c.Request.Context(), UIDFrom(c), q.ConvID, q.Cursor, q.Limit)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, resp)
}
