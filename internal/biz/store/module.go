package store

import (
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// NewModule 组装 store 模块的全部依赖并返回 HTTP Handler。
func NewModule(database *sqlx.DB, rdb *redis.Client) *Handler {
	return NewHandler(NewService(NewRepository(database), rdb))
}
