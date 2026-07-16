package router

import (
	"github.com/gin-gonic/gin"

	"github.com/dextea-v3/dextea-customer/api/internal/config"
	"github.com/dextea-v3/dextea-customer/api/internal/demo"
)

// Setup 构建并返回配置好的 gin 引擎。
func Setup(cfg *config.Config, demoHandler *demo.Handler) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// 各业务模块自行注册路由，router 只负责引擎与全局中间件。
	demoHandler.Register(r)

	return r
}
