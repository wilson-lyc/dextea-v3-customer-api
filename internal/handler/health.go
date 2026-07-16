package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler 提供健康检查相关的接口。
type HealthHandler struct{}

// NewHealthHandler 创建 HealthHandler 实例。
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Check 返回服务的健康状态。
//
//	@Summary	健康检查
//	@Tags		health
//	@Produce	json
//	@Success	200	{object}	map[string]string
//	@Router		/health [get]
func (h *HealthHandler) Check(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "dextea-customer-api",
	})
}
