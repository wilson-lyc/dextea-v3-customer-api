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
	cfg := config.Load()

	// 打开 MySQL 连接（sqlx），供各业务模块的 repository 共用。
	// 若未配置 DB_HOST / DB_NAME，database 为 nil，健康接口会显示 db:"disabled"。
	database, err := mysql.New(cfg.DatabaseDSN())
	if err != nil {
		log.Fatalf("open mysql error: %v", err)
	}
	if database != nil {
		defer database.Close()
	}

	// 打开 Redis 连接；若未配置 REDIS_ADDR，rdb 为 nil，健康接口会显示 redis:"disabled"。
	rdb, err := redis.New(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("open redis error: %v", err)
	}
	if rdb != nil {
		defer rdb.Close()
	}

	// 顾客模块：repository -> service -> handler。
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
