package main

import (
	"log"
	"net/http"

	"github.com/dextea-v3/dextea-customer/api/internal/biz"
	"github.com/dextea-v3/dextea-customer/api/internal/infra/config"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 配置自检
	if err := cfg.Validate(); err != nil {
		log.Fatalf("%s", err.Error())
	}

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
