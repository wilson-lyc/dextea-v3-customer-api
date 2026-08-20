package order

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
	"github.com/dextea-v3/dextea-customer/api/internal/common/response"
	applog "github.com/dextea-v3/dextea-customer/api/internal/infra/log"
	"github.com/dextea-v3/dextea-customer/api/internal/pkg/consts"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(r *gin.Engine) {
	g := r.Group("/api/v1/orders")

	// 订单预构建
	g.POST("/pre-build", h.PreBuild)

	// 创建订单
	g.POST("", h.Create)

	// 获取月订单列表
	g.GET("/monthly", h.GetMonthOrders)

	// 获取订单详情
	g.GET("/:orderId", h.GetDetail)

	// 获取订单支付状态
	g.GET("/:orderId/payment-status", h.GetPaymentStatus)

	// 取消订单
	g.POST("/:orderId/cancel", h.Cancel)

	// 标记订单制作完成
	g.POST("/:orderId/ready", h.MarkReady)

	// 标记订单已取餐
	g.POST("/:orderId/collect", h.MarkCollected)
}

func (h *Handler) PreBuild(c *gin.Context) {
	customerID, ok := h.customerID(c)
	if !ok {
		response.ErrorOf(c, bizerror.ErrUnauthorized)
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		applog.Warn(c.Request.Context(), "order pre-build read request body failed", zap.Error(err))
		response.ErrorOf(c, bizerror.NewWith(bizerror.ErrBadRequest, bizerror.WithMessage("读取请求体失败")))
		return
	}
	result, err := h.svc.Forward(c.Request.Context(), customerID, http.MethodPost, c.Request.URL.Path, "", body)
	if err != nil {
		response.ErrorOf(c, err)
		return
	}
	c.Data(result.StatusCode, result.ContentType, result.Body)
}

func (h *Handler) Create(c *gin.Context) {
	customerID, ok := h.customerID(c)
	if !ok {
		response.ErrorOf(c, bizerror.ErrUnauthorized)
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		applog.Warn(c.Request.Context(), "order create read request body failed", zap.Error(err))
		response.ErrorOf(c, bizerror.NewWith(bizerror.ErrBadRequest, bizerror.WithMessage("读取请求体失败")))
		return
	}
	result, err := h.svc.Forward(c.Request.Context(), customerID, http.MethodPost, c.Request.URL.Path, "", body)
	if err != nil {
		response.ErrorOf(c, err)
		return
	}
	c.Data(result.StatusCode, result.ContentType, result.Body)
}

func (h *Handler) GetMonthOrders(c *gin.Context) {
	customerID, ok := h.customerID(c)
	if !ok {
		response.ErrorOf(c, bizerror.ErrUnauthorized)
		return
	}
	result, err := h.svc.Forward(c.Request.Context(), customerID, http.MethodGet, c.Request.URL.Path, c.Request.URL.RawQuery, nil)
	if err != nil {
		response.ErrorOf(c, err)
		return
	}
	c.Data(result.StatusCode, result.ContentType, result.Body)
}

func (h *Handler) GetDetail(c *gin.Context) {
	customerID, ok := h.customerID(c)
	if !ok {
		response.ErrorOf(c, bizerror.ErrUnauthorized)
		return
	}
	result, err := h.svc.Forward(c.Request.Context(), customerID, http.MethodGet, c.Request.URL.Path, c.Request.URL.RawQuery, nil)
	if err != nil {
		response.ErrorOf(c, err)
		return
	}
	c.Data(result.StatusCode, result.ContentType, result.Body)
}

func (h *Handler) GetPaymentStatus(c *gin.Context) {
	customerID, ok := h.customerID(c)
	if !ok {
		response.ErrorOf(c, bizerror.ErrUnauthorized)
		return
	}
	result, err := h.svc.Forward(c.Request.Context(), customerID, http.MethodGet, c.Request.URL.Path, c.Request.URL.RawQuery, nil)
	if err != nil {
		response.ErrorOf(c, err)
		return
	}
	c.Data(result.StatusCode, result.ContentType, result.Body)
}

func (h *Handler) Cancel(c *gin.Context) {
	customerID, ok := h.customerID(c)
	if !ok {
		response.ErrorOf(c, bizerror.ErrUnauthorized)
		return
	}
	result, err := h.svc.Forward(c.Request.Context(), customerID, http.MethodPost, c.Request.URL.Path, "", nil)
	if err != nil {
		response.ErrorOf(c, err)
		return
	}
	c.Data(result.StatusCode, result.ContentType, result.Body)
}

func (h *Handler) MarkReady(c *gin.Context) {
	customerID, ok := h.customerID(c)
	if !ok {
		response.ErrorOf(c, bizerror.ErrUnauthorized)
		return
	}
	result, err := h.svc.Forward(c.Request.Context(), customerID, http.MethodPost, c.Request.URL.Path, "", nil)
	if err != nil {
		response.ErrorOf(c, err)
		return
	}
	c.Data(result.StatusCode, result.ContentType, result.Body)
}

func (h *Handler) MarkCollected(c *gin.Context) {
	customerID, ok := h.customerID(c)
	if !ok {
		response.ErrorOf(c, bizerror.ErrUnauthorized)
		return
	}
	result, err := h.svc.Forward(c.Request.Context(), customerID, http.MethodPost, c.Request.URL.Path, "", nil)
	if err != nil {
		response.ErrorOf(c, err)
		return
	}
	c.Data(result.StatusCode, result.ContentType, result.Body)
}

func (h *Handler) customerID(c *gin.Context) (int64, bool) {
	raw := c.GetHeader(consts.CustomerIDHeader)
	if raw == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}
