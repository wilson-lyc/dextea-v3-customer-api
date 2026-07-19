package area

import (
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"github.com/dextea-v3/dextea-customer/api/internal/config"
)

// NewModule 组装 area 模块的全部依赖并返回 HTTP Handler。
func NewModule(database *sqlx.DB, rdb *redis.Client, cfg *config.Config) *Handler {
	return NewHandler(NewService(cfg, NewRepository(database), rdb))
}
