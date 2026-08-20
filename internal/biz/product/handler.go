package product

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
	g := r.Group("/api/v1/products")
	g.GET("/detail", h.GetDetail)
	g.POST("/status", h.GetStoreStatus)
}

func (h *Handler) GetDetail(c *gin.Context) {
	var req GetProductDetailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		applog.Warn(c.Request.Context(), "product detail invalid request params", zap.String("error", fmt.Sprintf("%+v", err)))
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

func (h *Handler) GetStoreStatus(c *gin.Context) {
	var req GetProductStoreStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		applog.Warn(c.Request.Context(), "product store-status invalid request params", zap.String("error", fmt.Sprintf("%+v", err)))
		response.ErrorBiz(c, bizerror.New(&bizerror.BizError{Code: 40001, Message: "请求参数不合法"}))
		return
	}

	result, err := h.svc.GetStoreStatus(c.Request.Context(), req)
	if err != nil {
		response.ErrorOf(c, err)
		return
	}
	response.OK(c, result)
}
