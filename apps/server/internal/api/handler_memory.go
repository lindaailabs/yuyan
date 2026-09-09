package api

import (
	"github.com/gin-gonic/gin"

	"github.com/lindaailabs/yuyan/server/internal/service"
)

// MemoryHandler 记忆端点（/api/v1/pets/{id}/memories、/pet-memories，须经 AuthMiddleware）。
type MemoryHandler struct {
	mem *service.MemoryService
}

// NewMemoryHandler 构造。
func NewMemoryHandler(mem *service.MemoryService) *MemoryHandler {
	return &MemoryHandler{mem: mem}
}

// List GET /pets/:id/memories：查看宠物的长期记忆。
func (h *MemoryHandler) List(c *gin.Context) {
	petID, ok := paramID(c)
	if !ok {
		return
	}
	items, err := h.mem.List(c.Request.Context(), UIDFrom(c), petID)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, items)
}

// Delete DELETE /pet-memories/:id：软删除记忆（status=3，不物理删除）。
func (h *MemoryHandler) Delete(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	if err := h.mem.Delete(c.Request.Context(), UIDFrom(c), id); err != nil {
		FailErr(c, err)
		return
	}
	OK(c, nil)
}
