package order

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/dextea-v3/dextea-customer/api/internal/common/response"
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
		response.Error(c, http.StatusUnauthorized, 40100, "未登录或登录已过期")
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("[WARN] order pre-build read request body failed: %+v", err)
		response.Error(c, http.StatusBadRequest, 40001, "读取请求体失败")
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
		response.Error(c, http.StatusUnauthorized, 40100, "未登录或登录已过期")
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("[WARN] order create read request body failed: %+v", err)
		response.Error(c, http.StatusBadRequest, 40001, "读取请求体失败")
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
		response.Error(c, http.StatusUnauthorized, 40100, "未登录或登录已过期")
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
		response.Error(c, http.StatusUnauthorized, 40100, "未登录或登录已过期")
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
		response.Error(c, http.StatusUnauthorized, 40100, "未登录或登录已过期")
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
		response.Error(c, http.StatusUnauthorized, 40100, "未登录或登录已过期")
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
		response.Error(c, http.StatusUnauthorized, 40100, "未登录或登录已过期")
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
		response.Error(c, http.StatusUnauthorized, 40100, "未登录或登录已过期")
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
	raw := c.GetHeader(CustomerIDHeader)
	if raw == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}
