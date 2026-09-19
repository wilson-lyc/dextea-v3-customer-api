package biz

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/dextea-v3/dextea-customer/api/internal/biz/area"
	"github.com/dextea-v3/dextea-customer/api/internal/biz/customer"
	"github.com/dextea-v3/dextea-customer/api/internal/biz/menu"
	"github.com/dextea-v3/dextea-customer/api/internal/biz/order"
	"github.com/dextea-v3/dextea-customer/api/internal/biz/product"
	"github.com/dextea-v3/dextea-customer/api/internal/biz/store"
	"github.com/dextea-v3/dextea-customer/api/internal/infra/alipay"
	"github.com/dextea-v3/dextea-customer/api/internal/infra/config"
	"github.com/dextea-v3/dextea-customer/api/internal/infra/mysql"
	"github.com/dextea-v3/dextea-customer/api/internal/infra/redis"
	"github.com/dextea-v3/dextea-customer/api/internal/infra/storerpc"
	"github.com/dextea-v3/dextea-customer/api/internal/transport/router"
	"github.com/dextea-v3/dextea-customer/api/internal/transport/server"
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

	// 初始化支付宝客户端（独立于 customer 模块），失败时降级为 nil，不阻塞启动。
	alipayClient := alipay.NewFromConfig(cfg)

	// 业务模块注册
	productHandler, err := product.NewModule(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("init product rpc client: %w", err)
	}
	storeClient, err := storerpc.NewClient(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("init store rpc client: %w", err)
	}
	menuHandler, err := menu.NewModule(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("init menu rpc client: %w", err)
	}
	handlers := []server.Registerable{
		area.NewModule(storeClient, rdb, cfg),
		customer.NewModule(database, rdb, cfg, alipayClient),
		store.NewModule(storeClient),
		menuHandler,
		productHandler,
		order.NewModule(cfg),
	}

	engine := router.Setup(cfg, handlers...)

	cleanup := func() {
		// Store RPC 当前按请求建立短连接，保留统一清理位以便未来切换连接池。
		if database != nil {
			_ = database.Close()
		}
		if rdb != nil {
			_ = rdb.Close()
		}
	}

	return engine, cleanup, nil
}
