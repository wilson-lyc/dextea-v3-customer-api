package order

import (
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

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
// 确保所有转发请求都先经过 token 解析、Customer ID 提取并写入 X-Customer-Id 头。
// 各路由均绑定同一个通用转发 handler：拦截器完成身份注入后，请求原样转发下游，
// 不再关心下游各接口的具体路径。
func (h *Handler) Register(r *gin.Engine) {
	g := r.Group("/api/v1/orders")
	// 统一拦截：所有 Order 模块请求必须先完成身份解析并写入 X-Customer-Id。
	g.Use(ForwardInterceptor(h.cfg))

	// 1. 创建订单
	g.POST("", h.Forward)
	// 2. 预构建订单
	g.POST("/pre-build", h.Forward)
	// 3. 获取订单列表
	g.GET("", h.Forward)
	// 4. 获取订单详情
	g.GET("/:orderId", h.Forward)
	// 5. 获取订单状态
	g.GET("/:orderId/status", h.Forward)
}

// Forward 通用转发 handler：从拦截器上下文取已认证的顾客 id，将请求的
// method / path / query / body 原样交给 Service 转发到下游并透传响应。
// 请求路径、查询串、请求体中的具体业务字段均由下游订单服务解释，此处不做处理。
func (h *Handler) Forward(c *gin.Context) {
	customerID, ok := h.customerID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, 40100, "未登录或登录已过期")
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("[WARN] order forward read request body failed: %+v", err)
		response.Error(c, http.StatusBadRequest, 40001, "读取请求体失败")
		return
	}

	result, err := h.svc.Forward(c.Request.Context(), customerID, c.Request.Method, c.Request.URL.Path, c.Request.URL.RawQuery, body)
	if err != nil {
		response.ErrorOf(c, err)
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
