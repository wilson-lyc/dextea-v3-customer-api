package router

import (
	"github.com/gin-gonic/gin"

	"github.com/dextea-v3/dextea-customer/api/internal/config"
	"github.com/dextea-v3/dextea-customer/api/internal/middleware"
	"github.com/dextea-v3/dextea-customer/api/internal/server"
)

// Setup 构建并返回配置好的 gin 引擎。
//
// handlers 为可变参数，由各业务模块实现 server.Registerable 后注入，
// router 不再依赖任何具体业务类型，新增模块无需改动此处。
func Setup(cfg *config.Config, handlers ...server.Registerable) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	// Logger 记录访问日志；ExceptionInterceptor 作为全局异常拦截器，
	// 替代 gin.Recovery：不仅能捕获 panic，还会把系统异常清洗为统一的对外提示。
	r.Use(gin.Logger(), middleware.CORS(), middleware.ExceptionInterceptor())

	// 各业务模块自行注册路由，router 只负责引擎与全局中间件。
	for _, h := range handlers {
		h.Register(r)
	}

	return r
}
