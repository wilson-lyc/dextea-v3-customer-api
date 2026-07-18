package biz

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/dextea-v3/dextea-customer/api/internal/alipay"
	"github.com/dextea-v3/dextea-customer/api/internal/biz/area"
	"github.com/dextea-v3/dextea-customer/api/internal/biz/customer"
	"github.com/dextea-v3/dextea-customer/api/internal/biz/menu"
	"github.com/dextea-v3/dextea-customer/api/internal/biz/product"
	"github.com/dextea-v3/dextea-customer/api/internal/biz/store"
	"github.com/dextea-v3/dextea-customer/api/internal/config"
	"github.com/dextea-v3/dextea-customer/api/internal/mysql"
	"github.com/dextea-v3/dextea-customer/api/internal/redis"
	"github.com/dextea-v3/dextea-customer/api/internal/router"
	"github.com/dextea-v3/dextea-customer/api/internal/server"
)

func New(cfg *config.Config) (*gin.Engine, func(), error) {
	// 初始化 MySQL 连接
	database, err := mysql.New(cfg.DatabaseDSN())
	if err != nil {
		return nil, nil, fmt.Errorf("open mysql: %w", err)
	}

	// 初始化 Redis 连接
	rdb, err := redis.New(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		if database != nil {
			_ = database.Close()
		}
		return nil, nil, fmt.Errorf("open redis: %w", err)
	}

	// 初始化支付宝客户端（独立于 customer 模块）
	var alipayClient *alipay.Client
	if cfg.AlipayAppID != "" && cfg.AlipayPrivateKey != "" {
		isProduction := cfg.AlipayGateway == "" || cfg.AlipayGateway == "https://openapi.alipay.com/gateway.do"
		alipayClient, err = alipay.New(cfg.AlipayAppID, cfg.AlipayPrivateKey, cfg.AlipayPublicKey, isProduction)
		if err != nil {
			log.Printf("[WARN] 支付宝客户端初始化失败，支付宝登录将不可用: %v", err)
			// 不阻塞启动，允许降级运行
		}
	}

	// 业务模块注册
	handlers := []server.Registerable{
		area.NewModule(cfg.AmapAPIKey),
		customer.NewModule(database, rdb, cfg, alipayClient),
		store.NewModule(database, rdb),
		menu.NewModule(database),
		product.NewModule(database),
	}

	engine := router.Setup(cfg, handlers...)

	cleanup := func() {
		if database != nil {
			_ = database.Close()
		}
		if rdb != nil {
			_ = rdb.Close()
		}
	}

	return engine, cleanup, nil
}
