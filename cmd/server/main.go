package main

import (
	"log"
	"net/http"

	"github.com/dextea-v3/dextea-customer/api/internal/config"
	"github.com/dextea-v3/dextea-customer/api/internal/customer"
	"github.com/dextea-v3/dextea-customer/api/internal/mysql"
	"github.com/dextea-v3/dextea-customer/api/internal/redis"
	"github.com/dextea-v3/dextea-customer/api/internal/router"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 配置自检
	if err := cfg.Validate(); err != nil {
		log.Fatalf("%s", err.Error())
	}

	// 连接 MySQL 数据库
	database, err := mysql.New(cfg.DatabaseDSN())
	if err != nil {
		log.Fatalf("open mysql error: %v", err)
	}
	if database != nil {
		defer database.Close()
	}

	// 连接 Redis 数据库
	rdb, err := redis.New(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("open redis error: %v", err)
	}
	if rdb != nil {
		defer rdb.Close()
	}

	// 注册业务模块
	customerRepo := customer.NewRepository(database)
	customerSvc := customer.NewService(customerRepo, rdb, cfg)
	customerHandler := customer.NewHandler(customerSvc)

	r := router.Setup(cfg, customerHandler)

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
