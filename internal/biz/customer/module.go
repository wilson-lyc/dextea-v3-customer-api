package customer

import (
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"github.com/dextea-v3/dextea-customer/api/internal/alipay"
	"github.com/dextea-v3/dextea-customer/api/internal/config"
)

func NewModule(database *sqlx.DB, rdb *redis.Client, cfg *config.Config, alipayClient *alipay.Client) *Handler {
	return NewHandler(NewService(NewRepository(database), rdb, cfg, alipayClient))
}
