package app

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/dextea-v3/dextea-customer/api/internal/config"
	"github.com/dextea-v3/dextea-customer/api/internal/customer"
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

	// 业务模块注册
	handlers := []server.Registerable{
		customer.NewModule(database, rdb, cfg),
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
