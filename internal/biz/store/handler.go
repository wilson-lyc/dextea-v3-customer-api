package store

import (
	"log"

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
	store.GET("/search", h.Search)
	store.GET("/detail", h.GetDetail)
}

func (h *Handler) Nearby(c *gin.Context) {
	var req NearbyRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Printf("[WARN] store nearby invalid request params: %+v", err)
		response.FailBiz(c, bizerror.New(bizerror.CodeValidationFail))
		return
	}

	result, err := h.svc.Nearby(c.Request.Context(), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) Search(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Printf("[WARN] store search invalid request params: %+v", err)
		response.FailBiz(c, bizerror.New(bizerror.CodeValidationFail))
		return
	}

	result, err := h.svc.Search(c.Request.Context(), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) GetDetail(c *gin.Context) {
	var req GetDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Printf("[WARN] store detail invalid request params: %+v", err)
		response.FailBiz(c, bizerror.New(bizerror.CodeValidationFail))
		return
	}

	result, err := h.svc.GetDetail(c.Request.Context(), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.OK(c, result)
}
