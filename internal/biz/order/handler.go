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

// Register 注册 Order 模块与下游 Java 订单服务一一对应的 8 条路由。
// 顾客身份已由全局 Auth 中间件解析并写入 X-Customer-Id 请求头，此处及各 handler
// 不再解析 token，仅做读取与透传。
func (h *Handler) Register(r *gin.Engine) {
	g := r.Group("/api/v1/orders")

	// 1. 订单预构建
	g.POST("/pre-build", h.PreBuild)
	// 2. 创建订单
	g.POST("", h.Create)
	// 3. 获取月订单列表
	g.GET("/monthly", h.GetMonthOrders)
	// 4. 获取订单详情
	g.GET("/:orderId", h.GetDetail)
	// 5. 获取订单支付状态
	g.GET("/:orderId/payment-status", h.GetPaymentStatus)
	// 6. 取消订单
	g.POST("/:orderId/cancel", h.Cancel)
	// 7. 标记订单制作完成
	g.POST("/:orderId/ready", h.MarkReady)
	// 8. 标记订单已取餐
	g.POST("/:orderId/collect", h.MarkCollected)
}

// PreBuild 订单预构建：原样透传请求体到下游，顾客身份经 X-Customer-Id 头透传。
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

// Create 创建订单：原样透传请求体到下游，顾客身份经 X-Customer-Id 头透传。
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

// GetMonthOrders 获取月订单列表：透传查询参数，顾客身份经 X-Customer-Id 头透传。
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

// GetDetail 获取订单详情：orderId 取自路径参数，顾客身份经 X-Customer-Id 头透传。
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

// GetPaymentStatus 获取订单支付状态：orderId 取自路径参数，顾客身份经 X-Customer-Id 头透传。
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

// Cancel 取消订单：orderId 取自路径参数，顾客身份经 X-Customer-Id 头透传。
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

// MarkReady 标记订单制作完成：orderId 取自路径参数，顾客身份经 X-Customer-Id 头透传。
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

// MarkCollected 标记订单已取餐：orderId 取自路径参数，顾客身份经 X-Customer-Id 头透传。
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

// customerID 从请求头 X-Customer-Id 取出已认证的顾客 id（由全局 Auth 中间件写入）。
// 缺失时交由调用方返回 401。
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
