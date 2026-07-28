package order

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
	"github.com/dextea-v3/dextea-customer/api/internal/common/response"
	"github.com/dextea-v3/dextea-customer/api/internal/config"
)

type Handler struct {
	svc *Service
	cfg *config.Config
}

func NewHandler(svc *Service, cfg *config.Config) *Handler {
	return &Handler{svc: svc, cfg: cfg}
}

// Register 注册 Order 模块的全部 5 条路由，并统一挂载 ForwardInterceptor，
// 确保所有转发请求都先经过 token 解析、Customer ID 提取并写入 X-Custom-Id 头。
func (h *Handler) Register(r *gin.Engine) {
	g := r.Group("/api/v1/orders")
	// 统一拦截：所有 Order 模块请求必须先完成身份解析并写入 X-Custom-Id。
	g.Use(ForwardInterceptor(h.cfg))

	// 1. 创建订单
	g.POST("", func(c *gin.Context) { h.handlePost(c, h.svc.Create) })
	// 2. 预构建订单
	g.POST("/pre-build", func(c *gin.Context) { h.handlePost(c, h.svc.PreBuild) })
	// 3. 获取订单列表
	g.GET("", h.List)
	// 4. 获取订单详情
	g.GET("/:orderId", h.Detail)
	// 5. 获取订单状态
	g.GET("/:orderId/status", h.Status)
}

// handlePost 处理两个 POST 转发（创建 / 预构建），从拦截器上下文中取已认证的顾客 id，
// 将请求体转发到下游订单服务并透传响应。
func (h *Handler) handlePost(c *gin.Context, fn func(context.Context, int64, []byte) (*ForwardResult, error)) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("[WARN] order forward read request body failed: %+v", err)
		response.FailBadRequest(c, "读取请求体失败")
		return
	}

	customerID, ok := h.customerID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, bizerror.CodeUnauthorized.Code, bizerror.CodeUnauthorized.Message)
		return
	}

	result, err := fn(c.Request.Context(), customerID, body)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	c.Data(result.StatusCode, result.ContentType, result.Body)
}

// List 订单列表查询：GET /api/v1/orders
func (h *Handler) List(c *gin.Context) {
	customerID, ok := h.customerID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, bizerror.CodeUnauthorized.Code, bizerror.CodeUnauthorized.Message)
		return
	}
	result, err := h.svc.List(c.Request.Context(), customerID, c.Request.URL.RawQuery)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	c.Data(result.StatusCode, result.ContentType, result.Body)
}

// Detail 订单详情查询：GET /api/v1/orders/{orderId}
func (h *Handler) Detail(c *gin.Context) {
	customerID, ok := h.customerID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, bizerror.CodeUnauthorized.Code, bizerror.CodeUnauthorized.Message)
		return
	}
	orderID := c.Param("orderId")
	if orderID == "" {
		response.FailBadRequest(c, "订单 ID 不能为空")
		return
	}
	result, err := h.svc.Detail(c.Request.Context(), customerID, orderID, c.Request.URL.RawQuery)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	c.Data(result.StatusCode, result.ContentType, result.Body)
}

// Status 订单状态查询：GET /api/v1/orders/{orderId}/status
func (h *Handler) Status(c *gin.Context) {
	customerID, ok := h.customerID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, bizerror.CodeUnauthorized.Code, bizerror.CodeUnauthorized.Message)
		return
	}
	orderID := c.Param("orderId")
	if orderID == "" {
		response.FailBadRequest(c, "订单 ID 不能为空")
		return
	}
	result, err := h.svc.Status(c.Request.Context(), customerID, orderID, c.Request.URL.RawQuery)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	c.Data(result.StatusCode, result.ContentType, result.Body)
}

// customerID 从拦截器写入的 gin.Context 中取出已认证的顾客 id。
// 因 ForwardInterceptor 已保证存在，缺失时视为未通过身份解析，交由调用方返回 401。
func (h *Handler) customerID(c *gin.Context) (int64, bool) {
	if v, ok := c.Get(customIDContextKey); ok {
		if id, ok := v.(int64); ok {
			return id, true
		}
	}
	return 0, false
}
