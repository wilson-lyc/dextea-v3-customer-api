package store

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	applog "github.com/dextea-v3/dextea-customer/api/internal/infra/log"
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
		applog.Warn(c.Request.Context(), "store nearby invalid request params", zap.String("error", fmt.Sprintf("%+v", err)))
		response.ErrorBiz(c, bizerror.New(&bizerror.BizError{Code: 40001, Message: "请求参数不合法"}))
		return
	}

	result, err := h.svc.Nearby(c.Request.Context(), req)
	if err != nil {
		response.ErrorOf(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) Search(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		applog.Warn(c.Request.Context(), "store search invalid request params", zap.String("error", fmt.Sprintf("%+v", err)))
		response.ErrorBiz(c, bizerror.New(&bizerror.BizError{Code: 40001, Message: "请求参数不合法"}))
		return
	}

	result, err := h.svc.Search(c.Request.Context(), req)
	if err != nil {
		response.ErrorOf(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) GetDetail(c *gin.Context) {
	var req GetDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		applog.Warn(c.Request.Context(), "store detail invalid request params", zap.String("error", fmt.Sprintf("%+v", err)))
		response.ErrorBiz(c, bizerror.New(&bizerror.BizError{Code: 40001, Message: "请求参数不合法"}))
		return
	}

	result, err := h.svc.GetDetail(c.Request.Context(), req)
	if err != nil {
		response.ErrorOf(c, err)
		return
	}
	response.OK(c, result)
}
