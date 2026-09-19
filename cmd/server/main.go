package main

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/dextea-v3/dextea-customer/api/internal/biz"
	"github.com/dextea-v3/dextea-customer/api/internal/infra/config"
	"github.com/dextea-v3/dextea-customer/api/internal/infra/nacos"
	"github.com/dextea-v3/dextea-customer/api/internal/infra/otel"
)

// serviceVersion 由构建时注入（ldflags），缺省回退到 unknown。
var serviceVersion = "unknown"

func main() {
	// 加载配置（Nacos 配置中心为软依赖；缺省时回退 .env / 默认值）
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

	// 本服务向 Nacos 注册中心注册自身。
	// 「是否配置 Nacos」由 NACOS_ENABLED 与 NACOS_SERVER_ADDR 决定，
	//   - env 无连接参数，或 Nacos 不可达 → 视为未配置，本服务不注册，纯 .env 运行。
	//   - env 有连接参数且连接成功 → 注册本服务，关停时主动注销。
	listenPort, _ := strconv.ParseUint(cfg.Port, 10, 64)
	registrar, regErr := nacos.NewRegistrar(cfg.NacosConfig(), cfg.ServiceName, listenPort, map[string]string{
		"version": serviceVersion,
	})
	if regErr != nil {
		log.Printf("[WARN] nacos unavailable or misconfigured, treat as not-configured and skip self-registration: %v", regErr)
	} else if registrar != nil {
		if err := registrar.Register(); err != nil {
			// 连接参数在 env 中，但 Nacos 实际不可达：等同未配置，回退 .env，不注册。
			log.Printf("[WARN] nacos not reachable, treat as not-configured and skip self-registration: %v", err)
		} else {
			log.Printf("[INFO] registered %s (%s) to nacos", cfg.ServiceName, registrar.Addr())
			defer func() {
				if err := registrar.Deregister(); err != nil {
					log.Printf("[WARN] deregister from nacos failed: %v", err)
				}
			}()
		}
	} else {
		log.Printf("[INFO] nacos not configured in env, running without self-registration (pure .env mode)")
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
