package router

import (
	"github.com/gin-gonic/gin"

	"github.com/dextea-v3/dextea-customer/api/internal/config"
	"github.com/dextea-v3/dextea-customer/api/internal/handler"
)

// Setup 构建并返回配置好的 gin 引擎。
func Setup(cfg *config.Config) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	registerRoutes(r)

	return r
}

func registerRoutes(r *gin.Engine) {
	r.GET("/health", handler.NewHealthHandler().Check)
}
