package main

import (
	"context"
	"log"
	"net/http"

	"github.com/dextea-v3/dextea-customer/api/internal/biz"
	"github.com/dextea-v3/dextea-customer/api/internal/infra/config"
	"github.com/dextea-v3/dextea-customer/api/internal/infra/otel"
)

// serviceVersion 由构建时注入（ldflags），缺省回退到 unknown。
var serviceVersion = "unknown"

func main() {
	// 加载配置（Nacos 为强依赖，不可达或缺失配置将直接启动失败）
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("%s", err.Error())
	}

	// 配置自检
	if err := cfg.Validate(); err != nil {
		log.Fatalf("%s", err.Error())
	}

	// 初始化 OpenTelemetry：注册全局 propagator 与 TracerProvider，
	// 上报目标由环境变量 OTEL_EXPORTER_OTLP_ENDPOINT 决定。
	shutdown, err := otel.Setup(context.Background(), cfg.ServiceName, serviceVersion, cfg.Environment)
	if err != nil {
		log.Fatalf("init otel error: %v", err)
	}
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			log.Printf("otel shutdown error: %v", err)
		}
	}()

	// 组合根：完成基础设施连接与全部模块装配
	r, cleanup, err := biz.New(cfg)
	if err != nil {
		log.Fatalf("init app error: %v", err)
	}
	defer cleanup()



	addr := ":" + cfg.Port
	log.Printf("%s listening on %s (env=%s)", cfg.ServiceName, addr, cfg.Environment)

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
