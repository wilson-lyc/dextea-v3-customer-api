package demo

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler 属于「接口定义」层，负责把业务能力暴露为 HTTP 接口，
// 并完成请求上下文的传递与响应写出。
type Handler struct {
	svc *Service
}

// NewHandler 创建 Handler 实例。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register 将本模块的路由注册到 gin 引擎。
// 每个业务模块自行声明自己的接口，避免路由散落在全局 router 中。
func (h *Handler) Register(r *gin.Engine) {
	r.GET("/health", h.Health)
}

// Health 健康检查接口。
//
//	@Summary	健康检查
//	@Tags		demo
//	@Produce	json
//	@Success	200	{object}	HealthResponse
//	@Router		/health [get]
func (h *Handler) Health(c *gin.Context) {
	resp, err := h.svc.CheckHealth(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
