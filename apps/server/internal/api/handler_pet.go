package api

import (
	"github.com/gin-gonic/gin"

	"github.com/lindaailabs/yuyan/server/internal/model"
	"github.com/lindaailabs/yuyan/server/internal/service"
)

// PetHandler AI 宠物档案端点（/api/v1/pets/*，须经 AuthMiddleware）。
type PetHandler struct {
	pets *service.PetService
}

func NewPetHandler(pets *service.PetService) *PetHandler {
	return &PetHandler{pets: pets}
}

// Create POST /pets：创建宠物档案。
func (h *PetHandler) Create(c *gin.Context) {
	var req model.CreatePetInput
	if err := c.ShouldBindJSON(&req); err != nil {
		FailErr(c, errInvalidParam)
		return
	}
	resp, err := h.pets.Create(c.Request.Context(), UIDFrom(c), &req)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, resp)
}

// List GET /pets：我的宠物列表，一期 App 默认使用第一只。
func (h *PetHandler) List(c *gin.Context) {
	resp, err := h.pets.List(c.Request.Context(), UIDFrom(c))
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, resp)
}

// Detail GET /pets/:id：宠物档案。
func (h *PetHandler) Detail(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	resp, err := h.pets.Detail(c.Request.Context(), UIDFrom(c), id)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, resp)
}

// Update PUT /pets/:id：更新基础档案。
func (h *PetHandler) Update(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	var req model.UpdatePetInput
	if err := c.ShouldBindJSON(&req); err != nil {
		FailErr(c, errInvalidParam)
		return
	}
	resp, err := h.pets.Update(c.Request.Context(), UIDFrom(c), id, &req)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, resp)
}

// State GET /pets/:id/state：宠物主页状态。
func (h *PetHandler) State(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	resp, err := h.pets.State(c.Request.Context(), UIDFrom(c), id)
	if err != nil {
		FailErr(c, err)
		return
	}
	OK(c, resp)
}
