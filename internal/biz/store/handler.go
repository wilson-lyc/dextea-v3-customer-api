package store

import (
	"github.com/gin-gonic/gin"

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
	"github.com/dextea-v3/dextea-customer/api/internal/common/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(r *gin.Engine) {
	store := r.Group("/api/v1/stores")
	store.GET("/nearby", h.Nearby)
}

func (h *Handler) Nearby(c *gin.Context) {
	var req NearbyRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailBiz(c, bizerror.New(bizerror.CodeBadRequest, err.Error()))
		return
	}

	result, err := h.svc.Nearby(c.Request.Context(), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, result)
}
